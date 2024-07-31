package types

import (
	"encoding/json"
	"time"
)

type PgBlockSyncTrack struct {
	Id                    uint      `json:"id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Hash                  string    `json:"hash,omitempty"`
	Number                int64     `json:"number,omitempty"`
	Retrieved             bool      `json:"retrieved,omitempty"`
	Processed             bool      `json:"processed,omitempty"`
	ReceiptsProcessed     bool      `json:"receipts_processed,omitempty"`
	TransactionsProcessed bool      `json:"transactions_processed,omitempty"`
	TransactionCount      int64     `json:"transaction_count,omitempty"`
	ContractsProcessed    bool      `json:"contracts_processed,omitempty"`
}

func (s PgBlockSyncTrack) String() string {
	//bytes, _ := json.Marshal(s)
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}
