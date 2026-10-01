# [032] Brokerage import

> **Feature ID:** 032 · **Area:** Core (user-facing) · **Status:** implemented
>
> **Backend packages:** `internal/importer` · `internal/cashflow` · `internal/portfolio` · `internal/marketdata` · `transport/http/handlers/importer`
>
> **Frontend:** `web/src/lib/components/organisms/import-dialog` · `web/src/lib/components/molecules/activity-trail` · `web/src/lib/services/importer.ts`
>
> **Related features:** [002] CSV imports · [009] Cashflow · [013] Portfolio · [005] Market data · [001] Realtime updates

## Overview

A monthly broker account statement is one file with two audiences: the money movements
belong in Cashflow, the trades belong in Portfolio. This feature makes that one
handling — an import panel reachable from both pages — and makes repeating it safe,
because consecutive exports overlap.

## Domain

**One upload, two destinations.** The panel posts the same file to
`POST /imports/cashflow` and `POST /imports/portfolio`. They stay two independent
imports on purpose: each has its own parser, its own lifecycle and its own counters,
and one can be refused while the other succeeds. The panel polls both and reports them
side by side. Nothing is rolled back when only one half lands; re-uploading is safe.

**Overlap is recognised on content, not position.** Deduplication still keys on the row
`checksum`, but the checksum no longer contains the CSV line number — in next month's
export every row sits on a different line. It contains a *dedup sequence* instead: how
often that exact content occurred inside the same file (`importer.DedupSequencer`). A
row carried over from the previous export produces the same checksum and is counted as
already imported; two genuinely identical rows in one export produce sequences 1 and 2
and stay two transactions. The line number is still stored and still orders rows that
share a day.

The content the sequence is taken over includes the broker's own order reference where
the export carries one (DEGIRO's `Order Id` — on the cashflow side it already travels
in the row's note, on the portfolio side through the non-persisted
`TransactionData.ExternalRef`).

**Existing transactions are not converted.** The checksum formula changed, so rows
imported before this feature keep their old checksums. They are not re-detected, and
nothing migrates them.

**Linking is exact.** Every portfolio rebuild re-links positions to listings, so adding
a listing and running **Rebuild portfolio** links transactions that were imported
earlier. The match is `marketdata.Queries.ListingByIdentity`: the exact ISIN, or the
exact symbol when the instrument has no ISIN. A present-but-unknown ISIN does **not**
fall back to the symbol, and no match leaves the position unlinked rather than
attaching it to a near neighbour. An unlinked position is still a position; it just
contributes no market value.

**Products without a listing are named.** A completed portfolio import reports the
distinct instruments it brought in that no listing matches, so they are visible instead
of silently missing from performance. The list is derived at read time from the
import's transactions, so it is always current.

## Reading an import

`GET /imports/{import_id}` returns one import's status, its counters
(`total_rows` / `imported` / `duplicates` / `failed`), a classified `reason`, and —
for a completed portfolio import — `unlinked_products`. It is scoped to the session's
account: another account's import, and an import with no account (EOD uploads are
listing-scoped), are reported as `404`.

`reason` is `file_not_recognised` when the vendor's parser refused the whole file,
which happens when the headers are not the ones that export carries. That deserves a
different answer than a generic failure — a different file, not a retry — so the panel
says so and reports that nothing was changed.

The stored upload is removed once an import reaches a terminal state: the rows have
been taken over by the target feature and nothing reads the file again.

## Panel

The panel is `organisms/import-dialog`: a two-step dialog whose form (account +
file) becomes a result. Its `molecules/activity-trail` shows the four steps —
File → Cashflow → Portfolio → Performance — as a route before anything runs, then
fills them in one by one. Reachable from **Import CSV** on Cashflow and on Portfolio;
both open the same dialog.

The performance step reports that the rebuild *started*, not that it finished: the
rebuild is kicked off by the completed portfolio import ([013]) and runs on its own, so
the panel does not claim a finish it cannot observe.

## Code map

| Path | Responsibility |
| --- | --- |
| `internal/importer/sequence.go` | `DedupSequencer` — the occurrence count that replaced the line number |
| `internal/importer/queries.go` | Read side: one import's result, `FailureReason` |
| `internal/importer/{cashflow,portfolio}/processor.go` | Stamp the sequence; classify an unreadable file |
| `internal/cashflow/transaction.go`, `internal/portfolio/transaction.go` | Checksum over content + sequence |
| `internal/marketdata/queries.go` | `ListingByIdentity` — exact ISIN, else exact symbol |
| `internal/portfolio/builder.go` | Position → listing linking on every rebuild |
| `internal/portfolio/queries.go` | `UnlinkedProducts` for one import |
| `transport/http/handlers/importer/get.go` | `GET /imports/{import_id}` |
| `web/src/lib/components/organisms/import-dialog/` | The panel |
| `web/src/lib/services/importer.mock.ts` | Fixture-mode stand-in for the pipeline |

## Gaps / not implemented

- **Only DEGIRO.** ING and N26 have cashflow parsers but no portfolio parser, so the
  panel offers brokerage vendors only.
- **No import history.** The result is not retrievable once the panel is closed —
  `GET /imports/{import_id}` needs the id, and nothing lists past imports.
- **No post-processing.** Imported rows cannot be edited, merged, or rolled back.
- **Listings stay manual.** Nothing is created automatically, and there is no screen to
  pick a listing per transaction; the route is Listings → Rebuild portfolio.
- **No prices from the export.** A linked listing gets its history from its provider;
  nothing is derived from the statement.
