package postgres

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"src/types"
	"strconv"
)

type PgSlotSyncTrackRepository interface {
	Add(appDoc types.PgSlotSyncTrack) (string, error)
	List(count int) ([]*types.PgSlotSyncTrack, error)
	GetById(oId string) (*types.PgSlotSyncTrack, error)
	GetBySlotNumber(oId int) (*types.PgSlotSyncTrack, error)
	Update(appDoc types.PgSlotSyncTrack) (*types.PgSlotSyncTrack, error)
	Delete(oId string) (int64, error)
}

type pgSlotSyncTrackRepository struct {
	client       *gorm.DB
	indicesExist bool
}

func NewSlotSyncTrackRepository(client *PostgresDB) PgSlotSyncTrackRepository {
	return &pgSlotSyncTrackRepository{client: client.client, indicesExist: false}
}

func (a *pgSlotSyncTrackRepository) Add(appDoc types.PgSlotSyncTrack) (string, error) {

	result := a.client.Clauses(clause.OnConflict{DoNothing: true}).Create(&appDoc)

	if result.Error != nil {
		return "", result.Error
	}

	return strconv.Itoa(int(appDoc.ID)), nil
}

func (a *pgSlotSyncTrackRepository) List(count int) ([]*types.PgSlotSyncTrack, error) {

	var blocks []*types.PgSlotSyncTrack
	result := a.client.Find(&blocks)

	if result.Error != nil {
		return nil, result.Error
	}

	return blocks, nil
}

func (a *pgSlotSyncTrackRepository) GetById(id string) (*types.PgSlotSyncTrack, error) {

	var block *types.PgSlotSyncTrack

	a.client.First(&block, id)

	return block, nil
}

func (a *pgSlotSyncTrackRepository) GetBySlotNumber(num int) (*types.PgSlotSyncTrack, error) {

	var block *types.PgSlotSyncTrack

	a.client.Where("Number <> ?", num).Find(&block)

	return block, nil
}

func (a *pgSlotSyncTrackRepository) Update(appDoc types.PgSlotSyncTrack) (*types.PgSlotSyncTrack, error) {

	a.client.Save(appDoc)

	return &appDoc, nil
}

func (a *pgSlotSyncTrackRepository) Delete(id string) (int64, error) {

	a.client.Delete(&types.PgSlotSyncTrack{}, id)

	return 0, nil
}
