package consume

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
	protobuf2 "src/protobuf"
	"src/types"
	"src/utils"
	"strconv"
)

type TypeConsumer interface {
	MonitorChannel()
	GetChannel() chan *sarama.ConsumerMessage
	TopicLabel() string
	AttachAddressChannel(chan AddressChannelMessage)
}

type BlockConsumer struct {
	Channel         chan *sarama.ConsumerMessage
	Topic           string
	BlockRepository BlockRepository
}

func NewBlockConsumer(settings DatabaseSetting, idxConfig types.IdxConfigStruct, Topic string) *BlockConsumer {
	client, err := GetClient(settings, idxConfig)
	if err != nil {
		panic(err)
	}
	blockDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "blocks",
	}

	return &BlockConsumer{
		BlockRepository: NewBlockRepository(client, blockDbSettings),
		Channel:         make(chan *sarama.ConsumerMessage),
		Topic:           Topic,
	}
}

func (db *BlockConsumer) MonitorChannel() {
	receivedCount := 0
	for msg := range db.Channel {
		receivedCount++
		go func(message *sarama.ConsumerMessage) {
			var block protobuf2.Block
			err := proto.Unmarshal(message.Value, &block)

			_, err = db.BlockRepository.Add(*types.Block{}.MongoFromProtobufType(&block), context.Background())
			if err != nil {
				utils.Logger.Errorf("Error from consumer for block: %v", err)
				//return
			}
			if receivedCount%consumptionLogModulo == 0 {
				fmt.Printf("Processed %d Blocks\n", receivedCount)
			}
		}(msg)

	}
}

func (db *BlockConsumer) TopicLabel() string {
	return db.Topic
}

func (db *BlockConsumer) GetChannel() chan *sarama.ConsumerMessage {
	return db.Channel
}

func (db *BlockConsumer) AttachAddressChannel(chan AddressChannelMessage) {

}

type ReceiptConsumer struct {
	Channel           chan *sarama.ConsumerMessage
	Topic             string
	AddressChannel    chan AddressChannelMessage
	ReceiptRepository ReceiptRepository
	LogRepository     LogRepository
}

func NewReceiptConsumer(settings DatabaseSetting, idxConfig types.IdxConfigStruct, Topic string) *ReceiptConsumer {

	client, err := GetClient(settings, idxConfig)
	if err != nil {
		panic(err)
	}

	receiptDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "receipts",
	}

	logDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "logs",
	}

	return &ReceiptConsumer{
		Channel:           make(chan *sarama.ConsumerMessage),
		Topic:             Topic,
		AddressChannel:    make(chan AddressChannelMessage),
		ReceiptRepository: NewReceiptRepository(client, receiptDbSettings),
		LogRepository:     NewLogRepository(client, logDbSettings),
	}
}
func (db *ReceiptConsumer) MonitorChannel() {
	receivedCount := 0
	for msg := range db.Channel {
		receivedCount++
		go func(message *sarama.ConsumerMessage) {

			var receipt protobuf2.Receipt
			err := proto.Unmarshal(message.Value, &receipt)

			bnum, _ := strconv.ParseInt(receipt.BlockNumber, 16, 64)

			// indicates contract creation so, it wouldn't already exist
			if receipt.ContractAddress != "0x0" {
				db.AddressChannel <- AddressChannelMessage{
					action: "addressContract",
					value:  types.Address{Address: receipt.ContractAddress, IsContract: true, LastSeen: bnum},
				}
				//db.AddContractAddress() <- &types.Address{Address: receipt.ContractAddress, IsContract: true, LastSeen: bnum}
			}
			_, err = db.ReceiptRepository.Add(*types.Receipt{}.MongoFromProtobufType(&receipt), context.Background())

			if err != nil {
				utils.Logger.Errorf("Error from consumer for receipt: %v", err)
				//return
			}
			if receivedCount%consumptionLogModulo == 0 {
				fmt.Printf("Processed %d Receipts\n", receivedCount)
			}

		}(msg)

	}
}

func (db *ReceiptConsumer) TopicLabel() string {
	return db.Topic
}

func (db *ReceiptConsumer) GetChannel() chan *sarama.ConsumerMessage {
	return db.Channel
}

func (db *ReceiptConsumer) AttachAddressChannel(AddressChannel chan AddressChannelMessage) {
	db.AddressChannel = AddressChannel
}

