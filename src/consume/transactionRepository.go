package consume

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"src/types"
	"src/utils"
)

type TransactionRepository interface {
	Add(appDoc types.MongoTransaction, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*types.MongoTransaction, error)
	GetById(oId string, ctx context.Context) (*types.MongoTransaction, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type transactionRepository struct {
	client       *mongo.Client
	config       *DatabaseSetting
	indicesExist bool
}

func NewTransactionRepository(client *mongo.Client, config *DatabaseSetting) TransactionRepository {
	return &transactionRepository{client: client, config: config, indicesExist: false}
}

func (app *transactionRepository) AddIndex() (string, error) {

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

// TODO: identify fields that need to be indexed
func (app *transactionRepository) Add(appDoc types.MongoTransaction, ctx context.Context) (string, error) {
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

	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			utils.Logger.Fatal(err)
			return "", err
		} else {
			return "-3", nil
		}
	}

	typeCheck := reflect.ValueOf(insertResult.InsertedID)
	if typeCheck.IsValid() {
		if oidResult, ok := insertResult.InsertedID.(string); !ok {
			return "-2", err
		} else {
			return oidResult, nil
		}
	}
	return "0", nil
}

func (app *transactionRepository) List(count int, ctx context.Context) ([]*types.MongoTransaction, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var appDocs []*types.MongoTransaction

	for cursor.Next(ctx) {
		var elem types.MongoTransaction
		if err := cursor.Decode(&elem); err != nil {
			utils.Logger.Fatal(err)
			return nil, err
		}

		appDocs = append(appDocs, &elem)
	}

	utils.Logger.Info("Document Count:", len(appDocs))
	return appDocs, nil
}

func (app *transactionRepository) GetById(oId string, ctx context.Context) (*types.MongoTransaction, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *types.MongoTransaction

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
