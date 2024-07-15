package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	"os/signal"
	"src/engine"
	"src/kafka"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"syscall"
	"time"
)

type BeaconBlockRunner struct {
	priorRetrievalInProgress bool
	priorBeaconBlock         int
	currentBeaconBlock       int
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
}

func NewBeaconBlockRunner(producerFactory *kafka.ProducerProvider) BeaconBlockRunner {

	redisClient := redisdb.NewClient()

	return BeaconBlockRunner{
		priorBeaconBlock:         0,
		priorRetrievalInProgress: false,
		currentBeaconBlock:       0,
		producerFactory:          producerFactory,
		redis:                    *redisClient,
	}
}

func (b *BeaconBlockRunner) Demo() {
	//events := make(chan *sse.Event)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	events := make(chan *sse.Event)

	client := sse.NewClient("http://127.0.0.1:3500/eth/v1/events?topics=block")
	err := client.SubscribeChan("messages", events)
	if err != nil {
		return
	}
	type BBlock struct {
		Slot                string `json:"slot,omitempty"`
		Block               string `json:"block,omitempty"`
		ExecutionOptimistic bool   `json:"execution_optimistic,omitempty"`
	}

	for event := range events {
		fmt.Println(string(event.Data))

		data := BBlock{}
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
	}
	//client := sse.NewClient("http://127.0.0.1:3500/eth/v1/events?topics=block")
	////err := client.SubscribeChan("messages", events)
	////if err != nil {
	////	return
	////}
	//
	//
	//err := client.Subscribe("messages", func(msg *sse.Event) {
	//	// Got some data!
	//	fmt.Println(string(msg.Data))
	//
	//	data := BBlock{}
	//	_ = json.Unmarshal([]byte(string(msg.Data)), &data)
	//	println(data.Slot)
	//	select {
	//	case <-ctx.Done():
	//		stop()
	//		fmt.Println("signal received")
	//		syscall.Exit(1)
	//		return
	//	default:
	//	}
	//})
	if err != nil {
		panic(err)
		return
	}
}

func (b *BeaconBlockRunner) StartBeaconSync() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	bHeader := engine.GetBeaconHeader()
	currentSlotString := bHeader.Data[0].Header.Message.Slot
	err := b.redis.Set("BeaconSlotNumberOnSyncStart", currentSlotString)
	if err != nil {
		panic(err)
		return
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received")
		return
	default:
	}

	go b.getPriorSlots()
	b.getCurrentBeaconBlock()
}

func (b *BeaconBlockRunner) getCurrentBeaconBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	blockGen := engine.BeaconHeader()

	//fmt.Printf(lastBlock)
	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest Beacon block: %s", latestBlock.Slot)
		//engine.GetBeaconHeaderBySlot(latestBlock.Slot)

		go b.processBlobSideCars(latestBlock.Slot)

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}
	}

}

func (b *BeaconBlockRunner) processBlobSideCars(slot string) {
	sideCar := engine.GetBlobSideCars(slot)
	for _, blob := range sideCar.Data {
		completed := b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		if !completed {
			utils.Logger.Errorf("Blob Error for Blob index: %s, slot number: %s", blob.Index, slot)
			time.Sleep(50 * time.Millisecond)
			err := b.redis.Set("BlobError_"+blob.Index+"_"+slot, blob.Index)
			if err != nil {
				utils.Logger.Errorf("Error Setting Blob Error in Redis for Blob index: %s, slot number: %s", blob.Index, slot)
			}
			b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		}
	}
}

func (b *BeaconBlockRunner) processBeaconBlock(block *types.BeaconHeadersResponse) {

}

func (b *BeaconBlockRunner) getPriorSlots() {
	//dbAccess := r.newBlockSyncTrack()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	Num, _ := b.redis.Get("BeaconSlotNumberOnSyncStart")
	slotNumberOnSyncStart, _ := strconv.Atoi(Num)

	lastSlotRetrieved := 0
	val, err := b.redis.Get("lastPriorSlotRetrieved")
	if err != nil {
		err := b.redis.Set("lastPriorSlotRetrieved", lastSlotRetrieved)
		if err != nil {
			return
		}
	} else {
		lastSlotRetrieved, _ = strconv.Atoi(val)
	}

	for lastSlotRetrieved <= slotNumberOnSyncStart {
		fmt.Printf("Getting Blobs for Prior Slot %d\n", lastSlotRetrieved)
		go b.processBlobSideCars(strconv.Itoa(lastSlotRetrieved))

		//batchResponse, err := r.slotRetriever.GetSlotBatch(lastSlotRetrieved, batchEndSlot)
		//
		//if err != nil {
		//	goodRun = false
		//	continue
		//}
		//for _, response := range batchResponse {
		//	switch slot := response.Result.(type) {
		//	case *types.Block:
		//		_, err := dbAccess.Add(types.PgBlockSyncTrack{
		//			Number:                int64(slot.Number),
		//			Retrieved:             false,
		//			Processed:             false,
		//			ReceiptsProcessed:     false,
		//			TransactionsProcessed: false,
		//			ContractsProcessed:    false,
		//		})
		//
		//		if err != nil {
		//			// need to monitor to reduce batch size if db gets too slow
		//			// add error log here
		//			continue
		//		}
		//
		//		go r.processBlock(*slot)
		//	}
		//
		//}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}

		lastSlotRetrieved = lastSlotRetrieved + 1
		time.Sleep(100 * time.Millisecond)

	}

	utils.Logger.Info("exiting: getPriorSlot External")
}

