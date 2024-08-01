package mongodb

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"src/postgres"
	protobuf2 "src/protobuf"
	"src/types"
	"src/utils"
	"strconv"
)

var consumptionLogModulo = 500

type DatabaseCoordinator interface {
	AddBlock() chan<- *types.MongoBlock
	AddReceipt() chan<- *types.MongoReceipt
	AddLog() chan<- *types.MongoLog
	AddBlob() chan<- *types.MongoBlob
	AddTransaction() chan<- *types.MongoTransaction
	AddAddress() chan<- *types.Address
	AddAddressBalance() chan<- *types.Address
	AddAddressDetail() chan<- *types.Address
	AddContractAddress() chan<- *types.Address
	Close()
	ConvertToBlock(block *protobuf2.Block) *types.MongoBlock
	ConvertToReceipt(receipt *protobuf2.Receipt) *types.MongoReceipt
	ConvertToLog(logVal *protobuf2.Log) *types.MongoLog
	ConvertToBlob(blob *protobuf2.Blob) *types.MongoBlob
	ConvertToTransaction(tx *protobuf2.Transaction) *types.MongoTransaction
	ConvertToAddress(addr *protobuf2.AddressDetails) *types.Address
	MessageChannel() chan<- *sarama.ConsumerMessage
	ConsumeMessage() error
}

type databaseCoordinator struct {
	//ctx               context.Context
	AddressChecker        CheckAddress
	BlockRepository       BlockRepository
	ReceiptRepository     ReceiptRepository
	LogRepository         LogRepository
	BlobRepository        BlobRepository
	TransactionRepository TransactionRepository
	AddressRepository     postgres.AddressRepository
	blockChan             chan *types.MongoBlock
	receiptChan           chan *types.MongoReceipt
	logChan               chan *types.MongoLog
	addressChan           chan *types.Address
	addresUpdatesChan     chan *types.Address
	addressDetailChan     chan *types.Address
	addressBalanceChan    chan *types.Address
	addressContractChan   chan *types.Address
	blobChan              chan *types.MongoBlob
	transactionChan       chan *types.MongoTransaction
	messageChannel        chan *sarama.ConsumerMessage
}

func NewDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {
	return newDatabaseCoordinator(settings, idxConfig)
}

