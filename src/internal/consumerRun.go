package internal

import (
	"src/kafka"
	"src/types"
)

func ConsumerRun() {

	go func() {
		kafka.NewMongoDbConsumer([]string{types.RECEIPT_TOPIC})
	}()
	go func() {
		kafka.NewMongoDbConsumer([]string{types.TRANSACTION_TOPIC})
	}()
	go func() {
		kafka.NewMongoDbConsumer([]string{types.BLOB_TOPIC})
	}()
	go func() {
		kafka.NewMongoDbConsumer([]string{types.ADDRESS_TOPIC})
	}()

	kafka.NewMongoDbConsumer([]string{types.BLOCK_TOPIC})

	//kafka.NewMongoDbConsumer([]string{types.ADDRESS_TOPIC})
	//kafka.NewMongoDbConsumer([]string{types.TRANSACTION_TOPIC})

}
