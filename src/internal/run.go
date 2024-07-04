package internal

import (
	"fmt"
	"os"
	"os/signal"
	"src/engine"
	"src/kafka"
	"src/mongodb"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"syscall"
)

var settings = mongodb.DatabaseSetting{
	Url:        "mongodb://localhost:27017",
	DbName:     "blocks",
	Collection: "blocks",
}

var brokers = []string{"localhost:9092"}

func Run() {

	var priorRetrievalInProgress bool = false

	redisClient := redisdb.NewClient()

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	blockRetrievers := engine.NewBlockRetriever(*redisClient)

	fmt.Printf("producer ready: %t\n", producerFactory.Connected)
	blockGen := blockRetrievers.GetBlocks()

	lastBlock, _ := redisClient.Get("B1")
	fmt.Printf(lastBlock)
	for block := range blockGen {
		utils.Logger.Infof("latest block: %d", block.Number)
		formatBlockNumberInt := strconv.Itoa(int(block.Number))
		err := redisClient.Set("latestBlock", formatBlockNumberInt)
		if err != nil {
			panic(err)
		}

		err = redisClient.Set("B"+formatBlockNumberInt, formatBlockNumberInt)
		if err != nil {
			return
		}

		if !priorRetrievalInProgress && block.Number > 1 {
			priorRetrievalInProgress = true
			go func() {
				getPriorBlocks(*blockRetrievers, *redisClient)
			}()

		}
		producerFactory.Produce("Block", block)
		convertedBlock := types.Block{}.FromGoType(block)

		go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
			receiptChan := engine.GetBlockReceipts("http://localhost:8545", blockHash)
			for receipt := range receiptChan {
				for _, aReceipt := range receipt {
					producerFactory.Produce("Receipt", *aReceipt)
				}
			}
		}(convertedBlock.Hash)

	}
	//defer func(client *mongo.Client, ctx context.Context) {
	//	err := client.Disconnect(ctx)
	//	if err != nil {
	//
	//	}
	//}(client, context.TODO())
}

func getPriorBlocks(blockRetriever engine.BlockRetriever, redisClient redisdb.RedisClient) {

	//BlockNumToGetChan := make(chan int)
	//defer close(BlockNumToGetChan)
	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//blockRetriever.GetPastBlocks(BlockNumToGetChan)
	lastBlockAsString, _ := redisClient.Get("latestBlock")
	lastBlock, _ := strconv.Atoi(lastBlockAsString)
	utils.Logger.Infof("latest block known: %d", lastBlock)
	for i := 1; i < lastBlock; i++ {
		blockNumberKey := generateBlockNumberKey(i)
		utils.Logger.Infof("Check if prior block retrieved: %s", blockNumberKey)
		_, err := redisClient.Get(blockNumberKey)
		if err != nil {
			utils.Logger.Infof("Prior block needed: %s", blockNumberKey)
			//BlockNumToGetChan <- i
			block := blockRetriever.GetPastBlocks(i)

			producerFactory.Produce("Block", block)
			convertedBlock := types.Block{}.FromGoType(block)

			go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
				receiptChan := engine.GetBlockReceipts("http://localhost:8545", blockHash)
				for receipt := range receiptChan {
					for _, aReceipt := range receipt {
						producerFactory.Produce("Receipt", *aReceipt)
					}
				}
			}(convertedBlock.Hash)
			//	Dangerously assume save did not fail
			err = redisClient.Set(blockNumberKey, strconv.Itoa(i))
			if err != nil {
				panic(err)
			}
		}
		println("check 2")
		//sig := <-sigs
		//if sig == syscall.SIGINT || sig == syscall.SIGTERM {
		//	utils.Logger.Info("exiting: getPriorBlocks")
		//	return
		//}
	}

	utils.Logger.Info("exiting: getPriorBlock External")
}

func generateBlockNumberKey(blockNumber int) string {
	formatInt := strconv.Itoa(blockNumber)
	return "B" + formatInt
}
