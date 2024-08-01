package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"log"
	"os"
	"time"
)

type ConsumeHandler struct {
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (consumer ConsumeHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (consumer ConsumeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer ConsumeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	return nil
}

func ConsumerGroup() {

	brokerUri := os.Getenv("BROKER_URI")
	brokers := []string{brokerUri}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {

	}
	groupId := "mongo"

	config := sarama.NewConfig()
	config.Version = version
	config.Metadata.Timeout = 30 * time.Second
	config.ClientID = "Indexer"
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Create a new consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, config)

	handler := ConsumeHandler{}

	for {
		err := consumerGroup.Consume(context.Background(), []string{topic}, handler)
		if err != nil {
			log.Printf("Error consuming messages: %v", err)
		}
	}
}
