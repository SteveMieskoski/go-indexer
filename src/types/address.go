package types

import (
	"time"
)

type RawAddress struct {
	address    string
	nonce      string
	isContract bool
	balance    int64
}

type Address struct {
	Id         uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Address    string
	Nonce      string
	IsContract bool
	Balance    int64
}

//type MethodSignature struct {
//	gorm.Model
//	signature       string
//	contractAddress string
//}
