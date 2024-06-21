package kafka

import (
	"testing"
)

func Test_consumer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"consume1"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer()
		})
	}
}

//func TestConsumer_Cleanup(t *testing.T) {
//	type fields struct {
//		ready chan bool
//	}
//	type args struct {
//		in0 sarama.ConsumerGroupSession
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			consumer := &Consumer{
//				ready: tt.fields.ready,
//			}
//			if err := consumer.Cleanup(tt.args.in0); (err != nil) != tt.wantErr {
//				t.Errorf("Cleanup() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestConsumer_ConsumeClaim(t *testing.T) {
//	type fields struct {
//		ready chan bool
//	}
//	type args struct {
//		session sarama.ConsumerGroupSession
//		claim   sarama.ConsumerGroupClaim
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			consumer := &Consumer{
//				ready: tt.fields.ready,
//			}
//			if err := consumer.ConsumeClaim(tt.args.session, tt.args.claim); (err != nil) != tt.wantErr {
//				t.Errorf("ConsumeClaim() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func TestConsumer_Setup(t *testing.T) {
//	type fields struct {
//		ready chan bool
//	}
//	type args struct {
//		in0 sarama.ConsumerGroupSession
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			consumer := &Consumer{
//				ready: tt.fields.ready,
//			}
//			if err := consumer.Setup(tt.args.in0); (err != nil) != tt.wantErr {
//				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
//
//func Test_consumer(t *testing.T) {
//	tests := []struct {
//		name string
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			consumer()
//		})
//	}
//}
//
//func Test_toggleConsumptionFlow(t *testing.T) {
//	type args struct {
//		client   sarama.ConsumerGroup
//		isPaused *bool
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			toggleConsumptionFlow(tt.args.client, tt.args.isPaused)
//		})
//	}
//}
