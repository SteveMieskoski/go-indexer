package engine

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"os/signal"
	"src/kafka"
	"src/postgres"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type addressToCheckStruct struct {
	addressSet  mapset.Set[string]
	blockNumber int64
	txCount     int64
}

type BlockRunner struct {
	priorRetrievalInProgress bool
	lastBlock                int
	firstBlockSeen           int
	produceDelay             time.Duration
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
	redisTrack               redisdb.RedisClient
	blockRetriever           BlockRetriever
	blockSyncTrack           postgres.PgBlockSyncTrackRepository
	newBlockSyncTrack        func() postgres.PgBlockSyncTrackRepository
	pgRetryTrack             postgres.PgTrackForToRetryRepository
	//addressesToCheck         addressToCheckStruct
}

func NewBlockRunner(producerFactory *kafka.ProducerProvider) BlockRunner {

	redisClient := redisdb.NewClient(1)
	redisTrack := redisdb.NewClient(1)
	blockRetriever := NewBlockRetriever(*redisClient)
	blockSyncTracking := postgres.NewBlockSyncTrackRepository(postgres.NewClient())

	pgRetryTrack := postgres.NewTrackForToRetryRepository(postgres.NewClient())

	createNewBlockSyncTrack := func() postgres.PgBlockSyncTrackRepository {
		return postgres.NewBlockSyncTrackRepository(postgres.NewClient())
	}

	return BlockRunner{
		priorRetrievalInProgress: false,
		lastBlock:                0,
		firstBlockSeen:           0,
		redis:                    *redisClient,
		redisTrack:               *redisTrack,
		producerFactory:          producerFactory,
		blockRetriever:           *blockRetriever,
		blockSyncTrack:           blockSyncTracking,
		newBlockSyncTrack:        createNewBlockSyncTrack,
		produceDelay:             100,
		pgRetryTrack:             pgRetryTrack,
	}
}

func (r *BlockRunner) Demo() {
	//balance, blockNumber, err := r.blockRetriever.GetAddressBalance("0x02cD57cD479AFC7d4ba49275dC8F75706B3aaa27", 2003762)
	//if err != nil {
	//	return
	//}
	//fmt.Printf("%v, %v\n", balance, blockNumber)
	//parseInt, err := strconv.ParseInt(balance[2:], 16, 64)
	//
	//CollectedAddress := types.AddressBalance{
	//	Address:  "0x02cD57cD479AFC7d4ba49275dC8F75706B3aaa27",
	//	LastSeen: blockNumber,
	//	Balance:  parseInt,
	//}
	//
	//println(CollectedAddress.String())
}

func (r *BlockRunner) StartBlockSync() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", r.producerFactory.Connected)

	blockNumber := r.blockRetriever.LatestBlock()

	err := r.redis.Set("blockNumberOnSyncStart", blockNumber)
	if err != nil {
		return
	}
	err = r.redis.Set("priorCurrentBlock", blockNumber)
	if err != nil {
		return
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

	go r.getPriorBlocks() // TODO <- UNCOMMENT
	r.listenForCurrentBlock()
}

func (r *BlockRunner) listenForCurrentBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	dbAccess := r.newBlockSyncTrack()

	fmt.Printf("producer ready: %t\n", r.producerFactory.Connected)

	blockRetriever := NewBlockRetriever(r.redis)
	blockGen := blockRetriever.BlockHeaderChannel()
	var wg sync.WaitGroup

	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest block: %d", latestBlock.Number)

		block := blockRetriever.GetBlock(int(latestBlock.Number))
		_, err := dbAccess.Add(types.PgBlockSyncTrack{
			Number:                int64(block.Number),
			Hash:                  block.Hash.String(),
			Retrieved:             false,
			Processed:             false,
			ReceiptsProcessed:     false,
			TransactionsProcessed: false,
			ContractsProcessed:    false,
		})

		if err != nil {
			// need to monitor to reduce batch size if db gets too slow
			// add error log here
			utils.Logger.Errorln(err)
			continue
		}
		wg.Add(1)
		r.processBlock(block, &wg)

		val, err := r.redis.Get("priorCurrentBlock")
		if err != nil {
			utils.Logger.Errorln(err)
			continue
		}
		valNum, _ := strconv.Atoi(val)
		if int(latestBlock.Number)-valNum > 1 {
			go r.getPriorBlocksInRange(valNum+1, int(latestBlock.Number)-1)
		}

		err = r.redis.Set("priorCurrentBlock", int64(block.Number))
		if err != nil {
			utils.Logger.Errorln(err)
			continue
		}
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

}

