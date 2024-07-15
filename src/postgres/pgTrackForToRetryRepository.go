package postgres

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"src/types"
	"strconv"
)

type PgTrackForToRetryRepository interface {
	Add(appDoc types.PgTrackForToRetry) (string, error)
	List(count int) ([]*types.PgTrackForToRetry, error)
	GetById(oId string) (*types.PgTrackForToRetry, error)
	GetBySlotNumber(oId int) (*types.PgTrackForToRetry, error)
	Update(appDoc types.PgTrackForToRetry) (*types.PgTrackForToRetry, error)
	Delete(oId string) (int64, error)
}

type pgTrackForToRetryRepository struct {
	client       *gorm.DB
	indicesExist bool
}

func NewTrackForToRetryRepository(client *PostgresDB) PgTrackForToRetryRepository {
	return &pgTrackForToRetryRepository{client: client.client, indicesExist: false}
}

func (a *pgTrackForToRetryRepository) Add(appDoc types.PgTrackForToRetry) (string, error) {

	result := a.client.Clauses(clause.OnConflict{DoNothing: true}).Create(&appDoc)

	if result.Error != nil {
		return "", result.Error
	}

	return strconv.Itoa(int(appDoc.ID)), nil
}

func (a *pgTrackForToRetryRepository) List(count int) ([]*types.PgTrackForToRetry, error) {

	var blocks []*types.PgTrackForToRetry
	result := a.client.Find(&blocks)

	if result.Error != nil {
		return nil, result.Error
	}

	return blocks, nil
}

func (a *pgTrackForToRetryRepository) GetById(id string) (*types.PgTrackForToRetry, error) {

	var block *types.PgTrackForToRetry

	a.client.First(&block, id)

	return block, nil
}

func (a *pgTrackForToRetryRepository) GetBySlotNumber(num int) (*types.PgTrackForToRetry, error) {

	var block *types.PgTrackForToRetry

	a.client.Where("Number <> ?", num).Find(&block)

	return block, nil
}

func (a *pgTrackForToRetryRepository) Update(appDoc types.PgTrackForToRetry) (*types.PgTrackForToRetry, error) {

	a.client.Save(appDoc)

	return &appDoc, nil
}

func (a *pgTrackForToRetryRepository) Delete(id string) (int64, error) {

	a.client.Delete(&types.PgTrackForToRetry{}, id)

	return 0, nil
}
