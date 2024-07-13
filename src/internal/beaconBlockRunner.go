package internal

import (
	"context"
	"fmt"
	"os/signal"
	"src/engine"
	"src/kafka"
	"src/redisdb"
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

func (b *BeaconBlockRunner) getCurrentBeaconBlock() {

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
				completed := b.producerFactory.Produce("Blob", blob)
				if !completed {
					time.Sleep(50 * time.Millisecond)
					err := b.redis.Set("BlobError_"+blob.Index+"_"+currentSlotString, blob.Index)
					if err != nil {
						utils.Logger.Errorf("Error Setting Blob Error in Redis for Blob index: %s, slot number: %s", blob.Index, currentSlotString)
					}
					b.producerFactory.Produce("Blob", blob)
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
		fmt.Println("finished getting blobs for slot %d", currentSlot)

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

				b.producerFactory.Produce("Blob", *blob)
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

			b.producerFactory.Produce("Blob", *blob)
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
