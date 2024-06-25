package models

import (
	"github.com/ethereum/go-ethereum/common"
	"reflect"
	"src/protobuf"
	"src/types"
	base "src/utils"
	"testing"
)

func TestBlockFromGoType(t *testing.T) {
	type args struct {
		block types.Block
	}
	tests := []struct {
		name string
		args args
		want Block
	}{
		{
			name: "TestBlockFromGoType",
			args: args{
				block: types.Block{
					BaseFeePerGas: 0,
					BlockNumber:   0,
					Difficulty:    0,
					GasLimit:      0,
					GasUsed:       0,
					Hash: base.Hash{
						Hash: common.HexToHash("0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb"),
					},
					Miner: base.Address{
						Address: common.HexToAddress("0xb3d1002f77c20a96477d4a41d60853f8b9786393"),
					},
					ParentHash: base.Hash{
						Hash: common.HexToHash("0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb"),
					},
					Timestamp:    0,
					Transactions: nil,
					Uncles:       nil,
					Withdrawals:  nil,
					Number:       0,
				},
			},
			want: Block{
				Id:            "0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb",
				BaseFeePerGas: "0",
				BlockNumber:   "0",
				Difficulty:    "0",
				GasLimit:      "0",
				GasUsed:       "0",
				Hash:          "0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb",
				Miner:         "0xb3d1002f77c20a96477d4a41d60853f8b9786393",
				ParentHash:    "0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb",
				Timestamp:     0,
				Transactions:  nil,
				Uncles:        nil,
				Withdrawals:   nil,
				Number:        "0",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BlockFromGoType(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockFromGoType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockProtobufFromGoType(t *testing.T) {
	type args struct {
		block types.Block
	}
	tests := []struct {
		name string
		args args
		want protobuf.Block
	}{
		{
			name: "TestBlockFromGoType",
			args: args{
				block: types.Block{
					BaseFeePerGas: 0,
					BlockNumber:   0,
					Difficulty:    0,
					GasLimit:      0,
					GasUsed:       0,
					Hash: base.Hash{
						Hash: common.HexToHash("0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb"),
					},
					Miner: base.Address{
						Address: common.HexToAddress("0xb3d1002f77c20a96477d4a41d60853f8b9786393"),
					},
					ParentHash: base.Hash{
						Hash: common.HexToHash("0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb"),
					},
					Timestamp:    0,
					Transactions: nil,
					Uncles:       nil,
					Withdrawals:  nil,
					Number:       0,
				},
			},
			want: protobuf.Block{
				BaseFeePerGas: "0",
				BlockNumber:   "0",
				Difficulty:    "0",
				GasLimit:      "0",
				GasUsed:       "0",
				Hash:          "0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb",
				Miner:         "0xb3d1002f77c20a96477d4a41d60853f8b9786393",
				ParentHash:    "0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb",
				Timestamp:     0,
				Transactions:  nil,
				Uncles:        nil,
				Withdrawals:   nil,
				Number:        "0",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BlockProtobufFromGoType(tt.args.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockProtobufFromGoType() = %v, want %v", got, tt.want)
			}
		})
	}
}
