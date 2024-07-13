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

// Database settings (insert your own database name and connection URI)
//const dbName = "fiber_test"
//const mongoURI = "mongodb://user:password@localhost:27017/" + dbName

// Employee struct
type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

// Connect configures the MongoDB client and initializes the database connection.
// Source: https://www.mongodb.com/docs/drivers/go/current/quick-start/
//func Connect() error {
//	uri := os.Getenv("MONGO_URI")
//	dbName := os.Getenv("MONGO_DB")
//
//	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri+"/"+dbName))
//
//	defer func() {
//		if err := client.Disconnect(context.TODO()); err != nil {
//			panic(err)
//		}
//	}()
//
//	db := client.Database(dbName)
//
//	if err != nil {
//		return err
//	}
//
//	mg = MongoInstance{
//		Client: client,
//		Db:     db,
//	}
//
//	return nil
//}

func NewMongoInstance() (MongoInstance, error) {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri+"/"+dbName))
	if err != nil {
		return MongoInstance{}, err
	}

	//defer func() {
	//	if err := client.Disconnect(context.TODO()); err != nil {
	//		panic(err)
	//	}
	//}()

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

	// Get all employee records from MongoDB
	// Docs: https://docs.mongodb.com/manual/reference/command/find/
	app.Get("/employee", func(c *fiber.Ctx) error {
		// get all records as a cursor
		query := bson.D{{}}
		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var employees []Employee = make([]Employee, 0)

		// iterate the cursor and decode each item into an Employee
		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())

		}
		// return employees list in JSON format
		return c.JSON(employees)
	})

	// Get once employee records from MongoDB
	// Docs: https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents
	app.Get("/transaction/hash/:hash", func(c *fiber.Ctx) error {
		// get id by params
		_id := c.Params("hash")

		//_id, err := primitive.ObjectIDFromHex(params)
		//if err != nil {
		//	return c.Status(500).SendString(err.Error())
		//}
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

		println(address)
		// transaction receipts use the tx hash as the _id
		filter := bson.D{{Key: "from", Value: address}}

		cursor, err := m.Db.Collection("receipts").Find(c.Context(), filter)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var receipts []types.MongoReceipt = make([]types.MongoReceipt, 0)

		// iterate the cursor and decode each item into an Employee
		if err := cursor.All(c.Context(), &receipts); err != nil {
			return c.Status(500).SendString(err.Error())

		}
		// return employees list in JSON format
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
