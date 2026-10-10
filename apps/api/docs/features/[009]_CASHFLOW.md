# [009]–[012],[023],[032] Cashflow

> **Feature IDs:** 009 (querying) · 010 (analytics) · 011 (tagging) · 012 (ignore) · 023 (manual transactions) · 032 (date changes) · **Area:** Core (user-facing) · **Status:** live; routes registered in `cmd/ta11y/main.go`
>
> **Backend packages:** `internal/cashflow` · `transport/http/handlers/cashflow` · `internal/storage/sqlx_cashflow_store.go` · `internal/importer/cashflow`
>
> **Related features:** [002] CSV imports · [003] Account management · [010] shares the same store

## Overview

Cashflow is the bank/payment transaction subsystem. Transactions arrive two ways — bulk
from a CSV import ([002]) or manually entered — and are then queried, analyzed, tagged, and
ignored:

- **[009] Query** — list transactions with field filters (description/note/source contains,
  direction, tags, untagged), a fuzzy `q`, `hide_ignored`, date range, sorting, and offset
  pagination.
- **[010] Analytics** — monthly incoming/outgoing/net trends, and tag distribution
  (combined / incoming / outgoing). Both accept a date range and `include_ignored`.
- **[011] Tagging** — tag a single transaction, a selection of IDs, or everything matching a
  filter.
- **[012] Ignore** — mark a selection or a filter match as ignored / not-ignored; ignored
  rows drop out of analytics totals. Both paths are by-hand paths, so they also set
  `ignore_overridden`, which keeps an ignore rule ([033]) from undoing the decision later.
- **[023] Manual create** — bulk-create up to 100 manual transactions for an account, on any
  day up to and including today.
- **[032] Change date** — move one manually entered transaction to another day, today or
  earlier. Nothing else about it changes, and imported rows are refused.

## Domain model

```mermaid
classDiagram
    class Commands {
        -CommandStore cs
        -QueryStore qs
        -accountExistenceChecker aec
        +CreateMany(ctx, accID, importID, data) CreateManyResult
        +ChangeDate(ctx, accID, id, date) Transaction
        +TagByID(ctx, id, tag) error
        +TagByIDs(ctx, ids, tag) int
        +TagByFilter(ctx, tag, accID, filters) BulkTagResult
        +IgnoreByIDs(ctx, ids, ignored) int
        +IgnoreByFilter(ctx, filters, ignored) int
    }
    class Queries {
        -QueryStore qs
        +MonthlyAnalytics(ctx, filter) MonthlyAnalyticsPoint[]
        +TagDistribution(ctx, filter) TagDistribution
        +ListTransactions(ctx, query) TransactionListResult
    }
    class CommandStore {
        <<interface>>
        +CreateTransactions(ctx, txs) int
        +UpdateDate(ctx, accID, id, date, checksum) int
        +UpdateTagByIDs(ctx, ids, tag) int
        +UpdateTagByFilter(ctx, filters, tag) int
        +UpdateIgnoredByIDs(ctx, ids, ignored) int
        +UpdateIgnoredByFilter(ctx, filters, ignored) int
    }
    class QueryStore {
        <<interface>>
        +GetMonthlyAnalytics(ctx, filter) MonthlyAnalyticsPoint[]
        +GetTagDistribution(ctx, filter) TagDistribution
        +ListTransactions(ctx, query) TransactionListResult
        +CountByFilter(ctx, filters) int
        +GetTransaction(ctx, accID, id) Transaction
        +GetTransactionByChecksum(ctx, accID, checksum) Transaction
    }
    class accountExistenceChecker {
        <<interface>>
        +Exists(ctx, id) bool
    }
    class Transaction {
        +UUID ID
        +UUID AccountID
        +UUID ImportID
        +string Source
        +money.Price AmountCents
        +CashFlowDirection Direction
        +string Tag
        +bool Ignored
        +string Checksum
    }
    class SQLXCashflowStore

    Commands ..> CommandStore
    Commands ..> accountExistenceChecker : guards CreateMany
    Queries ..> QueryStore
    CommandStore <|.. SQLXCashflowStore
    QueryStore <|.. SQLXCashflowStore

    note for Transaction "Checksum = SHA-256 over description, note, source,\ndirection, amount, date, dedup sequence, account_id"
    note for SQLXCashflowStore "bulk insert is ON CONFLICT(checksum) DO NOTHING"
```

Enums: `CashFlowDirection` ∈ {`in`, `out`}; `AccountType` ∈ {`checking`, `savings`,
`credit`, `brokerage`}. Manual rows carry `source = "manual"` or `"manual:<vendor>"`;
imported rows carry the vendor name.

## Data model

