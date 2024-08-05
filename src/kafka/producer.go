package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/ethereum/go-ethereum/metrics"
	"log"
	"os"
	"os/signal"
	"src/types"
	"src/utils"
	//"strconv"
	"sync"
	"syscall"
)

var (
	//brokers   = []string{"localhost:9092"}
	version   = "7.0.0"
	topic     = "test"
	producers = 6
	verbose   = true

	recordsNumber int64 = 100

	recordsRate = metrics.GetOrRegisterMeter("records.rate", nil)
)

//type Producer struct {
//	producer  *kafka.Producer
//	Connected bool
//}

// pool of producers that ensure transactional-id is unique.
type ProducerProvider struct {
	transactionIdGenerator int32

	producersLock sync.Mutex
	producers     []sarama.AsyncProducer

	ProducerProvider func() sarama.AsyncProducer

	Connected bool
}

func GenerateKafkaConfig() *sarama.Config {
	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Idempotent = true
	config.Producer.Return.Errors = false
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Producer.Transaction.Retry.Backoff = 10
	config.Producer.Transaction.ID = "txn_producer"
	config.Producer.MaxMessageBytes = 5000000
	config.Net.MaxOpenRequests = 1
	return config
}

func NewProducerProvider(brokers []string, producerConfigurationProvider func() *sarama.Config, idxConfig types.IdxConfigStruct) *ProducerProvider {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	brokerUri := os.Getenv("BROKER_URI")
	brokers = []string{brokerUri}

	provider := &ProducerProvider{}
	provider.ProducerProvider = func() sarama.AsyncProducer {

		config := producerConfigurationProvider()
		suffix := provider.transactionIdGenerator
		// Append transactionIdGenerator to current config.Producer.Transaction.ID to ensure transaction-id uniqueness.
		if config.Producer.Transaction.ID != "" {
			provider.transactionIdGenerator++
			config.Producer.Transaction.ID = config.Producer.Transaction.ID + "-" + fmt.Sprint(suffix)
		}
		producer, err := sarama.NewAsyncProducer(brokers, config)
		if err != nil {
			return nil
		}
		return producer
	}

	provider.Connected = true
	return provider
}

// ProduceBlock TODO: figure out how to properly generalize this function
func (p *ProducerProvider) Produce(topic string, block []byte) bool {
	producer := p.borrow()
	defer p.release(producer)

	// Start kafka transaction
	err := producer.BeginTxn()
	if err != nil {
		utils.Logger.Errorf("unable to start txn %s\n", err)
		return false
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(block),
	}

	producer.Input() <- msg

	// commit transaction
	err = producer.CommitTxn()
	// TODO: Implement appropriate error handling and recovery
	if err != nil {
		utils.Logger.Errorf("Producer: unable to commit txn %s\n", err)
		for {
			if producer.TxnStatus()&sarama.ProducerTxnFlagFatalError != 0 {
				// fatal error. need to recreate producer.
				utils.Logger.Errorf("Producer: producer is in a fatal state, need to recreate it")
				break
			}
			// If producer is in abortable state, try to abort current transaction.
			if producer.TxnStatus()&sarama.ProducerTxnFlagAbortableError != 0 {
				err = producer.AbortTxn()
				if err != nil {
					// If an error occured just retry it.
					utils.Logger.Errorf("Producer: unable to abort transaction: %+v", err)
					continue
				}
				break
			}
			// if not you can retry
			err = producer.CommitTxn()
			if err != nil {
				utils.Logger.Errorf("Producer: unable to commit txn %s\n", err)
				continue
			}
		}
		return false
	}
	return true
}

func (p *ProducerProvider) borrow() (producer sarama.AsyncProducer) {
	p.producersLock.Lock()
	defer p.producersLock.Unlock()

	if len(p.producers) == 0 {
		for {
			producer = p.ProducerProvider()
			if producer != nil {
				return
			}
		}
	}

	index := len(p.producers) - 1
	producer = p.producers[index]
	p.producers = p.producers[:index]
	return
}

func (p *ProducerProvider) release(producer sarama.AsyncProducer) {
	p.producersLock.Lock()
	defer p.producersLock.Unlock()

	// If released producer is erroneous close it and don't return it to the producer pool.
	if producer.TxnStatus()&sarama.ProducerTxnFlagInError != 0 {
		// Try to close it
		_ = producer.Close()
		return
	}
	p.producers = append(p.producers, producer)
}

func (p *ProducerProvider) clear() {
	p.producersLock.Lock()
	defer p.producersLock.Unlock()

	for _, producer := range p.producers {
		producer.Close()
	}
	p.producers = p.producers[:0]
}
