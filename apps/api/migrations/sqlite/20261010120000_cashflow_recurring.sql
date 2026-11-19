-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20261010120000_cashflow_recurring.sql. See there
-- for why a recurring item is its own record, what match_key is for, and why
-- ending is a date rather than a delete.
CREATE TABLE cashflow_recurring_items (
    id         TEXT PRIMARY KEY,
    account_id TEXT     NOT NULL REFERENCES cashflow_accounts(account_id) ON DELETE CASCADE,
    name       TEXT     NOT NULL,
    direction  TEXT     NOT NULL CHECK (direction IN ('in', 'out')),
    rhythm     TEXT     NOT NULL CHECK (rhythm IN ('monthly', 'quarterly', 'yearly')),
    match_key  TEXT     NOT NULL,
    ended_from DATE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_recurring_items_account_name UNIQUE (account_id, name)
);

CREATE TABLE cashflow_recurring_links (
    id             TEXT PRIMARY KEY,
    item_id        TEXT     NOT NULL REFERENCES cashflow_recurring_items(id) ON DELETE CASCADE,
    transaction_id TEXT     NOT NULL UNIQUE REFERENCES transactions(id) ON DELETE CASCADE,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cashflow_recurring_dismissals (
    id         TEXT     PRIMARY KEY,
    account_id TEXT     NOT NULL REFERENCES cashflow_accounts(account_id) ON DELETE CASCADE,
    match_key  TEXT     NOT NULL,
    direction  TEXT     NOT NULL CHECK (direction IN ('in', 'out')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_recurring_dismissals UNIQUE (account_id, match_key, direction)
);

CREATE INDEX idx_recurring_items_account ON cashflow_recurring_items (account_id, direction);
CREATE INDEX idx_recurring_links_item ON cashflow_recurring_links (item_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_recurring_links_item;
DROP INDEX IF EXISTS idx_recurring_items_account;
DROP TABLE IF EXISTS cashflow_recurring_dismissals;
DROP TABLE IF EXISTS cashflow_recurring_links;
DROP TABLE IF EXISTS cashflow_recurring_items;

-- +goose StatementEnd
