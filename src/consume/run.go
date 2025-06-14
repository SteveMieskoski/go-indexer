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

	DbConsumer([]string{types.RECEIPT_TOPIC, types.TRANSACTION_TOPIC, types.BLOB_TOPIC, types.ADDRESS_TOPIC, types.BLOCK_TOPIC}, DbCoordinator, idxConfig)
	println("ConsumerRun")

}
