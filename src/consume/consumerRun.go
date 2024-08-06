package consume

import (
	"os"
	"src/types"
)

func ConsumerRun(idxConfig types.IdxConfigStruct) {

	uri := os.Getenv("MONGO_URI")
	var settings = DatabaseSetting{
		Url:        uri,
		DbName:     "blocks",
		Collection: "blocks", // default Collection Name. Overridden in consumer.go
	}
	DbCoordinator, _ := NewDatabaseCoordinator(settings, idxConfig)
	go DbConsumer([]string{types.RECEIPT_TOPIC}, DbCoordinator, idxConfig)
	go DbConsumer([]string{types.TRANSACTION_TOPIC}, DbCoordinator, idxConfig)
	go DbConsumer([]string{types.BLOB_TOPIC}, DbCoordinator, idxConfig)
	go DbConsumer([]string{types.ADDRESS_TOPIC}, DbCoordinator, idxConfig)

	DbConsumer([]string{types.BLOCK_TOPIC}, DbCoordinator, idxConfig)

	//kafka.DbConsumer([]string{types.ADDRESS_TOPIC}, DbCoordinator, idxConfig)
	//kafka.DbConsumer([]string{types.TRANSACTION_TOPIC}, DbCoordinator, idxConfig)

	//kafka.DbConsumer([]string{types.RECEIPT_TOPIC, types.TRANSACTION_TOPIC, types.BLOB_TOPIC, types.ADDRESS_TOPIC, types.BLOCK_TOPIC}, DbCoordinator, idxConfig)
	println("ConsumerRun")

}
