-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20260907130000_account_admin.sql. Booleans are 0/1
-- on SQLite. See the Postgres mirror for why this flag exists and its limits.
ALTER TABLE accounts ADD COLUMN admin INTEGER NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE accounts DROP COLUMN admin;

-- +goose StatementEnd
