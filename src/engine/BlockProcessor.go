package engine

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/golang/protobuf/proto"
	"src/kafka"
	"src/postgres"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"sync"
	"time"
)

type BlockProcessor struct {
	producerFactory *kafka.ProducerProvider
	blockRetriever  BlockRetriever
	blockSyncTrack  postgres.PgBlockSyncTrackRepository
	pgRetryTrack    postgres.PgTrackForToRetryRepository
	errorCount      int
	produceDelay    time.Duration
}

func NewBlockProcessor(producerFactory *kafka.ProducerProvider, idxConfig types.IdxConfigStruct) BlockProcessor {

	redisClient := redisdb.NewClient(1)
	blockRetriever := NewBlockRetriever(*redisClient)
	blockSyncTracking := postgres.NewBlockSyncTrackRepository(postgres.NewClient(idxConfig))

	pgRetryTrack := postgres.NewTrackForToRetryRepository(postgres.NewClient(idxConfig))

	fmt.Printf("producer ready: %t\n", producerFactory.Connected)

	return BlockProcessor{
		producerFactory: producerFactory,
		blockRetriever:  *blockRetriever,
		blockSyncTrack:  blockSyncTracking,
		produceDelay:    100,
		pgRetryTrack:    pgRetryTrack,
	}
}

func (r *BlockProcessor) processBlock(block types.Block, wg *sync.WaitGroup) bool {

	TransactionsProcessed := true
	ReceiptsProcessed := true
	pbBlock := types.Block{}.ProtobufFromGoType(block)
	blockToSend, err := proto.Marshal(&pbBlock)
	if err != nil {
		utils.Logger.Errorln(err)
	}
	completed := r.producerFactory.Produce(types.BLOCK_TOPIC, blockToSend)

	if completed {
		convertedBlock := types.Block{}.MongoFromGoType(block)

		TransactionsProcessed = r.processBlockTransactions(block, convertedBlock)

		ReceiptsProcessed = r.processBlockReceipts(convertedBlock, wg)

		updateComplete, err := r.updateSyncRecord(int(block.Number), ReceiptsProcessed, TransactionsProcessed)
		if err != nil {
			utils.Logger.Errorln(err)
		}

		if !updateComplete {
			completed = false
		}

	} else {
		r.errorCount += 1
		utils.Logger.Infof("Error producing BLOCK %v", block.Number)
		numb := strconv.Itoa(int(block.Number))
		_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
			DataType: types.BLOCK_TOPIC,
			BlockId:  numb,
		})
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}

	wg.Done()
	return completed && ReceiptsProcessed && TransactionsProcessed
}

func (r *BlockProcessor) processBlockTransactions(block types.Block, convertedBlock types.MongoBlock) bool {

	TransactionsProcessed := true
	for _, tx := range block.Transactions {
		pbTx := types.Transaction{}.ProtobufFromGoType(tx)
		txToSend, err := proto.Marshal(&pbTx)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		completedTx := r.producerFactory.Produce(types.TRANSACTION_TOPIC, txToSend)
		if !completedTx {
			r.errorCount += 1
			TransactionsProcessed = false
			_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
				DataType: types.TRANSACTION_TOPIC,
				BlockId:  convertedBlock.Number,
				RecordId: tx.Hash.String(),
			})
			if err != nil {
				utils.Logger.Errorln(err)
			}
			utils.Logger.Infof("Error producing transaction for block %s transaction hash: %s", convertedBlock.Number, tx.Hash)
			time.Sleep(r.produceDelay * time.Millisecond)
		}
	}
	return TransactionsProcessed
}

func (r *BlockProcessor) processBlockReceipts(convertedBlock types.MongoBlock, wg *sync.WaitGroup) bool {

	addressesToCheck := addressToCheckStruct{
		addressSet:  mapset.NewSet[string](),
		blockNumber: 0,
		txCount:     int64(len(convertedBlock.Transactions)),
	}

	ReceiptsProcessed := true
	receipts, err := GetBlockReceipts(convertedBlock.Hash)

	if err != nil {
		utils.Logger.Error(err)
		wg.Done()
		return false
	}

	for _, Receipt := range receipts {
		pbReceipt := types.Receipt{}.ProtobufFromGoType(*Receipt)
		receiptToSend, err := proto.Marshal(&pbReceipt)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		completedR := r.producerFactory.Produce(types.RECEIPT_TOPIC, receiptToSend)
		if !completedR {
			r.errorCount += 1
			ReceiptsProcessed = false
			_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
				DataType: types.RECEIPT_TOPIC,
				BlockId:  convertedBlock.Number,
				RecordId: Receipt.TransactionHash.String(),
			})
			if err != nil {
				utils.Logger.Errorln(err)
			}
			utils.Logger.Infof("Error producing receipt for block %s transaction hash: %s", convertedBlock.Number, Receipt.TransactionHash)
			time.Sleep(r.produceDelay * time.Millisecond)
		} else {
			addressesToCheck.addressSet.Add(Receipt.From.String())
			addressesToCheck.addressSet.Add(Receipt.To.String())
			for _, lgs := range Receipt.Logs {
				addressesToCheck.addressSet.Add(lgs.Address.String())
			}
		}
	}

	num, _ := strconv.Atoi(convertedBlock.Number)
	addressesToCheck.blockNumber = int64(num)

	r.processAddressesInBlock(addressesToCheck)

	return ReceiptsProcessed
}

