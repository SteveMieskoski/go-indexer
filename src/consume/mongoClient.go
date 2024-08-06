package consume

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"src/types"
	"src/utils"
)

const uri = "mongodb://localhost:27017"

type db struct {
	client *mongo.Client
}

type DatabaseSetting struct {
	Url        string
	DbName     string
	Collection string
}

func ConnectMongoDb() (*mongo.Client, error) {

	opts := options.Client().ApplyURI(uri) /*.SetServerAPIOptions(serverAPI)*/
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	utils.Logger.Info("Pinged your deployment. You successfully connected to MongoDB!")

	return client, nil
}

func GetClient(setting DatabaseSetting, idxConfig types.IdxConfigStruct) (*mongo.Client, error) {
	//serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(setting.Url) /*.SetServerAPIOptions(serverAPI)*/
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	if idxConfig.ClearConsumer {
		err := client.Database(setting.DbName).Collection(setting.Collection).Drop(context.TODO())
		if err != nil {
			panic(err)
			//return nil, err
		}
	}
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database(setting.DbName).RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	utils.Logger.Info("Pinged your deployment. You successfully connected to MongoDB!")

	return client, nil
}
