package models

import (
	"encoding/json"
	"src/protobuf"
	"src/types"
	"strconv"
)

type Receipt struct {
	Id                string `bson:"_id" json:"_id"`
	BlockHash         string `bson:"blockHash,omitempty" json:"blockHash,omitempty"`
	BlockNumber       string `bson:"blockNumber" json:"blockNumber"`
	ContractAddress   string `bson:"contractAddress,omitempty" json:"contractAddress,omitempty"`
	CumulativeGasUsed string `bson:"cumulativeGasUsed,omitempty" json:"cumulativeGasUsed,omitempty"`
	EffectiveGasPrice string `bson:"effectiveGasPrice,omitempty" json:"effectiveGasPrice,omitempty"`
	From              string `bson:"from,omitempty" json:"from,omitempty"`
	GasUsed           string `bson:"gasUsed" json:"gasUsed"`
	IsError           bool   `bson:"isError,omitempty" json:"isError,omitempty"`
	Logs              []Log  `bson:"logs" json:"logs"`
	Status            string `bson:"status" json:"status"`
	To                string `bson:"to,omitempty" json:"to,omitempty"`
	TransactionHash   string `bson:"transactionHash" json:"transactionHash"`
	TransactionIndex  string `bson:"transactionIndex" json:"transactionIndex"`
}

func (s Receipt) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func ReceiptFromGoType(receipt types.Receipt) Receipt {

	var convertedLogs []Log
	for _, log := range receipt.Logs {
		convertedLogs = append(convertedLogs, LogFromGoType(log))
	}

	return Receipt{
		Id:                receipt.TransactionHash.String(),
		BlockHash:         receipt.BlockHash.String(),
		BlockNumber:       strconv.FormatUint(uint64(receipt.BlockNumber), 10),
		ContractAddress:   receipt.ContractAddress.String(),
		CumulativeGasUsed: strconv.FormatUint(uint64(receipt.CumulativeGasUsed), 10),
		EffectiveGasPrice: strconv.FormatUint(uint64(receipt.EffectiveGasPrice), 10),
		From:              receipt.From.String(),
		GasUsed:           strconv.FormatUint(uint64(receipt.GasUsed), 10),
		IsError:           receipt.IsError,
		Logs:              convertedLogs,
		Status:            strconv.FormatUint(uint64(receipt.Status), 10),
		To:                receipt.To.String(),
		TransactionHash:   receipt.TransactionHash.String(),
		TransactionIndex:  strconv.FormatUint(uint64(receipt.TransactionIndex), 10),
	}

}

func ReceiptProtobufFromGoType(receipt types.Receipt) protobuf.Receipt {
	receiptString := ReceiptFromGoType(receipt)

	var convertedLogs []*protobuf.Receipt_Log

	for _, v := range receiptString.Logs {
		convertedLogs = append(convertedLogs, LogProtobufFromMongoType(v))
	}

	return protobuf.Receipt{

		BlockHash:         receipt.BlockHash.String(),
		BlockNumber:       strconv.FormatUint(uint64(receipt.BlockNumber), 10),
		ContractAddress:   receipt.ContractAddress.String(),
		CumulativeGasUsed: strconv.FormatUint(uint64(receipt.CumulativeGasUsed), 10),
		EffectiveGasPrice: strconv.FormatUint(uint64(receipt.EffectiveGasPrice), 10),
		From:              receipt.From.String(),
		GasUsed:           strconv.FormatUint(uint64(receipt.GasUsed), 10),
		IsError:           receipt.IsError,
		Logs:              convertedLogs,
		Status:            strconv.FormatUint(uint64(receipt.Status), 10),
		To:                receipt.To.String(),
		TransactionHash:   receipt.TransactionHash.String(),
		TransactionIndex:  strconv.FormatUint(uint64(receipt.TransactionIndex), 10),
	}
}

func ReceiptFromProtobufType(receipt protobuf.Receipt) *Receipt {

	var convertedLogs []Log
	for _, log := range receipt.Logs {
		convertedLogs = append(convertedLogs, LogMongoFromProtobufType(log))
	}

	return &Receipt{
		Id:                receipt.TransactionHash,
		BlockHash:         receipt.BlockHash,
		BlockNumber:       receipt.BlockNumber,
		ContractAddress:   receipt.ContractAddress,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		EffectiveGasPrice: receipt.EffectiveGasPrice,
		From:              receipt.From,
		GasUsed:           receipt.GasUsed,
		IsError:           receipt.IsError,
		Logs:              convertedLogs,
		Status:            receipt.Status,
		To:                receipt.From,
		TransactionHash:   receipt.TransactionHash,
		TransactionIndex:  receipt.TransactionIndex,
	}
}
