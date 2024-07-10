package postgres

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"src/types"
	"strconv"
)

type AddressRepository interface {
	Add(appDoc types.Address, ctx context.Context) (string, error)
	List(count int, ctx context.Context) ([]*types.Address, error)
	GetById(oId string, ctx context.Context) (*types.Address, error)
	Delete(oId string, ctx context.Context) (int64, error)
}

type addressRepository struct {
	client       *gorm.DB
	indicesExist bool
}

func NewAddressRepository(client *PostgresDB) AddressRepository {
	return &addressRepository{client: client.client, indicesExist: false}
}

func (a *addressRepository) Add(appDoc types.Address, ctx context.Context) (string, error) {

	result := a.client.Clauses(clause.OnConflict{DoNothing: true}).Create(&appDoc)

	if result.Error != nil {
		return "", result.Error
	}

	return strconv.Itoa(int(appDoc.ID)), nil
}

func (a *addressRepository) List(count int, ctx context.Context) ([]*types.Address, error) {

	var addresses []*types.Address
	result := a.client.Find(&addresses)

	if result.Error != nil {
		return nil, result.Error
	}

	return addresses, nil
}

func (a *addressRepository) GetById(id string, ctx context.Context) (*types.Address, error) {

	var address *types.Address

	a.client.First(&address, id)

	return address, nil
}

func (a *addressRepository) Delete(id string, ctx context.Context) (int64, error) {

	a.client.Delete(&types.Address{}, id)

	return 0, nil
}
