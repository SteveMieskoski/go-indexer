package models

type Block struct {
	Id            string        `bson:"_id" json:"_id"`
	BaseFeePerGas string        `bson:"baseFeePerGas" json:"baseFeePerGas"`
	BlockNumber   string        `bson:"blockNumber" json:"blockNumber"`
	Difficulty    string        `bson:"difficulty" json:"difficulty"`
	GasLimit      string        `bson:"gasLimit" json:"gasLimit"`
	GasUsed       string        `bson:"gasUsed" json:"gasUsed"`
	Hash          string        `bson:"hash" json:"hash"`
	Miner         string        `bson:"miner" json:"miner"`
	ParentHash    string        `bson:"parentHash" json:"parentHash"`
	Timestamp     uint64        `bson:"timestamp" json:"timestamp"`
	Transactions  []Transaction `bson:"transactions" json:"transactions"`
	Uncles        []string      `bson:"uncles,omitempty" json:"uncles,omitempty"`
	Withdrawals   []string      `bson:"withdrawals,omitempty" json:"withdrawals,omitempty"`
	Number        string        `bson:"number" json:"number"`
}
