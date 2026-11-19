-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Corporate actions [013]. A split changes the share count without changing the
-- money invested, so a portfolio that accumulates quantity from raw broker
-- transactions drifts from reality the moment one occurs: prices after the split
-- are post-split while the accumulated quantity is not.
--
-- The provider already reports the factor on every end-of-day row (1 on ordinary
-- days), so the split is recorded alongside the prices it applies to rather than
-- in a separate corporate-actions table -- no extra provider request, and the
-- factor is always in step with the price series it belongs to.
--
-- Rows that predate this column, and manually imported end-of-day files that
-- carry no factor, default to 1 and are therefore a no-op for the rebuild.
-- ---------------------------------------------------------------------------
ALTER TABLE marketdata.eods
    ADD COLUMN split_factor DOUBLE PRECISION NOT NULL DEFAULT 1;

-- Rebuilds look up the handful of split days inside a position's window, so the
-- index covers only rows that actually record a split.
CREATE INDEX idx_eods_splits ON marketdata.eods (listing_id, date)
    WHERE split_factor <> 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS marketdata.idx_eods_splits;
ALTER TABLE marketdata.eods DROP COLUMN IF EXISTS split_factor;

-- +goose StatementEnd
