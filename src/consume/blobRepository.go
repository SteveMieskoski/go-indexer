package consume

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"src/types"
	"src/utils"
)

type BlobRepository interface {
	Add(appDoc types.MongoBlob, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*types.MongoBlob, error)
	GetById(oId string, ctx context.Context) (*types.MongoBlob, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type blobRepository struct {
	client       *mongo.Client
	config       *DatabaseSetting
	indicesExist bool
}

func NewBlobRepository(client *mongo.Client, config *DatabaseSetting) BlobRepository {
	return &blobRepository{client: client, config: config, indicesExist: false}
}

// TODO: identify a unique property to use as the id
func (app *blobRepository) Add(appDoc types.MongoBlob, ctx context.Context) (string, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	insertResult, err := collection.InsertOne(ctx, appDoc)

	if errors.Is(err, mongo.ErrNilCursor) {
		return "-1", err
	}

	if oidResult, ok := insertResult.InsertedID.(string); !ok {
		return "-2", err
	} else {
		return oidResult, nil
	}

}

func (app *blobRepository) List(count int, ctx context.Context) ([]*types.MongoBlob, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*types.MongoBlob

	for cursor.Next(ctx) {
		var elem types.MongoBlob
		if err := cursor.Decode(&elem); err != nil {
			utils.Logger.Fatal(err)
			return nil, err
		}

		appDocs = append(appDocs, &elem)
	}

	cursor.Close(ctx)

	utils.Logger.Info("Document Count:", len(appDocs))
	return appDocs, nil
}

func (app *blobRepository) GetById(oId string, ctx context.Context) (*types.MongoBlob, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *types.MongoBlob

	collection.FindOne(ctx, filter).Decode(&appDoc)

	return appDoc, nil
}

func (app *blobRepository) Delete(oId string, ctx context.Context) (int64, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)
	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return 0, bson.ErrDecodeToNil
	}

	return result.DeletedCount, nil
}
