package types

import (
	"gorm.io/gorm"
)

type RawAddress struct {
	address    string
	nonce      string
	isContract bool
	balance    int64
}

type Address struct {
	gorm.Model
	Address    string `gorm:"uniqueIndex"`
	Nonce      string
	IsContract bool
	Balance    int64
}

type MethodSignature struct {
	gorm.Model
	signature       string
	contractAddress string
}
