package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"src/types"
	"sync"
)

//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"

type AClient struct {
	locked sync.Mutex
	client *gorm.DB
	index  int
}

type PostgresDB struct {
	client  *gorm.DB
	clients []AClient
}

func NewClient() *PostgresDB {
	dsn := os.Getenv("RAW_GO_POSTGRES_STRING")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&types.Address{}, &types.PgBlockSyncTrack{}, types.PgSlotSyncTrack{}, types.PgTrackForToRetry{})
	if err != nil {
		return nil
	}

	return &PostgresDB{client: db}
}

func (p *PostgresDB) getNewClient() *gorm.DB {
	dsn := os.Getenv("RAW_GO_POSTGRES_STRING")
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db
}

//func (p *PostgresDB) getClient() *gorm.DB {
//
//	return db
//}
//
//func (p *AClient) borrow() (client AClient) {
//	p.producersLock.Lock()
//	defer p.producersLock.Unlock()
//
//	if len(p.producers) == 0 {
//		for {
//			producer = p.ProducerProvider()
//			if producer != nil {
//				return
//			}
//		}
//	}
//
//	index := len(p.producers) - 1
//	producer = p.producers[index]
//	p.producers = p.producers[:index]
//	return
//}
//
//func (p *AClient) release(producer AClient) {
//	p.producersLock.Lock()
//	defer p.producersLock.Unlock()
//
//	// If released producer is erroneous close it and don't return it to the producer pool.
//	if producer.TxnStatus()&sarama.ProducerTxnFlagInError != 0 {
//		// Try to close it
//		_ = producer.Close()
//		return
//	}
//	p.producers = append(p.producers, producer)
//}
