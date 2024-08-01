package engine

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"os/signal"
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
	blockProcessor           BlockProcessor
	redis                    redisdb.RedisClient
	redisTrack               redisdb.RedisClient
	blockRetriever           BlockRetriever
	blockSyncTrack           postgres.PgBlockSyncTrackRepository
	newBlockSyncTrack        func() postgres.PgBlockSyncTrackRepository
	pgRetryTrack             postgres.PgTrackForToRetryRepository
	errorCount               int // track how many errors occur. If this reaches a high threshold and sinceLastError is low then possibly exit
	sinceLastError           int // track how many successful instances occurred since last error
	pauseRunner              *sync.WaitGroup
	//addressesToCheck         addressToCheckStruct
}

func NewBlockRunner(blockProcessor BlockProcessor, idxConfig types.IdxConfigStruct) BlockRunner {

	redisClient := redisdb.NewClient(1)
	redisTrack := redisdb.NewClient(1)
	blockRetriever := NewBlockRetriever(*redisClient)

	// Reset Block Tracking in Redis
	if idxConfig.ClearRedis {
		_, err := redisClient.Del("blockNumberOnSyncStart")
		if err != nil {
			utils.Logger.Errorln(err)
		}
		_, err = redisClient.Del("priorCurrentBlock")
		if err != nil {
			utils.Logger.Errorln(err)
		}
		_, err = redisClient.Del("lastPriorBlockRetrieved")
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}
	blockSyncTracking := postgres.NewBlockSyncTrackRepository(postgres.NewClient(idxConfig))

	pgRetryTrack := postgres.NewTrackForToRetryRepository(postgres.NewClient(idxConfig))

	createNewBlockSyncTrack := func() postgres.PgBlockSyncTrackRepository {
		return postgres.NewBlockSyncTrackRepository(postgres.NewClient(idxConfig))
	}

	var pr sync.WaitGroup

	return BlockRunner{
		priorRetrievalInProgress: false,
		lastBlock:                0,
		firstBlockSeen:           0,
		redis:                    *redisClient,
		redisTrack:               *redisTrack,
		blockProcessor:           blockProcessor,
		blockRetriever:           *blockRetriever,
		blockSyncTrack:           blockSyncTracking,
		newBlockSyncTrack:        createNewBlockSyncTrack,
		produceDelay:             100,
		pgRetryTrack:             pgRetryTrack,
		errorCount:               0,
		pauseRunner:              &pr,
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

	blockNumber := r.blockRetriever.LatestBlock()

	err := r.redis.Set("blockNumberOnSyncStart", blockNumber)
	if err != nil {
		return
	}
	err = r.redis.Set("priorCurrentBlock", blockNumber)
	if err != nil {
		return
	}

	errorCount, _ := r.redis.Get("retrievalErrorCount")
	if err != nil {
		errorCount = "0"
		err := r.redis.Set("retrievalErrorCount", errorCount)
		if err != nil {
			utils.Logger.Errorln(err)
			//return
		}
	}

	r.errorCount, _ = strconv.Atoi(errorCount)

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

	blockRetriever := NewBlockRetriever(r.redis)
	blockGen := blockRetriever.BlockHeaderChannel()
	var wg sync.WaitGroup

	for latestBlock := range blockGen {
		r.pauseRunner.Wait()
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
		r.blockProcessor.processBlock(block, &wg)

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
		err = r.redis.Set("retrievalErrorCount", r.errorCount)
		if err != nil {
			utils.Logger.Errorln(err)
		}

		if r.errorCount >= 100 {
			r.RetryFailedRetrievals()
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

	blockTimerCount := 0
	blockTimerAverage := 0

	for lastBlockRetrieved < blockNumberOnSyncStart {
		r.pauseRunner.Wait()
		start := time.Now()
		var wg sync.WaitGroup

		if !goodRun {
			blocksPerBatch = 10
		} /*else if blocksPerBatch < 210 {
			blocksPerBatch = blocksPerBatch + 10
		}*/

		// TODO: Modify this to use a rolling window for the average
		if goodRun && blockTimerAverage > 5 && blocksPerBatch > 20 {
			blocksPerBatch = blocksPerBatch - 10

		} else if goodRun && blockTimerAverage < 5 {
			blocksPerBatch = blocksPerBatch + 10
		}
		if blockTimerCount >= 5 {
			blockTimerCount = 0
			blockTimerAverage = 0
		}

		//if !goodRun {
		//	blocksPerBatch = 10
		//} else if blocksPerBatch < 300 {
		//	blocksPerBatch = blocksPerBatch + 10
		//} else if blocksPerBatch < 1000 {
		//	blocksPerBatch = blocksPerBatch + 1
		//}
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
					produceOk := r.blockProcessor.processBlock(*block, &wg)
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

		// monitor retrieval timing
		blockTimerCount += 1
		duration := time.Since(start).Seconds()
		blockTimerAverage = (blockTimerAverage + int(duration)) / blockTimerCount
		fmt.Printf("Batch Retreival took: %f for %d blocks\n", duration, batchEndBlock-lastBlockRetrieved)

		lastBlockRetrieved = batchEndBlock + 1

		err = r.redis.Set("lastPriorBlockRetrieved", lastBlockRetrieved)
		if err != nil {
			utils.Logger.Errorln(err)
			//return
		}
		err = r.redis.Set("retrievalErrorCount", r.errorCount)
		if err != nil {
			utils.Logger.Errorln(err)
		}
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
					produceOk := r.blockProcessor.processBlock(*block, &wg)
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
	r.pauseRunner.Add(1)

	defer func() {
		err := r.redis.Set("retrievalErrorCount", 0)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		r.pauseRunner.Done()
	}()

	blocksToRetry, _ := r.pgRetryTrack.GetByDataType(types.BLOCK_TOPIC)
	var wg sync.WaitGroup
	for _, tx := range blocksToRetry {
		wg.Add(1)
		bNum, _ := strconv.Atoi(tx.BlockId)
		block := r.blockRetriever.GetBlock(bNum)
		completed := r.blockProcessor.processBlock(block, &wg)
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
		completedTx := r.blockProcessor.producerFactory.Produce(types.TRANSACTION_TOPIC, transaction)
		if completedTx {
			bNum, _ := strconv.Atoi(tx.BlockId)
			_, err := r.blockProcessor.updateSyncTransactionsRecord(bNum, true)
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
		completedTx := r.blockProcessor.producerFactory.Produce(types.RECEIPT_TOPIC, txReceipt)
		if completedTx {
			bNum, _ := strconv.Atoi(receipt.BlockId)
			_, err := r.blockProcessor.updateSyncReceiptsRecord(bNum, true)
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
