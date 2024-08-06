package produce

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	"io"
	"net/http"
	"os"
	"os/signal"
	"src/types"
	"syscall"
)

func BeaconHeader() chan types.BBlock {

	url := os.Getenv("BEACON_RPC_URL")
	newBeaconBlock := make(chan types.BBlock)

	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		client := sse.NewClient(url + "/eth/v1/events?topics=block")

		events := make(chan *sse.Event)
		err := client.SubscribeChan("messages", events)
		if err != nil {
			return
		}

		for event := range events {
			fmt.Println(string(event.Data))

			data := types.BBlock{}
			_ = json.Unmarshal([]byte(string(event.Data)), &data)
			println(data.Slot)
			select {
			case <-ctx.Done():
				stop()
				fmt.Println("signal received")
				close(events)
				syscall.Exit(1)
				return
			default:
			}
			newBeaconBlock <- data
		}

		if err != nil {
			panic(err)
			return
		}
	}()

	return newBeaconBlock
}

func GetBeaconHeaderBySlot(slotNum string) *types.BeaconHeadersResponse {

	url := os.Getenv("BEACON_RPC_URL")
	res, err := http.Get(url + "/eth/v1/beacon/headers/" + slotNum)
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

	var bHeaders *types.BeaconHeadersResponse
	if err := json.Unmarshal([]byte(resBody), &bHeaders); err != nil {
		panic(err)
	}

	return bHeaders
}

func GetBeaconHeader() *types.BeaconHeadersResponse {

	url := os.Getenv("BEACON_RPC_URL")
	res, err := http.Get(url + "/eth/v1/beacon/headers")
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

	var bHeaders *types.BeaconHeadersResponse
	if err := json.Unmarshal([]byte(resBody), &bHeaders); err != nil {
		panic(err)
	}

	return bHeaders
}

func GetBlobSideCars(slot string) types.SidecarsResponse {

	url := os.Getenv("BEACON_RPC_URL")
	res, err := http.Get(url + "/eth/v1/beacon/blob_sidecars/" + slot)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("client: could not read response body: %s\n", err)
		}
	}(res.Body)

	var arr *types.SidecarsResponse
	if err := json.Unmarshal([]byte(resBody), &arr); err != nil {
		fmt.Printf("client: could not unmarshal body: %s\n", err)
		panic(err)
	}

	return *arr
}
