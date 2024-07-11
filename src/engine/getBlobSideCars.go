package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"src/types"
)

func GetBlobSideCars(slot string) types.SidecarsResponse {

	res, err := http.Get("http://127.0.0.1:3500/eth/v1/beacon/blob_sidecars/" + slot)
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

	return *arr
}
