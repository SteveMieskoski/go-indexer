package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	base "src/utils"
	"strconv"
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
}

type MongoLog struct {
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

func (s Log) FromGoType(lg Log) MongoLog {

	var convertedTopics []string
	for _, t := range lg.Topics {
		convertedTopics = append(convertedTopics, t.String())
	}

	return MongoLog{
		Address:          lg.Address.String(),
		BlockHash:        lg.BlockHash.String(),
		BlockNumber:      strconv.FormatUint(uint64(lg.BlockNumber), 10),
		Data:             lg.Data,
		LogIndex:         strconv.FormatUint(uint64(lg.LogIndex), 10),
		Timestamp:        uint64(lg.Timestamp),
		Topics:           convertedTopics,
		TransactionHash:  lg.TransactionHash.String(),
		TransactionIndex: strconv.FormatUint(uint64(lg.TransactionIndex), 10),
	}
}

func (s Log) ProtobufFromGoType(lg Log) protobufLocal.Receipt_Log {
	logString := s.FromGoType(lg)

	return protobufLocal.Receipt_Log{
		Address:          logString.Address,
		BlockHash:        logString.BlockHash,
		BlockNumber:      logString.BlockNumber,
		Data:             logString.Data,
		LogIndex:         logString.LogIndex,
		Timestamp:        logString.Timestamp,
		Topics:           logString.Topics,
		TransactionHash:  logString.TransactionHash,
		TransactionIndex: logString.TransactionIndex,
	}
}

func (s Log) MongoFromProtobufType(lg protobufLocal.Receipt_Log) MongoLog {

	return MongoLog{
		Address:          lg.Address,
		BlockHash:        lg.BlockHash,
		BlockNumber:      lg.BlockNumber,
		Data:             lg.Data,
		LogIndex:         lg.LogIndex,
		Timestamp:        lg.Timestamp,
		Topics:           lg.Topics,
		TransactionHash:  lg.TransactionHash,
		TransactionIndex: lg.TransactionIndex,
	}
}

func (s Log) ProtobufFromMongoType(lg MongoLog) *protobufLocal.Receipt_Log {

	return &protobufLocal.Receipt_Log{
		Address:          lg.Address,
		BlockHash:        lg.BlockHash,
		BlockNumber:      lg.BlockNumber,
		Data:             lg.Data,
		LogIndex:         lg.LogIndex,
		Timestamp:        lg.Timestamp,
		Topics:           lg.Topics,
		TransactionHash:  lg.TransactionHash,
		TransactionIndex: lg.TransactionIndex,
	}
}
