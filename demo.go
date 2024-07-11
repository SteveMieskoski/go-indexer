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

func main() {

	res, err := http.Get("http://127.0.0.1:3500/eth/v1/beacon/blob_sidecars/433")
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

	var arr *types.SidecarsResponse
	if err := json.Unmarshal([]byte(resBody), &arr); err != nil {
		panic(err)
	}
	//fmt.Println(arr)
	fmt.Printf("json map: %v\n", arr.Data[0].KzgCommitment)
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
