package mongodb

import (
	"context"
	"src/postgres"
	protobuf2 "src/protobuf"
	"src/types"
)

type DatabaseCoordinator interface {
	AddBlock() chan<- *types.MongoBlock
	AddReceipt() chan<- *types.MongoReceipt
	AddLog() chan<- *types.MongoLog
	AddBlob() chan<- *types.MongoBlob
	Close()
	ConvertToBlock(block protobuf2.Block) *types.MongoBlock
	ConvertToReceipt(receipt protobuf2.Receipt) *types.MongoReceipt
	ConvertToLog(logVal *protobuf2.Receipt_Log) *types.MongoLog
	ConvertToBlob(blob *protobuf2.Blob) *types.MongoBlob
}

type databaseCoordinator struct {
	//ctx               context.Context
	BlockRepository   BlockRepository
	ReceiptRepository ReceiptRepository
	LogRepository     LogRepository
	BlobRepository    BlobRepository
	AddressRepository postgres.AddressRepository
	blockChan         chan *types.MongoBlock
	receiptChan       chan *types.MongoReceipt
	logChan           chan *types.MongoLog
	addressChan       chan *types.Address
	blobChan          chan *types.MongoBlob
}

func NewDatabaseCoordinator(settings DatabaseSetting) (DatabaseCoordinator, error) {
	return newDatabaseCoordinator(settings)
}

func newDatabaseCoordinator(settings DatabaseSetting) (DatabaseCoordinator, error) {

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

	pg := postgres.NewClient()

	client, err := GetClient(settings)
	if err != nil {
		return nil, err
	}
	client2, err := GetClient(settings)
	if err != nil {
		return nil, err
	}
	client3, err := GetClient(settings)
	if err != nil {
		return nil, err
	}
	client4, err := GetClient(settings)
	if err != nil {
		return nil, err
	}

	AddressRepository := postgres.NewAddressRepository(pg)
	BlockRepository := NewBlockRepository(client, blockDbSettings)
	ReceiptRepository := NewReceiptRepository(client2, receiptDbSettings)
	LogRepository := NewLogRepository(client3, logDbSettings)
	BlobRepository := NewBlobRepository(client4, blobDbSettings)

	dbc := &databaseCoordinator{
		BlockRepository:   BlockRepository,
		ReceiptRepository: ReceiptRepository,
		LogRepository:     LogRepository,
		BlobRepository:    BlobRepository,
		AddressRepository: AddressRepository,
		blockChan:         make(chan *types.MongoBlock),
		receiptChan:       make(chan *types.MongoReceipt),
		logChan:           make(chan *types.MongoLog),
		blobChan:          make(chan *types.MongoBlob),
		addressChan:       make(chan *types.Address),
	}

	go func() {
		dbc.monitorBlockChannel()
	}()

	go func() {
		dbc.monitorReceiptChannel()
	}()

	go func() {
		dbc.monitorLogChannel()
	}()

	go func() {
		dbc.monitorAddressChannel()
	}()

	go func() {
		dbc.monitorBlobChannel()
	}()

	return dbc, nil
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

func (db *databaseCoordinator) AddAddress() chan<- *types.Address {
	return db.addressChan
}

func (db *databaseCoordinator) AddBlob() chan<- *types.MongoBlob {
	return db.blobChan
}

func (db *databaseCoordinator) Close() {
	close(db.blockChan)
	close(db.receiptChan)
	close(db.logChan)
	close(db.addressChan)
	close(db.blobChan)
}

func (db *databaseCoordinator) monitorBlockChannel() {

	for blk := range db.blockChan {
		_, err := db.BlockRepository.Add(*blk, context.Background())
		if err != nil {
			return
		}
	}
}

func (db *databaseCoordinator) monitorReceiptChannel() {
	for receipt := range db.receiptChan {
		db.AddAddress() <- &types.Address{Address: receipt.From}
		db.AddAddress() <- &types.Address{Address: receipt.To}
		if receipt.ContractAddress != "0x0" {
			db.AddAddress() <- &types.Address{Address: receipt.ContractAddress, IsContract: true}
		}
		_, err := db.ReceiptRepository.Add(*receipt, context.Background())

		if err != nil {
			return
		}
	}
}

func (db *databaseCoordinator) monitorLogChannel() {
	for log := range db.logChan {
		_, err := db.LogRepository.Add(*log, context.Background())
		if err != nil {
			return
		}
	}
}

func (db *databaseCoordinator) monitorBlobChannel() {
	for blob := range db.blobChan {
		_, err := db.BlobRepository.Add(*blob, context.Background())
		if err != nil {
			return
		}
	}
}

func (db *databaseCoordinator) monitorAddressChannel() {
	for address := range db.addressChan {
		println(address)
		println("Received address")
		_, err := db.AddressRepository.Add(*address, context.Background())
		if err != nil {
			return
		}
	}
}

func (db *databaseCoordinator) ConvertToBlock(block protobuf2.Block) *types.MongoBlock {
	return types.Block{}.MongoFromProtobufType(block)
}

func (db *databaseCoordinator) ConvertToReceipt(receipt protobuf2.Receipt) *types.MongoReceipt {
	return types.Receipt{}.MongoFromProtobufType(receipt)
}

func (db *databaseCoordinator) ConvertToLog(logVal *protobuf2.Receipt_Log) *types.MongoLog {
	return types.Log{}.MongoFromProtobufType(*logVal)
}

func (db *databaseCoordinator) ConvertToBlob(blob *protobuf2.Blob) *types.MongoBlob {
	return types.Blob{}.MongoFromProtobufType(*blob)
}
