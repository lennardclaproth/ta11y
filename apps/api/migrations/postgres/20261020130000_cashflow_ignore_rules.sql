-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Ignore rules [033]. Around a third of an imported statement is transfers
-- between the owner's own accounts and credit-card settlements; they say nothing
-- about income or spending but arrive as ordinary transactions and have to be
-- ignored by hand after every import.
--
-- A rule is deliberately the same shape as what the ledger can already filter on
-- -- text in the description or note, a direction, one bank or all -- so what it
-- catches can be previewed with the query the ledger itself uses, and nothing
-- about it is inferred.
-- ---------------------------------------------------------------------------
CREATE TABLE cashflow.ignore_rules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID         NOT NULL REFERENCES cashflow.accounts(account_id) ON DELETE CASCADE,
    name         VARCHAR(120) NOT NULL,
    match_field  VARCHAR(16)  NOT NULL CHECK (match_field IN ('description', 'note')),
    contains     VARCHAR(255) NOT NULL,
    -- NULL matches both directions; the rule simply does not narrow on it.
    direction    VARCHAR(8)   CHECK (direction IS NULL OR direction IN ('in', 'out')),
    -- The bank the rule is limited to, empty for every one. A cashflow
    -- transaction has no account of its own, only the source it was imported
    -- from, so that is what a rule can honestly scope on.
    source       VARCHAR(64)  NOT NULL DEFAULT '',
    enabled      BOOLEAN      NOT NULL DEFAULT TRUE,
    ignored_total INTEGER     NOT NULL DEFAULT 0,
    last_applied_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cashflow_ignore_rules_account ON cashflow.ignore_rules (account_id, enabled);

-- Which rule ignored a row, so a rule that is too wide is visible per transaction
-- rather than only in a total, and ignore_overridden: the row's ignored state was
-- decided by hand. Rules leave those alone, which is what makes restoring one
-- stick across later imports and later applications of the same rule.
ALTER TABLE cashflow.transactions
    ADD COLUMN ignored_by_rule_id UUID REFERENCES cashflow.ignore_rules(id) ON DELETE SET NULL,
    ADD COLUMN ignore_overridden  BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_cashflow_transactions_ignored_by_rule
    ON cashflow.transactions (import_id, ignored_by_rule_id)
    WHERE ignored_by_rule_id IS NOT NULL;

-- The import already counts what was new, duplicate and failed; what its rules
-- ignored is the fourth number its result page reads.
ALTER TABLE import.imports
    ADD COLUMN auto_ignored INTEGER NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE import.imports DROP COLUMN IF EXISTS auto_ignored;
DROP INDEX IF EXISTS cashflow.idx_cashflow_transactions_ignored_by_rule;
ALTER TABLE cashflow.transactions
    DROP COLUMN IF EXISTS ignore_overridden,
    DROP COLUMN IF EXISTS ignored_by_rule_id;
DROP INDEX IF EXISTS cashflow.idx_cashflow_ignore_rules_account;
DROP TABLE IF EXISTS cashflow.ignore_rules;

-- +goose StatementEnd
