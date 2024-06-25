package kafka

import (
	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type encoder struct {
	msg proto.Message
}

func (e *encoder) Encode() ([]byte, error) {
	return proto.Marshal(e.msg)
}
func (e *encoder) Length() int {
	return proto.Size(e.msg)
}

// ProtoEncoder encodes ProtocolBuffer messages natively on Sarama
// Utilizing this encoder will only encode the message during the producing time
func ProtoEncoder(msg proto.Message) sarama.Encoder {
	return &encoder{msg: msg}
}
