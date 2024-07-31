package engine

import (
	"context"
	"fmt"
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
	url := os.Getenv("HTTP_RPC_URL")
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
	url := os.Getenv("HTTP_RPC_URL")
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
func (b BlockRetriever) GetAddressBalance(address string, blockNumber int64) (string, string, int64, error) {

	if len(address) < 42 {
		//println("Invalid Address Length")
		return "", "", blockNumber, fmt.Errorf("======================== invalid address %v", address)
	}
	// Connect the client.
	url := os.Getenv("HTTP_RPC_URL")
	client, err := rpc.Dial(url)
	if err != nil {
		utils.Logger.Error("error connecting to RPC endpoint:", err)
		return "", "", blockNumber, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	blockNumberHex := strconv.FormatInt(blockNumber, 16)

	var balance string
	var txCount string

	err = client.CallContext(ctx, &balance, "eth_getBalance", address, "0x"+blockNumberHex)

	if err != nil {
		utils.Logger.Error("can't get Balances:", err)
		return "", "", blockNumber, err
	}

	err = client.CallContext(ctx, &txCount, "eth_getTransactionCount", address, "0x"+blockNumberHex)

	return balance, txCount, blockNumber, nil
}

func (b BlockRetriever) GetAddressDetailsBatch(addressList []string, blockNumber int64) ([]rpc.BatchElem, int64, error) {
	// Connect the client.
	url := os.Getenv("WS_RPC_URL")
	client, _ := rpc.Dial(url)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigs)

	blockNumberHex := strconv.FormatInt(blockNumber, 16)
	var lastBlock []rpc.BatchElem

	for _, addr := range addressList {

		var block string

		lastBlock = append(lastBlock, rpc.BatchElem{
			Method: "eth_getBalance",
			Args:   []interface{}{addr, "0x" + blockNumberHex},
			Result: &block,
		})
	}

	err := client.BatchCall(lastBlock)

	if err != nil {
		utils.Logger.Error("can't get Balances:", err)
		return nil, blockNumber, err
	}

	return lastBlock, blockNumber, nil
}

func (b BlockRetriever) GetTransaction(txHash string) types.Transaction {
	// Connect the client.
	url := os.Getenv("HTTP_RPC_URL")
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
	url := os.Getenv("HTTP_RPC_URL")
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
