package engine

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
	"os/signal"
	"src/types"
	"src/utils"
	"syscall"
)

func GetBlockReceipts(blockHash string) (types.Receipts, error) {
	url := os.Getenv("HTTP_RPC_URL")
	client, _ := rpc.Dial(url) // todo: move this external to the function (maybe)
	defer client.Close()

	subch := make(chan types.Receipts)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer close(subch)
	// Hold result in r of types.Receipts
	var r types.Receipts
	err := client.CallContext(ctx, &r, "eth_getBlockReceipts", blockHash)
	if err != nil {
		utils.Logger.Error("can't get latest block receipt:", err)
		utils.Logger.Errorf("can't get latest block receipt for hash: %s\n", blockHash)

		return r, err
	}
	//if len(r) > 1 {
	//	println(r[0].String())
	//}

	return r, nil
}
