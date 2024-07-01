package types

import (
	"encoding/json"
	base "src/utils"
)

type Block struct {
	BaseFeePerGas base.Gas       `json:"baseFeePerGas"`
	BlockNumber   base.Blknum    `json:"blockNumber"`
	Difficulty    base.Value     `json:"difficulty"`
	GasLimit      base.Gas       `json:"gasLimit"`
	GasUsed       base.Gas       `json:"gasUsed"`
	Hash          base.Hash      `json:"hash"`
	Miner         base.Address   `json:"miner"`
	ParentHash    base.Hash      `json:"parentHash"`
	Timestamp     base.Timestamp `json:"timestamp"`
	LogsBloom     string         `json:"logsBloom"`
	Transactions  []Transaction  `json:"transactions"`
	Uncles        []base.Hash    `json:"uncles,omitempty"`
	Withdrawals   []Withdrawal   `json:"withdrawals,omitempty"`
	// EXISTING_CODE
	Number base.Blknum `json:"number"`
	// EXISTING_CODE
}

func (s Block) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
