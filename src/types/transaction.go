package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	base "src/utils"
	"strconv"
)

type Transaction struct {
	BlockHash            Hash           `json:"blockHash"`
	BlockNumber          base.Blknum    `json:"blockNumber"`
	From                 base.Address   `json:"from"`
	Gas                  base.Gas       `json:"gas"`
	GasPrice             base.Gas       `json:"gasPrice"`
	GasUsed              base.Gas       `json:"gasUsed"`
	HasToken             bool           `json:"hasToken"`
	Hash                 Hash           `json:"hash"`
	Input                string         `json:"input"`
	IsError              bool           `json:"isError"`
	MaxFeePerGas         base.Gas       `json:"maxFeePerGas"`
	MaxPriorityFeePerGas base.Gas       `json:"maxPriorityFeePerGas"`
	Nonce                base.Value     `json:"nonce"`
	Receipt              *Receipt       `json:"receipt"`
	Timestamp            base.Timestamp `json:"timestamp"`
	To                   base.Address   `json:"to"`
	TransactionIndex     base.Txnum     `json:"transactionIndex"`
	TransactionType      string         `json:"type"`
	Value                base.Wei       `json:"value"`
	V                    string         `json:"v"`
	R                    string         `json:"r"`
	S                    string         `json:"s"`
	YParity              string         `json:"yParity"`
	Message              string         `json:"-"`
	AccessList           []AccessList   `json:"AccessList"`
}

type MongoTransaction struct {
	Id                   string            `bson:"_id" json:"_id"`
	BlockHash            string            `bson:"blockHash" json:"blockHash"`
	BlockNumber          string            `bson:"blockNumber" json:"blockNumber"`
	From                 string            `bson:"from" json:"from"`
	Gas                  string            `bson:"gas" json:"gas"`
	GasPrice             string            `bson:"gasPrice" json:"gasPrice"`
	GasUsed              string            `bson:"gasUsed" json:"gasUsed"`
	HasToken             bool              `bson:"hasToken" json:"hasToken"`
	Hash                 string            `bson:"hash" json:"hash"`
	Input                string            `bson:"input" json:"input"`
	IsError              bool              `bson:"isError" json:"isError"`
	MaxFeePerGas         string            `bson:"maxFeePerGas" json:"maxFeePerGas"`
	MaxPriorityFeePerGas string            `bson:"maxPriorityFeePerGas" json:"maxPriorityFeePerGas"`
	Nonce                uint64            `bson:"nonce" json:"nonce"`
	Timestamp            uint64            `bson:"timestamp" json:"timestamp"`
	To                   string            `bson:"to" json:"to"`
	TransactionIndex     string            `bson:"transactionIndex" json:"transactionIndex"`
	TransactionType      string            `bson:"type" json:"type"`
	Value                string            `bson:"value" json:"value"`
	Message              string            `bson:"message" json:"message"`
	V                    string            `bson:"v" json:"v"`
	R                    string            `bson:"r" json:"r"`
	S                    string            `bson:"s" json:"s"`
	YParity              string            `bson:"yParity" json:"yParity"`
	AccessList           []MongoAccessList `json:"AccessList"`
}

func (s MongoTransaction) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s Transaction) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s Transaction) MongoFromGoType(tx Transaction) MongoTransaction {

	var convertedAccessList []MongoAccessList
	for _, t := range tx.AccessList {
		convertedAccessList = append(convertedAccessList, AccessList{}.MongoFromGoType(t))
	}

	return MongoTransaction{
		Id:                   tx.Hash.String(),
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
		AccessList:           convertedAccessList,
	}
}

func (s Transaction) ProtobufFromGoType(tx Transaction) protobufLocal.Transaction {
	txString := s.MongoFromGoType(tx)

	var convertedAccessList []*protobufLocal.AccessList

	for _, v := range txString.AccessList {
		convertedAccessList = append(convertedAccessList, AccessList{}.ProtobufFromMongoType(v))
	}

	return protobufLocal.Transaction{
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
		To:                   txString.To,
		TransactionIndex:     txString.TransactionIndex,
		TransactionType:      txString.TransactionType,
		Value:                txString.Value,
		Message:              txString.Message,
		V:                    txString.V,
		R:                    txString.R,
		S:                    txString.S,
		YParity:              txString.YParity,
		AccessLists:          convertedAccessList,
	}
}

func (s Transaction) MongoFromProtobufType(tx protobufLocal.Transaction) *MongoTransaction {

	var convertedAccessList []MongoAccessList
	for _, log := range tx.AccessLists {
		result := AccessList{}.MongoFromProtobufType(*log)
		convertedAccessList = append(convertedAccessList, *result)
	}

	return &MongoTransaction{
		Id:                   tx.Hash,
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
		To:                   tx.To,
		TransactionIndex:     tx.TransactionIndex,
		TransactionType:      tx.TransactionType,
		Value:                tx.Value,
		Message:              tx.Message,
		V:                    tx.V,
		R:                    tx.R,
		S:                    tx.S,
		YParity:              tx.YParity,
		AccessList:           convertedAccessList,
	}
}

func (s Transaction) ProtobufFromMongoType(txString MongoTransaction) *protobufLocal.Transaction {

	var convertedAccessList []*protobufLocal.AccessList

	for _, v := range txString.AccessList {
		convertedAccessList = append(convertedAccessList, AccessList{}.ProtobufFromMongoType(v))
	}

	return &protobufLocal.Transaction{
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
		To:                   txString.To,
		TransactionIndex:     txString.TransactionIndex,
		TransactionType:      txString.TransactionType,
		Value:                txString.Value,
		Message:              txString.Message,
		V:                    txString.V,
		R:                    txString.R,
		S:                    txString.S,
		YParity:              txString.YParity,
		AccessLists:          convertedAccessList,
	}
}
