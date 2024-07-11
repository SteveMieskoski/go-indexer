package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
)

type SidecarsResponse struct {
	Data []*Blob `json:"data"`
}

type Blob struct {
	Index                       string                   `json:"index"`
	Blob                        string                   `json:"blob"`
	SignedBlockHeader           *SignedBeaconBlockHeader `json:"signed_block_header"`
	KzgCommitment               string                   `json:"kzg_commitment"`
	KzgProof                    string                   `json:"kzg_proof"`
	KzgCommitmentInclusionProof []string                 `json:"kzg_commitment_inclusion_proof"`
}

type MongoBlob struct {
	Index                       string                 `bson:"index" json:"index"`
	Blob                        string                 `bson:"blob" json:"blob"`
	SignedBlockHeader           MongoSignedBlockHeader `bson:"signed_block_header" json:"signed_block_header"`
	KzgCommitment               string                 `bson:"kzg_commitment" json:"kzg_commitment"`
	KzgProof                    string                 `bson:"kzg_proof" json:"kzg_proof"`
	KzgCommitmentInclusionProof []string               `bson:"kzg_commitment_inclusion_proof" json:"kzg_commitment_inclusion_proof"`
}

func (s Blob) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s Blob) FromGoType(blob Blob) MongoBlob {

	sbh := SignedBeaconBlockHeader{}.FromGoType(*blob.SignedBlockHeader)

	return MongoBlob{
		Index:                       blob.Index,
		Blob:                        blob.Blob,
		SignedBlockHeader:           sbh,
		KzgCommitment:               blob.KzgCommitment,
		KzgProof:                    blob.KzgProof,
		KzgCommitmentInclusionProof: blob.KzgCommitmentInclusionProof,
	}
}

func (s Blob) ProtobufFromGoType(blob Blob) protobufLocal.Blob {

	sbh := SignedBeaconBlockHeader{}.ProtobufFromGoType(*blob.SignedBlockHeader)

	return protobufLocal.Blob{
		Index:                       blob.Index,
		Blob:                        blob.Blob,
		SignedBlockHeader:           &sbh,
		KzgCommitment:               blob.KzgCommitment,
		KzgProof:                    blob.KzgProof,
		KzgCommitmentInclusionProof: blob.KzgCommitmentInclusionProof,
	}
}

func (s Blob) MongoFromProtobufType(blob protobufLocal.Blob) *MongoBlob {

	sbh := SignedBeaconBlockHeader{}.MongoFromProtobufType(*blob.SignedBlockHeader)

	return &MongoBlob{
		Index:                       blob.Index,
		Blob:                        blob.Blob,
		SignedBlockHeader:           *sbh,
		KzgCommitment:               blob.KzgCommitment,
		KzgProof:                    blob.KzgProof,
		KzgCommitmentInclusionProof: blob.KzgCommitmentInclusionProof,
	}
}

func (s Blob) ProtobufFromMongoType(blob MongoBlob) *protobufLocal.Blob {

	sbh := SignedBeaconBlockHeader{}.ProtobufFromMongoType(blob.SignedBlockHeader)

	return &protobufLocal.Blob{
		Index:                       blob.Index,
		Blob:                        blob.Blob,
		SignedBlockHeader:           sbh,
		KzgCommitment:               blob.KzgCommitment,
		KzgProof:                    blob.KzgProof,
		KzgCommitmentInclusionProof: blob.KzgCommitmentInclusionProof,
	}

}