func newDatabaseCoordinator(settings DatabaseSetting, idxConfig types.IdxConfigStruct) (DatabaseCoordinator, error) {

	blockDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "blocks",
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

	blobDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "blobs",
	}

	transactionDbSettings := &DatabaseSetting{
		Url:        settings.Url,
		DbName:     settings.DbName,
		Collection: "transactions",
	}

	pg := postgres.NewClient(idxConfig)

	client, err := GetClient(settings, idxConfig)
	if err != nil {
		return nil, err
	}

	client2, err := GetClient(settings, idxConfig)
	if err != nil {
		return nil, err
	}

	client3, err := GetClient(settings, idxConfig)
	if err != nil {
		return nil, err
	}

	client4, err := GetClient(settings, idxConfig)
	if err != nil {
		return nil, err
	}

	AddressRepository := postgres.NewAddressRepository(pg)
	TransactionRepository := NewTransactionRepository(client, transactionDbSettings)
	BlockRepository := NewBlockRepository(client, blockDbSettings)
	ReceiptRepository := NewReceiptRepository(client2, receiptDbSettings)
	LogRepository := NewLogRepository(client3, logDbSettings)
	BlobRepository := NewBlobRepository(client4, blobDbSettings)

	AddressCheck := NewAddressChecker()

	dbc := &databaseCoordinator{
		AddressChecker:        AddressCheck,
		BlockRepository:       BlockRepository,
		ReceiptRepository:     ReceiptRepository,
		LogRepository:         LogRepository,
		BlobRepository:        BlobRepository,
		TransactionRepository: TransactionRepository,
		AddressRepository:     AddressRepository,
		blockChan:             make(chan *types.MongoBlock),
		receiptChan:           make(chan *types.MongoReceipt),
		logChan:               make(chan *types.MongoLog),
		blobChan:              make(chan *types.MongoBlob),
		addressChan:           make(chan *types.Address),
		addresUpdatesChan:     make(chan *types.Address),
		addressBalanceChan:    make(chan *types.Address),
		addressContractChan:   make(chan *types.Address),
		addressDetailChan:     make(chan *types.Address),
		transactionChan:       make(chan *types.MongoTransaction),
		messageChannel:        make(chan *sarama.ConsumerMessage, 5),
	}

	go dbc.monitorBlockChannel()
	go dbc.monitorReceiptChannel()
	go dbc.monitorLogChannel()
	go dbc.monitorAddressChannel()
	go dbc.monitorAddAddressBalanceChannel()
	go dbc.monitorAddressUpdateChannel()
	go dbc.monitorAddAddressDetailChannel()
	go dbc.monitorAddContractAddressChannel()
	go dbc.monitorBlobChannel()
	go dbc.monitorTransactionChannel()

	go func() {
		err := dbc.ConsumeMessage()
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
func (db *databaseCoordinator) ConsumeMessage() error {

	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//defer stop()

	for message := range db.messageChannel {

		if message.Topic == types.BLOCK_TOPIC {
			var block protobuf2.Block
			err := proto.Unmarshal(message.Value, &block)
			if err != nil {
				return err
			}
			db.AddBlock() <- db.ConvertToBlock(&block)
		}

		if message.Topic == types.RECEIPT_TOPIC {
			var receipt protobuf2.Receipt
			err := proto.Unmarshal(message.Value, &receipt)
			if err != nil {
				return err
			}
			db.AddReceipt() <- db.ConvertToReceipt(&receipt)

			for _, logVal := range receipt.Logs {
				db.AddLog() <- db.ConvertToLog(logVal)
			}
		}

		if message.Topic == types.BLOB_TOPIC {
			var blob protobuf2.Blob
			err := proto.Unmarshal(message.Value, &blob)
			if err != nil {
				return err
			}
			db.AddBlob() <- db.ConvertToBlob(&blob)

		}

		if message.Topic == types.TRANSACTION_TOPIC {
			var tx protobuf2.Transaction
			err := proto.Unmarshal(message.Value, &tx)
			if err != nil {
				return err
			}
			db.AddTransaction() <- db.ConvertToTransaction(&tx)
		}

		if message.Topic == types.ADDRESS_TOPIC {

			var addr protobuf2.AddressDetails
			err := proto.Unmarshal(message.Value, &addr)
			if err != nil {
				return err
			}
			db.AddAddressBalance() <- db.ConvertToAddress(&addr)

		}
	}

	return nil
}

func (db *databaseCoordinator) AddBlock() chan<- *types.MongoBlock {
	return db.blockChan
}

func (db *databaseCoordinator) AddReceipt() chan<- *types.MongoReceipt {
	return db.receiptChan
}

func (db *databaseCoordinator) AddLog() chan<- *types.MongoLog {
	return db.logChan
}

func (db *databaseCoordinator) AddBlob() chan<- *types.MongoBlob {
	return db.blobChan
}

func (db *databaseCoordinator) AddTransaction() chan<- *types.MongoTransaction {
	return db.transactionChan
}

func (db *databaseCoordinator) AddAddress() chan<- *types.Address {
	return db.addressChan
}

func (db *databaseCoordinator) UpdateAddress() chan<- *types.Address {
	return db.addresUpdatesChan
}

func (db *databaseCoordinator) AddAddressBalance() chan<- *types.Address {
	return db.addressBalanceChan
}

func (db *databaseCoordinator) AddAddressDetail() chan<- *types.Address {
	return db.addressDetailChan
}

func (db *databaseCoordinator) AddContractAddress() chan<- *types.Address {
	return db.addressContractChan
}

func (db *databaseCoordinator) Close() {
	close(db.blockChan)
	close(db.receiptChan)
	close(db.logChan)
	close(db.addressChan)
	close(db.addresUpdatesChan)
	close(db.addressBalanceChan)
	close(db.addressDetailChan)
	close(db.addressContractChan)
	close(db.blobChan)
	close(db.transactionChan)
	close(db.messageChannel)
}

func (db *databaseCoordinator) monitorBlockChannel() {
	receivedCount := 0
	for blk := range db.blockChan {
		receivedCount++
		_, err := db.BlockRepository.Add(*blk, context.Background())
		if err != nil {
			utils.Logger.Errorf("Error from consumer for block: %v", err)
			//return
		}
		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Blocks\n", receivedCount)
		}
	}
}

func (db *databaseCoordinator) monitorReceiptChannel() {
	receivedCount := 0
	for receipt := range db.receiptChan {
		receivedCount++
		bnum, _ := strconv.ParseInt(receipt.BlockNumber, 16, 64)

		// indicates contract creation so, it wouldn't already exist
		if receipt.ContractAddress != "0x0" {
			db.AddContractAddress() <- &types.Address{Address: receipt.ContractAddress, IsContract: true, LastSeen: bnum}
		}
		_, err := db.ReceiptRepository.Add(*receipt, context.Background())

		if err != nil {
			utils.Logger.Errorf("Error from consumer for receipt: %v", err)
			//return
		}
		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Receipts\n", receivedCount)
		}

	}
}

