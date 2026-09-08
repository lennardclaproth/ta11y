---
description: Create a mirrored goose migration pair (Postgres + SQLite) for a schema change
argument-hint: "<snake_case_name> — what the schema change does"
allowed-tools: Bash, Read, Write, Edit, Grep, Glob
---

Create a migration: $ARGUMENTS

1. Confirm the change actually needs new schema. If an existing table can carry it, say so and stop.
2. Read the most recent migration in **both** `apps/api/migrations/postgres/` and
   `apps/api/migrations/sqlite/` to match style, and note how the dialects diverge: Postgres uses
   real schemas (`cashflow.transactions`), SQLite flattens to prefixed names (`cashflow_transactions`).
3. Run `make migrate-create name=<snake_case_name>` if it scaffolds into both directories; otherwise
   write both files by hand with the **same timestamp prefix and filename**.
4. Write real `-- +goose Up` and `-- +goose Down` sections in each. The down path must actually
   reverse the up path.
5. Never modify an existing migration file.
6. If the two dialects have to diverge beyond schema-qualification, add the same explanatory comment
   to both files.
7. If new tables or schemas were added, add their names to the constants in
   `apps/api/internal/storage/db.go` so `qualifyTable` handles them.
8. Verify: `git status --porcelain apps/api/migrations` shows exactly two new files, then
   `make build`. Say whether you ran the migration against a database or not.
