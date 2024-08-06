package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	base "src/utils"
	"strconv"
)

type Receipt struct {
	BlockHash         Hash         `json:"blockHash,omitempty"`
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
	TransactionHash   Hash         `json:"transactionHash"`
	TransactionIndex  base.Txnum   `json:"transactionIndex"`
}

type Receipts []*Receipt

type MongoReceipt struct {
	Id                string     `bson:"_id" json:"_id"`
	BlockHash         string     `bson:"blockHash,omitempty" json:"blockHash,omitempty"`
	BlockNumber       string     `bson:"blockNumber" json:"blockNumber"`
	ContractAddress   string     `bson:"contractAddress,omitempty" json:"contractAddress,omitempty"`
	CumulativeGasUsed string     `bson:"cumulativeGasUsed,omitempty" json:"cumulativeGasUsed,omitempty"`
	EffectiveGasPrice string     `bson:"effectiveGasPrice,omitempty" json:"effectiveGasPrice,omitempty"`
	From              string     `bson:"from,omitempty" json:"from,omitempty"`
	GasUsed           string     `bson:"gasUsed" json:"gasUsed"`
	IsError           bool       `bson:"isError,omitempty" json:"isError,omitempty"`
	Logs              []MongoLog `bson:"logs" json:"logs"`
	Status            string     `bson:"status" json:"status"`
	To                string     `bson:"to,omitempty" json:"to,omitempty"`
	TransactionHash   string     `bson:"transactionHash" json:"transactionHash"`
	TransactionIndex  string     `bson:"transactionIndex" json:"transactionIndex"`
}

func (s Receipt) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s Receipt) MongoFromGoType(receipt Receipt) MongoReceipt {

	var convertedLogs []MongoLog
	for _, log := range receipt.Logs {
		convertedLogs = append(convertedLogs, log.MongoFromGoType(log))
	}

	return MongoReceipt{
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

func (s Receipt) ProtobufFromGoType(receipt Receipt) protobufLocal.Receipt {

	receiptString := s.MongoFromGoType(receipt)

	var convertedLogs []*protobufLocal.Log

	for _, v := range receiptString.Logs {
		convertedLogs = append(convertedLogs, Log{}.ProtobufFromMongoType(v))
	}

	return protobufLocal.Receipt{

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

func (s Receipt) MongoFromProtobufType(receipt *protobufLocal.Receipt) *MongoReceipt {

	var convertedLogs []MongoLog
	for _, log := range receipt.Logs {
		result := Log{}.MongoFromProtobufType(*log)
		convertedLogs = append(convertedLogs, *result)
	}

	return &MongoReceipt{
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

func (s Receipt) ProtobufFromMongoType(receipt MongoReceipt) *protobufLocal.Receipt {

	var convertedLogs []*protobufLocal.Log

	for _, v := range receipt.Logs {
		convertedLogs = append(convertedLogs, Log{}.ProtobufFromMongoType(v))
	}

	return &protobufLocal.Receipt{

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
		To:                receipt.To,
		TransactionHash:   receipt.TransactionHash,
		TransactionIndex:  receipt.TransactionIndex,
	}

}
