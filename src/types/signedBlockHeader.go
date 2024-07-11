package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
)

type SignedBeaconBlockHeader struct {
	Message   BeaconBlockHeader `json:"message"`
	Signature string            `json:"signature"`
}

type MongoSignedBlockHeader struct {
	Message   MongoBeaconBlockHeader `bson:"message" json:"message"`
	Signature string                 `bson:"signature" json:"signature"`
}

func (s SignedBeaconBlockHeader) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s SignedBeaconBlockHeader) FromGoType(blob SignedBeaconBlockHeader) MongoSignedBlockHeader {

	return MongoSignedBlockHeader{
		Message:   BeaconBlockHeader{}.FromGoType(blob.Message),
		Signature: blob.Signature,
	}
}

func (s SignedBeaconBlockHeader) ProtobufFromGoType(blob SignedBeaconBlockHeader) protobufLocal.SignedBeaconBlockHeader {

	msg := BeaconBlockHeader{}.ProtobufFromGoType(blob.Message)

	return protobufLocal.SignedBeaconBlockHeader{
		BeaconHeader: &msg,
		Signature:    blob.Signature,
	}
}

func (s SignedBeaconBlockHeader) MongoFromProtobufType(blob protobufLocal.SignedBeaconBlockHeader) *MongoSignedBlockHeader {

	msg := BeaconBlockHeader{}.MongoFromProtobufType(*blob.BeaconHeader)
	return &MongoSignedBlockHeader{
		Message:   *msg,
		Signature: blob.Signature,
	}
}

func (s SignedBeaconBlockHeader) ProtobufFromMongoType(blob MongoSignedBlockHeader) *protobufLocal.SignedBeaconBlockHeader {

	msg := BeaconBlockHeader{}.ProtobufFromMongoType(blob.Message)

	return &protobufLocal.SignedBeaconBlockHeader{
		BeaconHeader: msg,
		Signature:    blob.Signature,
	}

}
