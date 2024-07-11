package types

import (
	"encoding/json"
	protobufLocal "src/protobuf"
)

type BeaconHeadersResponse struct {
	Data                []*HeaderResponseData `json:"data"`
	ExecutionOptimistic bool                  `json:"executionOptimistic"`
	Finalized           bool                  `json:"finalized"`
}

type MongoBeaconHeadersResponse struct {
	Data                []*MongoHeaderResponseData `json:"data"`
	ExecutionOptimistic bool                       `json:"executionOptimistic"`
	Finalized           bool                       `json:"finalized"`
}

type HeaderResponseData struct {
	Header    *SignedBeaconBlockHeader `json:"data"`
	Root      string                   `json:"root"`
	Canonical bool                     `json:"canonical"`
}

type MongoHeaderResponseData struct {
	Header    *MongoSignedBlockHeader `json:"data"`
	Root      string                  `json:"root"`
	Canonical bool                    `json:"canonical"`
}

func (s BeaconHeadersResponse) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s BeaconHeadersResponse) FromGoType(blob BeaconHeadersResponse) MongoBeaconHeadersResponse {

	var data []*MongoHeaderResponseData
	for _, block := range blob.Data {
		headerData := SignedBeaconBlockHeader{}.FromGoType(*block.Header)
		dat := MongoHeaderResponseData{
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

	var data []*MongoHeaderResponseData
	for _, block := range blob.Data {
		headerData := SignedBeaconBlockHeader{}.MongoFromProtobufType(*block.Header)
		dat := MongoHeaderResponseData{
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
