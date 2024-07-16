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
	BlockHeaderChannel() chan types.Block
	GetBlocks() chan types.Block
	GetPastBlocks(blockToGet chan int) chan types.Block
}

type BlockRetriever struct {
	RedisClient redisdb.RedisClient
}

func NewBlockRetriever(redisClient redisdb.RedisClient) *BlockRetriever {
	return &BlockRetriever{redisClient}
}

func (b BlockRetriever) BlockHeaderChannel() chan types.Block {

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

		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			sub.Unsubscribe()
			utils.Logger.Info("exiting: GetBlocks")
			defer close(subch)
			return
		}
		utils.Logger.Error("connection lost: ", <-sub.Err())
	}()

	return subch
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

		time.Sleep(5 * time.Millisecond)
		// The connection is established now.
		// Update the channel with the current block.
		//var lastBlock types.Block
		//err = client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "latest", true)
		//if err != nil {
		//	utils.Logger.Error("can't get latest block:", err)
		//	return
		//}
		//subch <- lastBlock
		sig := <-sigs
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			sub.Unsubscribe()
			utils.Logger.Info("exiting: GetBlocks")
			defer close(subch)
			return
		}
		utils.Logger.Error("connection lost: ", <-sub.Err())
	}()

	return subch
}

func (b BlockRetriever) GetBlock(blockToGet int) types.Block {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	// Ensure that subch receives the latest block.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock types.Block
	blockNumberToRetreive := strconv.FormatInt(int64(blockToGet), 16)
	err := client.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "0x"+blockNumberToRetreive, true)
	if err != nil {
		utils.Logger.Error("can't get block:", err)
	}

	utils.Logger.Infof("retrieved block %d", blockToGet)
	return lastBlock
}

func (b BlockRetriever) LatestBlock() int {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	// Ensure that subch receives the latest block.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The connection is established now.
	// Update the channel with the current block.
	var latestBlock string
	err := client.CallContext(ctx, &latestBlock, "eth_blockNumber")
	if err != nil {
		utils.Logger.Error("can't get block:", err)
	}
	parseInt, err := strconv.ParseInt(latestBlock[2:], 16, 0)

	return int(parseInt)
}

func (b BlockRetriever) GetBlockBatch(firstBlockToGet int, lastBlockToGet int) ([]rpc.BatchElem, error) {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	var lastBlock []rpc.BatchElem

	for i := firstBlockToGet; i < lastBlockToGet; i++ {
		blockNumberToRetreive := strconv.FormatInt(int64(i), 16)
		var block types.Block

		lastBlock = append(lastBlock, rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{"0x" + blockNumberToRetreive, true},
			Result: &block,
		})
	}

	err := client.BatchCall(lastBlock)

	if err != nil {
		utils.Logger.Error("can't get block:", err)
		return nil, err
	}

	//utils.Logger.Infof("retrieved block %d", blockToGet)
	return lastBlock, nil
}

func (b BlockRetriever) GetTransaction(txHash string) types.Transaction {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	// Ensure that subch receives the latest block.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock types.Transaction
	err := client.CallContext(ctx, &lastBlock, "eth_getTransactionByHash", txHash)
	if err != nil {
		utils.Logger.Error("can't get transaction:", err)
	}

	utils.Logger.Infof("retrieved transaction %s", txHash)
	return lastBlock
}

func (b BlockRetriever) GetTransactionReceipt(txHash string) types.Receipt {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	// Ensure that subch receives the latest block.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The connection is established now.
	// Update the channel with the current block.
	var lastBlock types.Receipt
	err := client.CallContext(ctx, &lastBlock, "eth_getTransactionReceipt", txHash)
	if err != nil {
		utils.Logger.Error("can't get receipt:", err)
	}

	utils.Logger.Infof("retrieved receipt %s", txHash)
	return lastBlock
}
