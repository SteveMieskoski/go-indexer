CREATE OR REPLACE FUNCTION updated_time()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';


CREATE TRIGGER update_time_1
    BEFORE UPDATE
    ON indexer.public.addresses
    FOR EACH ROW
EXECUTE PROCEDURE updated_time();

CREATE TRIGGER update_time_1
    BEFORE UPDATE
    ON indexer.public.pg_block_sync_tracks
    FOR EACH ROW
EXECUTE PROCEDURE updated_time();

CREATE TRIGGER update_time_1
    BEFORE UPDATE
    ON indexer.public.pg_slot_sync_tracks
    FOR EACH ROW
EXECUTE PROCEDURE updated_time();

CREATE TRIGGER update_time_1
    BEFORE UPDATE
    ON indexer.public.pg_track_for_to_retries
    FOR EACH ROW
EXECUTE PROCEDURE updated_time();


create table addresses
(
    id          bigserial
        primary key,
    created_at      timestamp with time zone default now(),
    updated_at      timestamp with time zone default now(),
    address     text,
    nonce       text,
    is_contract boolean,
    balance     bigint,
    lastSeen    bigint
);

create unique index idx_addresses_address
    on addresses (Address);

create table pg_block_sync_tracks
(
    id                     bigserial
        primary key,
    created_at      timestamp with time zone default now(),
    updated_at      timestamp with time zone default now(),
    hash                   text,
    number                 bigint,
    retrieved              boolean,
    processed              boolean,
    receipts_processed     boolean,
    transactions_processed boolean,
    transaction_count      bigint,
    contracts_processed    boolean
);

create unique index idx_pg_block_sync_tracks_number
    on pg_block_sync_tracks (number);

create table pg_slot_sync_tracks
(
    id              bigserial
        primary key,
    created_at      timestamp with time zone default now(),
    updated_at      timestamp with time zone default now(),
    hash            text,
    slot            bigint,
    retrieved       boolean,
    processed       boolean,
    blobs_processed boolean,
    blob_count      bigint
);

create unique index idx_pg_slot_sync_tracks_slot
    on pg_slot_sync_tracks (slot);


create table pg_track_for_to_retries
(
    id             bigserial
        primary key,
    created_at      timestamp with time zone default now(),
    updated_at      timestamp with time zone default now(),
    type_for_retry text,
    block_id       text,
    record_id      text
);