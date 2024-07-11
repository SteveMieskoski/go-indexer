package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"src/types"
)

//type BeaconBlockHeader struct {
//	Slot          string `json:"slot"`
//	ProposerIndex string `json:"proposer_index"`
//	ParentRoot    string `json:"parent_root"`
//	StateRoot     string `json:"state_root"`
//	BodyRoot      string `json:"body_root"`
//}
//
//type SignedBeaconBlockHeader struct {
//	Message   *BeaconBlockHeader `json:"message"`
//	Signature string             `json:"signature"`
//}
//
//type Sidecar struct {
//	Index                    string                   `json:"index"`
//	Blob                     string                   `json:"blob"`
//	SignedBeaconBlockHeader  *SignedBeaconBlockHeader `json:"signed_block_header"`
//	KzgCommitment            string                   `json:"kzg_commitment"`
//	KzgProof                 string                   `json:"kzg_proof"`
//	CommitmentInclusionProof []string                 `json:"kzg_commitment_inclusion_proof"`
//}
//
//type SidecarsResponse struct {
//	Data []*Sidecar `json:"data"`
//}

type SignedBlock struct {
	Message   json.RawMessage `json:"message"` // represents the block values based on the version
	Signature string          `json:"signature"`
}

type GetBlockResponse struct {
	Data *SignedBlock `json:"data"`
}

type BeaconBlockHeader struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

type SignedBeaconBlockHeader struct {
	Message   *BeaconBlockHeader `json:"message"`
	Signature string             `json:"signature"`
}

type SignedBeaconBlockHeaderContainer struct {
	Header    *SignedBeaconBlockHeader `json:"header"`
	Root      string                   `json:"root"`
	Canonical bool                     `json:"canonical"`
}

type GetBlockHeadersResponse struct {
	Data                []*SignedBeaconBlockHeaderContainer `json:"data"`
	ExecutionOptimistic bool                                `json:"execution_optimistic"`
	Finalized           bool                                `json:"finalized"`
}

func main() {

	res, err := http.Get("http://127.0.0.1:3500/eth/v1/beacon/headers")
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	//fmt.Printf("client: response body: %s\n", resBody[0:60000])

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}
	}(res.Body)

	var arr *types.BeaconHeadersResponse
	//var arr *SidecarsResponse
	if err := json.Unmarshal([]byte(resBody), &arr); err != nil {
		panic(err)
	}

	fmt.Printf("json map: %v\n", arr)
	fmt.Printf("json map: %v\n", arr.Data[0])
	//// TODO RESOLVE ->
	//// https://pkg.go.dev/github.com/prysmaticlabs/prysm/v4@v4.2.1/beacon-chain/rpc/eth/beacon#SignedBlock
	//// https://github.com/prysmaticlabs/prysm/blob/v4.2.1/beacon-chain/rpc/eth/beacon/structs.go#L138
	//var dat *types.SignedBeaconBlockHeader
	//if err := json.Unmarshal(*arr.Data[0], &dat); err != nil {
	//	panic(err)
	//}

	//fmt.Printf("json map: %v\n", arr.Data[0].Header)
	//fmt.Println(arr)
	//fmt.Printf("json map: %v\n", arr.Data[0].Root)
	//type Data struct {
	//	data []types.Blob
	//}
	////var data Data
	////var data map[string]interface{}
	//var data []interface{}
	//decoder := json.NewDecoder(res.Body)
	//
	//err = decoder.Decode(&data)
	//if err != nil {
	//	fmt.Printf("client: could not read response body: %s\n", err)
	//	return
	//}
	//
	////println(data[0].KzgCommitment)
	//fmt.Printf("json map: %v\n", data)
}
