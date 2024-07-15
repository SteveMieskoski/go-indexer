package types

import (
	"encoding/json"
	"gorm.io/gorm"
	protobufLocal "src/protobuf"
	base "src/utils"
	"strconv"
)

type DataModel[goType any, protoType any, modelType any] interface {
	MongoFromGoType(data goType) modelType
	ProtobufFromGoType(data goType) protoType
	MongoFromProtobufType(data protoType) *modelType
	ProtobufFromMongoType(data modelType) *protoType
}

type Block struct {
	BaseFeePerGas         base.Gas       `json:"baseFeePerGas"`
	BlobGasUsed           base.Gas       `json:"blobGasUsed"`
	ExcessBlobGas         base.Gas       `json:"excessBlobGas"`
	ParentBeaconBlockRoot base.Hash      `json:"parentBeaconBlockRoot"`
	ReceiptsRoot          base.Hash      `json:"receiptsRoot"`
	StateRoot             base.Hash      `json:"stateRoot"`
	BlockNumber           base.Blknum    `json:"blockNumber"`
	Difficulty            base.Value     `json:"difficulty"`
	GasLimit              base.Gas       `json:"gasLimit"`
	GasUsed               base.Gas       `json:"gasUsed"`
	Hash                  base.Hash      `json:"hash"`
	Miner                 base.Address   `json:"miner"`
	ParentHash            base.Hash      `json:"parentHash"`
	Timestamp             base.Timestamp `json:"timestamp"`
	LogsBloom             string         `json:"logsBloom"`
	Transactions          []Transaction  `json:"transactions"`
	Uncles                []base.Hash    `json:"uncles,omitempty"`
	Withdrawals           []Withdrawal   `json:"withdrawals,omitempty"`
	Number                base.Blknum    `json:"number"`
}

type MongoBlock struct {
	Id                    string             `bson:"_id" json:"_id"`
	BaseFeePerGas         string             `bson:"baseFeePerGas" json:"baseFeePerGas"`
	BlobGasUsed           string             `bson:"blobGasUsed" json:"blobGasUsed"`
	ExcessBlobGas         string             `bson:"excessBlobGas" json:"excessBlobGas"`
	ParentBeaconBlockRoot string             `bson:"parentBeaconBlockRoot" json:"parentBeaconBlockRoot"`
	ReceiptsRoot          string             `bson:"receiptsRoot" json:"receiptsRoot"`
	StateRoot             string             `bson:"stateRoot" json:"stateRoot"`
	BlockNumber           string             `bson:"blockNumber" json:"blockNumber"`
	Difficulty            string             `bson:"difficulty" json:"difficulty"`
	GasLimit              string             `bson:"gasLimit" json:"gasLimit"`
	GasUsed               string             `bson:"gasUsed" json:"gasUsed"`
	Hash                  string             `bson:"hash" json:"hash"`
	LogsBloom             string             `bson:"logsBloom" json:"logsBloom"`
	Miner                 string             `bson:"miner" json:"miner"`
	ParentHash            string             `bson:"parentHash" json:"parentHash"`
	Timestamp             uint64             `bson:"timestamp" json:"timestamp"`
	Transactions          []MongoTransaction `bson:"transactions" json:"transactions"`
	Uncles                []string           `bson:"uncles,omitempty" json:"uncles,omitempty"`
	Withdrawals           []string           `bson:"withdrawals,omitempty" json:"withdrawals,omitempty"`
	Number                string             `bson:"number" json:"number"`
}

type PgBlockNote struct {
	gorm.Model
	Hash      string `gorm:"uniqueIndex"`
	Timestamp uint64
	Number    string
	retrieved bool
}

func (s Block) String() string {
	//bytes, _ := json.Marshal(s)
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}

func (s Block) MongoFromGoType(block Block) MongoBlock {

	var convertedArray []MongoTransaction

	for _, v := range block.Transactions {
		convertedArray = append(convertedArray, v.MongoFromGoType(v))
	}

	return MongoBlock{
		Id:                    block.Hash.String(),
		BaseFeePerGas:         strconv.FormatUint(uint64(block.BaseFeePerGas), 10),
		BlockNumber:           strconv.FormatUint(uint64(block.Number), 10),
		Difficulty:            strconv.FormatUint(uint64(block.Difficulty), 10),
		GasLimit:              strconv.FormatUint(uint64(block.GasLimit), 10),
		GasUsed:               strconv.FormatUint(uint64(block.GasUsed), 10),
		BlobGasUsed:           strconv.FormatUint(uint64(block.BlobGasUsed), 10),
		ExcessBlobGas:         strconv.FormatUint(uint64(block.ExcessBlobGas), 10),
		ParentBeaconBlockRoot: block.ParentBeaconBlockRoot.String(),
		ReceiptsRoot:          block.ReceiptsRoot.String(),
		StateRoot:             block.StateRoot.String(),
		Hash:                  block.Hash.String(),
		LogsBloom:             block.LogsBloom,
		Miner:                 block.Miner.String(),
		ParentHash:            block.ParentHash.String(),
		Timestamp:             uint64(block.Timestamp.Int64()),
		Transactions:          convertedArray,
		Uncles:                nil,
		Withdrawals:           nil,
		Number:                strconv.FormatUint(uint64(block.Number), 10),
	}
}

