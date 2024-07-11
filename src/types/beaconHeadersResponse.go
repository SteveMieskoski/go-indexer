package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
)

//type SignedBeaconBlockHeaderContainer struct {
//	Header    *SignedBeaconBlockHeader `json:"header"`
//	Root      string                   `json:"root"`
//	Canonical bool                     `json:"canonical"`
//}
//
//type GetBlockHeadersResponse struct {
//	Data                []*SignedBeaconBlockHeaderContainer `json:"data"`
//	ExecutionOptimistic bool                                `json:"execution_optimistic"`
//	Finalized           bool                                `json:"finalized"`
//}

type BeaconHeadersResponse struct {
	Data                []*HeaderResponseContainer `json:"data"`
	ExecutionOptimistic bool                       `json:"execution_optimistic"`
	Finalized           bool                       `json:"finalized"`
}

type MongoBeaconHeadersResponse struct {
	Data                []*MongoHeaderResponseContainer `json:"data"`
	ExecutionOptimistic bool                            `json:"execution_optimistic"`
	Finalized           bool                            `json:"finalized"`
}

type HeaderResponseContainer struct {
	Header    *SignedBeaconBlockHeader `json:"header"`
	Root      string                   `json:"root"`
	Canonical bool                     `json:"canonical"`
}

type MongoHeaderResponseContainer struct {
	Header    *MongoSignedBlockHeader `json:"header"`
	Root      string                  `json:"root"`
	Canonical bool                    `json:"canonical"`
}

func (s HeaderResponseContainer) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s BeaconHeadersResponse) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s BeaconHeadersResponse) MongoFromGoType(blob BeaconHeadersResponse) MongoBeaconHeadersResponse {

	var data []*MongoHeaderResponseContainer
	for _, block := range blob.Data {
		//headerResponse := MongoHeaderResponseContainer{}
		//json.Unmarshal(*block.Header, &headerResponse)
		headerData := SignedBeaconBlockHeader{}.MongoFromGoType(*block.Header)
		dat := MongoHeaderResponseContainer{
			Header:    &headerData,
			Root:      block.Root,
			Canonical: block.Canonical,
		}
		data = append(data, &dat)
	}

	return MongoBeaconHeadersResponse{
		Data:                data,
		ExecutionOptimistic: blob.ExecutionOptimistic,
		Finalized:           blob.Finalized,
	}
}

func (s BeaconHeadersResponse) ProtobufFromGoType(blob BeaconHeadersResponse) protobufLocal.BeaconHeaderResponse {

	var data []*protobufLocal.BeaconHeaderResponse_Header
	for _, block := range blob.Data {
		headerData := SignedBeaconBlockHeader{}.ProtobufFromGoType(*block.Header)
		dat := protobufLocal.BeaconHeaderResponse_Header{
			Header:    &headerData,
			Root:      block.Root,
			Canonical: block.Canonical,
		}
		data = append(data, &dat)
	}

	return protobufLocal.BeaconHeaderResponse{
		Data:                data,
		ExecutionOptimistic: blob.ExecutionOptimistic,
		Finalized:           blob.Finalized,
	}
}

func (s BeaconHeadersResponse) MongoFromProtobufType(blob protobufLocal.BeaconHeaderResponse) *MongoBeaconHeadersResponse {

	var data []*MongoHeaderResponseContainer
	for _, block := range blob.Data {
		headerData := SignedBeaconBlockHeader{}.MongoFromProtobufType(*block.Header)
		dat := MongoHeaderResponseContainer{
			Header:    headerData,
			Root:      block.Root,
			Canonical: block.Canonical,
		}
		data = append(data, &dat)
	}

	return &MongoBeaconHeadersResponse{
		Data:                data,
		ExecutionOptimistic: blob.ExecutionOptimistic,
		Finalized:           blob.Finalized,
	}
}

func (s BeaconHeadersResponse) ProtobufFromMongoType(blob MongoBeaconHeadersResponse) *protobufLocal.BeaconHeaderResponse {

	var data []*protobufLocal.BeaconHeaderResponse_Header
	for _, block := range blob.Data {
		headerData := SignedBeaconBlockHeader{}.ProtobufFromMongoType(*block.Header)
		dat := protobufLocal.BeaconHeaderResponse_Header{
			Header:    headerData,
			Root:      block.Root,
			Canonical: block.Canonical,
		}
		data = append(data, &dat)
	}

	return &protobufLocal.BeaconHeaderResponse{
		Data:                data,
		ExecutionOptimistic: blob.ExecutionOptimistic,
		Finalized:           blob.Finalized,
	}

}
