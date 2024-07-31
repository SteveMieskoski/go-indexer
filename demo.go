package main

//type IdxConfig struct {
//	ClearKafka    bool `json:"clear_kafka"`
//	ClearPostgres bool `json:"clear_postgres"`
//	DisableBeacon bool `json:"disable_beacon"`
//	RunAsProducer bool `json:"run_as_producer"`
//}
//
//func (s IdxConfig) String() string {
//	//bytes, _ := json.Marshal(s)
//	bytes, _ := json.MarshalIndent(s, "", "   ")
//	return string(bytes)
//}

func main() {
	//
	//var config IdxConfig
	//
	//app := &cli.App{
	//	Flags: []cli.Flag{
	//		&cli.BoolFlag{
	//			Name:        "clear-kafka",
	//			Value:       false,
	//			Usage:       "thing",
	//			Destination: &config.ClearKafka,
	//		},
	//		&cli.BoolFlag{
	//			Name:        "clear-postgres",
	//			Value:       false,
	//			Usage:       "thing",
	//			Destination: &config.ClearPostgres,
	//		},
	//		&cli.BoolFlag{
	//			Name:        "disable-beacon",
	//			Value:       false,
	//			Usage:       "thing",
	//			Destination: &config.DisableBeacon,
	//		},
	//		&cli.BoolFlag{
	//			Name:        "run-as-producer",
	//			Value:       true,
	//			Usage:       "thing",
	//			Destination: &config.RunAsProducer,
	//		},
	//	},
	//	Action: func(cCtx *cli.Context) error {
	//		//name := "Nefertiti"
	//		//if cCtx.NArg() > 0 {
	//		//	name = cCtx.Args().Get(0)
	//		//}
	//		//config.ClearPostgres = cCtx.Bool("clear.kafka")
	//		//config.ClearKafka = cCtx.Bool("clear.postgres")
	//		//config.DisableBeacon = cCtx.Bool("disable-beacon")
	//		//config.RunAsProducer = cCtx.Bool("run-as-producer")
	//		//if cCtx.String("lang") == "spanish" {
	//		//	fmt.Println("Hola", name)
	//		//} else {
	//		//	fmt.Println("Hello", name)
	//		//}
	//		if cCtx.Bool("clear-kafka") {
	//			fmt.Println("clear.kafka")
	//		} else {
	//			fmt.Println("No clear.kafka")
	//		}
	//		if cCtx.Bool("clear-postgres") {
	//			fmt.Println("clear.postgres")
	//		} else {
	//			fmt.Println("No clear.postgres")
	//		}
	//		if cCtx.Bool("run-as-producer") {
	//			fmt.Println("Do run-as-producer")
	//		} else {
	//			fmt.Println("RUN NOT AS PRODUCER")
	//		}
	//		if cCtx.Bool("disable-beacon") {
	//			fmt.Println("Do disable-beacon")
	//		} else {
	//			clearKafka()
	//			fmt.Println("NO disable-beacon")
	//		}
	//		fmt.Println("Parsing CLI Flags End")
	//		println(config.String())
	//		return nil
	//	},
	//}
	//
	//if err := app.Run(os.Args); err != nil {
	//	log.Fatal(err)
	//}
	//
	//println(config.String())
}

//func clearKafka() {
//	fmt.Println("Clear Kafka")
//}
//
//func clearPostgres() {
//	fmt.Println("Clear Postgres")
//}
