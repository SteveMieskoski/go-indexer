package models

import (
	"encoding/json"
)

type Transaction struct {
	BlockHash            string   `bson:"blockHash" json:"blockHash"`
	BlockNumber          string   `bson:"blockNumber" json:"blockNumber"`
	From                 string   `bson:"from" json:"from"`
	Gas                  string   `bson:"gas" json:"gas"`
	GasPrice             string   `bson:"gasPrice" json:"gasPrice"`
	GasUsed              string   `bson:"gasUsed" json:"gasUsed"`
	HasToken             bool     `bson:"hasToken" json:"hasToken"`
	Hash                 string   `bson:"hash" json:"hash"`
	Input                string   `bson:"input" json:"input"`
	IsError              bool     `bson:"isError" json:"isError"`
	MaxFeePerGas         string   `bson:"maxFeePerGas" json:"maxFeePerGas"`
	MaxPriorityFeePerGas string   `bson:"maxPriorityFeePerGas" json:"maxPriorityFeePerGas"`
	Nonce                uint64   `bson:"nonce" json:"nonce"`
	Receipt              *Receipt `bson:"receipt" json:"receipt"`
	Timestamp            uint64   `bson:"timestamp" json:"timestamp"`
	To                   string   `bson:"to" json:"to"`
	TransactionIndex     string   `bson:"transactionIndex" json:"transactionIndex"`
	TransactionType      string   `bson:"type" json:"type"`
	Value                string   `bson:"value" json:"value"`
	Message              string   `bson:"message" json:"message"`
}

func (s Transaction) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
