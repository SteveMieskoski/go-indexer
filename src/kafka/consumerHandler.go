package kafka

import (
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"os"
	"os/signal"
	"src/mongodb"
	"src/postgres"
	protobuf2 "src/protobuf"
	"src/types"
	"src/utils"
	"syscall"
)

type ConsumerHandler struct {
	ready               chan bool
	terminateRun        bool
	DatabaseCoordinator mongodb.DatabaseCoordinator
	PostgresCoordinator postgres.PostgresDB
	PrimaryCoordinator  string
	IdxConfig           types.IdxConfigStruct
	ConsumerTopics      []string
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (consumer ConsumerHandler) Setup(sess sarama.ConsumerGroupSession) error {
	if consumer.IdxConfig.ClearConsumer {
		for _, aTopic := range consumer.ConsumerTopics {
			sess.ResetOffset(aTopic, 0, 0, "")
		}
	}

	// Mark the consumer as ready
	println("consumer.go:146 SETUP") // todo remove dev item
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (consumer ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
			/*start := time.Now()*/
			if !ok {
				utils.Logger.Infof("message channel was closed")
				return nil
			}

			//consumer.DatabaseCoordinator.MessageChannel() <- message

			switch message.Topic {
			case types.BLOCK_TOPIC:
				var block protobuf2.Block
				err := proto.Unmarshal(message.Value, &block)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddBlock() <- consumer.DatabaseCoordinator.ConvertToBlock(&block)
				break
			case types.RECEIPT_TOPIC:
				var receipt protobuf2.Receipt
				err := proto.Unmarshal(message.Value, &receipt)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddReceipt() <- consumer.DatabaseCoordinator.ConvertToReceipt(&receipt)

				for _, logVal := range receipt.Logs {
					consumer.DatabaseCoordinator.AddLog() <- consumer.DatabaseCoordinator.ConvertToLog(logVal)
				}
				break
			case types.BLOB_TOPIC:
				var blob protobuf2.Blob
				err := proto.Unmarshal(message.Value, &blob)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddBlob() <- consumer.DatabaseCoordinator.ConvertToBlob(&blob)
				break
			case types.TRANSACTION_TOPIC:
				var tx protobuf2.Transaction
				err := proto.Unmarshal(message.Value, &tx)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddTransaction() <- consumer.DatabaseCoordinator.ConvertToTransaction(&tx)
				break
			case types.ADDRESS_TOPIC:
				var addr protobuf2.AddressDetails
				err := proto.Unmarshal(message.Value, &addr)
				if err != nil {
					return err
				}
				consumer.DatabaseCoordinator.AddAddressBalance() <- consumer.DatabaseCoordinator.ConvertToAddress(&addr)
				break
			default:
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
