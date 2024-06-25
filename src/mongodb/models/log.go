package models

import (
	"encoding/json"
	"src/protobuf"
	"src/types"
	"strconv"
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

func LogFromGoType(lg types.Log) Log {

	var convertedTopics []string
	for _, t := range lg.Topics {
		convertedTopics = append(convertedTopics, t.String())
	}

	return Log{
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

func LogProtobufFromGoType(lg types.Log) protobuf.Receipt_Log {
	logString := LogFromGoType(lg)

	return protobuf.Receipt_Log{
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

func LogProtobufFromMongoType(lg Log) *protobuf.Receipt_Log {

	return &protobuf.Receipt_Log{
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

func LogMongoFromProtobufType(lg *protobuf.Receipt_Log) Log {

	return Log{
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

func LogMongoDirectFromProtobufType(lg *protobuf.Receipt_Log) *Log {

	return &Log{
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