func (db *ReceiptConsumer) extractLogs(logs []protobuf2.Log) {
	receivedCount := 0
	for _, log := range logs {
		receivedCount++
		_, err := db.LogRepository.Add(*types.Log{}.MongoFromProtobufType(log), context.Background())
		if err != nil {
			utils.Logger.Errorf("Error from consumer for log: %v", err)
			//return
		}

		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Logs\n", receivedCount)
		}
	}
	fmt.Printf("receivedCount: %d \n", receivedCount)
}

type TransactionConsumer struct {
	Channel               chan *sarama.ConsumerMessage
	Topic                 string
	TransactionRepository TransactionRepository
	AddressChannel        chan AddressChannelMessage
}

func NewTransactionConsumer(settings DatabaseSetting, idxConfig types.IdxConfigStruct, Topic string) *TransactionConsumer {
	client, err := GetClient(settings, idxConfig)
	if err != nil {
		panic(err)
	}
	transactionDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "transactions",
	}
	fmt.Printf("TransactionConsumer READY \n")
	return &TransactionConsumer{
		Channel:               make(chan *sarama.ConsumerMessage),
		Topic:                 Topic,
		AddressChannel:        make(chan AddressChannelMessage),
		TransactionRepository: NewTransactionRepository(client, transactionDbSettings),
	}
}

func (db *TransactionConsumer) TopicLabel() string {
	return db.Topic
}

func (db *TransactionConsumer) GetChannel() chan *sarama.ConsumerMessage {
	return db.Channel
}

func (db *TransactionConsumer) AttachAddressChannel(AddressChannel chan AddressChannelMessage) {
	db.AddressChannel = AddressChannel
}

func (db *TransactionConsumer) MonitorChannel() {
	receivedCount := 0
	for msg := range db.Channel {
		receivedCount++
		go func(message *sarama.ConsumerMessage) {
			fmt.Printf("TransactionConsumer %d \n", receivedCount)
			var tx protobuf2.Transaction
			err := proto.Unmarshal(message.Value, &tx)
			if err != nil {
				utils.Logger.Errorf("Error from unmarshal for transaction: %v", err)
				//return
			}

			bnum, _ := strconv.ParseInt(tx.BlockNumber, 16, 64)

			select {
			case db.AddressChannel <- AddressChannelMessage{
				action: "addressDetail",
				value:  types.Address{Address: tx.From, IsContract: false, Nonce: int64(tx.Nonce), LastSeen: bnum},
			}:
				fmt.Printf("addressDetail sent\n")
			case db.AddressChannel <- AddressChannelMessage{
				action: "addressOnly",
				value:  types.Address{Address: tx.To, LastSeen: bnum},
			}:
				fmt.Printf("addressOnly sent\n")
			default:
				fmt.Printf("AddressChannel not available for topic %s \n", message.Topic)
			}
			//db.AddressChannel <- AddressChannelMessage{
			//	action: "addressDetail",
			//	value:  types.Address{Address: tx.From, IsContract: false, Nonce: int64(tx.Nonce), LastSeen: bnum},
			//}
			//db.AddressChannel <- AddressChannelMessage{
			//	action: "addressOnly",
			//	value:  types.Address{Address: tx.To, LastSeen: bnum},
			//}
			//fmt.Printf("TransactionConsumer not available for topic %s \n", message.Topic)
			go func(transaction *types.MongoTransaction) {
				_, err := db.TransactionRepository.Add(*transaction, context.Background())

				if err != nil {
					utils.Logger.Errorf("Error from consumer for transaction: %v", err)
					//return
				}
			}(types.Transaction{}.MongoFromProtobufType(tx))

			if receivedCount%consumptionLogModulo == 0 {
				fmt.Printf("Processed %d Transactions\n", receivedCount)
			}
		}(msg)

	}
}

type BlobConsumer struct {
	Channel        chan *sarama.ConsumerMessage
	Topic          string
	BlobRepository BlobRepository
}

func NewBlobConsumer(settings DatabaseSetting, idxConfig types.IdxConfigStruct, Topic string) *BlobConsumer {
	client, err := GetClient(settings, idxConfig)
	if err != nil {
		panic(err)
	}
	blobDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "blobs",
	}
	return &BlobConsumer{
		Channel:        make(chan *sarama.ConsumerMessage),
		Topic:          Topic,
		BlobRepository: NewBlobRepository(client, blobDbSettings),
	}
}
func (db *BlobConsumer) TopicLabel() string {
	return db.Topic
}

func (db *BlobConsumer) GetChannel() chan *sarama.ConsumerMessage {
	return db.Channel
}

func (db *BlobConsumer) AttachAddressChannel(chan AddressChannelMessage) {

}

