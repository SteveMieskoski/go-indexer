package models

import (
	"encoding/json"
)

type Receipt struct {
	BlockHash         string `bson:"blockHash,omitempty" json:"blockHash,omitempty"`
	BlockNumber       string `bson:"blockNumber" json:"blockNumber"`
	ContractAddress   string `bson:"contractAddress,omitempty" json:"contractAddress,omitempty"`
	CumulativeGasUsed string `bson:"cumulativeGasUsed,omitempty" json:"cumulativeGasUsed,omitempty"`
	EffectiveGasPrice string `bson:"effectiveGasPrice,omitempty" json:"effectiveGasPrice,omitempty"`
	From              string `bson:"from,omitempty" json:"from,omitempty"`
	GasUsed           string `bson:"gasUsed" json:"gasUsed"`
	IsError           bool   `bson:"isError,omitempty" json:"isError,omitempty"`
	Logs              []Log  `bson:"logs" json:"logs"`
	Status            string `bson:"status" json:"status"`
	To                string `bson:"to,omitempty" json:"to,omitempty"`
	TransactionHash   string `bson:"transactionHash" json:"transactionHash"`
	TransactionIndex  string `bson:"transactionIndex" json:"transactionIndex"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s Receipt) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
