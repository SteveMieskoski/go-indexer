package internal

import (
	"fmt"
	"src/engine"
	"src/kafka"
	"src/mongodb"
	"src/mongodb/models"
	"src/utils"
)

var settings = mongodb.DatabaseSetting{
	Url:        "mongodb://localhost:27017",
	DbName:     "blocks",
	Collection: "blocks",
}

var brokers = []string{"localhost:9092"}

func Run() {

	//client, err := mongodb.GetClient(settings)
	//if err != nil {
	//	return
	//}
	//
	//dbSettings := &mongodb.DatabaseSetting{
	//	Url:        "mongodb://localhost:27017",
	//	DbName:     "test",
	//	Collection: "block",
	//}
	//receiptDbSettings := *dbSettings
	//receiptDbSettings.Collection = "receipts"
	//BlockRepository := documents.NewBlockRepository(client, dbSettings)
	//ReceiptRepository := documents.NewReceiptRepository(client, &receiptDbSettings)

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)
	//producerFactory.ProducerProvider()
	//producerFactory.ProducerProvider()

	fmt.Printf("producer ready: %t\n", producerFactory.Connected)
	blockGen := engine.GetBlocks("ws://localhost:8546")

	for block := range blockGen {
		utils.Logger.Infof("latest block:", block.Number)
		producerFactory.ProduceBlock("Block", block)
		convertedBlock := models.BlockFromGoType(block)

		go func(blockHash string) { // TODO: make sure to clean up goroutine (i.e. connections)
			receiptChan := engine.GetBlockReceipts("http://localhost:8545", blockHash)
			for receipt := range receiptChan {
				for _, aReceipt := range receipt {
					producerFactory.ProduceReceipt("Receipt", *aReceipt)
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
