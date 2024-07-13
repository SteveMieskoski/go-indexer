package main

import (
	"flag"
	"github.com/joho/godotenv"
	"src/internal"
	"src/utils"
)

var (
	mode = flag.String("mode", "produce", "Mode to run in: \"produce\" to produce, \"consume\" to consume")
)

func main() {
	flag.Parse()
	utils.InitializeLogger()
	err := godotenv.Load("../.env")
	if err != nil {
		utils.Logger.Fatalf("Error loading .env file")
	}

	println(*mode) // todo remove dev item
	if *mode == "consume" {
		utils.Logger.Info("starting as consumer")
		internal.ConsumerRun()
		// []string{"Block", "Receipt", "Blob"}
		//kafka.NewMongoDbConsumer()
	}

	if *mode == "produce" {
		utils.Logger.Info("starting as producer")
		internal.Run()
	}
}