// break out into it's own type and possibly use context to hold the block number
func (r *BlockRunner) processBlock(block types.Block, wg *sync.WaitGroup) bool {

	TransactionsProcessed := true
	ReceiptsProcessed := true
	completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)

	if completed {
		convertedBlock := types.Block{}.MongoFromGoType(block)

		//utils.Logger.Infof("Block %s contains %d Transactions", convertedBlock.Number, len(convertedBlock.Transactions))

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

func (r *BlockRunner) processBlockTransactions(block types.Block, convertedBlock types.MongoBlock) bool {
	//addressesToCheck := addressToCheckStruct{
	//	addressSet:  mapset.NewSet[string](),
	//	blockNumber: 0,
	//	txCount:     int64(len(block.Transactions)),
	//}

	TransactionsProcessed := true
	for _, tx := range block.Transactions {
		completedTx := r.producerFactory.Produce(types.TRANSACTION_TOPIC, tx)
		if !completedTx {
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
		} /*else {
			addressesToCheck.addressSet.Add(tx.From.String())
			addressesToCheck.addressSet.Add(tx.To.String())
		}*/
	}
	//addressesToCheck.blockNumber = int64(block.Number)

	//r.processAddressesInBlock(addressesToCheck)
	return TransactionsProcessed
}

func (r *BlockRunner) processBlockReceipts(convertedBlock types.MongoBlock, wg *sync.WaitGroup) bool {

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
		completedR := r.producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
		if !completedR {
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
func (r *BlockRunner) processAddressesInBlock(addressesToCheck addressToCheckStruct) {

	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//defer stop()

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
			_ = r.producerFactory.Produce(types.ADDRESS_TOPIC, CollectedAddress)
		}

	}

}

func (r *BlockRunner) getPriorBlocks() {

	dbAccess := r.newBlockSyncTrack()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//var wg sync.WaitGroup

	Num, _ := r.redis.Get("blockNumberOnSyncStart")
	blockNumberOnSyncStart, _ := strconv.Atoi(Num)

	lastBlockRetrieved := 0
	val, err := r.redis.Get("lastPriorBlockRetrieved")
	if err != nil {
		err := r.redis.Set("lastPriorBlockRetrieved", lastBlockRetrieved)
		if err != nil {
			utils.Logger.Errorln(err)
			//return
		}
	} else {
		lastBlockRetrieved, _ = strconv.Atoi(val)
	}

	blocksPerBatch := 10
	goodRun := true
	//duration := 0

	for lastBlockRetrieved < blockNumberOnSyncStart {

		//start := time.Now()
		var wg sync.WaitGroup

		if !goodRun {
			blocksPerBatch = 10
		} else if blocksPerBatch < 300 {
			blocksPerBatch = blocksPerBatch + 10
		} /*else if blocksPerBatch < 1000 {
			blocksPerBatch = blocksPerBatch + 1
		}*/
		// was getting errors, maybe from above, need to check with the mem error
		//if duration > 900000000 && blocksPerBatch > 20 {
		//	blocksPerBatch = blocksPerBatch - 5
		//}

		batchEndBlock := lastBlockRetrieved + blocksPerBatch

		if batchEndBlock >= blockNumberOnSyncStart {
			batchEndBlock = blockNumberOnSyncStart
		}
		fmt.Printf("Indexing prior blocs from %d to %d\n", lastBlockRetrieved, batchEndBlock)
		batchResponse, err := r.blockRetriever.GetBlockBatch(lastBlockRetrieved, batchEndBlock)

		if err != nil {
			goodRun = false
			continue
		}

		wg.Add(len(batchResponse))
		for _, response := range batchResponse {
			switch block := response.Result.(type) {
			case *types.Block:
				_, err := dbAccess.Add(types.PgBlockSyncTrack{
					Number:                int64(block.Number),
					Hash:                  block.Hash.String(),
					Retrieved:             false,
					Processed:             false,
					ReceiptsProcessed:     false,
					TransactionsProcessed: false,
					ContractsProcessed:    false,
					TransactionCount:      int64(len(block.Transactions)),
				})

				if err != nil {
					// need to monitor to reduce batch size if db gets too slow
					// add error log here
					continue
				}

				go func() {
					produceOk := r.processBlock(*block, &wg)
					if produceOk {
						if r.produceDelay >= 200 {
							r.produceDelay = r.produceDelay - 100
						}
						goodRun = true
					} else {
						r.produceDelay = r.produceDelay + 100
						goodRun = false
					}
				}()

			}

		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}

		wg.Wait()
		//duration = int(time.Since(start))
		//println(duration)
		lastBlockRetrieved = batchEndBlock + 1
	}

	utils.Logger.Info("exiting: getPriorBlock External")
	r.RetryFailedRetrievals()
}

// TODO: this should probably be used as the runner for getting past blocks
func (r *BlockRunner) getPriorBlocksInRange(startBlock int, endBlock int) {

	dbAccess := r.newBlockSyncTrack()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lastBlockRetrieved := startBlock

	blocksPerBatch := 10
	goodRun := true

	for lastBlockRetrieved < endBlock {

		var wg sync.WaitGroup

		if !goodRun {
			blocksPerBatch = 10
		} else if blocksPerBatch < 300 {
			blocksPerBatch = blocksPerBatch + 10
		}

		batchEndBlock := lastBlockRetrieved + blocksPerBatch

		if batchEndBlock >= endBlock {
			batchEndBlock = endBlock
		}
		fmt.Printf("Indexing prior blocs from %d to %d\n", lastBlockRetrieved, batchEndBlock)
		batchResponse, err := r.blockRetriever.GetBlockBatch(lastBlockRetrieved, batchEndBlock)

		if err != nil {
			goodRun = false
			continue
		}

		wg.Add(len(batchResponse))
		for _, response := range batchResponse {
			switch block := response.Result.(type) {
			case *types.Block:
				_, err := dbAccess.Add(types.PgBlockSyncTrack{
					Number:                int64(block.Number),
					Hash:                  block.Hash.String(),
					Retrieved:             false,
					Processed:             false,
					ReceiptsProcessed:     false,
					TransactionsProcessed: false,
					ContractsProcessed:    false,
					TransactionCount:      int64(len(block.Transactions)),
				})

				if err != nil {
					// need to monitor to reduce batch size if db gets too slow
					// add error log here
					continue
				}

				go func() {
					produceOk := r.processBlock(*block, &wg)
					if produceOk {
						if r.produceDelay >= 200 {
							r.produceDelay = r.produceDelay - 100
						}
						goodRun = true
					} else {
						r.produceDelay = r.produceDelay + 100
						goodRun = false
					}
				}()

			}

		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}

		wg.Wait()
		lastBlockRetrieved = batchEndBlock + 1
	}

	fmt.Printf("Last Block Retreived at EXIT %d \n", lastBlockRetrieved)
	utils.Logger.Info("exiting: getPriorBlock External")
	r.RetryFailedRetrievals()
}

func (r *BlockRunner) RetryFailedRetrievals() {
	utils.Logger.Info("RetryFailedRetrievals START")
	blocksToRetry, _ := r.pgRetryTrack.GetByDataType(types.BLOCK_TOPIC)
	var wg sync.WaitGroup
	for _, tx := range blocksToRetry {
		wg.Add(1)
		bNum, _ := strconv.Atoi(tx.BlockId)
		block := r.blockRetriever.GetBlock(bNum)
		completed := r.processBlock(block, &wg)
		//completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)
		if completed {
			_, err := r.pgRetryTrack.Delete(tx.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
		wg.Wait()
	}

	transactionsToRetry, _ := r.pgRetryTrack.GetByDataType(types.TRANSACTION_TOPIC)

	for _, tx := range transactionsToRetry {
		println(tx.RecordId)
		transaction := r.blockRetriever.GetTransaction(tx.RecordId)
		completedTx := r.producerFactory.Produce(types.TRANSACTION_TOPIC, transaction)
		if completedTx {
			bNum, _ := strconv.Atoi(tx.BlockId)
			_, err := r.updateSyncTransactionsRecord(bNum, true)
			if err != nil {
				utils.Logger.Errorln(err)
			}
			_, err = r.pgRetryTrack.Delete(tx.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}

	receiptsToRetry, _ := r.pgRetryTrack.GetByDataType(types.RECEIPT_TOPIC)

	for _, receipt := range receiptsToRetry {
		txReceipt := r.blockRetriever.GetTransactionReceipt(receipt.RecordId)
		completedTx := r.producerFactory.Produce(types.RECEIPT_TOPIC, txReceipt)
		if completedTx {
			bNum, _ := strconv.Atoi(receipt.BlockId)
			_, err := r.updateSyncReceiptsRecord(bNum, true)
			if err != nil {
				utils.Logger.Errorln(err)
			}
			_, err = r.pgRetryTrack.Delete(receipt.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}
	utils.Logger.Info("RetryFailedRetrievals END")
}

func (r *BlockRunner) updateSyncRecord(blockNumber int, ReceiptsProcessed bool, TransactionsProcessed bool) (bool, error) {
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

func (r *BlockRunner) updateSyncTransactionsRecord(blockNumber int, TransactionsProcessed bool) (bool, error) {
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

func (r *BlockRunner) updateSyncReceiptsRecord(blockNumber int, ReceiptsProcessed bool) (bool, error) {
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
