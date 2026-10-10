# [034] Daily-priced asset holdings

> **Feature ID:** 034 · **Area:** Core (user-facing) · **Status:** implemented
>
> **Backend packages:** `internal/assets` · `internal/marketdata` · `internal/marketdata/alphavantage` · `transport/http/handlers/assets` · `transport/http/handlers/marketdata` · `internal/storage/sqlx_assets_store.go`
>
> **Frontend:** `web/src/routes/assets/[classId]` · `web/src/lib/components/molecules/holding-performance-row` · `web/src/lib/components/organisms/add-asset-item-dialog` · `web/src/lib/components/organisms/asset-class-drawer`
>
> **Related features:** [024] Asset management · [005] Market data · [017] Event-driven messaging

## Overview

An asset item either carries a worth the user sets by hand ([024]) or is linked to a
market-data listing, in which case its worth is **derived** each day from the quantity held
times that day's closing price. The user records **purchases** (date, quantity, unit price);
everything else — the item's current worth, its history back to the first purchase, and its
contribution to net worth — follows from them.

Only instruments quoted in euro can be linked, because the app converts no currencies. Alpha
Vantage supplies crypto in euro; gold and silver have no euro source and remain manual items.

## Domain model

- **`assets.Asset.ListingID`** (nullable) is the link. `NULL` reads as "manual", which is how
  every pre-existing item behaves. `Asset.IsDailyPriced()` is the one predicate the rest of the
  feature branches on.
- **`assets.Purchase`** — one acquisition: account, item, day (UTC), quantity (float, because
  crypto is held in fractions), unit price (`money.Price`). There is no sell counterpart, so the
  quantity held on a day is the sum of the purchases up to it.
- **`marketdata.Quote`** — one instrument offered as a daily-priced holding: symbol, name, kind,
  currency, and whether it is `Selectable`. An unselectable quote carries a `Reason` fit for
  display. `ListingID`/`Price` are nil until somebody tracks the instrument.

## Lifecycle

1. **Pick** — `marketdata.Quotes.Search` merges the catalogues of the sources that can fetch a
   price with a static table of blocked instruments. No provider request is made.
2. **Adopt** — `CreateDailyPricedAsset` resolves the symbol through `Quotes.Track`, which reuses
   the global listing if one exists and otherwise creates it with `DeferPriceSync`. The item is
   stored with zero worth and **no opening mutation**: unlike a manual item its history is not
   entered, it is derived.
3. **Derive** — the command publishes `assets.snapshots.rebuild.requested`. The handler runs
   `HoldingsSyncer.SyncAccount` *before* `Builder.RebuildAll`, because the syncer writes the
   mutations the snapshot rebuild reads.
4. **Read** — `Queries.ClassDetails` reports holdings apart from manual items;
   `Queries.HoldingDetails` adds the purchases.

Adding a purchase repeats steps 3–4.

## Processing rules

`planHoldingMutations` (`internal/assets/holdings.go`) walks every day from the earliest
purchase in the account to today and emits one **SET** mutation per daily-priced item per day:

- **Valuation** — quantity held that day × the last known price. A day the provider has no price
  for reuses the previous day's, which covers weekends, holidays and a provider that is behind.
  Before any price is known the item is valued at what was paid for it, so a purchase entered
  ahead of the price feed never reads as worth nothing.
- **Class total** — a mutation records the class total the snapshot builder reads, and a class
  can hold manual items next to linked ones. The total written is every daily-priced item in the
  class that day **plus** the manual worth carried forward to it; otherwise a price move would
  silently erase the manual items beside it.
- **Idempotence** — the derived rows are deleted and rewritten wholesale on every run
  (`DeleteDerivedMutations`), so the rebuild is safe to repeat.
- **Price window** — prices are read through `marketdata.Queries.GetEODByListing` from the item's
  first purchase onwards, which both bounds the provider fetch and refreshes a stale listing on
  the way. A listing that cannot be read leaves the holding at what was paid rather than failing
  a rebuild the rest of the account depends on.

