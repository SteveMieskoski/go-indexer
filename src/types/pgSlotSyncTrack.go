package types

import (
	"encoding/json"
	"time"
)

type PgSlotSyncTrack struct {
	Id             uint      `json:"id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Hash           string    `json:"hash,omitempty"`
	Slot           int64     `json:"slot,omitempty"`
	Retrieved      bool      `json:"retrieved,omitempty"`
	Processed      bool      `json:"processed,omitempty"`
	BlobsProcessed bool      `json:"blobs_processed,omitempty"`
	BlobCount      int64     `json:"blob_count,omitempty"`
}

func (s PgSlotSyncTrack) String() string {
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}
