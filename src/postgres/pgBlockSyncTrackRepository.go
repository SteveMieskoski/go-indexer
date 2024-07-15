package postgres

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"src/types"
	"strconv"
)

type PgBlockSyncTrackRepository interface {
	Add(appDoc types.PgBlockSyncTrack) (string, error)
	List(count int) ([]*types.PgBlockSyncTrack, error)
	GetById(oId string) (*types.PgBlockSyncTrack, error)
	GetByBlockNumber(oId int) (*types.PgBlockSyncTrack, error)
	Update(appDoc types.PgBlockSyncTrack) (*types.PgBlockSyncTrack, error)
	Delete(oId string) (int64, error)
}

type pgBlockSyncTrackRepository struct {
	client       *gorm.DB
	indicesExist bool
}

func NewBlockSyncTrackRepository(client *PostgresDB) PgBlockSyncTrackRepository {
	return &pgBlockSyncTrackRepository{client: client.client, indicesExist: false}
}

func (a *pgBlockSyncTrackRepository) Add(appDoc types.PgBlockSyncTrack) (string, error) {

	result := a.client.Clauses(clause.OnConflict{DoNothing: true}).Create(&appDoc)

	if result.Error != nil {
		return "", result.Error
	}

	return strconv.Itoa(int(appDoc.ID)), nil
}

func (a *pgBlockSyncTrackRepository) List(count int) ([]*types.PgBlockSyncTrack, error) {

	var blocks []*types.PgBlockSyncTrack
	result := a.client.Find(&blocks)

	if result.Error != nil {
		return nil, result.Error
	}

	return blocks, nil
}

func (a *pgBlockSyncTrackRepository) GetById(id string) (*types.PgBlockSyncTrack, error) {

	var block *types.PgBlockSyncTrack

	a.client.First(&block, id)

	return block, nil
}

func (a *pgBlockSyncTrackRepository) GetByBlockNumber(num int) (*types.PgBlockSyncTrack, error) {

	var block *types.PgBlockSyncTrack

	a.client.Where("Number <> ?", num).Find(&block)

	return block, nil
}

func (a *pgBlockSyncTrackRepository) Update(appDoc types.PgBlockSyncTrack) (*types.PgBlockSyncTrack, error) {

	a.client.Save(appDoc)

	return &appDoc, nil
}

func (a *pgBlockSyncTrackRepository) Delete(id string) (int64, error) {

	a.client.Delete(&types.PgBlockSyncTrack{}, id)

	return 0, nil
}
