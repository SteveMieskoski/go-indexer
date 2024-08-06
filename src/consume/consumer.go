package consume

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"log"
	"os"
	"os/signal"
	"src/postgres"
	"src/types"
	"src/utils"
	"sync"
	"syscall"
	"time"
)

var (
	version   = "7.0.0"
	topic     = "test"
	producers = 6
	verbose   = true

	recordsNumber int64 = 100

	//recordsRate = metrics.GetOrRegisterMeter("records.rate", nil)
)

// Consumer Sarama consumer group consumer
type Consumer struct {
	ready               chan bool
	terminateRun        bool
	DatabaseCoordinator DatabaseCoordinator
	PostgresCoordinator postgres.PostgresDB
	PrimaryCoordinator  string
	IdxConfig           types.IdxConfigStruct
	ConsumerTopics      []string
}

func DbConsumer(topics []string, DbCoordinator DatabaseCoordinator, idxConfig types.IdxConfigStruct) {

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	keepRunning := true
	utils.Logger.Info("Starting a new Sarama consumer")

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	version, err := sarama.ParseKafkaVersion(version)
	groupId := "mongo"

	if err != nil {
		utils.Logger.Panicf("Error parsing Kafka version: %v", err)
	}

	brokerUri := os.Getenv("BROKER_URI")
	brokers := []string{brokerUri}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version
	config.Metadata.Timeout = 30 * time.Second
	config.ClientID = "Indexer"
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := ConsumerHandler{
		ready:          make(chan bool),
		IdxConfig:      idxConfig,
		ConsumerTopics: topics,
	}

	consumer.DatabaseCoordinator = DbCoordinator
	consumer.PrimaryCoordinator = "MONGO"

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, groupId, config)

	if err != nil {
		utils.Logger.Panicf("Error creating consumer group client: %v", err)
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
					utils.Logger.Panicf("Error from consumer: %v", err)
					return
				}
				utils.Logger.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
			select {
			case <-sigterm:
				utils.Logger.Info("consumer loop terminating: via signal")
				return
			default:
			}
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	utils.Logger.Info("Sarama consumer up and running!...")

	for keepRunning {
		select {
		case <-ctx.Done():
			utils.Logger.Info("NewConsumer terminating: context cancelled")
			keepRunning = false
			consumer.terminateRun = true
		case <-sigterm:
			utils.Logger.Info("NewConsumer terminating: via signal")
			keepRunning = false
			consumer.terminateRun = true
			cancel()
			err := client.Close()
			if err != nil {
				return
			}
			return
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
		utils.Logger.Info("Resuming consumption")
	} else {
		client.PauseAll()
		utils.Logger.Info("Resuming consumption")
	}

	*isPaused = !*isPaused
}
