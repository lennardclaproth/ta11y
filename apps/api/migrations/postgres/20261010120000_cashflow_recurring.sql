-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Recurring items [033]. A subscription, fixed cost or recurring income is not a
-- property of a single transaction but of the series of transactions that keep
-- arriving for the same counterparty, so it is its own record that transactions
-- are linked to -- never a tag, which would move the tag distribution, and never
-- a flag on the row, which could carry neither a name, a rhythm nor an end.
--
-- match_key is the normalised fingerprint of the statement description the item
-- was started from. "To whom" exists only inside that description and is spelled
-- differently per bank, so the key is what recognises the same counterparty in a
-- later import; the name stays whatever the user chose or confirmed.
--
-- ended_from is the first day of the month an item stops being expected. Ending
-- is a date, not a delete: every transaction linked before it stays linked.
-- ---------------------------------------------------------------------------
CREATE TABLE cashflow.recurring_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID         NOT NULL REFERENCES cashflow.accounts(account_id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    direction  VARCHAR(8)   NOT NULL CHECK (direction IN ('in', 'out')),
    rhythm     VARCHAR(16)  NOT NULL CHECK (rhythm IN ('monthly', 'quarterly', 'yearly')),
    match_key  VARCHAR(255) NOT NULL,
    ended_from DATE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_recurring_items_account_name UNIQUE (account_id, name)
);

-- A transaction belongs to at most one item, so the unique key is the
-- transaction alone: linking the same row twice is refused by the database
-- rather than by a check the next caller could forget.
CREATE TABLE cashflow.recurring_links (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id        UUID        NOT NULL REFERENCES cashflow.recurring_items(id) ON DELETE CASCADE,
    transaction_id UUID        NOT NULL UNIQUE REFERENCES cashflow.transactions(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- A dismissed suggestion does not come back, so the refusal is recorded against
-- the fingerprint it was made about rather than against the transactions, which
-- keep arriving.
CREATE TABLE cashflow.recurring_dismissals (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID         NOT NULL REFERENCES cashflow.accounts(account_id) ON DELETE CASCADE,
    match_key  VARCHAR(255) NOT NULL,
    direction  VARCHAR(8)   NOT NULL CHECK (direction IN ('in', 'out')),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_recurring_dismissals UNIQUE (account_id, match_key, direction)
);

CREATE INDEX idx_recurring_items_account ON cashflow.recurring_items (account_id, direction);
CREATE INDEX idx_recurring_links_item ON cashflow.recurring_links (item_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS cashflow.idx_recurring_links_item;
DROP INDEX IF EXISTS cashflow.idx_recurring_items_account;
DROP TABLE IF EXISTS cashflow.recurring_dismissals;
DROP TABLE IF EXISTS cashflow.recurring_links;
DROP TABLE IF EXISTS cashflow.recurring_items;

-- +goose StatementEnd
