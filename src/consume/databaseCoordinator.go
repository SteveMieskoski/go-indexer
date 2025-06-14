package consume

import (
	"fmt"
	"github.com/IBM/sarama"
	"src/types"
	"src/utils"
	"time"
)

var consumptionLogModulo = 500

type DatabaseCoordinator interface {
	Close()
	MessageChannel() chan<- *sarama.ConsumerMessage
	ReceiptChannel() chan int64
	ConsumeMessage(consumeChannels map[string]chan *sarama.ConsumerMessage) error
}

type databaseCoordinator struct {
	consumers             map[string]TypeConsumer
	consumeChannels       map[string]chan *sarama.ConsumerMessage
	messageChannel        chan *sarama.ConsumerMessage
	messageReceiptChannel chan int64
}

func NewDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {
	return newDatabaseCoordinator(settings, idxConfig)
}

func newDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {

	AddressConsumerChannel := NewAddressChannelConsumer(idxConfig)

	go AddressConsumerChannel.MonitorChannel()

	consumerList := []TypeConsumer{
		NewBlockConsumer(settings, idxConfig, types.BLOCK_TOPIC),
		NewReceiptConsumer(settings, idxConfig, types.RECEIPT_TOPIC),
		NewTransactionConsumer(settings, idxConfig, types.TRANSACTION_TOPIC),
		NewBlobConsumer(settings, idxConfig, types.BLOB_TOPIC),
		NewAddressConsumer(idxConfig, types.ADDRESS_TOPIC)}

	consumeChannels := make(map[string]chan *sarama.ConsumerMessage)
	consumers := make(map[string]TypeConsumer)

	for _, cons := range consumerList {
		consumeChannels[cons.TopicLabel()] = cons.GetChannel()
		cons.AttachAddressChannel(AddressConsumerChannel.AddressChannel)
		consumers[cons.TopicLabel()] = cons
		go cons.MonitorChannel()
	}

	dbc := &databaseCoordinator{
		consumers:             consumers,
		consumeChannels:       consumeChannels,
		messageChannel:        make(chan *sarama.ConsumerMessage, len(consumerList)*5),
		messageReceiptChannel: make(chan int64),
	}

	dbc.consumeChannels = consumeChannels

	go func() {
		err := dbc.ConsumeMessage(consumeChannels)
		if err != nil {
			utils.Logger.Errorf("Error from consumer for log: %v", err)
		}
	}()

	return dbc, nil
}

func (db *databaseCoordinator) MessageChannel() chan<- *sarama.ConsumerMessage {
	return db.messageChannel
}

func (db *databaseCoordinator) ReceiptChannel() chan int64 {
	return db.messageReceiptChannel
}

// TODO: I like this idea, but it appears like the channel is getting saturated from the consumer
func (db *databaseCoordinator) ConsumeMessage(consumeChannels map[string]chan *sarama.ConsumerMessage) error {

	var counter int64 = 0
	for message := range db.messageChannel {
		counter++
		go db.sendOrRetry(consumeChannels[message.Topic], message, counter, false, 1)
	}

	return nil
}

func (db *databaseCoordinator) sendOrRetry(sender chan *sarama.ConsumerMessage, message *sarama.ConsumerMessage, counter int64, confirmSent bool, tryCount int) {
	if tryCount > 20 {
		return
	}
	select {
	case sender <- message:
		if !confirmSent {
			db.messageReceiptChannel <- counter // don't get hung up if the consuming channel is blocking
		}

	default:

		if !confirmSent {
			fmt.Printf("Receiving channel not available for topic %s \n", message.Topic)
			db.messageReceiptChannel <- counter // don't get hung up if the consuming channel is blocking
		} else {
			fmt.Printf("Try %d - Receiving channel not available for topic %s \n", tryCount, message.Topic)
		}
		var holder sarama.ConsumerMessage
		holder = *message
		retryDelay := time.Duration(tryCount * 100)
		time.Sleep(retryDelay * time.Millisecond)
		tryCount++
		db.sendOrRetry(sender, &holder, 0, true, tryCount)
	}
}

func (db *databaseCoordinator) Close() {

	close(db.messageChannel)
}
