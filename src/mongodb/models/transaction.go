package models

import (
	"encoding/json"
	"src/protobuf"
	"src/types"
	"strconv"
)

type Transaction struct {
	BlockHash            string `bson:"blockHash" json:"blockHash"`
	BlockNumber          string `bson:"blockNumber" json:"blockNumber"`
	From                 string `bson:"from" json:"from"`
	Gas                  string `bson:"gas" json:"gas"`
	GasPrice             string `bson:"gasPrice" json:"gasPrice"`
	GasUsed              string `bson:"gasUsed" json:"gasUsed"`
	HasToken             bool   `bson:"hasToken" json:"hasToken"`
	Hash                 string `bson:"hash" json:"hash"`
	Input                string `bson:"input" json:"input"`
	IsError              bool   `bson:"isError" json:"isError"`
	MaxFeePerGas         string `bson:"maxFeePerGas" json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `bson:"maxPriorityFeePerGas" json:"maxPriorityFeePerGas"`
	Nonce                uint64 `bson:"nonce" json:"nonce"`
	Timestamp            uint64 `bson:"timestamp" json:"timestamp"`
	To                   string `bson:"to" json:"to"`
	TransactionIndex     string `bson:"transactionIndex" json:"transactionIndex"`
	TransactionType      string `bson:"type" json:"type"`
	Value                string `bson:"value" json:"value"`
	Message              string `bson:"message" json:"message"`
	V                    string `bson:"v" json:"v"`
	R                    string `bson:"r" json:"r"`
	S                    string `bson:"s" json:"s"`
	YParity              string `bson:"yParity" json:"yParity"`
}

func (s Transaction) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func TransactionFromGoType(tx types.Transaction) Transaction {

	return Transaction{
		BlockHash:            tx.BlockHash.String(),
		BlockNumber:          strconv.FormatUint(uint64(tx.BlockNumber), 10),
		From:                 tx.From.String(),
		Gas:                  strconv.FormatUint(uint64(tx.Gas), 10),
		GasPrice:             strconv.FormatUint(uint64(tx.GasPrice), 10),
		GasUsed:              strconv.FormatUint(uint64(tx.GasUsed), 10),
		HasToken:             false,
		Hash:                 tx.Hash.String(),
		Input:                tx.Input,
		IsError:              false,
		MaxFeePerGas:         strconv.FormatUint(uint64(tx.MaxFeePerGas), 10),
		MaxPriorityFeePerGas: strconv.FormatUint(uint64(tx.MaxPriorityFeePerGas), 10),
		Nonce:                uint64(tx.Nonce),
		Timestamp:            uint64(tx.Timestamp.Int64()),
		To:                   tx.To.String(),
		TransactionIndex:     strconv.FormatUint(uint64(tx.TransactionIndex), 10),
		TransactionType:      tx.TransactionType,
		Value:                tx.Value.String(),
		V:                    tx.V,
		R:                    tx.R,
		S:                    tx.S,
		YParity:              tx.YParity,
		Message:              "",
	}
}

func TransactionProtobufFromGoType(tx types.Transaction) protobuf.Block_Transaction {

	txString := TransactionFromGoType(tx)

	return protobuf.Block_Transaction{
		BlockHash:            txString.BlockHash,
		BlockNumber:          txString.BlockNumber,
		From:                 txString.From,
		Gas:                  txString.Gas,
		GasPrice:             txString.GasPrice,
		GasUsed:              txString.GasUsed,
		HasToken:             txString.HasToken,
		Hash:                 txString.Hash,
		Input:                txString.Input,
		IsError:              txString.IsError,
		MaxFeePerGas:         txString.MaxPriorityFeePerGas,
		MaxPriorityFeePerGas: txString.MaxFeePerGas,
		Nonce:                txString.Nonce,
		Timestamp:            txString.Timestamp,
		To:                   txString.TransactionType,
		TransactionIndex:     txString.TransactionIndex,
		TransactionType:      txString.TransactionType,
		Value:                txString.Value,
		Message:              txString.Message,
		V:                    txString.TransactionType,
		R:                    txString.V,
		S:                    txString.TransactionType,
		YParity:              txString.YParity,
	}

}

func TransactionProtobufFromMongoType(tx Transaction) *protobuf.Block_Transaction {

	return &protobuf.Block_Transaction{
		BlockHash:            tx.BlockHash,
		BlockNumber:          tx.BlockNumber,
		From:                 tx.From,
		Gas:                  tx.Gas,
		GasPrice:             tx.GasPrice,
		GasUsed:              tx.GasUsed,
		HasToken:             tx.HasToken,
		Hash:                 tx.Hash,
		Input:                tx.Input,
		IsError:              tx.IsError,
		MaxFeePerGas:         tx.MaxPriorityFeePerGas,
		MaxPriorityFeePerGas: tx.MaxFeePerGas,
		Nonce:                tx.Nonce,
		Timestamp:            tx.Timestamp,
		To:                   tx.TransactionType,
		TransactionIndex:     tx.TransactionIndex,
		TransactionType:      tx.TransactionType,
		Value:                tx.Value,
		Message:              tx.Message,
		V:                    tx.TransactionType,
		R:                    tx.V,
		S:                    tx.TransactionType,
		YParity:              tx.YParity,
	}

}

func TransactionMongoFromProtobufType(tx *protobuf.Block_Transaction) Transaction {

	return Transaction{
		BlockHash:            tx.BlockHash,
		BlockNumber:          tx.BlockNumber,
		From:                 tx.From,
		Gas:                  tx.Gas,
		GasPrice:             tx.GasPrice,
		GasUsed:              tx.GasUsed,
		HasToken:             tx.HasToken,
		Hash:                 tx.Hash,
		Input:                tx.Input,
		IsError:              tx.IsError,
		MaxFeePerGas:         tx.MaxPriorityFeePerGas,
		MaxPriorityFeePerGas: tx.MaxFeePerGas,
		Nonce:                tx.Nonce,
		Timestamp:            tx.Timestamp,
		To:                   tx.TransactionType,
		TransactionIndex:     tx.TransactionIndex,
		TransactionType:      tx.TransactionType,
		Value:                tx.Value,
		Message:              tx.Message,
		V:                    tx.TransactionType,
		R:                    tx.V,
		S:                    tx.TransactionType,
		YParity:              tx.YParity,
	}

}
