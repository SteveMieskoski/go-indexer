package main

import (
	"flag"
	"src/internal"
	"src/kafka"
	"src/utils"
)

var (
	mode = flag.String("mode", "produce", "Mode to run in: \"produce\" to produce, \"consume\" to consume")
)

func main() {
	flag.Parse()
	utils.InitializeLogger()

	println(*mode) // todo remove dev item
	if *mode == "consume" {
		utils.Logger.Info("starting as consumer")
		kafka.NewConsumer([]string{"Block", "Receipt"})
	}

	if *mode == "produce" {
		utils.Logger.Info("starting as producer")
		internal.Run()
	}
}
