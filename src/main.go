package main

import (
	//"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"src/consume"
	"src/produce"
	"src/types"
	"src/utils"
)

var (
	IdxConfig = types.IdxConfigStruct{}
)

func main() {
	utils.InitializeLogger()
	err := godotenv.Load("../.env")
	//err := godotenv.Load("../.env-remote")
	if err != nil {
		utils.Logger.Fatalf("Error loading .env file")
		err := godotenv.Load("../.env-example")
		if err != nil {
			utils.Logger.Fatalf("Error loading .env-example file")
		}
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
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Running with Config: %s\n\n", IdxConfig.String())

	if IdxConfig.RunAsProducer {
		utils.Logger.Info("starting as producer")
		produce.Run(IdxConfig)
	} else {
		utils.Logger.Info("starting as consumer")
		consume.ConsumerRun(IdxConfig)
	}
}
