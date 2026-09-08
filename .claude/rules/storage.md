---
paths:
  - "apps/api/internal/storage/**/*.go"
---

# Storage (`internal/storage`)

`sqlx`-based, one consolidated `sqlx_<aggregate>_store.go` per aggregate. Stores implement the
**feature-owned** interfaces directly (asserted with `var _ feature.Iface = (*Store)(nil)`); they do
not define their own parallel abstractions.

Both Postgres (deployed) and SQLite (local dev) are served from the same code.

## Patterns to follow

- `DB` wraps `*sqlx.DB`.
- Transactions: `DB.WithTx(ctx, fn)` stashes the `*sqlx.Tx` in the context; stores call
  `db.GetExecutor(ctx)` so the same query transparently uses the transaction or the pool.
  Never take a `*sqlx.Tx` parameter to work around this.
- `qualifyTable` prefixes a schema (e.g. `cashflow.transactions`) **only on Postgres**. SQLite is
  schemaless and gets the flattened, prefixed bare name (`cashflow_transactions`). Schema and table
  names are constants in `db.go` — add new ones there.
- Follow the existing rebind, duplicate-key mapping, and `sql.ErrNoRows` handling; map driver errors
  to the owning feature's sentinel errors rather than leaking `database/sql` errors upward.
- Queries used by a `Queries` type stay read-only; writes go through the store methods a `Commands`
  type depends on.
