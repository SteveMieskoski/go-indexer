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
	"syscall"
	"time"
)

type BlockRunner struct {
	priorRetrievalInProgress bool
	lastBlock                int
	firstBlockSeen           int
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
	blockRetriever           engine.BlockRetriever
	blockSyncTrack           postgres.PgBlockSyncTrackRepository
	newBlockSyncTrack        func() postgres.PgBlockSyncTrackRepository
}

func NewBlockRunner(producerFactory *kafka.ProducerProvider) BlockRunner {

	redisClient := redisdb.NewClient()
	blockRetriever := engine.NewBlockRetriever(*redisClient)
	blockSyncTracking := postgres.NewBlockSyncTrackRepository(postgres.NewClient())

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

	//fmt.Printf(lastBlock)
	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest block: %d", latestBlock.Number)

		block := blockRetriever.GetBlock(int(latestBlock.Number))

		//r.producerFactory.Produce(types.BLOCK_TOPIC, block)
		r.processBlock(block)
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

}

func (r *BlockRunner) processBlock(block types.Block) {

	completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)

	if completed {
		convertedBlock := types.Block{}.MongoFromGoType(block)

		utils.Logger.Infof("Block %s contains %d Transactions", convertedBlock.Number, len(convertedBlock.Transactions))

		for _, tx := range block.Transactions {
			r.producerFactory.Produce(types.TRANSACTION_TOPIC, tx)
		}

		receipts, err := engine.GetBlockReceipts(convertedBlock.Hash)

		if err != nil {
			panic(err)
			return
		}
		for _, Receipt := range receipts {
			r.producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
		}

		_, err = r.blockSyncTrack.GetByBlockNumber(int(block.Number)) //TODO: STOPPED HERE <<<<<<<<<<<<<<<<<<<<<<<<<<
		blockEntry, err := r.blockSyncTrack.GetByBlockNumber(int(block.Number))
		blockEntry.Processed = true
		blockEntry.Retrieved = true
		blockEntry.ReceiptsProcessed = true
		blockEntry.TransactionsProcessed = true
		_, err = r.blockSyncTrack.Update(*blockEntry)

	} else {
		utils.Logger.Infof("producerFactory.Produce FAILED")
		//go func() {
		//	r.getPriorBlock(*blockRetriever, int(block.Number))
		//}()

	}
}

func (r *BlockRunner) getPriorBlocks() {
	dbAccess := r.newBlockSyncTrack()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

				go r.processBlock(*block)
			}

		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}

		lastBlockRetrieved = batchEndBlock + 1
		time.Sleep(200 * time.Millisecond)

	}

	utils.Logger.Info("exiting: getPriorBlock External")
}
