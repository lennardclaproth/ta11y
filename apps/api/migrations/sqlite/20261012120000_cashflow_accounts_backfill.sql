-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Cashflow account projection backfill [009]. Every cashflow transaction references
-- cashflow_accounts, but nothing created a row there: portfolio and assets build their
-- projection on account.created, cashflow did not, so every import and manual
-- transaction failed on its account foreign key. New accounts now get the row from
-- the cashflow account.created handler; this creates it for accounts that already exist.
--
-- Divergence from Postgres: SQLite has no UUID default, so the id is generated here as
-- a random version-4 UUID in the same lowercase, hyphenated text form the app writes.
-- ---------------------------------------------------------------------------
INSERT INTO cashflow_accounts (id, account_id)
SELECT
    lower(hex(randomblob(4))) || '-' ||
    lower(hex(randomblob(2))) || '-4' ||
    substr(lower(hex(randomblob(2))), 2) || '-' ||
    substr('89ab', 1 + (abs(random()) % 4), 1) || substr(lower(hex(randomblob(2))), 2) || '-' ||
    lower(hex(randomblob(6))),
    id
FROM accounts
WHERE true
ON CONFLICT (account_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Intentionally a no-op: a backfilled row cannot be told apart from one the
-- account.created handler wrote, and deleting it cascades to every cashflow
-- transaction of that account.
SELECT 1;

-- +goose StatementEnd
