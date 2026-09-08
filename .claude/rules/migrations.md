---
paths:
  - "apps/api/migrations/**"
---

# Migrations (goose, two dialects)

Every schema change needs **matching files in both** `apps/api/migrations/postgres/` and
`apps/api/migrations/sqlite/`, with the **same timestamped filename**. The schema is one
from-scratch init per dialect plus incremental files; Postgres uses real schemas, SQLite flattens
them into prefixed table names (`portfolio_accounts`, `asset_classes`).

## Hard rules

- **Do not modify an existing migration.** Add a new timestamped one.
- Only create migrations when the task explicitly calls for a schema change.
- `make migrate-create name=<n>` scaffolds; `make migrate-up` / `make migrate-down` /
  `make migrate-status` need `DATABASE_URL`.
- In dev the app auto-creates the database and migrates on startup, so a new pair of files is picked
  up by `make run`.
- Write a real `-- +goose Down` — the down path is exercised by `make migrate-down`.
- Keep the two dialects semantically equivalent; note any unavoidable divergence in a comment in
  both files.