```mermaid
erDiagram
    ACCOUNTS          ||--o| CASHFLOW_ACCOUNTS     : "account_id (CASCADE)"
    CASHFLOW_ACCOUNTS ||--o{ CASHFLOW_TRANSACTIONS : "account_id (CASCADE)"
    IMPORTS           ||--o{ CASHFLOW_TRANSACTIONS : "import_id (CASCADE, nullable)"

    CASHFLOW_TRANSACTIONS {
        uuid id PK
        uuid account_id FK
        uuid import_id FK "nullable (manual rows = NULL)"
        uuid ignored_by_rule_id FK "nullable - the rule that ignored it [033]"
        string description
        string source
        bigint amount_cents
        string direction "CHECK in|out"
        date date
        string tag "default ''"
        string account_type "CHECK checking|savings|credit|brokerage, nullable"
        bool ignored "default false"
        bool ignore_overridden "default false - decided by hand [033]"
        int row_number
        string checksum UK "unique - dedup key"
    }
    CASHFLOW_ACCOUNTS {
        uuid id PK
        uuid account_id FK "NOT NULL, unique"
    }
    ACCOUNTS { uuid id PK }
    IMPORTS  { uuid id PK }
```

Physical names: Postgres `cashflow.transactions` / `cashflow.accounts`; SQLite `transactions`
/ `cashflow_accounts`. A partial analytics index covers `(date, direction, tag) WHERE ignored = FALSE`.
Note `cashflow.transactions.account_id` references the projection (`cashflow.accounts.account_id`),
not `account.accounts` directly.

## Transaction classification states

`tag` and `ignored` are independent classifications a transaction moves through after it
exists:

```mermaid
stateDiagram-v2
    state Transaction {
        [*] --> Untagged
        Untagged --> Tagged : set tag (single / selection / filter)
        Tagged --> Tagged : retag
        Tagged --> Untagged : clear tag
        --
        [*] --> Active
        Active --> Ignored : ignore (selection / filter / rule [033])
        Ignored --> Active : un-ignore
    }
```

Ignored transactions are excluded from analytics totals unless the request sets
`include_ignored`. Which rows are ignored is also decided by ignore rules ([033]); how
ignored rows count is unchanged.

## Filter-based tagging flow

The filter-based tag/ignore commands are the only ones with a (declared) sync-vs-async
decision. Today the command always runs synchronously.

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant H as TagByFilter handler
    participant C as cashflow.Commands
    participant Q as QueryStore
    participant S as CommandStore

    Client->>H: POST /cashflow/transactions/tag/filter {tag, account_id?, filters}
    H->>H: map to app filters (validate direction / date range)
    H->>C: TagByFilter(tag, accID, filters)
    C->>Q: CountByFilter(filters)
    Q-->>C: totalMatched
    C->>S: UpdateTagByFilter(filters, tag)
    S-->>C: updatedCount
    C-->>H: BulkTagResult{Mode: sync, UpdatedCount, TotalMatched}
    alt Mode == sync (always, today)
        H-->>Client: 200 {updated_count, status}
    else Mode == async (declared, NOT implemented)
        H-->>Client: 202 {status: "scheduled background bulk tag job for N"}
    end
