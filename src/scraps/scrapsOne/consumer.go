package scrapsOne

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"src/mongodb"
	protobuf2 "src/protobuf"
	"sync"
	"syscall"
	"time"
)

// Sarama configuration options
var (
	group    = "one"
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

	keepRunning := true
	log.Println("Starting a new Sarama consumer")

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	}

	//config.Consumer.Group.ResetInvalidOffsets = true
	config.Consumer.Group.Rebalance.Timeout = 120 * time.Second
	config.Consumer.MaxWaitTime = 1 * time.Second

	config.ClientID = "Indexer"
	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	consumptionIsPaused := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, topics, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("NewConsumer terminating: context cancelled")
			keepRunning = false
			consumer.terminateRun = true
		case <-sigterm:
			log.Println("NewConsumer terminating: via signal")
			keepRunning = false
			consumer.terminateRun = true
			cancel()
			err := client.Close()
			if err != nil {
				return
			}
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}

	*isPaused = !*isPaused
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

	for {
		if consumer.terminateRun {
			DatabaseCoordinator.Close()
			session.Context().Done()
			return nil
		}
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}

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
			}

			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
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
}
