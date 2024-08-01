package main

import (
	//"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"src/internal"
	"src/types"
	"src/utils"
)

var (
	//mode      = flag.String("mode", "produce", "Mode to run in: \"produce\" to produce, \"consume\" to consume")
	IdxConfig = types.IdxConfigStruct{}
)

func main() {
	//flag.Parse()
	utils.InitializeLogger()
	err := godotenv.Load("../.env")
	//err := godotenv.Load("../.env-remote")
	if err != nil {
		utils.Logger.Fatalf("Error loading .env file")
	}

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "clear-kafka",
				Value:       false,
				Usage:       "thing",
				Destination: &IdxConfig.ClearKafka,
			},
			&cli.BoolFlag{
				Name:        "clear-postgres",
				Value:       false,
				Usage:       "thing",
				Destination: &IdxConfig.ClearPostgres,
			},
			&cli.BoolFlag{
				Name:        "clear-redis",
				Value:       false,
				Usage:       "thing",
				Destination: &IdxConfig.ClearRedis,
			},
			&cli.BoolFlag{
				Name:        "clear-consumer",
				Value:       false,
				Usage:       "thing",
				Destination: &IdxConfig.ClearConsumer,
			},
			&cli.BoolFlag{
				Name:        "disable-beacon",
				Value:       false,
				Usage:       "thing",
				Destination: &IdxConfig.DisableBeacon,
			},
			&cli.BoolFlag{
				Name:  "full-reset",
				Value: false,
				Usage: "thing",
			},
			&cli.BoolFlag{
				Name:        "run-as-producer",
				Value:       true,
				Usage:       "thing",
				Destination: &IdxConfig.RunAsProducer,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Bool("full-reset") {
				fmt.Println("Do run-as-producer")
				IdxConfig.ClearKafka = true
				IdxConfig.ClearPostgres = true
				IdxConfig.ClearRedis = true
			}
			//if cCtx.Bool("clear-kafka") {
			//	fmt.Println("clear.kafka")
			//} else {
			//	fmt.Println("No clear.kafka")
			//}
			//if cCtx.Bool("clear-postgres") {
			//	fmt.Println("clear.postgres")
			//} else {
			//	fmt.Println("No clear.postgres")
			//}
			//if cCtx.Bool("run-as-producer") {
			//	fmt.Println("Do run-as-producer")
			//} else {
			//	fmt.Println("RUN NOT AS PRODUCER")
			//}
			//if cCtx.Bool("disable-beacon") {
			//	fmt.Println("Do disable-beacon")
			//} else {
			//	fmt.Println("NO disable-beacon")
			//}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Running with Config: %s\n\n", IdxConfig.String())

	if IdxConfig.RunAsProducer {
		utils.Logger.Info("starting as producer")
		internal.Run(IdxConfig)
	} else {
		utils.Logger.Info("starting as consumer")
		internal.ConsumerRun(IdxConfig)
	}

	//println(*mode) // todo remove dev item
	//if *mode == "consume" {
	//	utils.Logger.Info("starting as consumer")
	//	internal.ConsumerRun()
	//	// []string{"Block", "Receipt", "Blob"}
	//	//kafka.NewMongoDbConsumer()
	//}
	//
	//if *mode == "produce" {
	//	utils.Logger.Info("starting as producer")
	//	internal.Run()
	//}
}
