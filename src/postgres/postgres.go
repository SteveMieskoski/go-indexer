package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"src/types"
)

//dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"

type PostgresDB struct {
	client *gorm.DB
}

func NewClient() *PostgresDB {
	dsn := os.Getenv("RAW_GO_POSTGRES_STRING")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&types.Address{})
	if err != nil {
		return nil
	}

	return &PostgresDB{client: db}
}
