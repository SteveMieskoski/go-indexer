package internal

import (
	"os"
	"src/produce"

	//"src/kafka"
	//"src/mongodb"
	"src/types"
)

//var settings = mongodb.DatabaseSetting{
//	Url:        "mongodb://localhost:27017",
//	DbName:     "blocks",
//	Collection: "blocks",
//}

//var brokers = []string{"localhost:9092"}

func Run(idxConfig types.IdxConfigStruct) {

	brokers := os.Getenv("BROKER_URI")
	//producerFactory := kafka.NewProducerProvider([]string{brokers}, kafka.GenerateKafkaConfig, idxConfig)

	blockProcessor := produce.NewBlockProcessor(produce.NewProducerProvider([]string{brokers}, produce.GenerateKafkaConfig, idxConfig), idxConfig)
	runs := produce.NewBlockRunner(blockProcessor, idxConfig)

	//beaconBlockRunner.StartBeaconSync()
	//
	//if !idxConfig.DisableBeacon {
	//	go func() {
	//		beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider([]string{brokers}, kafka.GenerateKafkaConfig, idxConfig), idxConfig)
	//		beaconBlockRunner.StartBeaconSync()
	//	}()
	//}

	//runs.Demo()
	runs.StartBlockSync()

	//beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))
	//beaconBlockRunner.StartBeaconSync()

}
