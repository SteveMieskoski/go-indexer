package internal

import (
	"context"
	"fmt"
	"os/signal"
	"src/engine"
	"src/kafka"
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
}

func NewBlockRunner() BlockRunner {

	return BlockRunner{
		priorRetrievalInProgress: false,
		lastBlock:                0,
		firstBlockSeen:           0,
		//requestHeader: make(chan *bool, 1),
	}
}

func (r *BlockRunner) getCurrentBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient := redisdb.NewClient()

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	blockRetrievers := engine.NewBlockRetriever(*redisClient)

	fmt.Printf("producer ready: %t\n", producerFactory.Connected)
	blockGen := blockRetrievers.GetBlocks()

	//lastBlock, _ := redisClient.Get("B1")
	//fmt.Printf(lastBlock)
	for block := range blockGen {
		utils.Logger.Infof("latest block: %d", block.Number)
		formatBlockNumberInt := strconv.Itoa(int(block.Number))
		err := redisClient.Set("latestBlock", formatBlockNumberInt)
		if err != nil {
			panic(err)
		}
		blkNmKey := generateBlockNumberKey(int(block.Number))
		err = redisClient.Set(blkNmKey, true)
		if err != nil {
			return
		}

		if int(block.Number) > r.lastBlock {
			producerFactory.Produce("Block", block)
			convertedBlock := types.Block{}.MongoFromGoType(block)

			go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
				receipts, err := engine.GetBlockReceipts("http://localhost:8545", blockHash)
				if err != nil {
					panic(err)
					return
				}
				for _, Receipt := range receipts {
					producerFactory.Produce("Receipt", *Receipt)
				}
			}(convertedBlock.Hash)

			r.lastBlock = int(block.Number)
		}

		utils.Logger.Infof("PRIOR RETRIEVAL STARTED: %t \n", r.priorRetrievalInProgress)

		if !r.priorRetrievalInProgress && block.Number > 1 {
			r.firstBlockSeen = int(block.Number)
			r.priorRetrievalInProgress = true
			go func() {
				r.getPriorBlocks(*blockRetrievers, *redisClient)
			}()

		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}
	}

}

func (r *BlockRunner) getPriorBlocks(blockRetriever engine.BlockRetriever, redisClient redisdb.RedisClient) {

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//blockRetriever.GetPastBlocks(BlockNumToGetChan)
	lastBlock := r.firstBlockSeen
	utils.Logger.Infof("latest block known: %d", lastBlock)
	lastBlockRetrieved := 0
	completed := true

	for lastBlockRetrieved < r.firstBlockSeen {
		blockNumberKey := generateBlockNumberKey(lastBlockRetrieved)
		utils.Logger.Infof("Check if prior block retrieved: %s", blockNumberKey)
		_, err := redisClient.Get(blockNumberKey)
		if err != nil {
			utils.Logger.Infof("Prior block needed: %s", blockNumberKey)
			//BlockNumToGetChan <- i
			block := blockRetriever.GetPastBlocks(lastBlockRetrieved)

			completed = producerFactory.Produce("Block", block)
			if completed {
				convertedBlock := types.Block{}.MongoFromGoType(block)

				go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
					receipts, err := engine.GetBlockReceipts("http://localhost:8545", blockHash)
					if err != nil {
						panic(err)
						return
					}
					for _, Receipt := range receipts {
						producerFactory.Produce("Receipt", *Receipt)
					}
				}(convertedBlock.Hash)
				//	Dangerously assume save did not fail
				err = redisClient.Set(blockNumberKey, strconv.Itoa(lastBlockRetrieved))
				if err != nil {
					panic(err)
				}
			}

		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}

		if completed {
			utils.Logger.Infof("Prior block %d ingested", lastBlockRetrieved)
			lastBlockRetrieved++
		} else {
			utils.Logger.Infof("Prior block %d ingestion failed", lastBlockRetrieved)
			time.Sleep(300 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)

	}

	utils.Logger.Info("exiting: getPriorBlock External")
}

func generateBlockNumberKey(blockNumber int) string {
	formatInt := strconv.Itoa(blockNumber)
	return "B" + formatInt
}
