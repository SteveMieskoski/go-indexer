package mongoQuery

import (
	"context"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"src/types"

	"github.com/gofiber/fiber/v2"
)

// MongoInstance contains the Mongo client and database objects
type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

func NewMongoInstance() (MongoInstance, error) {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri+"/"+dbName))
	if err != nil {
		return MongoInstance{}, err
	}

	db := client.Database(dbName)

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return mg, nil
}

func (m MongoInstance) MongoTest() {
	// Connect to the database

	// Create a Fiber app
	app := fiber.New()

	// Docs: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
	app.Get("/transaction/hash/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		println(_id)
		// transaction receipts use the tx hash as the _id
		filter := bson.D{{Key: "_id", Value: _id}}

		var result types.MongoReceipt

		if err := m.Db.Collection("receipts").FindOne(c.Context(), filter).Decode(&result); err != nil {
			println(err)
			return c.Status(500).SendString("Something went wrong.")
		}

		return c.Status(fiber.StatusOK).JSON(result)
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

		cursor, err := m.Db.Collection("receipts").Find(c.Context(), filter, &fOpt)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var receipts []types.MongoReceipt = make([]types.MongoReceipt, 0)

		if err := cursor.All(c.Context(), &receipts); err != nil {
			return c.Status(500).SendString(err.Error())

		}

		return c.JSON(receipts)

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
