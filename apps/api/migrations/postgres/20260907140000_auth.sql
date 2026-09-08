-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Authentication [030]. Deliberately minimal: an account is an id plus an email,
-- and every identity provider contributes only a pseudonymous (issuer, subject)
-- pair. No ID, access or refresh token is ever persisted -- the ID token is
-- verified once at the OIDC callback and discarded, after which the session
-- below is the only credential the API knows about.
-- ---------------------------------------------------------------------------

-- email is nullable so accounts predating authentication (the bootstrapped one)
-- keep working. NULLs do not collide in a unique index, so several may coexist.
ALTER TABLE account.accounts ADD COLUMN email VARCHAR(320);
CREATE UNIQUE INDEX uq_accounts_email_lower ON account.accounts (LOWER(email));

-- One row per (provider, user). Provisioning sets accounts.name to the email, so
-- the pre-existing UNIQUE on name stays satisfiable without a constraint change.
CREATE TABLE account.identities (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id    UUID         NOT NULL REFERENCES account.accounts(id) ON DELETE CASCADE,
    issuer        VARCHAR(255) NOT NULL,
    subject       VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMPTZ,
    CONSTRAINT uq_identities_issuer_subject UNIQUE (issuer, subject)
);

CREATE INDEX idx_identities_account ON account.identities (account_id);

-- Opaque server-side sessions. The cookie carries 32 random bytes; only their
-- SHA-256 (lowercase hex) is stored, so a database leak yields no usable
-- credential. Revocation is a row update, which is why these are not JWTs.
CREATE TABLE account.sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID        NOT NULL REFERENCES account.accounts(id) ON DELETE CASCADE,
    token_hash   CHAR(64)    NOT NULL UNIQUE,
    issued_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_sessions_account ON account.sessions (account_id);
CREATE INDEX idx_sessions_expires_at ON account.sessions (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS account.sessions;
DROP TABLE IF EXISTS account.identities;
DROP INDEX IF EXISTS account.uq_accounts_email_lower;
ALTER TABLE account.accounts DROP COLUMN IF EXISTS email;

-- +goose StatementEnd
