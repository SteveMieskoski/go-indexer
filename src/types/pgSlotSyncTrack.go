package types

import "time"

type PgSlotSyncTrack struct {
	Id             uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Hash           string
	Slot           int64
	Retrieved      bool
	Processed      bool
	BlobsProcessed bool
	BlobCount      int64
}