func (b *BeaconBlockRunner) getCurrentBeaconBlock2() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	for {
		bHeader := engine.GetBeaconHeader()

		currentSlotString := bHeader.Data[0].Header.Message.Slot
		currentSlot, _ := strconv.Atoi(currentSlotString)

		if currentSlot > b.currentBeaconBlock {
			sideCar := engine.GetBlobSideCars(strconv.Itoa(currentSlot))

			for _, blob := range sideCar.Data {
				completed := b.producerFactory.Produce(types.BLOB_TOPIC, blob)
				if !completed {
					time.Sleep(50 * time.Millisecond)
					err := b.redis.Set("BlobError_"+blob.Index+"_"+currentSlotString, blob.Index)
					if err != nil {
						utils.Logger.Errorf("Error Setting Blob Error in Redis for Blob index: %s, slot number: %s", blob.Index, currentSlotString)
					}
					b.producerFactory.Produce(types.BLOB_TOPIC, blob)
				}
			}
			blockNumberKey := generateSlotNumberKey(currentSlot)
			err := b.redis.Set(blockNumberKey, true)
			if err != nil {
				panic(err)
			}
			err = b.redis.Set("latestBeaconBlock", currentSlot)
			if err != nil {
				panic(err)
			}
		}

		// catch up to chain head
		// does not run before syncing on start
		if currentSlot-b.currentBeaconBlock > 1 && b.priorRetrievalInProgress {
			go func() {
				b.getPriorBeaconBlocks(currentSlot, currentSlot-b.currentBeaconBlock)
			}()

		}

		// Runs once at the start of syncing
		if currentSlot-b.priorBeaconBlock > 1 && !b.priorRetrievalInProgress {
			b.priorRetrievalInProgress = true
			go func() {
				b.getPriorBeaconBlocks(currentSlot, currentSlot-b.priorBeaconBlock)
			}()

		}

		b.currentBeaconBlock = currentSlot
		fmt.Printf("finished getting blobs for slot %d\n", currentSlot)

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received: getCurrentBeaconBlock")
			return
		default:

		}

		//TODO: tie this more to checking when a new block head arrives, and looking to see if there
		//TODO: is a gap between the prior beacon block and the currently reported head.
		time.Sleep(10000 * time.Millisecond)
	}

}

func (b *BeaconBlockRunner) getPriorBeaconBlocks(currentSlot int, skippedSlotCount int) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for i := currentSlot - skippedSlotCount; i < currentSlot; i++ {
		blockNumberKey := generateSlotNumberKey(i)
		utils.Logger.Infof("Check if prior Beacon block retrieved: %d", i)
		_, redisErr := b.redis.Get(blockNumberKey)
		if redisErr != nil {
			sideCar := engine.GetBlobSideCars(strconv.Itoa(i))

			for _, blob := range sideCar.Data {

				b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
			}
			blockNumberKey := generateSlotNumberKey(currentSlot)
			err := b.redis.Set(blockNumberKey, true)
			if err != nil {
				panic(err)
			}
		}

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received: getPriorBeaconBlocks")
			return
		default:

		}
		time.Sleep(100 * time.Millisecond)
	}

	utils.Logger.Info("exiting: getPriorBeaconBlocks External")

}

func (b *BeaconBlockRunner) getPriorBeaconBlock(slot int) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	blockNumberKey := generateSlotNumberKey(slot)
	utils.Logger.Infof("Check if prior Beacon block retrieved: %d", slot)
	_, redisErr := b.redis.Get(blockNumberKey)
	if redisErr != nil {
		sideCar := engine.GetBlobSideCars(strconv.Itoa(slot))

		for _, blob := range sideCar.Data {

			b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		}
		blockNumberKey := generateSlotNumberKey(slot)
		err := b.redis.Set(blockNumberKey, true)
		if err != nil {
			panic(err)
		}
	}

	select {
	case <-ctx.Done():
		stop()
		fmt.Println("signal received: getPriorBeaconBlocks")
		return
	default:

	}

	utils.Logger.Info("exiting: getPriorBeaconBlocks External")

}

func generateSlotNumberKey(blockNumber int) string {
	formatInt := strconv.Itoa(blockNumber)
	return "S" + formatInt
}
