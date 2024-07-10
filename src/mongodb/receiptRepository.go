package mongodb

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

type ReceiptRepository interface {
	Add(appDoc types.MongoReceipt, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*types.MongoReceipt, error)
	GetById(oId string, ctx context.Context) (*types.MongoReceipt, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type receiptRepository struct {
	client       *mongo.Client
	config       *DatabaseSetting
	indicesExist bool
}

func NewReceiptRepository(client *mongo.Client, config *DatabaseSetting) ReceiptRepository {
	return &receiptRepository{client: client, config: config, indicesExist: false}
}

func (app *receiptRepository) AddIndex() (string, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	indexModel := mongo.IndexModel{
		Keys: bson.D{{"from", 1}},
	}
	name, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	app.indicesExist = true

	return name, nil
}

func (app *receiptRepository) Add(appDoc types.MongoReceipt, ctx context.Context) (string, error) {

	//if !app.indicesExist {
	//	_, err := app.AddIndex()
	//	if err != nil {
	//
	//		return "", err
	//	}
	//	app.indicesExist = true
	//}
	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	insertResult, err := collection.InsertOne(ctx, appDoc)

	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			utils.Logger.Fatal(err)
			return "", err
		} else {
			return "-3", nil
		}
	}

	utils.Logger.Info("ReceiptRepository - ErrNilCursor Check")
	if errors.Is(err, mongo.ErrNilCursor) {
		return "-1", err
	}

	utils.Logger.Info("ReceiptRepository - Get Inserted Document _Id Check")
	typeCheck := reflect.ValueOf(insertResult.InsertedID)
	if typeCheck.IsValid() {
		if oidResult, ok := insertResult.InsertedID.(string); !ok {
			return "-2", err
		} else {
			return oidResult, nil
		}
	}

	utils.Logger.Error("receiptRepository.go:84", "INVALID TYPE CHECK RECEIPT REPOSITORY") // todo remove dev item
	return "0", nil

}

func (app *receiptRepository) List(count int, ctx context.Context) ([]*types.MongoReceipt, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	// TODO: look at using below logging library
	//logrus.Infof("FindOptions %d, DbName %s, Url %s", count, app.config.DbName, app.config.Url)

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*types.MongoReceipt
	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cursor.Next(ctx) {
		// create a value into which the single document can be decoded
		var elem types.MongoReceipt
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

func (app *receiptRepository) GetById(oId string, ctx context.Context) (*types.MongoReceipt, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	var appDoc *types.MongoReceipt

	collection.FindOne(ctx, filter).Decode(&appDoc)

	return appDoc, nil
}

func (app *receiptRepository) Delete(oId string, ctx context.Context) (int64, error) {

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)
	filter := bson.D{primitive.E{Key: "_id", Value: oId}}

	result, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return 0, bson.ErrDecodeToNil
	}

	return result.DeletedCount, nil
}
