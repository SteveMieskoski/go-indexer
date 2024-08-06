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

// TODO: need to use a postgres table or other intermediary to extract contract code from the tx
// TODO: that created the contract and is reported in the receipt
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

	if errors.Is(err, mongo.ErrNilCursor) {
		return "-1", err
	}

	typeCheck := reflect.ValueOf(insertResult.InsertedID)
	if typeCheck.IsValid() {
		if oidResult, ok := insertResult.InsertedID.(string); !ok {
			return "-2", err
		} else {
			return oidResult, nil
		}
	}

	utils.Logger.Error("receiptRepository.go:84", "INVALID TYPE CHECK RECEIPT REPOSITORY")
	return "0", nil

}

func (app *receiptRepository) List(count int, ctx context.Context) ([]*types.MongoReceipt, error) {

	findOptions := options.Find()
	findOptions.SetLimit(int64(count))

	collection := app.client.Database(app.config.DbName).Collection(app.config.Collection)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}

	var appDocs []*types.MongoReceipt

	for cursor.Next(ctx) {
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
