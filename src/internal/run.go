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

	producerFactory := kafka.NewProducerProvider(brokers, kafka.GenerateKafkaConfig)

	runs := NewBlockRunner(producerFactory)
	beaconBlockRunner := NewBeaconBlockRunner(producerFactory)

	go func() {
		beaconBlockRunner.getCurrentBeaconBlock()
	}()

	runs.getCurrentBlock()
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
