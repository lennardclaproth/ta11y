-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20261010120000_cashflow_ignore_rules.sql. See there
-- for why a rule is shaped like a ledger filter and what ignore_overridden protects.
-- SQLite adds one column per ALTER TABLE, so the two transaction columns are split.
CREATE TABLE cashflow_ignore_rules (
    id              TEXT PRIMARY KEY,
    account_id      TEXT     NOT NULL REFERENCES cashflow_accounts(account_id) ON DELETE CASCADE,
    name            TEXT     NOT NULL,
    match_field     TEXT     NOT NULL CHECK (match_field IN ('description', 'note')),
    contains        TEXT     NOT NULL,
    direction       TEXT     CHECK (direction IS NULL OR direction IN ('in', 'out')),
    source          TEXT     NOT NULL DEFAULT '',
    enabled         INTEGER  NOT NULL DEFAULT 1,
    ignored_total   INTEGER  NOT NULL DEFAULT 0,
    last_applied_at DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cashflow_ignore_rules_account ON cashflow_ignore_rules (account_id, enabled);

ALTER TABLE transactions ADD COLUMN ignored_by_rule_id TEXT REFERENCES cashflow_ignore_rules(id) ON DELETE SET NULL;
ALTER TABLE transactions ADD COLUMN ignore_overridden INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_cashflow_transactions_ignored_by_rule
    ON transactions (import_id, ignored_by_rule_id)
    WHERE ignored_by_rule_id IS NOT NULL;

ALTER TABLE imports ADD COLUMN auto_ignored INTEGER NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE imports DROP COLUMN auto_ignored;
DROP INDEX IF EXISTS idx_cashflow_transactions_ignored_by_rule;
ALTER TABLE transactions DROP COLUMN ignore_overridden;
ALTER TABLE transactions DROP COLUMN ignored_by_rule_id;
DROP INDEX IF EXISTS idx_cashflow_ignore_rules_account;
DROP TABLE IF EXISTS cashflow_ignore_rules;

-- +goose StatementEnd
