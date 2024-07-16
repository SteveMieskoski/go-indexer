package internal

import (
	"context"
	"fmt"
	"os/signal"
	"src/engine"
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

type BlockRunner struct {
	priorRetrievalInProgress bool
	lastBlock                int
	firstBlockSeen           int
	produceDelay             time.Duration
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
	blockRetriever           engine.BlockRetriever
	blockSyncTrack           postgres.PgBlockSyncTrackRepository
	newBlockSyncTrack        func() postgres.PgBlockSyncTrackRepository
	pgRetryTrack             postgres.PgTrackForToRetryRepository
}

func NewBlockRunner(producerFactory *kafka.ProducerProvider) BlockRunner {

	redisClient := redisdb.NewClient()
	blockRetriever := engine.NewBlockRetriever(*redisClient)
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
		producerFactory:          producerFactory,
		blockRetriever:           *blockRetriever,
		blockSyncTrack:           blockSyncTracking,
		newBlockSyncTrack:        createNewBlockSyncTrack,
		produceDelay:             100,
		pgRetryTrack:             pgRetryTrack,
	}
}

func (r *BlockRunner) Demo() {
	r.RetryFailedRetrievals()
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

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

	go r.getPriorBlocks()
	r.getCurrentBlock()
}

func (r *BlockRunner) getCurrentBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	dbAccess := r.newBlockSyncTrack()

	fmt.Printf("producer ready: %t\n", r.producerFactory.Connected)

	blockRetriever := engine.NewBlockRetriever(r.redis)
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
			continue
		}
		wg.Add(1)
		r.processBlock(block, &wg)
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

}

func (r *BlockRunner) processBlock(block types.Block, wg *sync.WaitGroup) bool {

	TransactionsProcessed := true
	ReceiptsProcessed := true
	completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)

	if completed {
		convertedBlock := types.Block{}.MongoFromGoType(block)

		utils.Logger.Infof("Block %s contains %d Transactions", convertedBlock.Number, len(convertedBlock.Transactions))

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
			}
		}

		receipts, err := engine.GetBlockReceipts(convertedBlock.Hash)

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
			}
		}

		_, err = r.blockSyncTrack.GetByBlockNumber(int(block.Number))
		blockEntry, err := r.blockSyncTrack.GetByBlockNumber(int(block.Number))
		if err != nil {
			completed = false
		}
		if blockEntry != nil {
			blockEntry.Processed = ReceiptsProcessed && TransactionsProcessed
			blockEntry.Retrieved = true
			blockEntry.ReceiptsProcessed = ReceiptsProcessed
			blockEntry.TransactionsProcessed = TransactionsProcessed
			_, err = r.blockSyncTrack.Update(*blockEntry)
		} else {
			utils.Logger.Infof("POSTGRES Error Updating Record for block %v", block.Number)
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
			return
		}
	} else {
		lastBlockRetrieved, _ = strconv.Atoi(val)
	}

	blocksPerBatch := 10
	goodRun := true

	for lastBlockRetrieved < blockNumberOnSyncStart {

		var wg sync.WaitGroup

		if !goodRun {
			blocksPerBatch = 10
		} else if blocksPerBatch < 300 {
			blocksPerBatch = blocksPerBatch + 10
		}

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

	utils.Logger.Info("exiting: getPriorBlock External")
	r.RetryFailedRetrievals()
}

func (r *BlockRunner) RetryFailedRetrievals() {
	utils.Logger.Info("RetryFailedRetrievals START")
	blocksToRetry, _ := r.pgRetryTrack.GetByDataType(types.BLOCK_TOPIC)

	for _, tx := range blocksToRetry {
		bNum, _ := strconv.Atoi(tx.BlockId)
		block := r.blockRetriever.GetBlock(bNum)

		completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)
		if completed {
			_, err := r.pgRetryTrack.Delete(tx.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}

	transactionsToRetry, _ := r.pgRetryTrack.GetByDataType(types.TRANSACTION_TOPIC)

	for _, tx := range transactionsToRetry {
		println(tx.RecordId)
		transaction := r.blockRetriever.GetTransaction(tx.RecordId)
		completedTx := r.producerFactory.Produce(types.TRANSACTION_TOPIC, transaction)
		if completedTx {
			print(transaction.String())
			_, err := r.pgRetryTrack.Delete(tx.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}

	receiptsToRetry, _ := r.pgRetryTrack.GetByDataType(types.RECEIPT_TOPIC)

	for _, receipt := range receiptsToRetry {
		println(receipt.RecordId)
		txReceipt := r.blockRetriever.GetTransactionReceipt(receipt.RecordId)
		completedTx := r.producerFactory.Produce(types.RECEIPT_TOPIC, txReceipt)
		if completedTx {
			print(txReceipt.String())
			_, err := r.pgRetryTrack.Delete(receipt.Id)
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}
	utils.Logger.Info("RetryFailedRetrievals END")
}
