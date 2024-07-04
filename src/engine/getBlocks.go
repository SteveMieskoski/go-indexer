package engine

import (
	"context"
	"github.com/ethereum/go-ethereum/rpc"
	"os"
	"os/signal"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"syscall"
	"time"
)

type IBlockRetriever interface {
	GetBlocks() chan types.Block
	GetPastBlocks(blockToGet chan int) chan types.Block
}

type BlockRetriever struct {
	RedisClient redisdb.RedisClient
}

func NewBlockRetriever(redisClient redisdb.RedisClient) *BlockRetriever {
	return &BlockRetriever{redisClient}
}

func (b BlockRetriever) GetBlocks() chan types.Block {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	//WS_RPC_URL
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
			utils.Logger.Info("exiting: GetBlocks")
			close(subch)
			return
		}
		utils.Logger.Error("connection lost: ", <-sub.Err())
	}()

	return subch
}

func (b BlockRetriever) GetPastBlocks(blockToGet int) types.Block {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//subch := make(chan types.Block)

	println("check 3")
	// Ensure that subch receives the latest block.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	println("check 4")
	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock types.Block
	blockNumberToRetreive := strconv.FormatInt(int64(blockToGet), 16)
	println(blockNumberToRetreive)
	err := client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "0x"+blockNumberToRetreive, true)
	if err != nil {
		utils.Logger.Error("can't get latest block:", err)
	}
	//subch <- lastBlock
	//sig := <-sigs
	//if sig == syscall.SIGINT || sig == syscall.SIGTERM {
	//	utils.Logger.Info("exiting: GetPastBlocks")
	//	//close(subch)
	//}
	utils.Logger.Info("return prior block")
	return lastBlock
}
