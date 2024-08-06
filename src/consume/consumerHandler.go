package consume

import (
	"github.com/IBM/sarama"
	"os"
	"os/signal"
	"src/types"
	"src/utils"
	"syscall"
	"time"
)

type ConsumerHandler struct {
	ready               chan bool
	terminateRun        bool
	DatabaseCoordinator DatabaseCoordinator
	PostgresCoordinator PostgresDB
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
		//if consumer.terminateRun {
		//	switch consumer.PrimaryCoordinator {
		//	case "MONGO":
		//		consumer.DatabaseCoordinator.Close()
		//	case "POSTGRES":
		//	default:
		//	}
		//
		//	session.Context().Done()
		//	return nil
		//}
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				utils.Logger.Infof("message channel was closed")
				return nil
			}

			consumer.DatabaseCoordinator.MessageChannel() <- message

			<-consumer.DatabaseCoordinator.ReceiptChannel()
			time.Sleep(10 * time.Millisecond) // add a slight delay as the processing side (postgres) is taking just a little too long and can cause the next entry to get lost
			//println(val)
			utils.Logger.Infof("continue")

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
