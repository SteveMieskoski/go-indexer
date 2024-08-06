package produce

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"src/types"
	"src/utils"
	"strconv"
)

type PgBlockSyncTrackRepository interface {
	Add(appDoc types.PgBlockSyncTrack) (string, error)
	GetByBlockNumber(oId int) (*types.PgBlockSyncTrack, error)
	Update(appDoc types.PgBlockSyncTrack) (*types.PgBlockSyncTrack, error)
	Delete(oId string) (int64, error)
}

type pgBlockSyncTrackRepository struct {
	client       *pgxpool.Pool
	indicesExist bool
}

func NewBlockSyncTrackRepository(client *PostgresDB) PgBlockSyncTrackRepository {
	return &pgBlockSyncTrackRepository{client: client.client, indicesExist: false}
}

func (a *pgBlockSyncTrackRepository) Add(appDoc types.PgBlockSyncTrack) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into pg_block_sync_tracks ("Hash", "Number", "Retrieved", "Processed",
                                  "ReceiptsProcessed", "TransactionsProcessed", "TransactionCount", "ContractsProcessed")
			 values ($1, $2, $3, $4, $5, $6, $7, $8);`,
		appDoc.Hash, appDoc.Number, appDoc.Retrieved, appDoc.Processed,
		appDoc.ReceiptsProcessed, appDoc.TransactionsProcessed,
		appDoc.TransactionCount, appDoc.ContractsProcessed)

	if err != nil {
		utils.Logger.Errorln(err)
		utils.Logger.Infof("Postgres Error for BLOCK: %s\n", appDoc.String())
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *pgBlockSyncTrackRepository) GetByBlockNumber(num int) (*types.PgBlockSyncTrack, error) {

	var block types.PgBlockSyncTrack

	err := a.client.QueryRow(context.Background(),
		`select * from pg_block_sync_tracks where "Number" = $1;`, num).Scan(&block.Id, &block.CreatedAt,
		&block.UpdatedAt, &block.Hash, &block.Number, &block.Retrieved, &block.Processed,
		&block.ReceiptsProcessed, &block.TransactionsProcessed, &block.TransactionCount, &block.ContractsProcessed)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Logger.Infof("No Rows in Get Block Result for block number %d\n", num)
			// there were no rows, but otherwise no error occurred
		} else {
			utils.Logger.Errorln(err)

		}
	}

	return &block, nil
}

func (a *pgBlockSyncTrackRepository) Update(appDoc types.PgBlockSyncTrack) (*types.PgBlockSyncTrack, error) {

	var updateString = `
		update pg_block_sync_tracks 
		set "Retrieved" = $1, "Processed" = $2, "ReceiptsProcessed" = $3, "TransactionsProcessed" = $4,
    		"ContractsProcessed" = $5, "TransactionCount" = $6
		where "Number" = $7;`

	_, err := a.client.Exec(context.Background(), updateString,
		appDoc.Retrieved, appDoc.Processed, appDoc.ReceiptsProcessed,
		appDoc.TransactionsProcessed, appDoc.ContractsProcessed, appDoc.TransactionCount, appDoc.Number)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// there were no rows, but otherwise no error occurred
			utils.Logger.Infof("No Rows in Update Result for block number %d\n", appDoc.Number)
		} else {
			utils.Logger.Errorln(err)

		}
	}

	return &appDoc, nil
}

func (a *pgBlockSyncTrackRepository) Delete(id string) (int64, error) {

	_, err := a.client.Exec(context.Background(),
		`delete from pg_block_sync_tracks where "Id" = $1;`, id)

	if err != nil {
		utils.Logger.Errorln(err)
	}

	return 0, nil
}
