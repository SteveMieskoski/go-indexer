package internal

import (
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

	runs := NewBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))

	//beaconBlockRunner.StartBeaconSync()
	//
	go func() {
		beaconBlockRunner := NewBeaconBlockRunner(kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig))
		beaconBlockRunner.StartBeaconSync()
	}()

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
