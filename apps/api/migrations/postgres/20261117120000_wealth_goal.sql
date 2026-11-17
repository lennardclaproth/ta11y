-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Monthly wealth growth goal [033].
--
-- A transaction gets a purpose: it is either income the goal is measured
-- against, a contribution towards wealth, or neither. This is a third axis next
-- to the tag and the ignored flag -- a contribution to a savings account is
-- usually ignored for the cashflow totals and still has to count here -- so it
-- is its own column rather than a reserved tag value. The empty string is the
-- unassigned state, which keeps the column NOT NULL and lets existing rows
-- default into it without a backfill.
--
-- Only outgoing money can be a contribution and only incoming money can be
-- income; that rule is enforced in the cashflow package, where the direction is
-- already known, rather than by a cross-column constraint here.
--
-- The index carries account_id and date because the standing always groups an
-- account's rows into calendar months before it looks at the purpose.
-- ---------------------------------------------------------------------------
ALTER TABLE cashflow.transactions
    ADD COLUMN purpose VARCHAR(16) NOT NULL DEFAULT ''
        CHECK (purpose IN ('', 'income', 'wealth'));

CREATE INDEX idx_transactions_purpose ON cashflow.transactions (account_id, date, purpose);

-- ---------------------------------------------------------------------------
-- The goal itself. One row per month in which the share was set or corrected,
-- not one row per account: adjusting the goal applies from the month of the
-- adjustment and leaves earlier months judged by the share that held then. The
-- unique constraint makes a second adjustment within the same month a
-- correction of the first.
-- ---------------------------------------------------------------------------
CREATE SCHEMA IF NOT EXISTS wealthgoal;

CREATE TABLE wealthgoal.goals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id     UUID        NOT NULL REFERENCES account.accounts(id) ON DELETE CASCADE,
    share_percent  INTEGER     NOT NULL CHECK (share_percent BETWEEN 0 AND 100),
    effective_from DATE        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_wealth_goals_account_month UNIQUE (account_id, effective_from)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS wealthgoal.goals;
DROP SCHEMA IF EXISTS wealthgoal;

DROP INDEX IF EXISTS cashflow.idx_transactions_purpose;
ALTER TABLE cashflow.transactions DROP COLUMN IF EXISTS purpose;

-- +goose StatementEnd
