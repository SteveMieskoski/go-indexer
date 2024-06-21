package engine

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
	"os/signal"
	"src/types"
	"syscall"
	"time"
)

//type Connection struct {
//	Chain string
//}

//	type Block struct {
//		Number *hexutil.Big
//	}

//func createWsClient(url string) {
//	// Connect the client.
//	client, _ := rpc.Dial(url)
//
//	subch := make(chan Block)
//
//	// Ensure that subch receives the latest block.
//	go func() {
//		for i := 0; ; i++ {
//			if i > 5 {
//				//time.Sleep(2 * time.Second)
//				return
//			}
//			subscribeBlocks(client, subch)
//		}
//	}()
//
//	// Print events from the subscription as they arrive.
//	for block := range subch {
//		fmt.Println("latest block:", block.Number)
//	}
//}

func createWsClient(url string) {
	// Connect the client.
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	subch := make(chan types.Block)

	go func() {
		// Ensure that subch receives the latest block.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Subscribe to new blocks.
		sub, err := client.EthSubscribe(ctx, subch, "newHeads")
		if err != nil {
			fmt.Println("subscribe error:", err)
			return
		}

		// The connection is established now.
		// Update the channel with the current block.
		var lastBlock types.Block
		err = client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "latest", true)
		if err != nil {
			fmt.Println("can't get latest block:", err)
			return
		}

		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			sub.Unsubscribe()
			println("exiting")
			close(subch)
			return
		}
		fmt.Println("connection lost: ", <-sub.Err())
	}()

	// Print events from the subscription as they arrive.
	for block := range subch {
		fmt.Println("latest block:", block.Number)
	}
}

// subscribeBlocks runs in its own goroutine and maintains
// a subscription for new blocks.
func subscribeBlocks(client *rpc.Client, subch chan types.Block) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Subscribe to new blocks.
	sub, err := client.EthSubscribe(ctx, subch, "newHeads")
	if err != nil {
		fmt.Println("subscribe error:", err)
		return
	}

	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock types.Block
	err = client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "latest", true)
	if err != nil {
		fmt.Println("can't get latest block:", err)
		return
	}
	subch <- lastBlock

	// The subscription will deliver events to the channel. Wait for the
	// subscription to end for any reason, then loop around to re-establish
	// the connection.
	fmt.Println("connection lost: ", <-sub.Err())

	ctx.Done()
}
