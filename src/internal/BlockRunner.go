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
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
}

func NewBlockRunner(producerFactory *kafka.ProducerProvider) BlockRunner {

	redisClient := redisdb.NewClient()

	return BlockRunner{
		priorRetrievalInProgress: false,
		lastBlock:                0,
		firstBlockSeen:           0,
		redis:                    *redisClient,
		producerFactory:          producerFactory,
		//requestHeader: make(chan *bool, 1),
	}
}

func (r *BlockRunner) getCurrentBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	blockRetrievers := engine.NewBlockRetriever(r.redis)

	fmt.Printf("producer ready: %t\n", r.producerFactory.Connected)
	blockGen := blockRetrievers.GetBlocks()

	lastKnownBlock, err := r.redis.Get("latestBlock")
	databasePopulated := true
	if err != nil {
		databasePopulated = false
	} else {
		r.lastBlock, _ = strconv.Atoi(lastKnownBlock)
	}
	//fmt.Printf(lastBlock)
	for ablock := range blockGen {
		utils.Logger.Infof("latest block: %d", ablock.Number)

		block := blockRetrievers.GetBlock(int(ablock.Number))

		formatBlockNumberInt := strconv.Itoa(int(block.Number))
		err := r.redis.Set("latestBlock", formatBlockNumberInt)
		if err != nil {
			panic(err)
		}
		blkNmKey := generateBlockNumberKey(int(block.Number))
		err = r.redis.Set(blkNmKey, true)
		if err != nil {
			return
		}

		if int(block.Number) > r.lastBlock {
			// todo | set up a retry mechanism or
			// todo | need a way to check (i.e. bloom filters on consumer side maybe) that no blocks were missed
			completed := r.producerFactory.Produce(types.BLOCK_TOPIC, block)

			if completed {
				err := r.redis.Set("missingBlock", int(block.Number))
				if err != nil {
					utils.Logger.Errorf("failed to set missing block: %v", err)
				}

				convertedBlock := types.Block{}.MongoFromGoType(block)

				utils.Logger.Infof("Transactions In Current Block: %d", len(convertedBlock.Transactions))

				go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
					receipts, err := engine.GetBlockReceipts("http://localhost:8545", blockHash)
					if err != nil {
						panic(err)
						return
					}
					for _, Receipt := range receipts {
						r.producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
					}
				}(convertedBlock.Hash)

				go func(transactions []types.Transaction) {
					for _, tx := range transactions {
						r.producerFactory.Produce(types.TRANSACTION_TOPIC, tx)
					}
				}(block.Transactions)

				r.lastBlock = int(block.Number)
			} else {
				go func() {
					r.getPriorBlock(*blockRetrievers, int(block.Number))
				}()

			}

		}

		//utils.Logger.Infof("PRIOR RETRIEVAL STARTED: %t \n", r.priorRetrievalInProgress)

		if !r.priorRetrievalInProgress && block.Number > 1 {
			// todo Need to scan redis to find the last old block retrieved
			if !databasePopulated {
				r.firstBlockSeen = int(block.Number)
			} else {
				r.firstBlockSeen = int(block.Number)
			}

			r.priorRetrievalInProgress = true
			go func() {
				r.getPriorBlocks(*blockRetrievers)
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

func (r *BlockRunner) getPriorBlocks(blockRetriever engine.BlockRetriever) {

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//blockRetriever.GetBlock(BlockNumToGetChan)
	lastBlock := r.firstBlockSeen
	utils.Logger.Infof("latest block known: %d", lastBlock)
	lastBlockRetrieved := 0
	completed := true

	for lastBlockRetrieved < r.firstBlockSeen {
		blockNumberKey := generateBlockNumberKey(lastBlockRetrieved)
		utils.Logger.Infof("Check if prior block retrieved: %s", blockNumberKey)
		_, err := r.redis.Get(blockNumberKey)
		if err != nil {
			utils.Logger.Infof("Prior block needed: %s", blockNumberKey)
			//BlockNumToGetChan <- i
			block := blockRetriever.GetBlock(lastBlockRetrieved)

			completed = producerFactory.Produce(types.BLOCK_TOPIC, block)
			if completed {
				convertedBlock := types.Block{}.MongoFromGoType(block)
				utils.Logger.Infof("Transactions In Past Block: %d", len(convertedBlock.Transactions))

				go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
					receipts, err := engine.GetBlockReceipts("http://localhost:8545", blockHash)
					if err != nil {
						panic(err)
						return
					}
					for _, Receipt := range receipts {
						producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
					}
				}(convertedBlock.Hash)

				go func(transactions []types.Transaction) {
					for _, tx := range transactions {
						producerFactory.Produce(types.TRANSACTION_TOPIC, tx)
					}
				}(block.Transactions)

				//	Dangerously assume save did not fail
				err = r.redis.Set(blockNumberKey, strconv.Itoa(lastBlockRetrieved))
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

func (r *BlockRunner) getPriorBlock(blockRetriever engine.BlockRetriever, blockNumber int) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	time.Sleep(50 * time.Millisecond)
	completed := true

	blockNumberKey := generateBlockNumberKey(blockNumber)
	utils.Logger.Infof("Check if prior block retrieved: %s", blockNumberKey)
	_, err := r.redis.Get(blockNumberKey)
	if err != nil {
		utils.Logger.Infof("Prior block needed: %s", blockNumberKey)
		//BlockNumToGetChan <- i
		block := blockRetriever.GetBlock(blockNumber)

		completed = r.producerFactory.Produce(types.BLOCK_TOPIC, block)
		if completed {
			convertedBlock := types.Block{}.MongoFromGoType(block)
			utils.Logger.Infof("Transactions In Past Block: %d", len(convertedBlock.Transactions))

			go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
				receipts, err := engine.GetBlockReceipts("http://localhost:8545", blockHash)
				if err != nil {
					panic(err)
					return
				}
				for _, Receipt := range receipts {
					r.producerFactory.Produce(types.RECEIPT_TOPIC, *Receipt)
				}
			}(convertedBlock.Hash)

			go func(transactions []types.Transaction) {
				for _, tx := range transactions {
					r.producerFactory.Produce(types.TRANSACTION_TOPIC, tx)
				}
			}(block.Transactions)

			//	Dangerously assume save did not fail
			err = r.redis.Set(blockNumberKey, strconv.Itoa(blockNumber))
			if err != nil {
				panic(err)
			}

			_, err := r.redis.Del("missingBlock")
			if err != nil {
				utils.Logger.Errorf("missing block reset failed: %v", err)
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
		utils.Logger.Infof("Prior block %d ingested", blockNumber)
	}

	utils.Logger.Info("exiting: getPriorBlock External: For single block")
}

func generateBlockNumberKey(blockNumber int) string {
	formatInt := strconv.Itoa(blockNumber)
	return "B" + formatInt
}
