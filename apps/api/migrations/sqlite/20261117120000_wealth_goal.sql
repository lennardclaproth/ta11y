-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20261010090000_wealth_goal.sql. See there for why
-- the purpose is its own column next to the tag and the ignored flag, and why the
-- goal is stored per month rather than once per account.
--
-- SQLite has no schemas, so wealthgoal.goals is flattened to wealth_goals; the
-- storage layer resolves the name per dialect.
ALTER TABLE transactions ADD COLUMN purpose TEXT NOT NULL DEFAULT ''
    CHECK (purpose IN ('', 'income', 'wealth'));

CREATE INDEX idx_transactions_purpose ON transactions (account_id, date, purpose);

CREATE TABLE wealth_goals (
    id             TEXT PRIMARY KEY,
    account_id     TEXT     NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    share_percent  INTEGER  NOT NULL CHECK (share_percent BETWEEN 0 AND 100),
    effective_from DATE     NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_wealth_goals_account_month UNIQUE (account_id, effective_from)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS wealth_goals;

-- The index has to go first: SQLite refuses to drop a column an index covers.
DROP INDEX IF EXISTS idx_transactions_purpose;
ALTER TABLE transactions DROP COLUMN purpose;

-- +goose StatementEnd
