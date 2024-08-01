package mongoQuery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"src/types"
)

// DbInstance contains the Mongo client and database objects
type DbInstance struct {
	Client   *mongo.Client
	Db       *mongo.Database
	PgClient *pgxpool.Pool
}

var mg DbInstance

func NewMongoInstance() (DbInstance, error) {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri+"/"+dbName))
	if err != nil {
		return DbInstance{}, err
	}

	db := client.Database(dbName)

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("RAW_GO_POSTGRES_STRING"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	//go func() {
	//	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//	defer stop()
	//	defer dbpool.Close()
	//
	//	select {
	//	case <-ctx.Done():
	//		fmt.Println("signal received, Closing DB connection...")
	//		stop()
	//		dbpool.Close()
	//		return
	//	}
	//}()

	mg = DbInstance{
		Client:   client,
		Db:       db,
		PgClient: dbpool,
	}

	return mg, nil
}

type searchResponse struct {
	HashType string `json:"hashType"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func (m DbInstance) MongoTest() {
	// Connect to the database

	// Create a Fiber app
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/search/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		println(_id)
		// transaction receipts use the tx hash as the _id
		txFilter := bson.D{{Key: "hash", Value: _id}}

		var result types.MongoReceipt

		var resultType searchResponse

		if err := m.Db.Collection("transactions").FindOne(c.Context(), txFilter).Decode(&result); err != nil {
			println(err)
			blockFilter := bson.D{{Key: "_id", Value: _id}}
			var block types.MongoBlock

			if err := m.Db.Collection("blocks").FindOne(c.Context(), blockFilter).Decode(&block); err != nil {
				println(err)
				address := types.Address{}
				// TODO: Fix Ingestion to define nonce and isContract on all instances
				row := m.PgClient.QueryRow(context.Background(),
					`SELECT "Address", "Balance" FROM indexer.public.addresses WHERE "Address" = $1`, _id)

				if err := row.Scan(&address.Address, &address.Balance); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						// there were no rows, but otherwise no error occurred
						return c.Status(500).JSON(errorResponse{Message: "Not Found"})
					} else {
						return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})

					}
				}

				resultType = searchResponse{HashType: "address"}

			} else {
				resultType = searchResponse{HashType: "block"}

			}
		} else {
			resultType = searchResponse{HashType: "transaction"}
		}

		return c.Status(fiber.StatusOK).JSON(resultType)
	})

	// Docs: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
	app.Get("/transaction/hash/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		println(_id)
		// transaction receipts use the tx hash as the _id
		filter := bson.D{{Key: "hash", Value: _id}}

		var result types.MongoReceipt

		if err := m.Db.Collection("transactions").FindOne(c.Context(), filter).Decode(&result); err != nil {
			println(err)
			return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})
		}

		return c.Status(fiber.StatusOK).JSON(result)
	})

	app.Get("/receipt/hash/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		println(_id)
		// transaction receipts use the tx hash as the _id
		filter := bson.D{{Key: "_id", Value: _id}}

		var result types.MongoReceipt

		if err := m.Db.Collection("receipts").FindOne(c.Context(), filter).Decode(&result); err != nil {
			println(err)
			return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})
		}

		return c.Status(fiber.StatusOK).JSON(result)
	})

	app.Get("/log/hash/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		println(_id)
		// transaction receipts use the tx hash as the _id
		filter := bson.D{{Key: "transactionHash", Value: _id}}

		//var result types.MongoLog

		fOpt := options.FindOptions{}
		fOpt.SetSort(bson.D{{"logIndex", 1}})

		cursor, err := m.Db.Collection("logs").Find(c.Context(), filter, &fOpt)
		if err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})
		}

		var logs []types.MongoLog = make([]types.MongoLog, 0)

		if err := cursor.All(c.Context(), &logs); err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})

		}

		return c.JSON(logs)

		//if err := m.Db.Collection("logs").FindOne(c.Context(), filter).Decode(&result); err != nil {
		//	println(err)
		//	return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})
		//}
		//
		//return c.Status(fiber.StatusOK).JSON(result)
	})

	app.Get("/block/hash/:hash", func(c *fiber.Ctx) error {
		_id := c.Params("hash")
		filter := bson.D{{Key: "_id", Value: _id}}

		var block types.MongoBlock

		if err := m.Db.Collection("blocks").FindOne(c.Context(), filter).Decode(&block); err != nil {
			println(err)
			return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})
		}

		return c.Status(fiber.StatusOK).JSON(block)

	})

	app.Get("/address/:address", func(c *fiber.Ctx) error {

		addressParam := c.Params("address")
		address := types.Address{}
		// TODO: Fix Ingestion to define nonce and isContract on all instances
		row := m.PgClient.QueryRow(context.Background(),
			`SELECT "Address", "Balance" FROM indexer.public.addresses WHERE "Address" = $1`, addressParam)

		if err := row.Scan(&address.Address, &address.Balance); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// there were no rows, but otherwise no error occurred
				return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})
			} else {
				return c.Status(500).JSON(errorResponse{Message: "Something went wrong."})

			}
		}
		return c.Status(fiber.StatusOK).JSON(address)
	})

	app.Get("/transactions/address/:address", func(c *fiber.Ctx) error {
		// get id by params
		address := c.Params("address")
		limit := 20
		page := 1
		println(address)
		filter := bson.D{{Key: "from", Value: address}}

		l := int64(limit)
		skip := int64(page*limit - limit)
		fOpt := options.FindOptions{Limit: &l, Skip: &skip}
		fOpt.SetSort(bson.D{{"blockNumber", -1}})

		cursor, err := m.Db.Collection("receipts").Find(c.Context(), filter, &fOpt)
		if err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})
		}

		var receipts []types.MongoReceipt = make([]types.MongoReceipt, 0)

		if err := cursor.All(c.Context(), &receipts); err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})

		}

		return c.JSON(receipts)

	})

	app.Get("/recentblocks", func(c *fiber.Ctx) error {
		// get id by params
		limit := 5
		filter := bson.D{}

		l := int64(limit)
		skip := int64(0)
		fOpt := options.FindOptions{Limit: &l, Skip: &skip}
		fOpt.SetSort(bson.D{{"number", -1}})

		cursor, err := m.Db.Collection("blocks").Find(c.Context(), filter, &fOpt)
		if err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})
		}

		var blocks []types.MongoBlock = make([]types.MongoBlock, 0)

		if err := cursor.All(c.Context(), &blocks); err != nil {
			return c.Status(500).JSON(errorResponse{Message: err.Error()})

		}

		return c.JSON(blocks)

	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if c.Get("host") == "localhost:3000" {
			c.Locals("Host", "Localhost:3000")
			return c.Next()
		}
		return c.Status(403).SendString("Request origin not allowed")
	})

	// Upgraded websocket request
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		fmt.Println(c.Locals("Host")) // "Localhost:3000"
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)
			err = c.WriteMessage(mt, msg)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))

	//app.Get("/transaction/logs/:hash", func(c *fiber.Ctx) error {
	//	// get id by params
	//	hash := c.Params("hash")
	//
	//	println(hash)
	//	// transaction receipts use the tx hash as the _id
	//	filter := bson.D{{Key: "from", Value: hash}}
	//
	//	cursor, err := m.Db.Collection("receipts").Find(c.Context(), filter)
	//	if err != nil {
	//		return c.Status(500).SendString(err.Error())
	//	}
	//
	//	var receipts []types.MongoReceipt = make([]types.MongoReceipt, 0)
	//
	//	// iterate the cursor and decode each item into an Employee
	//	if err := cursor.All(c.Context(), &receipts); err != nil {
	//		return c.Status(500).SendString(err.Error())
	//
	//	}
	//	// return employees list in JSON format
	//	return c.JSON(receipts)
	//
	//})

	log.Fatal(app.Listen(":3000"))
}
