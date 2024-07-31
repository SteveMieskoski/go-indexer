package internal

import (
	"os"
	"src/engine"
	"src/kafka"

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

	runs := engine.NewBlockRunner(kafka.NewProducerProvider([]string{brokers}, kafka.GenerateKafkaConfig, idxConfig), idxConfig)

	//beaconBlockRunner.StartBeaconSync()
	//
	if !idxConfig.DisableBeacon {
		go func() {
			beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider([]string{brokers}, kafka.GenerateKafkaConfig, idxConfig), idxConfig)
			beaconBlockRunner.StartBeaconSync()
		}()
	}

	//runs.Demo()
	runs.StartBlockSync()

	//beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))
	//beaconBlockRunner.StartBeaconSync()

}
