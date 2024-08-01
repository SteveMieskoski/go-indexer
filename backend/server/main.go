package main

import (
	"backend/server/mongoQuery"
	"context"
	"github.com/joho/godotenv"
	"src/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		utils.Logger.Fatalf("Error loading .env file")
	}

	mongo, err := mongoQuery.NewMongoInstance()
	if err != nil {
		panic(err)
		return
	}
	defer func() {
		if err := mongo.Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	mongo.MongoTest()
	//// Fiber instance
	//app := fiber.New()
	//
	//// Routes
	//app.Get("/", hello)
	//
	//// Start server
	//log.Fatal(app.Listen(":3000"))
}

// Handler
func hello(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}
