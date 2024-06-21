package models

import (
	"encoding/json"
)

type Log struct {
	Address          string   `bson:"address" json:"address"`
	BlockHash        string   `bson:"blockHash" json:"blockHash"`
	BlockNumber      string   `bson:"blockNumber" json:"blockNumber"`
	Data             string   `bson:"data,omitempty" json:"data,omitempty"`
	LogIndex         string   `bson:"logIndex" json:"logIndex"`
	Timestamp        uint64   `bson:"timestamp,omitempty" json:"timestamp,omitempty"`
	Topics           []string `bson:"topics,omitempty" json:"topics,omitempty"`
	TransactionHash  string   `bson:"transactionHash" json:"transactionHash"`
	TransactionIndex string   `bson:"transactionIndex" json:"transactionIndex"`
}

func (s Log) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
