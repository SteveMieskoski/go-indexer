package documents

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"src/mongodb"
	"src/mongodb/models"
)

//var collectionName = "blocks"

type TransactionRepository interface {
	Add(appDoc models.Transaction, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*models.Transaction, error)
	GetById(oId string, ctx context.Context) (*models.Transaction, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type transactionRepository struct {
	client *mongo.Client
	config *mongodb.DatabaseSetting
}

func NewTransactionRepository(client *mongo.Client, config *mongodb.DatabaseSetting) TransactionRepository {
	return &transactionRepository{client: client, config: config}
}

func (app *transactionRepository) Add(appDoc models.Transaction, ctx context.Context) (string, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	insertResult, err := collection.InsertOne(ctx, appDoc)

	//println("Rcheck") // TODO: use an actual logging library
	if errors.Is(err, mongo.ErrNilCursor) {
		return "-1", err
	}

	//println("Rcheck2") // TODO: use an actual logging library
	if oidResult, ok := insertResult.InsertedID.(string); !ok {
		return "-2", err
	} else {
		return oidResult, nil
	}

}

func (app *transactionRepository) List(count int, ctx context.Context) ([]*models.Transaction, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	//logrus.Infof("FindOptions %d, DbName %s, Url %s", count, app.config.DbName, app.config.Url)

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*models.Transaction
	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {
		// create a value into which the single document can be decoded
		var elem models.Transaction
		if err := cursor.Decode(&elem); err != nil {
			logrus.Fatal(err)
			return nil, err
		}

		appDocs = append(appDocs, &elem)
	}

	cursor.Close(ctx)

	//logrus.Infof("AppDocs Count:", len(appDocs))
	return appDocs, nil
}

func (app *transactionRepository) GetById(oId string, ctx context.Context) (*models.Transaction, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *models.Transaction

	collection.FindOne(ctx, filter).Decode(&appDoc)

	return appDoc, nil
}

func (app *transactionRepository) Delete(oId string, ctx context.Context) (int64, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)
	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return 0, bson.ErrDecodeToNil
	}

	return result.DeletedCount, nil
}
