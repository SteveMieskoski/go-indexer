package types

import (
	"gorm.io/gorm"
)

type Address struct {
	gorm.Model
	address    string
	nonce      string
	isContract bool
	balance    int64
}