// The batch call to erigon (at least) is returning incorrectly.
func (r *BlockProcessor) processAddressesInBlock(addressesToCheck addressToCheckStruct) {

	setIterator := addressesToCheck.addressSet.ToSlice()

	for _, addr := range setIterator {
		// excluding null address. such as with contract creation
		if addr != "0x0" {
			balanceHex, txCountHex, blockNumber, err := r.blockRetriever.GetAddressBalance(addr, addressesToCheck.blockNumber)
			if err != nil {
				utils.Logger.Errorln(err)
				return
			}
			balance, err1 := strconv.ParseInt(balanceHex[2:], 16, 64)
			if err1 != nil {
				balance = 0
			}

			txCount, err2 := strconv.ParseInt(txCountHex[2:], 16, 64)
			if err2 != nil {
				txCount = 0
			}
			//fmt.Printf("%v -> %d @ %v \n", addressToLookUp, balance, blockNumber)

			CollectedAddress := types.AddressBalance{
				Address:  addr,
				LastSeen: blockNumber,
				Balance:  balance,
				Nonce:    txCount,
			}
			pbAddress := types.AddressBalance{}.ProtobufFromGoType(CollectedAddress)
			addressToSend, err := proto.Marshal(&pbAddress)
			if err != nil {
				utils.Logger.Errorln(err)
			}
			completed := r.producerFactory.Produce(types.ADDRESS_TOPIC, addressToSend)
			if !completed {
				r.errorCount += 1
			}
		}

	}

}

func (r *BlockProcessor) updateSyncRecord(blockNumber int, ReceiptsProcessed bool, TransactionsProcessed bool) (bool, error) {
	blockEntry, err := r.blockSyncTrack.GetByBlockNumber(blockNumber)
	if err != nil {
		utils.Logger.Errorln(err)
		return false, err
	}
	if blockEntry != nil {
		blockEntry.Processed = ReceiptsProcessed && TransactionsProcessed
		blockEntry.Retrieved = true
		blockEntry.ReceiptsProcessed = ReceiptsProcessed
		blockEntry.TransactionsProcessed = TransactionsProcessed
		_, err = r.blockSyncTrack.Update(*blockEntry)
		if err != nil {
			utils.Logger.Errorln(err)
			utils.Logger.Infof("POSTGRES Error Updating Record for block %v", blockNumber)
			return false, err
		}
	} else {
		utils.Logger.Infof("POSTGRES RETURNED NIL FOR BLOCK SYNC RECORD OF BLOCK  %d", blockNumber)
	}

	return true, nil
}

func (r *BlockProcessor) updateSyncTransactionsRecord(blockNumber int, TransactionsProcessed bool) (bool, error) {
	blockEntry, err := r.blockSyncTrack.GetByBlockNumber(blockNumber)
	if err != nil {
		utils.Logger.Errorln(err)
		return false, err
	}
	if blockEntry != nil {
		blockEntry.Processed = blockEntry.ReceiptsProcessed && TransactionsProcessed
		blockEntry.Retrieved = true
		blockEntry.TransactionsProcessed = TransactionsProcessed
		_, err = r.blockSyncTrack.Update(*blockEntry)
		if err != nil {
			utils.Logger.Errorln(err)
			utils.Logger.Infof("POSTGRES Error Updating Record for block %v", blockNumber)
			return false, err
		}
	} else {
		utils.Logger.Infof("POSTGRES RETURNED NIL FOR BLOCK SYNC RECORD OF BLOCK  %d", blockNumber)
	}

	return true, nil
}

func (r *BlockProcessor) updateSyncReceiptsRecord(blockNumber int, ReceiptsProcessed bool) (bool, error) {
	blockEntry, err := r.blockSyncTrack.GetByBlockNumber(blockNumber)
	if err != nil {
		utils.Logger.Errorln(err)
		return false, err
	}
	if blockEntry != nil {
		blockEntry.Processed = ReceiptsProcessed && blockEntry.TransactionsProcessed
		blockEntry.Retrieved = true
		blockEntry.ReceiptsProcessed = ReceiptsProcessed
		_, err = r.blockSyncTrack.Update(*blockEntry)
		if err != nil {
			utils.Logger.Errorln(err)
			utils.Logger.Infof("POSTGRES Error Updating Record for block %v", blockNumber)
			return false, err
		}
	} else {
		utils.Logger.Infof("POSTGRES RETURNED NIL FOR BLOCK SYNC RECORD OF BLOCK  %d", blockNumber)
	}

	return true, nil
}
