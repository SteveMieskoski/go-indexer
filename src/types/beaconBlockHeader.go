package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
)

type BeaconBlockHeader struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

type MongoBeaconBlockHeader struct {
	Slot          string `bson:"slot" json:"slot"`
	ProposerIndex string `bson:"proposer_index" json:"proposer_index"`
	ParentRoot    string `bson:"parent_root" json:"parent_root"`
	StateRoot     string `bson:"state_root" json:"state_root"`
	BodyRoot      string `bson:"body_root" json:"body_root"`
}

func (s BeaconBlockHeader) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s BeaconBlockHeader) FromGoType(blob BeaconBlockHeader) MongoBeaconBlockHeader {

	return MongoBeaconBlockHeader{
		Slot:          blob.Slot,
		ProposerIndex: blob.ProposerIndex,
		ParentRoot:    blob.ParentRoot,
		StateRoot:     blob.StateRoot,
		BodyRoot:      blob.BodyRoot,
	}
}

func (s BeaconBlockHeader) ProtobufFromGoType(blob BeaconBlockHeader) protobufLocal.BeaconBlockHeader {

	return protobufLocal.BeaconBlockHeader{
		Slot:          blob.Slot,
		ProposerIndex: blob.ProposerIndex,
		ParentRoot:    blob.ParentRoot,
		StateRoot:     blob.StateRoot,
		BodyRoot:      blob.BodyRoot,
	}
}

func (s BeaconBlockHeader) MongoFromProtobufType(blob protobufLocal.BeaconBlockHeader) *MongoBeaconBlockHeader {

	return &MongoBeaconBlockHeader{
		Slot:          blob.Slot,
		ProposerIndex: blob.ProposerIndex,
		ParentRoot:    blob.ParentRoot,
		StateRoot:     blob.StateRoot,
		BodyRoot:      blob.BodyRoot,
	}
}

func (s BeaconBlockHeader) ProtobufFromMongoType(blob MongoBeaconBlockHeader) *protobufLocal.BeaconBlockHeader {

	return &protobufLocal.BeaconBlockHeader{
		Slot:          blob.Slot,
		ProposerIndex: blob.ProposerIndex,
		ParentRoot:    blob.ParentRoot,
		StateRoot:     blob.StateRoot,
		BodyRoot:      blob.BodyRoot,
	}

}
