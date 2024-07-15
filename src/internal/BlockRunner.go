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
	transactionsToRetry      []types.Transaction
	receiptsToRetry          []types.Receipt
	blocksToRetry            []string
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
		transactionsToRetry:      []types.Transaction{},
		receiptsToRetry:          []types.Receipt{},
		blocksToRetry:            []string{},
		pgRetryTrack:             pgRetryTrack,
		//requestHeader: make(chan *bool, 1),
	}
}

func (r *BlockRunner) Demo() {

	//blockNumber := r.blockRetriever.LatestBlock()
	//println(blockNumber)
	//parseInt, err := strconv.ParseInt(blockNumber[2:], 16, 64)
	//println(parseInt)
	//if err != nil {
	//	return
	//}

	//blocks := r.blockRetriever.GetBlockBatch(10, 20)
	//
	//for _, block := range blocks {
	//	switch t := block.Result.(type) {
	//	case *types.Block:
	//		println(t.Number)
	//	}
	//	//bytes, _ := json.MarshalIndent(block.Result, "", "   ")
	//	//println(string(bytes))
	//}
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

	fmt.Printf("producer ready: %t\n", r.producerFactory.Connected)

	blockRetriever := engine.NewBlockRetriever(r.redis)
	blockGen := blockRetriever.BlockHeaderChannel()
	var wg sync.WaitGroup
	//fmt.Printf(lastBlock)
	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest block: %d", latestBlock.Number)

		block := blockRetriever.GetBlock(int(latestBlock.Number))
		wg.Add(1)
		//r.producerFactory.Produce(types.BLOCK_TOPIC, block)
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
				completed = false
				TransactionsProcessed = false
				//r.transactionsToRetry = append(r.transactionsToRetry, tx)
				_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
					DataType: "Transaction",
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
			utils.Logger.Infof("Send on resultChan Err: %t\n", false)
			//resultChan <- false
			utils.Logger.Error(err)
			wg.Done()
			return false
		}
		for _, Receipt := range receipts {
			completedR := r.producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
			if !completedR {
				completed = false
				ReceiptsProcessed = false
				//r.receiptsToRetry = append(r.receiptsToRetry, *Receipt)
				_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
					DataType: "Receipt",
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

		_, err = r.blockSyncTrack.GetByBlockNumber(int(block.Number)) //TODO: STOPPED HERE <<<<<<<<<<<<<<<<<<<<<<<<<<
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
		//go func() {
		//	r.getPriorBlock(*blockRetriever, int(block.Number))
		//}()
		numb := strconv.Itoa(int(block.Number))
		//r.blocksToRetry = append(r.blocksToRetry, numb)
		_, err := r.pgRetryTrack.Add(types.PgTrackForToRetry{
			DataType: "Block",
			BlockId:  numb,
		})
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}
	println("WAIT DONE ==========================================")
	wg.Done()
	return completed
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
		//ctx = context.WithValue(ctx, "backoff", false)
		if !goodRun {
			blocksPerBatch = 10
		} else if blocksPerBatch < 50 {
			blocksPerBatch = blocksPerBatch + 10
		}

		batchEndBlock := lastBlockRetrieved + blocksPerBatch

		if batchEndBlock >= blockNumberOnSyncStart {
			batchEndBlock = blockNumberOnSyncStart
		}

		batchResponse, err := r.blockRetriever.GetBlockBatch(lastBlockRetrieved, batchEndBlock)

		if err != nil {
			goodRun = false
			continue
		}
		println(len(batchResponse))
		wg.Add(len(batchResponse))
		for _, response := range batchResponse {
			switch block := response.Result.(type) {
			case *types.Block:
				_, err := dbAccess.Add(types.PgBlockSyncTrack{
					Number:                int64(block.Number),
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
					} else {
						r.produceDelay = r.produceDelay + 100
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

		println("WAITING ===================================================")
		wg.Wait()
		lastBlockRetrieved = batchEndBlock + 1
		//time.Sleep(r.produceDelay * time.Millisecond)
		//println("WAITING ===================================================")
		//wg.Wait()
	}

	//wg.Wait()
	utils.Logger.Info("exiting: getPriorBlock External")
}
