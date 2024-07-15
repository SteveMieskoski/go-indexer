package types

import "gorm.io/gorm"

type PgSlotSyncTrack struct {
	gorm.Model
	Hash                  string
	Number                int64 `gorm:"uniqueIndex"`
	Retrieved             bool
	Processed             bool
	ReceiptsProcessed     bool
	TransactionsProcessed bool
	ContractsProcessed    bool
}
