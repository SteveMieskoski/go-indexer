package utils

import (
	"github.com/ethereum/go-ethereum/crypto"
)

func Keccak256(data []byte) []byte {
	return crypto.Keccak256(data)
}

//
//func Uint256() {
//	number = abi.U256(big.NewInt(3))
//}
