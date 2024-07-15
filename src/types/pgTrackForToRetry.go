package types

import "gorm.io/gorm"

type PgTrackForToRetry struct {
	gorm.Model
	DataType string
	BlockId  string
	RecordId string
}
