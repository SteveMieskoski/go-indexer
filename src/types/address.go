package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
	"time"
)

type RawAddress struct {
	address    string
	nonce      string
	isContract bool
	balance    int64
}

type Address struct {
	Id         uint      `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Address    string    `json:"address"`
	Nonce      int64     `json:"nonce"`
	IsContract bool      `json:"isContract"`
	Balance    int64     `json:"balance"`
	LastSeen   int64     `json:"lastSeen"`
}

func (s Address) String() string {
	//bytes, _ := json.Marshal(s)
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}

//type MethodSignature struct {
//	gorm.Model
//	signature       string
//	contractAddress string
//}

type AddressBalance struct {
	Address  string `json:"address"`
	Balance  int64  `json:"balance"`
	LastSeen int64  `json:"lastSeen"`
}

func (s AddressBalance) String() string {
	//bytes, _ := json.Marshal(s)
	bytes, _ := json.MarshalIndent(s, "", "   ")
	return string(bytes)
}

// AddressDetails
func (s AddressBalance) ProtobufFromGoType(addr AddressBalance) protobufLocal.AddressDetails {

	return protobufLocal.AddressDetails{
		Address:  addr.Address,
		Balance:  uint64(addr.Balance),
		LastSeen: uint64(addr.LastSeen),
	}
}

func (s AddressBalance) GoFromProtobufType(addr protobufLocal.AddressDetails) *AddressBalance {

	return &AddressBalance{
		Address: addr.Address,
		//Nonce:    int64(addr.Nonce),
		Balance:  int64(addr.Balance),
		LastSeen: int64(addr.LastSeen),
	}
}

func (s AddressBalance) GoAddressFromProtobufType(addr protobufLocal.AddressDetails) *Address {

	return &Address{
		Address:  addr.Address,
		Balance:  int64(addr.Balance),
		LastSeen: int64(addr.LastSeen),
	}
}
