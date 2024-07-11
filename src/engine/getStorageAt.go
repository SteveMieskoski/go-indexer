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

func GetStorageAt(url string, blockHash string) chan types.Receipts {
	client, _ := rpc.Dial(url) // todo: move this external to the function (maybe)
	defer client.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	subch := make(chan types.Receipts)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer close(subch)
		// Hold result in r of types.Receipts
		var r types.Receipts
		err := client.CallContext(ctx, &r, "eth_getBlockReceipts", blockHash)
		if err != nil {
			utils.Logger.Error("can't get latest block receipt:", err)
			utils.Logger.Errorf("can't get latest block receipt for hash: %s\n", blockHash)

			return
		}
		if len(r) > 1 {
			println(r[0].String())
		}

		subch <- r

		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			println("exiting: GetBlockReceipts")
			return
		}
	}()

	return subch
}
