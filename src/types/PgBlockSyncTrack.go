package types

import "time"

type PgBlockSyncTrack struct {
	Id                    uint
	CreatedAt             time.Time
	UpdatedAt             time.Time
	Hash                  string
	Number                int64
	Retrieved             bool
	Processed             bool
	ReceiptsProcessed     bool
	TransactionsProcessed bool
	TransactionCount      int64
	ContractsProcessed    bool
}
