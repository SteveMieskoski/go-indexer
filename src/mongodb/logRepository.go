package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"src/mongodb/models"
	"src/utils"
)

type LogRepository interface {
	Add(appDoc models.Log, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*models.Log, error)
	GetById(oId string, ctx context.Context) (*models.Log, error)
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

func (app *logRepository) Add(appDoc models.Log, ctx context.Context) (string, error) {
	//if !app.indicesExist {
	//	_, err := app.AddIndex()
	//	if err != nil {
	//		return "", err
	//	}
	//	app.indicesExist = true
	//}

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	insertResult, err := collection.InsertOne(ctx, appDoc)

	utils.Logger.Info("LogRepository - ErrNilCursor Check")
	if errors.Is(err, mongo.ErrNilCursor) {
		return "-1", err
	}

	utils.Logger.Info("LogRepository - Get Inserted Document _Id Check")
	if oidResult, ok := insertResult.InsertedID.(string); !ok {
		return "-2", err
	} else {
		return oidResult, nil
	}

}

func (app *logRepository) List(count int, ctx context.Context) ([]*models.Log, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*models.Log
	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {
		// create a value into which the single document can be decoded
		var elem models.Log
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

func (app *logRepository) GetById(oId string, ctx context.Context) (*models.Log, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *models.Log

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
