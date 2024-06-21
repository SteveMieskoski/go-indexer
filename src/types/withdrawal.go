package types

import (
	"encoding/json"
	base "src/utils"
)

type Withdrawal struct {
	Address        base.Address   `json:"address"`
	Amount         base.Wei       `json:"amount"`
	BlockNumber    base.Blknum    `json:"blockNumber"`
	Index          base.Value     `json:"index"`
	Timestamp      base.Timestamp `json:"timestamp"`
	ValidatorIndex base.Value     `json:"validatorIndex"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s Withdrawal) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
