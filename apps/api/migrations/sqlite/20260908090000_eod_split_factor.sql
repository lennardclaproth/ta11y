-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20260908090000_eod_split_factor.sql. See there for
-- why the split factor lives on the end-of-day row.
ALTER TABLE eods ADD COLUMN split_factor REAL NOT NULL DEFAULT 1;

CREATE INDEX idx_eods_splits ON eods (listing_id, date) WHERE split_factor <> 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_eods_splits;
ALTER TABLE eods DROP COLUMN split_factor;

-- +goose StatementEnd
