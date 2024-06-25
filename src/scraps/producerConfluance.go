package scraps

//
//import (
//	"fmt"
//	"github.com/IBM/sarama"
//	"github.com/confluentinc/confluent-kafka-go/kafka"
//	"github.com/ethereum/go-ethereum/metrics"
//	"os"
//	"os/signal"
//	"sync"
//	"syscall"
//)
//
//var (
//	brokers   = "localhost:9092"
//	version   = "7.0.0"
//	topic     = "test"
//	producers = 1
//	verbose   = true
//
//	recordsNumber int64 = 100
//
//	recordsRate = metrics.GetOrRegisterMeter("records.rate", nil)
//)
//
//type Producer struct {
//	producer  *kafka.Producer
//	connected bool
//}
//// pool of producers that ensure transactional-id is unique.
//type producerProvider struct {
//	transactionIdGenerator int32
//
//	producersLock sync.Mutex
//	producers     []sarama.AsyncProducer
//
//	producerProvider func() sarama.AsyncProducer
//}
//
//func (k *Producer) connect() {
//	p, err := kafka.NewProducer(&kafka.ConfigMap{
//		// User-specific properties that you must set
//		"bootstrap.servers": brokers,
//
//		// Fixed properties
//		"acks": "all"})
//
//	if err != nil {
//		fmt.Printf("Failed to create producer: %s", err)
//		os.Exit(1)
//	}
//
//	sigs := make(chan os.Signal, 1)
//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
//
//	k.producer = p
//	k.connected = true
//
//	go func() {
//		for e := range p.Events() {
//			switch ev := e.(type) {
//			case *kafka.Message:
//				if ev.TopicPartition.Error != nil {
//					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
//				} else {
//					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
//						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
//				}
//			}
//		}
//	}()
//
//	sig := <-sigs
//	if sig == syscall.SIGINT || sig == syscall.SIGTERM {
//		println("exiting")
//		k.producer.Close()
//		return
//	}
//}
//
//func (k *Producer) Produce(topic string, message []byte) {
//	if !k.connected {
//		k.connect()
//	}
//	err := k.producer.Produce(&kafka.Message{
//		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
//		Key:            []byte(key),
//		Value:          []byte(data),
//	}, nil)
//	if err != nil {
//		return
//	}
//}
