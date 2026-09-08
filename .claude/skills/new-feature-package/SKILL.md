---
name: new-feature-package
description: Add a feature package or a new Commands/Queries method in the Go API, with feature-owned store interfaces, errors, events and the storage implementation. Use when backend behaviour needs a new capability rather than just a new route.
---

# Add backend feature behaviour

Feature packages live in `apps/api/internal/<feature>`. Read a populated one first — `assets` and
`cashflow` are the fullest examples.

## The boundary

- **`Commands`** — writes: mutate state, publish events.
- **`Queries`** — reads.

Both depend on **small interfaces the feature itself declares** (`creator`, `queryStore`, ...),
listing only the methods that feature needs. `internal/storage` implements those interfaces
directly, asserted with `var _ <feature>.Iface = (*SQLXStore)(nil)`.

Do **not** add adapter interfaces, reflection bridges, or duplicate domain types to avoid changing
the real boundary — change the boundary. Do not add a transport-side interface when the concrete
`Commands` or `Queries` type already expresses the dependency.

## Steps

1. **Domain types** in the feature package. Reuse `internal/money` for prices and decimals,
   `internal/date` for parsing, ranges and formatting, and `internal/sorting` for sort fields and
   directions — search for an existing helper before writing one.
2. **Sentinel errors** in that package's `errors.go`. Transport maps these to status codes.
3. **The store interface**, declared in the feature package beside the type that uses it — the
   smallest set of methods that feature needs.
4. **The method** on `Commands` or `Queries`: `ctx` first, account id explicit. Enforce domain
   invariants here, not in the handler.
5. **Storage implementation** in `internal/storage/sqlx_<aggregate>_store.go`. Use
   `db.GetExecutor(ctx)` so it works inside `DB.WithTx`, `qualifyTable` for the Postgres schema
   prefix (SQLite gets the flattened bare name), and the existing rebind, duplicate-key and
   `sql.ErrNoRows` handling. Add the `var _` assertion.
6. **Events**, if other features must react: define the payload in the feature's `events.go`,
   publish from `Commands`, and register the subscriber in the startup wiring — not by chaining a
   publish from inside another handler. Handlers live in `transport/messaging/handlers/<feature>`
   and stay thin.
7. **Bootstrap**, if startup must seed data: go through the owning feature package, never through
   new storage-level startup wiring.

## Finish

Unit-test the new behaviour (`commands_test.go` alongside), then `make build`, `make test`,
`make lint`. Document exported types and functions. Update `FEATURES.md` and `CHANGELOG.md` when
this is a real capability change.