func (s Block) ProtobufFromGoType(block Block) protobufLocal.Block {
	blockString := s.MongoFromGoType(block)

	var convertedArray []*protobufLocal.Transaction

	for _, v := range blockString.Transactions {
		convertedArray = append(convertedArray, Transaction{}.ProtobufFromMongoType(v))
	}

	return protobufLocal.Block{
		BaseFeePerGas:         blockString.BaseFeePerGas,
		BlockNumber:           blockString.Number,
		Difficulty:            blockString.Difficulty,
		GasLimit:              blockString.GasLimit,
		GasUsed:               blockString.GasUsed,
		BlobGasUsed:           blockString.BlobGasUsed,
		ExcessBlobGas:         blockString.ExcessBlobGas,
		ParentBeaconBlockRoot: blockString.ParentBeaconBlockRoot,
		ReceiptsRoot:          blockString.ReceiptsRoot,
		StateRoot:             blockString.StateRoot,
		Hash:                  blockString.Hash,
		LogsBloom:             blockString.LogsBloom,
		Miner:                 blockString.Miner,
		ParentHash:            blockString.ParentHash,
		Timestamp:             blockString.Timestamp,
		Transactions:          convertedArray,
		Uncles:                blockString.Uncles,
		Withdrawals:           blockString.Withdrawals,
		Number:                blockString.Number,
	}
}

func (s Block) MongoFromProtobufType(block protobufLocal.Block) *MongoBlock {

	var convertedArray []MongoTransaction

	for _, v := range block.Transactions {
		tx := Transaction{}.MongoFromProtobufType(*v)
		convertedArray = append(convertedArray, *tx)
	}

	return &MongoBlock{
		Id:                    block.Hash,
		BaseFeePerGas:         block.BaseFeePerGas,
		BlockNumber:           block.Number,
		Difficulty:            block.Difficulty,
		GasLimit:              block.GasLimit,
		GasUsed:               block.GasUsed,
		BlobGasUsed:           block.BlobGasUsed,
		ExcessBlobGas:         block.ExcessBlobGas,
		ParentBeaconBlockRoot: block.ParentBeaconBlockRoot,
		ReceiptsRoot:          block.ReceiptsRoot,
		StateRoot:             block.StateRoot,
		Hash:                  block.Hash,
		LogsBloom:             block.LogsBloom,
		Miner:                 block.Miner,
		ParentHash:            block.ParentHash,
		Timestamp:             block.Timestamp,
		Transactions:          convertedArray,
		Uncles:                block.Uncles,
		Withdrawals:           block.Withdrawals,
		Number:                block.Number,
	}
}

func (s Block) ProtobufFromMongoType(blockString MongoBlock) *protobufLocal.Block {

	var convertedArray []*protobufLocal.Transaction

	for _, v := range blockString.Transactions {
		convertedArray = append(convertedArray, Transaction{}.ProtobufFromMongoType(v))
	}

	return &protobufLocal.Block{
		BaseFeePerGas:         blockString.BaseFeePerGas,
		BlockNumber:           blockString.Number,
		Difficulty:            blockString.Difficulty,
		GasLimit:              blockString.GasLimit,
		GasUsed:               blockString.GasUsed,
		BlobGasUsed:           blockString.BlobGasUsed,
		ExcessBlobGas:         blockString.ExcessBlobGas,
		ParentBeaconBlockRoot: blockString.ParentBeaconBlockRoot,
		ReceiptsRoot:          blockString.ReceiptsRoot,
		StateRoot:             blockString.StateRoot,
		Hash:                  blockString.Hash,
		LogsBloom:             blockString.LogsBloom,
		Miner:                 blockString.Miner,
		ParentHash:            blockString.ParentHash,
		Timestamp:             blockString.Timestamp,
		Transactions:          convertedArray,
		Uncles:                blockString.Uncles,
		Withdrawals:           blockString.Withdrawals,
		Number:                blockString.Number,
	}
}