A linked item refuses `UpdateAssetWorth` (`ErrAssetDailyPriced`): a hand-set worth would be
overwritten by the next rebuild.

## Price source

`internal/marketdata/alphavantage` is the first fetcher for the Alpha Vantage provider, which
existed with credentials and quota but no way to pull data. It implements
`marketdata.EODFetcher` over `DIGITAL_CURRENCY_DAILY` with `market=EUR` — one request per symbol
for the whole history, which matters on a free tier of a few dozen requests a day. The provider
answers errors and quota refusals with HTTP 200 and a message field, so the body is inspected
rather than the status; the token is booked either way to keep the local count honest.

`currencies.go` holds the supported crypto table and the `CODE/EUR` symbol convention.

## Events

No new topics. The feature reuses `assets.snapshots.rebuild.requested` (published by the assets
write side) and `assets.snapshots.rebuilt`.

## Code map

| Concern | File |
| --- | --- |
| Purchase, validation, read models | `internal/assets/holding.go` |
| Derivation of worth per day | `internal/assets/holdings.go` |
| Write side (create, add purchase, manual guard) | `internal/assets/commands.go` |
| Read side (class holdings, holding details) | `internal/assets/queries.go` |
| Instrument catalogue and adoption | `internal/marketdata/quotes.go` |
| Alpha Vantage fetcher + supported crypto | `internal/marketdata/alphavantage/` |
| Purchases, derived/manual mutations, linked items | `internal/storage/sqlx_assets_store.go` |
| Schema (`items.listing_id`, `assets.purchases`) | `apps/api/migrations/{postgres,sqlite}/20261010130000_asset_holdings.sql` |
| HTTP | `transport/http/handlers/assets/holding.go` · `transport/http/handlers/marketdata/quotes.go` |
| Holdings sync before snapshot rebuild | `transport/messaging/handlers/assets/snapshots_rebuild_requested.go` |

### Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/marketdata/quotes` | Instruments that can back a holding, blocked ones included with a reason |
| `POST` | `/assets/holdings` | Add a daily-priced item with its first purchase |
| `POST` | `/assets/{asset_id}/purchases` | Record another acquisition |
| `GET` | `/assets/holdings/{asset_id}` | One holding with its purchases and value-against-paid series |

`GET /assets/classes/{class_id}` gained a `holdings` array alongside `assets`.

## Gaps / not implemented

- **No sells.** A holding only grows. Disposing of part of one has no representation.
- **No currency conversion**, by decision. Gold and silver are listed in the picker as blocked
  rows with the reason; they stay manual items until a euro source exists.
- **No conversion of an existing manual item** into a linked one, and no unlinking.
- **The supported crypto table is static** (`alphavantage/currencies.go`). The provider publishes
  a far longer list as a CSV download; adding a currency is a one-line change here instead.
- **No costs, interest, staking income or realized profit**, and no price alerts.
- **The derived rebuild is per account and whole-history.** It rewrites every derived mutation
  for the account on each request rather than updating the days that changed.
- **Adoption is not atomic with the item.** `Quotes.Track` creates the listing before the
  transaction that stores the item and its first purchase, so a failure inside that transaction
  leaves an unreferenced listing behind. Listings are global and meant to be reused by the next
  adopter, so the row is idle rather than wrong.
- **First-time adoption of a symbol races.** `Track` reads the listing and creates it if absent
  without a lock; two accounts adopting the same instrument at the same instant can both miss,
  and the loser's `ErrListingAlreadyExists` surfaces as a 500 instead of re-reading the winner's
  listing.
- **Listing deletion is not guarded.** `items.listing_id` carries no foreign key (listings are
  global reference data, items are account data), so a deleted listing leaves holdings valued at
  what was paid, which the read side reports as a missing symbol rather than hiding.