func (db *BlobConsumer) MonitorChannel() {
	receivedCount := 0
	for msg := range db.Channel {
		receivedCount++
		go func(message *sarama.ConsumerMessage) {
			var blob protobuf2.Blob
			err := proto.Unmarshal(message.Value, &blob)
			_, err = db.BlobRepository.Add(*types.Blob{}.MongoFromProtobufType(blob), context.Background())
			// Handle errors better
			if err != nil {
				utils.Logger.Errorf("Error from consumer for blob: %v", err)
				//return
			}

			if receivedCount%consumptionLogModulo == 0 {
				fmt.Printf("Processed %d Blobs\n", receivedCount)
			}
		}(msg)

	}
}

type AddressChannelMessage struct {
	action string
	value  types.Address
}

type AddressChannelConsumer struct {
	AddressChannel    chan AddressChannelMessage
	AddressRepository AddressRepository
}

func NewAddressChannelConsumer(idxConfig types.IdxConfigStruct) *AddressChannelConsumer {

	pg := NewClient(idxConfig)
	return &AddressChannelConsumer{
		AddressChannel:    make(chan AddressChannelMessage),
		AddressRepository: NewAddressRepository(pg),
	}
}

func (db *AddressChannelConsumer) MonitorChannel() {
	for message := range db.AddressChannel {
		go func(msg AddressChannelMessage) {
			switch msg.action {
			case "addressOnly":
				_, err := db.AddressRepository.AddAddressOnly(msg.value)
				// Handle errors better
				if err != nil {
					utils.Logger.Info("%v", err)
					//return
				}
				break
			case "addressDetail":
				_, err := db.AddressRepository.AddAddressDetail(msg.value)
				if err != nil {
					utils.Logger.Info("%v", err)
					//return
				}
				break
			case "addressContract":
				_, err := db.AddressRepository.AddContractAddress(msg.value)
				if err != nil {
					utils.Logger.Info("%v", err)
					//return
				}
				break

			}
		}(message)

	}
}

type AddressConsumer struct {
	Channel           chan *sarama.ConsumerMessage
	Topic             string
	AddressRepository AddressRepository
}

func NewAddressConsumer(idxConfig types.IdxConfigStruct, Topic string) *AddressConsumer {
	pg := NewClient(idxConfig)
	return &AddressConsumer{
		Channel:           make(chan *sarama.ConsumerMessage),
		Topic:             Topic,
		AddressRepository: NewAddressRepository(pg),
	}
}

func (db *AddressConsumer) TopicLabel() string {
	return db.Topic
}

func (db *AddressConsumer) GetChannel() chan *sarama.ConsumerMessage {
	return db.Channel
}

func (db *AddressConsumer) AttachAddressChannel(chan AddressChannelMessage) {

}

func (db *AddressConsumer) MonitorChannel() {
	for message := range db.Channel {
		go func(msg *sarama.ConsumerMessage) {
			var addr protobuf2.AddressDetails
			err := proto.Unmarshal(msg.Value, &addr)

			_, err = db.AddressRepository.AddAddressBalance(*types.AddressBalance{}.GoAddressFromProtobufType(addr))
			// Handle errors better
			if err != nil {
				utils.Logger.Errorf("Error from monitorAddAddressBalanceChannel: %v", err)
			}
		}(message)

	}
}

//type AddressDetailConsumer struct {
//	TypeConsumer
//	AddressRepository AddressRepository
//}
//
//func (db *AddressDetailConsumer) monitorChannel() {
//	for message := range db.Channel {
//		_, err := db.AddressRepository.AddAddressDetail(*address)
//		// Handle errors better
//		if err != nil {
//			utils.Logger.Info("%v", err)
//			//return
//		}
//	}
//}
//
//type ContractAddressConsumer struct {
//	TypeConsumer
//	AddressRepository AddressRepository
//}
//
//func (db *ContractAddressConsumer) monitorChannel() {
//	for message := range db.Channel {
//		_, err := db.AddressRepository.AddContractAddress(*address)
//		// Handle errors better
//		if err != nil {
//			utils.Logger.Errorf("Error from monitorAddContractAddressChannel: %v", err)
//		}
//	}
//}

//type AddressUpdateConsumer struct {
//	Channel           chan *sarama.ConsumerMessage
//	AddressRepository AddressRepository
//}
//
//func (db *AddressUpdateConsumer) monitorChannel() {
//	for message := range db.Channel {
//		err := db.AddressRepository.Update(*address)
//		// Handle errors better
//		if err != nil {
//			utils.Logger.Errorf("Error from monitorAddressUpdateChannel: %v", err)
//		}
//	}
//}
