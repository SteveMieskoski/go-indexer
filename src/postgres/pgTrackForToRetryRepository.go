package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"src/types"
	"src/utils"
	"strconv"
)

type PgTrackForToRetryRepository interface {
	Add(appDoc types.PgTrackForToRetry) (string, error)
	GetByDataType(dt string) ([]types.PgTrackForToRetry, error)
	Delete(oId uint) (int64, error)
}

type pgTrackForToRetryRepository struct {
	client       *pgxpool.Pool
	indicesExist bool
}

func NewTrackForToRetryRepository(client *PostgresDB) PgTrackForToRetryRepository {
	return &pgTrackForToRetryRepository{client: client.client, indicesExist: false}
}

func (a *pgTrackForToRetryRepository) Add(appDoc types.PgTrackForToRetry) (string, error) {

	_, err := a.client.Exec(context.Background(),
		`insert into pg_track_for_to_retries ("DataType", "BlockId", "RecordId")
			values ($1, $2, $3);`, appDoc.DataType, appDoc.BlockId, appDoc.RecordId)
	if err != nil {
		utils.Logger.Errorln(err)
	}

	return strconv.Itoa(int(appDoc.Id)), nil
}

func (a *pgTrackForToRetryRepository) GetByDataType(dt string) ([]types.PgTrackForToRetry, error) {

	println(dt)

	rows, err := a.client.Query(context.Background(),
		`select * from pg_track_for_to_retries where "DataType" = $1;`, dt)
	blocks, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.PgTrackForToRetry])
	if err != nil {
		utils.Logger.Errorln(err)
	}
	return blocks, nil
}

func (a *pgTrackForToRetryRepository) Delete(id uint) (int64, error) {

	_, err := a.client.Exec(context.Background(),
		`delete from pg_track_for_to_retries where "Id" = $1;`, id)
	if err != nil {
		utils.Logger.Errorln(err)
	}
	return 0, nil
}
