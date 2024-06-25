package types

import (
	"encoding/json"
	base "src/utils"
)

type Transaction struct {
	BlockHash            base.Hash      `json:"blockHash"`
	BlockNumber          base.Blknum    `json:"blockNumber"`
	From                 base.Address   `json:"from"`
	Gas                  base.Gas       `json:"gas"`
	GasPrice             base.Gas       `json:"gasPrice"`
	GasUsed              base.Gas       `json:"gasUsed"`
	HasToken             bool           `json:"hasToken"`
	Hash                 base.Hash      `json:"hash"`
	Input                string         `json:"input"`
	IsError              bool           `json:"isError"`
	MaxFeePerGas         base.Gas       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas base.Gas       `json:"maxPriorityFeePerGas"`
	Nonce                base.Value     `json:"nonce"`
	Receipt              *Receipt       `json:"receipt"`
	Timestamp            base.Timestamp `json:"timestamp"`
	To                   base.Address   `json:"to"`
	TransactionIndex     base.Txnum     `json:"transactionIndex"`
	TransactionType      string         `json:"type"`
	Value                base.Wei       `json:"value"`
	V                    string         `json:"v"`
	R                    string         `json:"r"`
	S                    string         `json:"s"`
	YParity              string         `json:"yParity"`
	Message              string         `json:"-"`
	// EXISTING_CODE
}

func (s Transaction) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
