package internal

import (
	"src/engine"
	"src/kafka"
	"src/mongodb"
)

var settings = mongodb.DatabaseSetting{
	Url:        "mongodb://localhost:27017",
	DbName:     "blocks",
	Collection: "blocks",
}

var brokers = []string{"localhost:9092"}

func Run() {

	runs := engine.NewBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))

	//beaconBlockRunner.StartBeaconSync()
	//
	//go func() {
	//	beaconBlockRunner := engine.NewBeaconBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))
	//	beaconBlockRunner.StartBeaconSync()
	//}()

	//runs.Demo()
	runs.StartBlockSync()
	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//defer stop()
	//
	//select {
	//case <-ctx.Done():
	//	stop()
	//	fmt.Println("signal received")
	//	return
	//}

}
