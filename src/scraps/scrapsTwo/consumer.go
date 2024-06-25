package scrapsTwo

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"src/mongodb"
	protobuf2 "src/protobuf"
	"syscall"
)

// Sarama configuration options
var (
	groupID  = "one"
	assignor = "roundrobin"
	oldest   = true
)

var (
	brokers   = []string{"localhost:9092"}
	version   = "7.0.0"
	topic     = "test"
	producers = 2
	verbose   = true

	recordsNumber int64 = 100

	//recordsRate = metrics.GetOrRegisterMeter("records.rate", nil)
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready        chan bool
	terminateRun bool
}

// need topics[topic] = handler
func NewConsumer(topics []string) {

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true

	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	//<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")
	// Consume messages
	for {
		handler := consumer

		err := client.Consume(ctx, topics, handler)
		if err != nil {
			panic(err)
		}

		select {
		case <-ctx.Done():
			log.Println("NewConsumer terminating: context cancelled")
			consumer.terminateRun = true
		case <-sigterm:
			log.Println("NewConsumer terminating: via signal")
			consumer.terminateRun = true
			cancel()
			err := client.Close()
			if err != nil {
				return
			}
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	var settings = mongodb.DatabaseSetting{
		Url:        "mongodb://localhost:27017",
		DbName:     "blocks",
		Collection: "blocks",
	}

	DatabaseCoordinator, _ := mongodb.NewDatabaseCoordinator(settings)

	for message := range claim.Messages() {

		var captured = false
		if message.Topic == "Block" {
			println("BLOCK")
			var block protobuf2.Block
			err := proto.Unmarshal(message.Value, &block)
			if err != nil {
				return err
			}
			DatabaseCoordinator.AddBlock() <- DatabaseCoordinator.ConvertToBlock(block)
			println("Consumed block", block.String())
			log.Printf("Message claimed: BlockNumber = %s, timestamp = %v, topic = %s", block.Number, message.Timestamp, message.Topic)
			captured = true
		}

		if message.Topic == "Receipt" {
			println("RECEIPT")
			var receipt protobuf2.Receipt
			err := proto.Unmarshal(message.Value, &receipt)
			if err != nil {
				return err
			}
			DatabaseCoordinator.AddReceipt() <- DatabaseCoordinator.ConvertToReceipt(receipt)

			for _, logVal := range receipt.Logs {
				DatabaseCoordinator.AddLog() <- DatabaseCoordinator.ConvertToLog(logVal)
			}
			println("Consumed receipt", receipt.String())
			log.Printf("Message claimed: TransactionHash = %s, timestamp = %v, topic = %s", receipt.TransactionHash, message.Timestamp, message.Topic)
			captured = true
		}
		if captured {
			session.MarkMessage(message, "")
		}
		select {
		case <-sigterm:
			log.Println("terminating: via signal: sigterm")
			DatabaseCoordinator.Close()
			session.Context().Done()
			return nil
		case <-sigusr1:
			log.Println("terminating: via signal: sigusr1")
			DatabaseCoordinator.Close()
			session.Context().Done()
			return nil
		}
	}

	return nil
}
