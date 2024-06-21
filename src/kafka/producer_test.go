package kafka

import (
	//"src/types"
	//base "src/utils"
	"testing"
)

func Test_connect(t *testing.T) {

	//s := make(chan string)
	connect()
	//block := types.Block{
	//	BaseFeePerGas: 0x1cc97b,
	//	BlockNumber:   0x2f,
	//	Difficulty:    0x0,
	//	GasLimit:      0x1c9c380,
	//	GasUsed:       0x17fc7,
	//	Hash:          base.HexToHash("0x29c1249c7a63f649da4130ad4ad9b5faab2c42ad98b7f8f762562f4b7fdd1ebb"),
	//	Miner:         base.HexToAddress("0x123463a4b065722e99115d6c222f267d9cabb524"),
	//	ParentHash:    base.HexToHash("0xed2c6b7308c11c1855d807ab4c5483cd7fe0e8711ab94d7918aea2757429ef41"),
	//	Timestamp:     0x6673b01b,
	//	Transactions:  nil,
	//	Uncles:        nil,
	//	Withdrawals:   nil,
	//	Number:        0x2f,
	//}
	//
	//s <- block.String()
}

//func Test_connect(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			connect()
//		})
//	}
//}

//func Test_newProducerProvider(t *testing.T) {
//	type args struct {
//		brokers                       []string
//		producerConfigurationProvider func() *sarama.Config
//	}
//	tests := []struct {
//		name string
//		args args
//		want *producerProvider
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := newProducerProvider(tt.args.brokers, tt.args.producerConfigurationProvider); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("newProducerProvider() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_produceTestRecord(t *testing.T) {
//	type args struct {
//		producerProvider *producerProvider
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			produceTestRecord(tt.args.producerProvider)
//		})
//	}
//}
//
//func Test_producerProvider_borrow(t *testing.T) {
//	type fields struct {
//		transactionIdGenerator int32
//		producersLock          sync.Mutex
//		producers              []sarama.AsyncProducer
//		producerProvider       func() sarama.AsyncProducer
//	}
//	tests := []struct {
//		name         string
//		fields       fields
//		wantProducer sarama.AsyncProducer
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &producerProvider{
//				transactionIdGenerator: tt.fields.transactionIdGenerator,
//				producersLock:          tt.fields.producersLock,
//				producers:              tt.fields.producers,
//				producerProvider:       tt.fields.producerProvider,
//			}
//			if gotProducer := p.borrow(); !reflect.DeepEqual(gotProducer, tt.wantProducer) {
//				t.Errorf("borrow() = %v, want %v", gotProducer, tt.wantProducer)
//			}
//		})
//	}
//}
//
//func Test_producerProvider_clear(t *testing.T) {
//	type fields struct {
//		transactionIdGenerator int32
//		producersLock          sync.Mutex
//		producers              []sarama.AsyncProducer
//		producerProvider       func() sarama.AsyncProducer
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &producerProvider{
//				transactionIdGenerator: tt.fields.transactionIdGenerator,
//				producersLock:          tt.fields.producersLock,
//				producers:              tt.fields.producers,
//				producerProvider:       tt.fields.producerProvider,
//			}
//			p.clear()
//		})
//	}
//}
//
//func Test_producerProvider_release(t *testing.T) {
//	type fields struct {
//		transactionIdGenerator int32
//		producersLock          sync.Mutex
//		producers              []sarama.AsyncProducer
//		producerProvider       func() sarama.AsyncProducer
//	}
//	type args struct {
//		producer sarama.AsyncProducer
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			p := &producerProvider{
//				transactionIdGenerator: tt.fields.transactionIdGenerator,
//				producersLock:          tt.fields.producersLock,
//				producers:              tt.fields.producers,
//				producerProvider:       tt.fields.producerProvider,
//			}
//			p.release(tt.args.producer)
//		})
//	}
//}
