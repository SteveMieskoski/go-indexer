package kafka

import (
	"github.com/IBM/sarama"
	"os"
	"src/types"
	"src/utils"
	"strconv"
	"time"
)

func ResetKafka() {
	// DELETES/CLEARS EXISTING TOPICS
	utils.Logger.Infof("DELETING/CLEARING EXISTING TOPICS")

	brokerUri := os.Getenv("BROKER_URI")
	broker := sarama.NewBroker(brokerUri)
	err := broker.Open(nil)
	if err != nil {
		panic(err)
	}

	versionNum, _ := strconv.ParseInt(version, 10, 0)
	_, err = broker.DeleteTopics(&sarama.DeleteTopicsRequest{
		Version: int16(versionNum),
		Topics:  []string{types.TRANSACTION_TOPIC, types.RECEIPT_TOPIC, types.BLOCK_TOPIC, types.LOG_TOPIC, types.BLOB_TOPIC, types.ADDRESS_TOPIC},
	})
	if err != nil {
		utils.Logger.Errorf("Producer: unable to delete topics %s\n", err)
	}
	//utils.Logger.Infof("waiting for Kafka to finish clearing")
	//time.Sleep(10 * time.Second)
	tpcs := make(map[string]*sarama.TopicDetail)
	tpcs[types.TRANSACTION_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	tpcs[types.RECEIPT_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	tpcs[types.BLOCK_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	tpcs[types.LOG_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	tpcs[types.BLOB_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	tpcs[types.ADDRESS_TOPIC] = &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: -1,
	}
	broker.CreateTopics(&sarama.CreateTopicsRequest{
		Version:      int16(versionNum),
		TopicDetails: tpcs,
		Timeout:      5 * time.Second,
		ValidateOnly: false,
	})
	//broker.GetMetadata()
	utils.Logger.Infof("waiting for Kafka to finish resetting")
	time.Sleep(8 * time.Second)
	// DeleteTopicsRequest
}
