-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Provider listing catalogue: a local cache of upstream ticker search results.
-- Deliberately separate from marketdata.listings, which holds only instruments
-- the user actually tracks -- GET /marketdata/listings is unpaginated and feeds
-- the portfolio listing picker, so catalogue rows must not land there.
-- name is nullable: upstream returns null/empty names for some tickers, and
-- those rows are surfaced but cannot be adopted (listings.name is NOT NULL).
-- ---------------------------------------------------------------------------
CREATE TABLE marketdata.provider_listings (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source        VARCHAR(20)  NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    symbol        VARCHAR(50)  NOT NULL,
    name          VARCHAR(255),
    exchange_name VARCHAR(255),
    exchange_mic  VARCHAR(20),
    has_eod       BOOLEAN      NOT NULL DEFAULT FALSE,
    has_intraday  BOOLEAN      NOT NULL DEFAULT FALSE,
    fetched_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_provider_listings_source_symbol UNIQUE (source, symbol)
);

CREATE INDEX idx_provider_listings_symbol_lower ON marketdata.provider_listings (LOWER(symbol));
CREATE INDEX idx_provider_listings_name_lower ON marketdata.provider_listings (LOWER(COALESCE(name, '')));

-- ---------------------------------------------------------------------------
-- Catalogue seed-sync runs. Bounded by design: the seed pages the provider's
-- popularity-ordered empty search, so a handful of pages yields the
-- most-traded instruments for a handful of API credits. No resume/offset
-- state -- a bounded run is cheap enough to simply repeat.
-- ---------------------------------------------------------------------------
CREATE TABLE marketdata.catalogue_syncs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source         VARCHAR(20)  NOT NULL CHECK (source IN ('alpha_vantage', 'market_stack', 'brandnewday')),
    status         VARCHAR(20)  NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    pages_fetched  INTEGER      NOT NULL DEFAULT 0,
    rows_upserted  INTEGER      NOT NULL DEFAULT 0,
    upstream_total INTEGER,
    last_error     TEXT,
    started_at     TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at    TIMESTAMPTZ
);

CREATE INDEX idx_catalogue_syncs_source_started ON marketdata.catalogue_syncs (source, started_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS marketdata.catalogue_syncs;
DROP TABLE IF EXISTS marketdata.provider_listings;

-- +goose StatementEnd
