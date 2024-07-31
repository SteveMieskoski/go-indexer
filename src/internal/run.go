package internal

import (
	"src/engine"
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

	runs := engine.NewBlockRunner(idxConfig)

	//beaconBlockRunner.StartBeaconSync()
	//
	if !idxConfig.DisableBeacon {
		go func() {
			beaconBlockRunner := engine.NewBeaconBlockRunner(idxConfig)
			beaconBlockRunner.StartBeaconSync()
		}()
	}

	//runs.Demo()
	runs.StartBlockSync()

	//beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))
	//beaconBlockRunner.StartBeaconSync()

}
