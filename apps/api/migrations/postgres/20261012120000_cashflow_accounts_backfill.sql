-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Cashflow account projection backfill [009]. Every cashflow transaction references
-- cashflow.accounts, but nothing created a row there: portfolio and assets build their
-- projection on account.created, cashflow did not, so every import and manual
-- transaction failed on transactions_account_id_fkey. New accounts now get the row from
-- the cashflow account.created handler; this creates it for accounts that already exist.
-- ---------------------------------------------------------------------------
INSERT INTO cashflow.accounts (account_id)
SELECT id FROM account.accounts
ON CONFLICT (account_id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Intentionally a no-op: a backfilled row cannot be told apart from one the
-- account.created handler wrote, and deleting it cascades to every cashflow
-- transaction of that account.
SELECT 1;

-- +goose StatementEnd
