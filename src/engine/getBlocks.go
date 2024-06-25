package engine

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
	"os/signal"
	"src/types"
	"src/utils"
	"syscall"
	"time"
)

func GetBlocks(url string) chan types.Block {
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
			utils.Logger.Error("subscribe error:", err)
			return
		}

		// The connection is established now.
		// Update the channel with the current block.
		var lastBlock types.Block
		err = client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "latest", true)
		if err != nil {
			utils.Logger.Error("can't get latest block:", err)
			return
		}
		subch <- lastBlock
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			sub.Unsubscribe()
			utils.Logger.Info("exiting")
			close(subch)
			return
		}
		utils.Logger.Error("connection lost: ", <-sub.Err())
	}()

	return subch
}
