package mongodb

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

type LogRepository interface {
	Add(appDoc types.MongoLog, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*types.MongoLog, error)
	GetById(oId string, ctx context.Context) (*types.MongoLog, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type logRepository struct {
	client       *mongo.Client
	config       *DatabaseSetting
	indicesExist bool
}

func NewLogRepository(client *mongo.Client, config *DatabaseSetting) LogRepository {
	return &logRepository{client: client, config: config, indicesExist: false}
}

func (app *logRepository) AddIndex() (string, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	indexModel := mongo.IndexModel{
		Keys: bson.D{{"title", 1}},
	}
	name, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	app.indicesExist = true

	return name, nil
}

func (app *logRepository) Add(appDoc types.MongoLog, ctx context.Context) (string, error) {
	//if !app.indicesExist {
	//	_, err := app.AddIndex()
	//	if err != nil {
	//		return "", err
	//	}
	//	app.indicesExist = true
	//}

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

func (app *logRepository) List(count int, ctx context.Context) ([]*types.MongoLog, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*types.MongoLog

	for cursor.Next(ctx) {
		var elem types.MongoLog
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

func (app *logRepository) GetById(oId string, ctx context.Context) (*types.MongoLog, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *types.MongoLog

	collection.FindOne(ctx, filter).Decode(&appDoc)

	return appDoc, nil
}

func (app *logRepository) Delete(oId string, ctx context.Context) (int64, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)
	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return 0, bson.ErrDecodeToNil
	}

	return result.DeletedCount, nil
}
