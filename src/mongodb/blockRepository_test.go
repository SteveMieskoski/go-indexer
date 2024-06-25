package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"src/mongodb/models"
	"testing"
)

// const uri = "mongodb://localhost:27017"
var settings = DatabaseSetting{
	Url:        "mongodb://localhost:27017",
	DbName:     "test",
	Collection: "test",
}

func TestNewBlockRepository(t *testing.T) {
	client, err := GetClient(settings)
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("mongodb.Disconnect() error = %v", err)
		}
	}(client, context.TODO())
	if err != nil {
		t.Errorf("Failed to connect to mongoDb")
	}
	type args struct {
		client *mongo.Client
		config *DatabaseSetting
	}
	tests := []struct {
		name string
		args args
		want BlockRepository
	}{
		{
			name: "success",
			args: args{
				client: client,
				config: &DatabaseSetting{
					Url:        "",
					DbName:     "test",
					Collection: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockRepository(tt.args.client, tt.args.config); got == nil {
				t.Errorf("NewBlockRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blockRepository_Add(t *testing.T) {
	client, err := GetClient(settings)
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("mongodb.Disconnect() error = %v", err)
		}
	}(client, context.TODO())
	if err != nil {
		t.Errorf("Failed to connect to mongoDb")
	}
	type fields struct {
		client *mongo.Client
		config *DatabaseSetting
	}
	type args struct {
		appDoc models.Block
		ctx    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: client,
				config: &DatabaseSetting{
					Url:        "",
					DbName:     "test",
					Collection: "test",
				},
			},
			args: args{
				appDoc: models.Block{
					Id:            "123",
					BaseFeePerGas: "",
					BlockNumber:   "",
					Difficulty:    "",
					GasLimit:      "",
					GasUsed:       "",
					Hash:          "",
					Miner:         "",
					ParentHash:    "",
					Timestamp:     0,
					Transactions:  nil,
					Uncles:        nil,
					Withdrawals:   nil,
					Number:        "",
				},
				ctx: context.TODO(),
			},
			want:    "123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &blockRepository{
				client: tt.fields.client,
				config: tt.fields.config,
			}
			got, err := app.Add(tt.args.appDoc, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Add() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blockRepository_GetById(t *testing.T) {
	client, err := GetClient(settings)
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("mongodb.Disconnect() error = %v", err)
		}
	}(client, context.TODO())
	if err != nil {
		t.Errorf("Failed to connect to mongoDb")
	}
	type fields struct {
		client *mongo.Client
		config *DatabaseSetting
	}
	type args struct {
		oId string
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Block
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: client,
				config: &settings,
			},
			args: args{
				oId: "123",
				ctx: context.TODO(),
			},
			want: &models.Block{
				Id:            "123",
				BaseFeePerGas: "",
				BlockNumber:   "",
				Difficulty:    "",
				GasLimit:      "",
				GasUsed:       "",
				Hash:          "",
				Miner:         "",
				ParentHash:    "",
				Timestamp:     0,
				Transactions:  nil,
				Uncles:        nil,
				Withdrawals:   nil,
				Number:        "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &blockRepository{
				client: tt.fields.client,
				config: tt.fields.config,
			}
			got, err := app.GetById(tt.args.oId, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_blockRepository_List(t *testing.T) {
	client, err := GetClient(settings)
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("mongodb.Disconnect() error = %v", err)
		}
	}(client, context.TODO())
	if err != nil {
		t.Errorf("Failed to connect to mongoDb")
	}
	type fields struct {
		client *mongo.Client
		config *DatabaseSetting
	}
	type args struct {
		count int
		ctx   context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*models.Block
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: client,
				config: &settings,
			},
			args: args{
				count: 1,
				ctx:   context.TODO(),
			},
			want: []*models.Block{
				{
					Id:            "123",
					BaseFeePerGas: "",
					BlockNumber:   "",
					Difficulty:    "",
					GasLimit:      "",
					GasUsed:       "",
					Hash:          "",
					Miner:         "",
					ParentHash:    "",
					Timestamp:     0,
					Transactions:  nil,
					Uncles:        nil,
					Withdrawals:   nil,
					Number:        "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &blockRepository{
				client: tt.fields.client,
				config: tt.fields.config,
			}
			got, err := app.List(tt.args.count, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// 6675d042e56a685b510ebeab
func Test_blockRepository_Delete(t *testing.T) {
	client, err := GetClient(settings)
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("mongodb.Disconnect() error = %v", err)
		}
	}(client, context.TODO())
	if err != nil {
		t.Errorf("Failed to connect to mongoDb")
	}
	type fields struct {
		client *mongo.Client
		config *DatabaseSetting
	}
	type args struct {
		oId string
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: client,
				config: &settings,
			},
			args: args{
				oId: "123",
				ctx: context.TODO(),
			},
			want: 1,
		},
		{
			name: "success",
			fields: fields{
				client: client,
				config: &DatabaseSetting{
					Url:        "",
					DbName:     "test",
					Collection: "test",
				},
			},
			args: args{
				oId: "123",
				ctx: context.TODO(),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &blockRepository{
				client: tt.fields.client,
				config: tt.fields.config,
			}
			got, err := app.Delete(tt.args.oId, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Delete() got = %v, want %v", got, tt.want)
			}
		})
	}
}