func (db *databaseCoordinator) monitorLogChannel() {
	receivedCount := 0
	for log := range db.logChan {
		receivedCount++
		_, err := db.LogRepository.Add(*log, context.Background())
		if err != nil {
			utils.Logger.Errorf("Error from consumer for log: %v", err)
			//return
		}

		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Logs\n", receivedCount)
		}
	}
}

func (db *databaseCoordinator) monitorTransactionChannel() {
	receivedCount := 0
	for tx := range db.transactionChan {
		receivedCount++
		bnum, _ := strconv.ParseInt(tx.BlockNumber, 16, 64)

		db.AddAddressDetail() <- &types.Address{Address: tx.From, IsContract: false, Nonce: int64(tx.Nonce), LastSeen: bnum}
		db.AddAddress() <- &types.Address{Address: tx.To, LastSeen: bnum}

		go func(transaction *types.MongoTransaction) {
			_, err := db.TransactionRepository.Add(*transaction, context.Background())

			if err != nil {
				utils.Logger.Errorf("Error from consumer for transaction: %v", err)
				//return
			}
		}(tx)

		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Transactions\n", receivedCount)
		}

	}
}

func (db *databaseCoordinator) monitorBlobChannel() {
	receivedCount := 0
	for blob := range db.blobChan {
		receivedCount++
		_, err := db.BlobRepository.Add(*blob, context.Background())
		// Handle errors better
		if err != nil {
			utils.Logger.Errorf("Error from consumer for blob: %v", err)
			//return
		}

		if receivedCount%consumptionLogModulo == 0 {
			fmt.Printf("Processed %d Blobs\n", receivedCount)
		}
	}
}

func (db *databaseCoordinator) monitorAddressChannel() {
	for address := range db.addressChan {
		go func(addr *types.Address) {
			_, err := db.AddressRepository.AddAddressOnly(*addr)
			// Handle errors better
			if err != nil {
				utils.Logger.Info("%v", err)
				//return
			}
		}(address)

	}
}

func (db *databaseCoordinator) monitorAddAddressBalanceChannel() {
	for address := range db.addressBalanceChan {
		go func(addr *types.Address) {
			_, err := db.AddressRepository.AddAddressBalance(*addr)
			// Handle errors better
			if err != nil {
				utils.Logger.Errorf("Error from monitorAddAddressBalanceChannel: %v", err)
			}
		}(address)

	}
}

func (db *databaseCoordinator) monitorAddAddressDetailChannel() {
	for address := range db.addressDetailChan {
		_, err := db.AddressRepository.AddAddressDetail(*address)
		// Handle errors better
		if err != nil {
			utils.Logger.Info("%v", err)
			//return
		}
	}
}

func (db *databaseCoordinator) monitorAddContractAddressChannel() {
	for address := range db.addressContractChan {
		_, err := db.AddressRepository.AddContractAddress(*address)
		// Handle errors better
		if err != nil {
			utils.Logger.Errorf("Error from monitorAddContractAddressChannel: %v", err)
		}
	}
}

func (db *databaseCoordinator) monitorAddressUpdateChannel() {
	for address := range db.addresUpdatesChan {
		err := db.AddressRepository.Update(*address)
		// Handle errors better
		if err != nil {
			utils.Logger.Errorf("Error from monitorAddressUpdateChannel: %v", err)
		}
	}
}

func (db *databaseCoordinator) ConvertToBlock(block *protobuf2.Block) *types.MongoBlock {
	return types.Block{}.MongoFromProtobufType(block)
}

func (db *databaseCoordinator) ConvertToReceipt(receipt *protobuf2.Receipt) *types.MongoReceipt {
	return types.Receipt{}.MongoFromProtobufType(receipt)
}

func (db *databaseCoordinator) ConvertToLog(logVal *protobuf2.Log) *types.MongoLog {
	return types.Log{}.MongoFromProtobufType(*logVal)
}

func (db *databaseCoordinator) ConvertToBlob(blob *protobuf2.Blob) *types.MongoBlob {
	return types.Blob{}.MongoFromProtobufType(*blob)
}

func (db *databaseCoordinator) ConvertToTransaction(tx *protobuf2.Transaction) *types.MongoTransaction {
	return types.Transaction{}.MongoFromProtobufType(*tx)
}

func (db *databaseCoordinator) ConvertToAddress(addr *protobuf2.AddressDetails) *types.Address {
	return types.AddressBalance{}.GoAddressFromProtobufType(*addr)
}
