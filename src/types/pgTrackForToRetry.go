package types

import (
	"time"
)

type PgTrackForToRetry struct {
	Id        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DataType  string
	BlockId   string
	RecordId  string
}
