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

func NewBeaconBlockRunner(redis redisdb.RedisClient, producerFactory *kafka.ProducerProvider) BeaconBlockRunner {

	return BeaconBlockRunner{
		priorBeaconBlock:         0,
		priorRetrievalInProgress: false,
		currentBeaconBlock:       0,
		producerFactory:          producerFactory,
		redis:                    redis,
	}

}

func (b *BeaconBlockRunner) getCurrentBeaconBlock() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	for {
		bHeader := engine.GetBeaconHeader()

		dataLen := len(bHeader.Data)
		if dataLen > 0 {
			println("OK")
		}

		println("HERE 1")
		currentSlotString := bHeader.Data[0].Header.Message.Slot
		currentSlot, _ := strconv.Atoi(currentSlotString)

		println("HERE 2")

		if currentSlot > b.currentBeaconBlock {
			sideCar := engine.GetBlobSideCars(strconv.Itoa(currentSlot))

			for _, blob := range sideCar.Data {
				b.producerFactory.Produce("Blob", blob)
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
			utils.Logger.Infof("currentSlot-b.currentBeaconBlock = %d", currentSlot-b.currentBeaconBlock)
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

		time.Sleep(5000 * time.Millisecond)
	}

}

func (b *BeaconBlockRunner) getPriorBeaconBlocks(currentSlot int, skippedSlotCount int) {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	utils.Logger.Infof("currentSlot - skippedSlotCount = %d", currentSlot-skippedSlotCount)
	for i := currentSlot - skippedSlotCount; i < currentSlot; i++ {
		blockNumberKey := generateSlotNumberKey(i)
		utils.Logger.Infof("Check if prior beacon block retrieved: %d", i)
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

func generateSlotNumberKey(blockNumber int) string {
	formatInt := strconv.Itoa(blockNumber)
	return "S" + formatInt
}
