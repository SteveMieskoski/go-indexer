package types

import (
	"encoding/json"
	base "src/utils"
)

type Receipt struct {
	BlockHash         base.Hash    `json:"blockHash,omitempty"`
	BlockNumber       base.Blknum  `json:"blockNumber"`
	ContractAddress   base.Address `json:"contractAddress,omitempty"`
	CumulativeGasUsed base.Gas     `json:"cumulativeGasUsed,omitempty"`
	EffectiveGasPrice base.Gas     `json:"effectiveGasPrice,omitempty"`
	From              base.Address `json:"from,omitempty"`
	GasUsed           base.Gas     `json:"gasUsed"`
	IsError           bool         `json:"isError,omitempty"`
	Logs              []Log        `json:"logs"`
	Status            base.Value   `json:"status"`
	To                base.Address `json:"to,omitempty"`
	TransactionHash   base.Hash    `json:"transactionHash"`
	TransactionIndex  base.Txnum   `json:"transactionIndex"`
}

type Receipts []*Receipt

func (s Receipt) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}
