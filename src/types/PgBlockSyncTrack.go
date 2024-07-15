package types

import "gorm.io/gorm"

type PgBlockSyncTrack struct {
	gorm.Model
	Hash                  string
	Number                int64 `gorm:"uniqueIndex"`
	Retrieved             bool
	Processed             bool
	ReceiptsProcessed     bool
	TransactionsProcessed bool
	TransactionCount      int64
	ContractsProcessed    bool
}
