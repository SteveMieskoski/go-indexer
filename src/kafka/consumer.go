package kafka

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"src/mongodb"
	"src/postgres"
	protobuf2 "src/protobuf"
	"src/types"
	"src/utils"
	"sync"
	"syscall"
	"time"
)

// Sarama configuration options
//var (
//	group    = "one"
//	assignor = "roundrobin"
//	oldest   = true
//)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready               chan bool
	terminateRun        bool
	DatabaseCoordinator mongodb.DatabaseCoordinator
	PostgresCoordinator postgres.PostgresDB
	PrimaryCoordinator  string
}

// need topics[topic] = handler
func NewPostgresConsumer(topics []string, idxConfig types.IdxConfigStruct) {

	keepRunning := true
	utils.Logger.Info("Starting a new Sarama consumer")

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	version, err := sarama.ParseKafkaVersion(version)

	if err != nil {
		utils.Logger.Panicf("Error parsing Kafka version: %v", err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version
	config.Metadata.Timeout = 20 * time.Second
	config.ClientID = "Indexer"
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	postgresConsumer := postgres.NewClient(idxConfig)

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	consumer.PostgresCoordinator = *postgresConsumer
	consumer.PrimaryCoordinator = "POSTGRES"

	brokerUri := os.Getenv("BROKER_URI")
	brokers := []string{brokerUri}
	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, "postgres", config)

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
					return
				}
				utils.Logger.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	utils.Logger.Info("Sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

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

func NewMongoDbConsumer(topics []string, idxConfig types.IdxConfigStruct) {

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

	uri := os.Getenv("MONGO_URI")
	var settings = mongodb.DatabaseSetting{
		Url:        uri,
		DbName:     "blocks",
		Collection: "blocks", // default Collection Name. Overridden in consumer.go
	}

	brokerUri := os.Getenv("BROKER_URI")
	brokers := []string{brokerUri}
	// --------- reset offsets (start) --------------------------
	//
	//broker := sarama.NewBroker(brokerUri)
	//err = broker.Open(nil)
	//if err != nil {
	//	panic(err)
	//}
	//
	//versionNum, _ := strconv.ParseInt(version.String(), 10, 0)
	//
	//_, err = broker.DeleteOffsets(&sarama.DeleteOffsetsRequest{
	//	Version: int16(versionNum),
	//	Group:   groupId,
	//})
	//if err != nil {
	//	utils.Logger.Info("Error deleting offsets: %v", err)
	//	return
	//}
	// --------- reset offsets (end) --------------------------
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
	consumer := Consumer{
		ready: make(chan bool),
	}
	DbCoordinator, _ := mongodb.NewDatabaseCoordinator(settings, idxConfig)
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

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	println("consumer.go:146 SETUP") // todo remove dev item
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

	for {
		if consumer.terminateRun {
			switch consumer.PrimaryCoordinator {
			case "MONGO":
				consumer.DatabaseCoordinator.Close()
			case "POSTGRES":
			default:
			}

			session.Context().Done()
			return nil
		}
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				utils.Logger.Infof("message channel was closed")
				return nil
			}

			if message.Topic == types.BLOCK_TOPIC {
				var block protobuf2.Block
				err := proto.Unmarshal(message.Value, &block)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddBlock() <- consumer.DatabaseCoordinator.ConvertToBlock(block)
				//utils.Logger.Infof("Message claimed: BlockNumber = %s, timestamp = %v, topic = %s", block.Number, message.Timestamp, message.Topic)
			}

			if message.Topic == types.RECEIPT_TOPIC {
				var receipt protobuf2.Receipt
				err := proto.Unmarshal(message.Value, &receipt)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddReceipt() <- consumer.DatabaseCoordinator.ConvertToReceipt(receipt)

				for _, logVal := range receipt.Logs {
					consumer.DatabaseCoordinator.AddLog() <- consumer.DatabaseCoordinator.ConvertToLog(logVal)
				}
				//println("Consumed receipt", receipt.String())
				//utils.Logger.Infof("Message claimed: TransactionHash = %s, timestamp = %v, topic = %s", receipt.TransactionHash, message.Timestamp, message.Topic)
			}

			if message.Topic == types.BLOB_TOPIC {
				var blob protobuf2.Blob
				err := proto.Unmarshal(message.Value, &blob)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddBlob() <- consumer.DatabaseCoordinator.ConvertToBlob(&blob)

				//utils.Logger.Infof("Message claimed: Blob = %s, timestamp = %v, topic = %s", blob.GetKzgCommitment(), message.Timestamp, message.Topic)
			}

			if message.Topic == types.TRANSACTION_TOPIC {
				var tx protobuf2.Transaction
				err := proto.Unmarshal(message.Value, &tx)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddTransaction() <- consumer.DatabaseCoordinator.ConvertToTransaction(&tx)

				//utils.Logger.Infof("Message claimed: Tx Hash = %s, timestamp = %v, topic = %s", tx.Hash, message.Timestamp, message.Topic)

			}

			if message.Topic == types.ADDRESS_TOPIC {
				var addr protobuf2.AddressDetails
				err := proto.Unmarshal(message.Value, &addr)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddAddressBalance() <- consumer.DatabaseCoordinator.ConvertToAddress(&addr)

				//utils.Logger.Infof("Message claimed: Address = %s, timestamp = %v, topic = %s", addr.Address, message.Timestamp, message.Topic)

			}

			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		case <-sigterm:
			utils.Logger.Info("terminating: via signal: sigterm")
			consumer.DatabaseCoordinator.Close()
			session.Context().Done()
			return nil
		case <-sigusr1:
			utils.Logger.Info("terminating: via signal: sigusr1")
			consumer.DatabaseCoordinator.Close()
			session.Context().Done()
			return nil
		}
	}
}
