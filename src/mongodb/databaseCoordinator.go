package mongodb

import (
	"context"
	"src/mongodb/models"
	protobuf2 "src/protobuf"
	"src/types"
)

type DatabaseCoordinator interface {
	AddBlock() chan<- *types.MongoBlock
	AddReceipt() chan<- *models.Receipt
	AddLog() chan<- *models.Log
	Close()
	ConvertToBlock(block protobuf2.Block) *types.MongoBlock
	ConvertToReceipt(receipt protobuf2.Receipt) *models.Receipt
	ConvertToLog(logVal *protobuf2.Receipt_Log) *models.Log
}

type databaseCoordinator struct {
	//ctx               context.Context
	BlockRepository   BlockRepository
	ReceiptRepository ReceiptRepository
	LogRepository     LogRepository
	blockChan         chan *types.MongoBlock
	receiptChan       chan *models.Receipt
	logChan           chan *models.Log
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

	BlockRepository := NewBlockRepository(client, blockDbSettings)
	ReceiptRepository := NewReceiptRepository(client2, receiptDbSettings)
	LogRepository := NewLogRepository(client3, logDbSettings)

	dbc := &databaseCoordinator{
		BlockRepository:   BlockRepository,
		ReceiptRepository: ReceiptRepository,
		LogRepository:     LogRepository,
		blockChan:         make(chan *types.MongoBlock),
		receiptChan:       make(chan *models.Receipt),
		logChan:           make(chan *models.Log),
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

	return dbc, nil
}

func (db *databaseCoordinator) AddBlock() chan<- *types.MongoBlock {
	return db.blockChan
}

func (db *databaseCoordinator) AddReceipt() chan<- *models.Receipt {
	return db.receiptChan
}

func (db *databaseCoordinator) AddLog() chan<- *models.Log {
	return db.logChan
}

func (db *databaseCoordinator) Close() {
	close(db.blockChan)
	close(db.receiptChan)
	close(db.logChan)
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

func (db *databaseCoordinator) ConvertToBlock(block protobuf2.Block) *types.MongoBlock {
	return types.Block{}.MongoFromProtobufType(block)
}

func (db *databaseCoordinator) ConvertToReceipt(receipt protobuf2.Receipt) *models.Receipt {
	return models.ReceiptFromProtobufType(receipt)
}

func (db *databaseCoordinator) ConvertToLog(logVal *protobuf2.Receipt_Log) *models.Log {
	return models.LogMongoDirectFromProtobufType(logVal)
}
