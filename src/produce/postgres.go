package produce

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"os/signal"
	"src/types"
	"src/utils"

	//"src/utils"
	"syscall"
)

var createAddressesTable = `
drop table if exists addresses;
create table addresses
(
    "Id"             bigserial
        primary key,
    "CreatedAt"       	timestamp with time zone default now(),
    "UpdatedAt"      	timestamp with time zone default now(),
    "Address"     		text,
    "Nonce"       		bigint,
    "IsContract"		boolean,
    "Balance"    		bigint,
    "LastSeen"    		bigint
);
create unique index idx_addresses_address
    on addresses ("Address")`

var createBlockSyncTable = `
drop table if exists pg_block_sync_tracks;
create table pg_block_sync_tracks
(
    "Id"             		bigserial primary key,
    "CreatedAt"       	    timestamp with time zone default now(),
    "UpdatedAt"      		timestamp with time zone default now(),
    "Hash"                  text,
    "Number"                bigint,
    "Retrieved"             boolean,
    "Processed"             boolean,
    "ReceiptsProcessed"     boolean,
    "TransactionsProcessed" boolean,
    "TransactionCount"      bigint,
    "ContractsProcessed"    boolean
);
create unique index idx_pg_block_sync_tracks_number
    on pg_block_sync_tracks ("Number");`

var createSlotSyncTable = `
drop table if exists pg_slot_sync_tracks;
create table pg_slot_sync_tracks
(
    "Id"             bigserial primary key,
    "CreatedAt"      timestamp with time zone default now(),
    "UpdatedAt"      timestamp with time zone default now(),
    "Hash"           text,
    "Slot"           bigint,
    "Retrieved"      boolean,
    "Processed"      boolean,
    "BlobsProcessed" boolean,
    "BlobCount"      bigint
);
create unique index idx_pg_slot_sync_tracks_slot
    on pg_slot_sync_tracks ("Slot");`

var createTrackForRetryTable = `
drop table if exists pg_track_for_to_retries;
create table pg_track_for_to_retries
(
    "Id"            bigserial primary key,
    "CreatedAt"     timestamp with time zone default now(),
    "UpdatedAt"     timestamp with time zone default now(),
    "DataType"  text,
    "BlockId"       text,
    "RecordId"      text
)`

type PostgresDB struct {
	client *pgxpool.Pool
}

func NewClient(idxConfig types.IdxConfigStruct) *PostgresDB {

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("RAW_GO_POSTGRES_STRING"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	go func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		defer dbpool.Close()

		select {
		case <-ctx.Done():
			fmt.Println("signal received, Closing DB connection...")
			return
		}
	}()

	if idxConfig.ClearConsumer {
		_, err = dbpool.Exec(context.Background(), createAddressesTable)
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}
	if idxConfig.ClearPostgres {
		// DROPS AND REGENERATES TABLES
		_, err = dbpool.Exec(context.Background(), createAddressesTable)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		_, err = dbpool.Exec(context.Background(), createBlockSyncTable)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		_, err = dbpool.Exec(context.Background(), createSlotSyncTable)
		if err != nil {
			utils.Logger.Errorln(err)
		}
		_, err = dbpool.Exec(context.Background(), createTrackForRetryTable)
		if err != nil {
			utils.Logger.Errorln(err)
		}
	}

	return &PostgresDB{client: dbpool}
}
