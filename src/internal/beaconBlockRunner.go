package internal

import (
	"context"
	"fmt"
	"os/signal"
	"src/engine"
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
	blobsToRetry             map[string][]string
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
		blobsToRetry:             make(map[string][]string),
		pgRetryTrack:             pgRetryTrack,
	}
}

func (b *BeaconBlockRunner) Demo() {

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

	//resultChan := make(chan bool)
	//defer close(resultChan)

	fmt.Printf("producer ready: %t\n", b.producerFactory.Connected)

	blockGen := engine.BeaconHeader()

	for latestBlock := range blockGen {
		utils.Logger.Infof("Recieved latest Beacon block: %s", latestBlock.Slot)
		var wg sync.WaitGroup

		sideCar := engine.GetBlobSideCars(latestBlock.Slot)

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

func (b *BeaconBlockRunner) processBlobSideCars(slot string, sideCar types.SidecarsResponse, wg *sync.WaitGroup) {

	for _, blob := range sideCar.Data {
		completed := b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		if !completed {
			utils.Logger.Errorf("Blob Error for Blob index: %s, slot number: %s", blob.Index, slot)
			//b.blobsToRetry[slot] = append(b.blobsToRetry[slot], blob.Index)
			_, err := b.pgRetryTrack.Add(types.PgTrackForToRetry{
				DataType: "Blob",
				BlockId:  slot,
				RecordId: blob.Index,
			})
			if err != nil {
				utils.Logger.Errorln(err)
			}
			//err := b.redis.Set("BlobError_"+blob.Index+"_"+slot, blob.Index)
			//if err != nil {
			//	utils.Logger.Errorf("Error Setting Blob Error in Redis for Blob index: %s, slot number: %s", blob.Index, slot)
			//}
			//b.producerFactory.Produce(types.BLOB_TOPIC, *blob)
		}
	}

	wg.Done()
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

	for lastSlotRetrieved <= slotNumberOnSyncStart {
		fmt.Printf("Getting Blobs for Prior Slot %d\n", lastSlotRetrieved)

		var wg sync.WaitGroup
		sideCar := engine.GetBlobSideCars(strconv.Itoa(lastSlotRetrieved))

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

		//select {
		//case produceOk := <-resultChan:
		//	if produceOk {
		//		if b.produceDelay >= 200 {
		//			b.produceDelay = b.produceDelay - 100
		//		}
		//	} else {
		//		b.produceDelay = b.produceDelay + 100
		//	}
		//case <-ctx.Done():
		//	stop()
		//	fmt.Println("signal received")
		//	return
		//default:
		//}

		lastSlotRetrieved = lastSlotRetrieved + 1
		//time.Sleep(b.produceDelay * time.Millisecond)
		println("WAITING ===================================================")
		wg.Wait()
	}

	utils.Logger.Info("exiting: getPriorSlot External")
}
