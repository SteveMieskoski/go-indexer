package consume

import (
	"fmt"
	"github.com/IBM/sarama"
	"src/types"
	"src/utils"
)

var consumptionLogModulo = 500

type DatabaseCoordinator interface {
	Close()
	MessageChannel() chan<- *sarama.ConsumerMessage
	ConsumeMessage(consumeChannels map[string]chan *sarama.ConsumerMessage) error
}

type databaseCoordinator struct {
	//BlockConsumer          *BlockConsumer
	//ReceiptConsumer        *ReceiptConsumer
	//TransactionConsumer    *TransactionConsumer
	//BlobConsumer           *BlobConsumer
	//AddressChannelConsumer *AddressChannelConsumer
	//AddressConsumer        *AddressConsumer
	consumers       map[string]TypeConsumer
	consumeChannels map[string]chan *sarama.ConsumerMessage
	messageChannel  chan *sarama.ConsumerMessage
}

func NewDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {
	return newDatabaseCoordinator(settings, idxConfig)
}

func newDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {

	AddressConsumerChannel := NewAddressChannelConsumer(idxConfig)

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
		//if cons.TopicLabel() == types.RECEIPT_TOPIC || cons.TopicLabel() == types.TRANSACTION_TOPIC {
		//	cons.AttachAddressChannel(AddressConsumerChannel.AddressChannel)
		//}
		cons.AttachAddressChannel(AddressConsumerChannel.AddressChannel)
		consumers[cons.TopicLabel()] = cons
	}

	dbc := &databaseCoordinator{
		//BlockConsumer:          NewBlockConsumer(settings, idxConfig, types.BLOCK_TOPIC),
		//ReceiptConsumer:        NewReceiptConsumer(settings, idxConfig, types.RECEIPT_TOPIC),
		//TransactionConsumer:    NewTransactionConsumer(settings, idxConfig, types.TRANSACTION_TOPIC),
		//BlobConsumer:           NewBlobConsumer(settings, idxConfig, types.BLOB_TOPIC),
		//AddressChannelConsumer: NewAddressChannelConsumer(idxConfig),
		//AddressConsumer:        NewAddressConsumer(idxConfig, types.ADDRESS_TOPIC),
		consumers:       consumers,
		consumeChannels: consumeChannels,
		messageChannel:  make(chan *sarama.ConsumerMessage, 5),
	}

	//consumeChannels := make(map[string]chan *sarama.ConsumerMessage)
	//consumeChannels[types.BLOCK_TOPIC] = dbc.BlockConsumer.Channel
	//consumeChannels[types.RECEIPT_TOPIC] = dbc.ReceiptConsumer.Channel
	//consumeChannels[types.TRANSACTION_TOPIC] = dbc.TransactionConsumer.Channel
	//consumeChannels[types.BLOB_TOPIC] = dbc.BlobConsumer.Channel
	//consumeChannels[types.ADDRESS_TOPIC] = dbc.AddressConsumer.Channel

	//dbc.ReceiptConsumer.AddressChannel = dbc.AddressChannelConsumer.AddressChannel
	//dbc.TransactionConsumer.AddressChannel = dbc.AddressChannelConsumer.AddressChannel

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

// TODO: I like this idea, but it appears like the channel is getting saturated from the consumer
func (db *databaseCoordinator) ConsumeMessage(consumeChannels map[string]chan *sarama.ConsumerMessage) error {

	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//defer stop()

	for message := range db.messageChannel {

		if message.Topic == types.TRANSACTION_TOPIC {
			fmt.Printf("messageChannel %s \n", message.Topic)
			consumeChannels[types.TRANSACTION_TOPIC] <- message
		} else {
			go func() {
				//fmt.Printf("messageChannel %s \n", message.Topic)
				consumeChannels[message.Topic] <- message
			}()
		}

	}

	return nil
}

func (db *databaseCoordinator) Close() {

	close(db.messageChannel)
}
