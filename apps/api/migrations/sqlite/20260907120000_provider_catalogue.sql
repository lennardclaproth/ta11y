-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20260907120000_provider_catalogue.sql. SQLite has
-- no schemas, so the marketdata.* tables keep their bare names (see qualify*
-- helpers in internal/storage/db.go). UUIDs are TEXT and booleans are 0/1.
-- See the Postgres mirror for why the catalogue is a separate table and why
-- name is nullable.

CREATE TABLE provider_listings (
    id            TEXT     PRIMARY KEY,
    source        TEXT     NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    symbol        TEXT     NOT NULL,
    name          TEXT,
    exchange_name TEXT,
    exchange_mic  TEXT,
    has_eod       INTEGER  NOT NULL DEFAULT 0,
    has_intraday  INTEGER  NOT NULL DEFAULT 0,
    fetched_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_provider_listings_source_symbol UNIQUE (source, symbol)
);

CREATE INDEX idx_provider_listings_symbol_lower ON provider_listings (LOWER(symbol));
CREATE INDEX idx_provider_listings_name_lower ON provider_listings (LOWER(COALESCE(name, '')));

CREATE TABLE catalogue_syncs (
    id             TEXT     PRIMARY KEY,
    source         TEXT     NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    status         TEXT     NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    pages_fetched  INTEGER  NOT NULL DEFAULT 0,
    rows_upserted  INTEGER  NOT NULL DEFAULT 0,
    upstream_total INTEGER,
    last_error     TEXT,
    started_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at    DATETIME
);

CREATE INDEX idx_catalogue_syncs_source_started ON catalogue_syncs (source, started_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS catalogue_syncs;
DROP TABLE IF EXISTS provider_listings;

-- +goose StatementEnd
