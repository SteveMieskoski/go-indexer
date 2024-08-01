package postgres

import (
	"context"
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"src/types"
	"src/utils"
	"strconv"
)

type PgSlotSyncTrackRepository interface {
	Add(appDoc types.PgSlotSyncTrack) (string, error)
	GetBySlotNumber(oId int) (*types.PgSlotSyncTrack, error)
	Update(appDoc types.PgSlotSyncTrack) (*types.PgSlotSyncTrack, error)
	Delete(oId string) (int64, error)
}

type pgSlotSyncTrackRepository struct {
	client       *pgxpool.Pool
	indicesExist bool
}

func NewSlotSyncTrackRepository(client *PostgresDB) PgSlotSyncTrackRepository {
	return &pgSlotSyncTrackRepository{client: client.client, indicesExist: false}
}

func (a *pgSlotSyncTrackRepository) Add(appDoc types.PgSlotSyncTrack) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into pg_slot_sync_tracks ("Hash", "Slot", "Retrieved", "Processed", "BlobsProcessed", "BlobCount")
values ($1, $2, $3, $4, $5, $6);`, appDoc.Hash, appDoc.Slot, appDoc.Retrieved, appDoc.Processed, appDoc.BlobsProcessed, appDoc.BlobCount)
	if err != nil {
		utils.Logger.Errorln(err)
		utils.Logger.Infof("Postgres Error for SLOT: %d\n", appDoc.Slot)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *pgSlotSyncTrackRepository) GetBySlotNumber(num int) (*types.PgSlotSyncTrack, error) {

	var block types.PgSlotSyncTrack

	err := a.client.QueryRow(context.Background(),
		`select * from pg_slot_sync_tracks where "Slot" = $1;`, num).Scan(&block.Id, &block.CreatedAt, &block.UpdatedAt, &block.Hash, &block.Slot, &block.Retrieved, &block.Processed, &block.BlobsProcessed, &block.BlobCount)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
		} else {
			panic(err)

		}
	}

	return &block, nil
}

func (a *pgSlotSyncTrackRepository) Update(appDoc types.PgSlotSyncTrack) (*types.PgSlotSyncTrack, error) {

	var updateString = `
		update pg_block_sync_tracks 
		set "Retrieved" = $1, "Processed" = $2, "BlobsProcessed" = $3
		where "Slot" = $4;`

	_, err := a.client.Exec(context.Background(), updateString, appDoc.Retrieved, appDoc.Processed,
		appDoc.BlobsProcessed, appDoc.Slot)

	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred
		} else {
			panic(err)

		}
	}

	return &appDoc, nil
}

func (a *pgSlotSyncTrackRepository) Delete(id string) (int64, error) {

	_, err := a.client.Exec(context.Background(),
		`delete from pg_block_sync_tracks where "Id" = $1;`, id)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return 0, nil
}
