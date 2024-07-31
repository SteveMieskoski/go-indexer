package internal

import (
	"os"
	"src/kafka"
	"src/mongodb"
	"src/types"
)

func ConsumerRun(idxConfig types.IdxConfigStruct) {

	uri := os.Getenv("MONGO_URI")
	var settings = mongodb.DatabaseSetting{
		Url:        uri,
		DbName:     "blocks",
		Collection: "blocks", // default Collection Name. Overridden in consumer.go
	}
	DbCoordinator, _ := mongodb.NewDatabaseCoordinator(settings, idxConfig)
	go func() {
		kafka.NewMongoDbConsumer([]string{types.RECEIPT_TOPIC}, DbCoordinator, idxConfig)
	}()
	go func() {
		kafka.NewMongoDbConsumer([]string{types.TRANSACTION_TOPIC}, DbCoordinator, idxConfig)
	}()
	//go func() {
	//	kafka.NewMongoDbConsumer([]string{types.BLOB_TOPIC}, idxConfig)
	//}()
	//go func() {
	//	kafka.NewMongoDbConsumer([]string{types.ADDRESS_TOPIC}, idxConfig)
	//}()
	//
	//kafka.NewMongoDbConsumer([]string{types.BLOCK_TOPIC}, idxConfig)

	kafka.NewMongoDbConsumer([]string{types.ADDRESS_TOPIC}, DbCoordinator, idxConfig)
	//kafka.NewMongoDbConsumer([]string{types.TRANSACTION_TOPIC}, idxConfig)

	println("ConsumerRun")

}
