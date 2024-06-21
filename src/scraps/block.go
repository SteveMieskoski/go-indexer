package scraps

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

type BlockRepository interface {
	Add(appDoc models.Block, ctx context.Context) (primitive.ObjectID, error)
	List(count int, ctx context.Context) ([]*models.Block, error)
	GetById(oId primitive.ObjectID, ctx context.Context) (*models.Block, error)
	Delete(oId primitive.ObjectID, ctx context.Context) (int64, error)
}

type blockRepository struct {
	client *mongo.Client
	config *mongodb.DatabaseSetting
}

func NewBlockRepository(client *mongo.Client, config *mongodb.DatabaseSetting) BlockRepository {
	return &blockRepository{client: client, config: config}
}

func (app *blockRepository) Add(appDoc models.Block, ctx context.Context) (primitive.ObjectID, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	insertResult, err := collection.InsertOne(ctx, appDoc)

	println("check")
	if !errors.Is(err, mongo.ErrNilCursor) {
		return primitive.NilObjectID, err
	}

	if oidResult, ok := insertResult.InsertedID.(primitive.ObjectID); !ok {
		return primitive.NilObjectID, err
	} else {
		return oidResult, nil
	}

}

func (app *blockRepository) List(count int, ctx context.Context) ([]*models.Block, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	//logrus.Infof("FindOptions %d, DbName %s, Url %s", count, app.config.DbName, app.config.Url)

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*models.Block
	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {
		// create a value into which the single document can be decoded
		var elem models.Block
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

func (app *blockRepository) GetById(oId primitive.ObjectID, ctx context.Context) (*models.Block, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *models.Block

	collection.FindOne(ctx, filter).Decode(&appDoc)

	return appDoc, nil
}

func (app *blockRepository) Delete(oId primitive.ObjectID, ctx context.Context) (int64, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)
	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return 0, bson.ErrDecodeToNil
	}

	return result.DeletedCount, nil
}
