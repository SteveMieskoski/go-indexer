package engine

import (
	"context"
	"fmt"
	"os/signal"
	"src/kafka"
	"src/postgres"
	"src/redisdb"
	"src/types"
	"src/utils"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type BeaconBlockRunner struct {
	priorRetrievalInProgress bool
	priorBeaconBlock         int
	currentBeaconBlock       int
	producerFactory          *kafka.ProducerProvider
	redis                    redisdb.RedisClient
	pgSlotSyncTrack          postgres.PgSlotSyncTrackRepository
	pgRetryTrack             postgres.PgTrackForToRetryRepository
	produceDelay             time.Duration
}

func NewBeaconBlockRunner(producerFactory *kafka.ProducerProvider) BeaconBlockRunner {

	redisClient := redisdb.NewClient()
	pgSlotSyncTrack := postgres.NewSlotSyncTrackRepository(postgres.NewClient())
	pgRetryTrack := postgres.NewTrackForToRetryRepository(postgres.NewClient())

	return BeaconBlockRunner{
		priorBeaconBlock:         0,
		priorRetrievalInProgress: false,
		currentBeaconBlock:       0,
		producerFactory:          producerFactory,
		redis:                    *redisClient,
		pgSlotSyncTrack:          pgSlotSyncTrack,
		produceDelay:             100,
		pgRetryTrack:             pgRetryTrack,
	}
}

func (b *BeaconBlockRunner) Demo() {

}

func (b *BeaconBlockRunner) StartBeaconSync() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	bHeader := GetBeaconHeader()
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

	blockGen := BeaconHeader()

	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest Beacon block: %s", latestBlock.Slot)
		var wg sync.WaitGroup

		sideCar := GetBlobSideCars(latestBlock.Slot)

		slotNum, _ := strconv.Atoi(latestBlock.Slot)
		_, errr := b.pgSlotSyncTrack.Add(types.PgSlotSyncTrack{
			Slot:           int64(slotNum),
			Retrieved:      true,
			Processed:      false,
			BlobsProcessed: false,
			BlobCount:      int64(len(sideCar.Data)),
		})

		if errr != nil {
			utils.Logger.Errorln(errr)
		}
		wg.Add(1)
		go b.processBlobSideCars(latestBlock.Slot, sideCar, &wg)

		select {
		case <-ctx.Done():
			stop()
			fmt.Println("signal received")
			return
		default:
		}
	}
}

func (b *BeaconBlockRunner) processBlobSideCars(slot string, sideCar types.SidecarsResponse, wg *sync.WaitGroup) bool {

	completedOk := true

	for _, blob := range sideCar.Data {
		completed := b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		if !completed {
			if completedOk {
				completedOk = false
			}
			utils.Logger.Errorf("Blob Error for Blob index: %s, slot number: %s", blob.Index, slot)

			_, err := b.pgRetryTrack.Add(types.PgTrackForToRetry{
				DataType: types.BLOB_TOPIC,
				BlockId:  slot,
				RecordId: blob.Index,
			})
			if err != nil {
				utils.Logger.Errorln(err)
			}
		}
	}

	wg.Done()
	return completedOk
}

func (b *BeaconBlockRunner) processBeaconBlock(block *types.BeaconHeadersResponse) {

}

func (b *BeaconBlockRunner) getPriorSlots() {
	_, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	resultChan := make(chan bool)
	defer close(resultChan)

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

	// TODO: See TODO below
	//slotsPerBatch := 10
	//completedOk := true

	for lastSlotRetrieved <= slotNumberOnSyncStart {

		fmt.Printf("Getting Blobs for Prior Slot %d\n", lastSlotRetrieved)

		var wg sync.WaitGroup

		sideCar := GetBlobSideCars(strconv.Itoa(lastSlotRetrieved))

		wg.Add(1)
		_, errr := b.pgSlotSyncTrack.Add(types.PgSlotSyncTrack{
			Slot:           int64(lastSlotRetrieved),
			Retrieved:      true,
			Processed:      false,
			BlobsProcessed: false,
			BlobCount:      int64(len(sideCar.Data)),
		})

		if errr != nil {
			utils.Logger.Errorln(errr)
		}
		go b.processBlobSideCars(strconv.Itoa(lastSlotRetrieved), sideCar, &wg)

		lastSlotRetrieved = lastSlotRetrieved + 1

		wg.Wait()

		// TODO: Revisit sending a number of calls all at once, but need to sync or respond to the
		// TODO: pace of the BlockRunner because when both are trying to commit to Kafka several fail
		//batchEndSlot := lastSlotRetrieved + slotsPerBatch
		//
		//if batchEndSlot >= slotNumberOnSyncStart {
		//	batchEndSlot = slotNumberOnSyncStart
		//}
		//
		//fmt.Printf("Getting Blobs for Prior Slot %d\n", lastSlotRetrieved)
		//
		//var wg sync.WaitGroup
		//
		//for i := lastSlotRetrieved; i < batchEndSlot; i++ {
		//
		//	sideCar := engine.GetBlobSideCars(strconv.Itoa(i))
		//
		//	wg.Add(1)
		//	_, errr := b.pgSlotSyncTrack.Add(types.PgSlotSyncTrack{
		//		Slot:           int64(i),
		//		Retrieved:      true,
		//		Processed:      false,
		//		BlobsProcessed: false,
		//		BlobCount:      int64(len(sideCar.Data)),
		//	})
		//
		//	if errr != nil {
		//		utils.Logger.Errorln(errr)
		//	}
		//	go func() {
		//		completedOk = b.processBlobSideCars(strconv.Itoa(i), sideCar, &wg)
		//	}()
		//
		//	if !completedOk {
		//		slotsPerBatch = 1
		//	} else if slotsPerBatch < 5 {
		//		slotsPerBatch = slotsPerBatch + 1
		//	}
		//}
		//
		//lastSlotRetrieved = batchEndSlot + 1
		//
		//wg.Wait()
	}

	utils.Logger.Info("exiting: getPriorSlot External")
	b.RetryFailedRetrievals()
}

func (b *BeaconBlockRunner) RetryFailedRetrievals() {
	utils.Logger.Info("RetryFailedRetrievals Beacon START")
	blobsToRetry, _ := b.pgRetryTrack.GetByDataType(types.BLOB_TOPIC)
	for _, blobRecord := range blobsToRetry {
		sideCar := GetBlobSideCars(blobRecord.BlockId)
		for _, blob := range sideCar.Data {
			if blob.Index == blobRecord.RecordId {
				completed := b.producerFactory.Produce(types.BLOB_TOPIC, blob)
				if completed {
					_, err := b.pgRetryTrack.Delete(blobRecord.Id)
					if err != nil {
						utils.Logger.Errorln(err)
					}
				}
			}
		}
	}
	utils.Logger.Info("RetryFailedRetrievals Beacon END")
}
