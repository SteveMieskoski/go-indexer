package types

import (
	"encoding/json"
	base "src/utils"
)

type Log struct {
	Address          base.Address   `json:"address"`
	BlockHash        base.Hash      `json:"blockHash"`
	BlockNumber      base.Blknum    `json:"blockNumber"`
	Data             string         `json:"data,omitempty"`
	LogIndex         base.Lognum    `json:"logIndex"`
	Timestamp        base.Timestamp `json:"timestamp,omitempty"`
	Topics           []base.Hash    `json:"topics,omitempty"`
	TransactionHash  base.Hash      `json:"transactionHash"`
	TransactionIndex base.Txnum     `json:"transactionIndex"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s Log) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
