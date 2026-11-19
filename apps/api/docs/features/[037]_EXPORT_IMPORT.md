# [037] Export and import

Each dataset exports to its own JSON file and imports from one: cashflow, portfolio and asset
classes for the signed-in account, and end-of-day prices for administrators. An import merges —
what is already there is left alone — so the same file can be imported twice, into the account it
came from or into another one on another instance.

## Files

Every file opens with `format: "ta11y.backup"`, `version: 1`, `kind` and `exported_at`. Records
refer to each other with file-local `ref`s, never with ids of the instance they came from.
Vendors are named by name, listings by `{symbol, source}`.

| Kind | Holds | Left out (rebuilt) |
| --- | --- | --- |
| `cashflow` | transactions (tag, note, ignored state, the rule that ignored them), their imports, ignore rules, recurring items with their links, refused suggestions | — |
| `portfolio` | transactions and their imports | positions, snapshots |
| `assets` | hand-kept classes, items (with listing), worth changes of unlisted items, purchases | the portfolio class, mutations of daily-priced items, snapshots |
| `eods` | every listing — with or without prices — with its name and instrument data (type, ticker, ISIN, description, exchange, region, currency, active), and every day's prices with its split factor | provider rows, API keys, the cached provider catalogue |

## Import

1. **Check.** The file is merged in a transaction that is rolled back (`dry_run`): the counts —
   new and already here, per kind of record — come from the same path a real import takes.
2. **Refuse** with 409, naming them, when the new records need vendors or listings this instance
   lacks. Nothing is written. A file of another kind, or no ta11y file at all, is a 422.
3. **Import.** One transaction: the whole file lands or nothing does.

What counts as already here:

- **Transactions** by checksum. Checksums digest the account and an unstored dedup sequence;
  into another account the sequence is recovered by reproducing the stored checksum under the
  source account, then the checksum is re-derived for the target, so a later CSV import of the
  same statement there still dedupes. A row edited after import (a digested field changed) falls
  back to a checksum keyed to the target — a second restore still finds it, a CSV import won't.
- **Imports** are recreated only for files that bring at least one new row, as completed or failed.
- **Ignore rules** by what they match (field, text, direction, bank), not by name.
- **Recurring items** by name; links and refused suggestions skip what exists.
- **Asset classes** by name, **items** by class and name. An item the account already holds keeps
  its own history: worth changes carry running totals that two histories would corrupt.
- **Listings** by symbol and source. One that exists keeps its own data; a missing one is
  created from the file's instrument data, before its prices, and linked to this instance's
  provider for its source when there is one. Its accumulated range is set to the file's first
  and last day, so an API source's first end-of-day read only syncs the days after the file.
  A listing without instrument data (a file from before listings travelled) must exist, or
  the file is refused naming it.
- **End-of-day prices** by listing and day.

## Events

`backup.portfolio.restored` → portfolio rebuild (which drives the assets sync);
`backup.assets.restored` → assets snapshot rebuild. A cashflow import derives nothing.

## Code map

- `internal/backup` — file format, export (`Queries`), merge (`Commands`, `restore*.go`), validation.
- `internal/storage/sqlx_backup_store.go` — full-column reads and writes, one transaction per import.
- `internal/cashflow`, `internal/portfolio` — `Transaction.RekeyedChecksum`.
- `transport/http/handlers/backup` — `GET|POST /backup/{cashflow,portfolio,assets}/{export,import}`
  (protected), `/backup/eods/{export,import}` (admin only).
- `transport/messaging/handlers/{portfolio,assets}/restored.go` — the rebuilds.
- `web`: `organisms/export-import-dialog`, opened from the account menu; `services/backup.ts`.

## Gaps

- An EOD import does not revalue portfolios or daily-priced items, the same as an EOD CSV upload.
- After an import the open page is not refreshed; it shows the new records on its next load.
- A cashflow file restores the ignore rule a row was ignored by only when that rule is in the file.
