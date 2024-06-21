package models

import (
	"encoding/json"
)

type Withdrawal struct {
	Address        string `json:"address" json:"address"`
	Amount         string `json:"amount" json:"amount"`
	BlockNumber    string `json:"blockNumber" json:"blockNumber"`
	Index          uint64 `json:"index" json:"index"`
	Timestamp      uint64 `json:"timestamp" json:"timestamp"`
	ValidatorIndex uint64 `json:"validatorIndex" json:"validatorIndex"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s Withdrawal) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