```

## Endpoints

| Capability | Method + route | Key inputs | Success |
| --- | --- | --- | --- |
| [009] Query | `GET /cashflow/transactions` | query: `limit/offset`, `sort_by/sort_order`, `q`, `description/note/source`, `direction`, `tags`, `untagged`, `hide_ignored`, `from/to` | 200 `{pagination, data[]}` |
| [023] Manual create | `POST /cashflow/transactions/manual` | body: `transactions[]` (`date`, `amount`, `type`, `description`, `note`, `tag`, `vendor?`) | 201 `{created_count, data[]}` |
| [032] Change date | `POST /cashflow/transactions/date` | body: `id`, `date` | 200 `{id, date}` |
| [010] Monthly | `GET /cashflow/analytics/monthly` | query: `from`, `to`, `include_ignored` | 200 `{data[]}` |
| [010] Tag distribution | `GET /cashflow/analytics/tags` | query: `from`, `to`, `include_ignored` | 200 `{combined, incoming, outgoing}` |
| [011] Tag one | `POST /cashflow/transactions/tag` | body: `id`, `tag` | 200 |
| [011] Tag selection | `POST /cashflow/transactions/tag/selection` | body: `tag`, `ids[]` | 200 `{updated_count, status}` |
| [011] Tag filter | `POST /cashflow/transactions/tag/filter` | body: `tag`, `account_id?`, `filters` | 200 (sync) / 202 (async, dead) |
| [012] Ignore selection | `POST /cashflow/transactions/ignore/selection` | body: `ignored?` (default true), `ids[]` | 200 `{updated_count, status}` |
| [012] Ignore filter | `POST /cashflow/transactions/ignore/filter` | body: `ignored?`, `filters` | 200 |

**Error mapping (common):** decode / validation / `ParseDirection` / bad sort / bad date range /
a date after today → 400; manual create `ErrAccountNotFound` and an unknown transaction id → 404;
manual create with any duplicate, or a date change that would produce one → 409; changing the
date of an imported row → 422; store errors → 500. Tagging a non-existent ID returns 200 (zero
rows updated).

## Processing rules

- **Dedup.** `NewTransaction` computes a SHA-256 `checksum` over description, note, source,
  direction, amount-cents, date (`YYYYMMDD`), a **dedup sequence**, and account ID. Bulk
  inserts use `ON CONFLICT (checksum) DO NOTHING`; `Imported` = rows actually inserted,
  `Duplicates` = the remainder. The sequence is how often identical content occurred inside
  the same import file ([032]) — *not* the CSV row number, which moves when exports overlap —
  so the same logical transaction on a different CSV row is recognised, while two identical
  rows in one file stay two transactions. Manual entries carry no sequence and fall back to
  their generated row number, so manual deduplication is unchanged. Rows imported before this
  change keep their old checksums and are not converted.
- **Manual create.** Capped at 100 rows (`ErrTransactionLimitExceeded`); `tag` is required;
  `source` = `manual` or `manual:<vendor>`; any duplicate in the batch yields a 409. The date
  must be today or earlier — a cashflow transaction records something that already happened.
- **Moving a transaction ([032]).** Only rows whose `source` is `manual` / `manual:<vendor>`
  can be moved, and only to today or earlier. Because the checksum carries the date, it is
  recomputed for the new day and looked up within the account first; a hit that is not the row
  itself is refused rather than written (the unique index would refuse it anyway). Imported
  rows keep their statement date: moving one would change the identity the next import of the
  same file compares against, so it would insert the old row again.
- **Filtering.** `description`/`note`/`source` use case-insensitive `LIKE %v%`; `direction`
  exact; `tags` OR-matched; `untagged` = empty tag; `hide_ignored` = `ignored = false`;
  `import_id` / `ignored_by_rule` exact ([033]); `from`/`to` bound `date`; `q` fuzzy-matches
  description/note/tag. Conditions are AND-joined.
- **Sorting/pagination.** Uses `internal/sorting`; sortable fields are `date` (default, DESC),
  `description`, `note`, `tag`, `source`, `amount`. Offset pagination with default limit 100.
- **Analytics.** Monthly buckets by month (`DATE_TRUNC` / `STRFTIME`) summing in/out and
  `net = incoming − outgoing`. Tag distribution groups by tag (untagged → `"untagged"`) split
  by direction, ordered by total desc. Both exclude `ignored` unless `include_ignored`.

## Events

Cashflow publishes **no** events. Imported rows are written synchronously inside the importer's
cashflow processor (which calls `Commands.CreateMany` with the import ID); cashflow does not
subscribe to `import.completed`.

## Code map

| Path | Responsibility |
| --- | --- |
| `internal/cashflow/commands.go` | Write side: `CreateMany`, `ChangeDate`, tag/ignore commands, `CommandStore`, bulk-tag result/mode |
| `internal/cashflow/queries.go` | Read side: list + analytics, `QueryStore`, sort fields, `ParseTransactionSort` |
| `internal/cashflow/filters.go` | App-level `TransactionFilters` for bulk mutations |
| `internal/cashflow/transaction.go` | `Transaction`, checksum, `CashFlowDirection`/`AccountType`, `CsvParser` |
| `internal/cashflow/errors.go` | Manual-create + account errors |
| `transport/http/handlers/cashflow/*.go` | Query, manual create, change date, analytics, tag (×3), ignore (×2) handlers + DTOs |
| `internal/storage/sqlx_cashflow_store.go` | `SQLXCashflowStore` (both interfaces): inserts, list/count, tag/ignore, analytics SQL |
| `internal/importer/cashflow/` | Import processor + vendor CSV parsers (ING / DeGiro / N26) |

## Gaps / not implemented

- **Async bulk tagging is not implemented.** `TagByFilter`/`IgnoreByFilter` always run
  synchronously and return `Mode: sync`; the HTTP 202 branch and `TagByFilterModeAsync` are
  currently dead. The old `[022]` auto-tagging agent (`TaggerJob`, `agent.enabled`) was
  removed with the `internal/jobs` package; `internal/cashflow/auto_tagger.go` remains as
  orphaned scaffolding with no caller.
- **No cashflow account projection wiring.** `cashflow.accounts` exists in the schema but no
  `account.created` handler or cashflow-account store populates it in the refactored tree.
- **Only the date can be changed after the fact.** Amount, description, tag and direction are
  fixed once a transaction exists, and there is no delete, so correcting anything else means
  re-entering the transaction.
