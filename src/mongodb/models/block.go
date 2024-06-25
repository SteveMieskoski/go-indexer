package models

import (
	"encoding/json"
	"src/protobuf"
	"src/types"
	"strconv"
)

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

func (s Block) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func BlockFromGoType(block types.Block) Block {

	var convertedArray []Transaction

	for _, v := range block.Transactions {
		convertedArray = append(convertedArray, TransactionFromGoType(v))
	}

	return Block{
		Id:            block.Hash.String(),
		BaseFeePerGas: strconv.FormatUint(uint64(block.BaseFeePerGas), 10),
		BlockNumber:   strconv.FormatUint(uint64(block.Number), 10),
		Difficulty:    strconv.FormatUint(uint64(block.Difficulty), 10),
		GasLimit:      strconv.FormatUint(uint64(block.GasLimit), 10),
		GasUsed:       strconv.FormatUint(uint64(block.GasUsed), 10),
		Hash:          block.Hash.String(),
		Miner:         block.Miner.String(),
		ParentHash:    block.ParentHash.String(),
		Timestamp:     uint64(block.Timestamp.Int64()),
		Transactions:  convertedArray,
		Uncles:        nil,
		Withdrawals:   nil,
		Number:        strconv.FormatUint(uint64(block.Number), 10),
	}
}

func BlockProtobufFromGoType(block types.Block) protobuf.Block {
	blockString := BlockFromGoType(block)

	var convertedArray []*protobuf.Block_Transaction

	for _, v := range blockString.Transactions {
		convertedArray = append(convertedArray, TransactionProtobufFromMongoType(v))
	}

	return protobuf.Block{
		BaseFeePerGas: blockString.BaseFeePerGas,
		BlockNumber:   blockString.Number,
		Difficulty:    blockString.Difficulty,
		GasLimit:      blockString.GasLimit,
		GasUsed:       blockString.GasUsed,
		Hash:          blockString.Hash,
		Miner:         blockString.Miner,
		ParentHash:    blockString.ParentHash,
		Timestamp:     blockString.Timestamp,
		Transactions:  convertedArray,
		Uncles:        blockString.Uncles,
		Withdrawals:   blockString.Withdrawals,
		Number:        blockString.Number,
	}
}

func BlockFromProtobufType(block protobuf.Block) *Block {

	var convertedArray []Transaction

	for _, v := range block.Transactions {
		convertedArray = append(convertedArray, TransactionMongoFromProtobufType(v))
	}

	return &Block{
		Id:            block.Hash,
		BaseFeePerGas: block.BaseFeePerGas,
		BlockNumber:   block.Number,
		Difficulty:    block.Difficulty,
		GasLimit:      block.GasLimit,
		GasUsed:       block.GasUsed,
		Hash:          block.Hash,
		Miner:         block.Miner,
		ParentHash:    block.ParentHash,
		Timestamp:     block.Timestamp,
		Transactions:  convertedArray,
		Uncles:        block.Uncles,
		Withdrawals:   block.Withdrawals,
		Number:        block.Number,
	}
}
