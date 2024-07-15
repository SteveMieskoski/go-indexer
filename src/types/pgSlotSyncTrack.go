package types

import "gorm.io/gorm"

type PgSlotSyncTrack struct {
	gorm.Model
	Hash           string
	Slot           int64 `gorm:"uniqueIndex"`
	Retrieved      bool
	Processed      bool
	BlobsProcessed bool
	BlobCount      int64
}
