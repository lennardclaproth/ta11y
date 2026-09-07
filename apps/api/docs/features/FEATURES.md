# Features

A concise overview of what My Finances Tracker does. Each entry is the application
capability, not its implementation; see the code under `apps/api/internal/<feature>`
and `apps/api/transport` for details.

Each capability links to a **detailed feature doc** (with class, ERD, state-machine, and
sequence diagrams) in this same `features/` directory. Tightly-related capabilities share a
doc (e.g. all cashflow and all market-data features).

## Core (user-facing)

### Web portal [020]
Cashflow, assets, portfolio, and admin pages share ta11y's editorial shell: a ruled masthead,
visible desktop navigation, a labeled mobile menu, serif headings, flat analytics sections,
and square paper table surfaces. Narrow screens scroll through analytics to a dedicated table
region; wide tables scroll horizontally without widening the page. Existing financial actions,
filters, date selection, and account controls remain available. Cashflow, Portfolio and Assets
place labeled creation actions with plus icons above the ledger, alongside its title or view tabs.

### Account management
Users create and list accounts. An account is the shared scope that imports, cashflow,
portfolio, and assets all hang off of.

→ Details: [Account management](%5B003%5D_ACCOUNT_MANAGEMENT.md)

### CSV imports
Users upload a vendor CSV for an account and it is accepted immediately, stored durably,
and processed asynchronously. Three explicit import types are supported — cashflow,
portfolio, and end-of-day market data — with vendor-specific parsers (DeGiro, ING, N26
for cashflow; DeGiro for portfolio; BrandNewDay for EOD).

→ Details: [CSV imports](%5B002%5D_CSV_IMPORTS.md)

### Cashflow insights
Users query bank/payment transactions with filtering, sorting, and pagination, and add
manual transactions. They can view monthly and per-tag analytics, tag transactions
(individually, by selection, or by filter), and ignore/unignore transactions to exclude
them from totals.

→ Details: [Cashflow](%5B009%5D_CASHFLOW.md)

### Portfolio management
Users read portfolio snapshots, current positions, and transaction history, add manual
portfolio transactions, and request asynchronous portfolio rebuilds that recompute
positions and snapshots from transactions and market data.

→ Details: [Portfolio](%5B013%5D_PORTFOLIO.md)

### Asset management
Users manage account-scoped asset classes and the items within them, set or adjust each
item's worth, inspect class details and mutations, and read total net-worth snapshots
across all classes.

→ Details: [Assets](%5B024%5D_ASSETS.md)

### Asset sync
When a portfolio rebuild completes, the computed portfolio worth is synced into a
read-only "portfolio" asset class and the total asset snapshots are rebuilt, so net worth
always reflects the latest portfolio value.

→ Details: [Assets](%5B024%5D_ASSETS.md)

### Realtime updates (WebSocket)
Clients subscribe to an account-scoped WebSocket (`/ws/accounts/{account_id}`) and receive
push notifications when long-running work finishes — import completed, portfolio rebuilt,
and asset snapshots rebuilt — so the UI can refresh without polling.

→ Details: [Realtime updates](%5B001%5D_REALTIME_UPDATES.md)

## Market data & admin

Admin screens (listings, dailies, provider credentials) are shown only to accounts flagged
`admin`; the bootstrapped account is one. Non-admin accounts do not see them in the
navigation and get an unavailable notice on a direct link. The API does not enforce the
flag, so this hides screens rather than protecting them -- it needs real authentication
before the API is reachable beyond localhost.

### Listing management
Admins create, update, list, and search market-data listings. Listings are the canonical
instruments referenced by portfolio transactions and end-of-day pricing. In the web app,
open **Listings** from the navigation, enable admin mode when prompted, then choose
**Add listing**. The form accepts a name, symbol and supported data source, with optional
instrument metadata. It preserves input on errors and identifies duplicate symbol/source
pairs. Demo-mode additions remain available to lists and searches until a page reload.

Listing search covers both the instruments you track and a locally cached copy of the
provider's ticker catalogue, selected with `scope` (`tracked` by default, plus `catalogue`
and `all`); tracked rows always sort first and each row reports whether it is `tracked`
and whether it can be adopted.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### Provider catalogue search
**Listings → Browse catalogue** searches the provider's instrument catalogue and adds
entries to your listings in bulk. Searching the local cache is free; **Search provider**
makes one metered provider request and merges the results into the cache.

The cache is filled two ways because it can never be assumed complete: the provider
matches substrings across its whole universe, so a search for `ASM` has far more matches
(125) than `ASML` (31), and cached results for one query say nothing about another. A
bounded seed run therefore caches the provider's most-traded instruments — its catalogue
comes back in popularity order, so a small page budget goes a long way — and each explicit
provider search tops the cache up for the terms you actually use. A provider search fetches
a single page: results are relevance-ranked, so the intended instrument is on the first
page, and the UI says so when the provider matched more than it returned.

Rows the provider gives no name for, or no end-of-day history for, are shown but cannot be
adopted — a listing requires a name, and one without price history would never yield
prices. Adopting does not fetch price history; that happens the first time the listing's
prices are viewed, so adding a batch cannot stall the request or drain the request budget.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### EOD market data
Admins retrieve end-of-day price history, upload manual EOD price files, and check the
processing status of an upload.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### Provider support
The API supports both API-based market-data providers (MarketStack, Alpha Vantage) and a
manual provider; the configured provider mode controls listing synchronization and whether
manual EOD uploads are allowed.

### Provider credentials
**Credentials** (admin) lists each provider with its base URL, request quota and whether an
API key is configured, and lets you set or replace a key. Stored keys are never included in
the listing -- only a hint like `****9876` -- so reading one back is a separate, logged
action. Manual providers ingest uploaded files and have no credentials to configure.

Keys supplied through the environment are seeded at startup, so a key that is still set in
the environment is re-added on restart alongside anything changed here.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### Vendor listing
Lists the active import vendors along with capability metadata, so clients can show which
vendors support CSV imports and which are disabled.

→ Details: [Vendor listing](%5B004%5D_VENDORS.md)

## Platform

### Event-driven messaging
An in-memory event bus fans domain events out to many features — for example, creating an
account projects it into portfolio, cashflow, assets, and importer, and completion events
bridge into the realtime WebSocket notifications.

→ Details: [Event-driven messaging](%5B017%5D_EVENT_DRIVEN_MESSAGING.md)

### Observability
Every request carries an identifier and structured logs with trace propagation and
sensitive-data redaction, wired through Elastic APM across HTTP and SQL.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Database bootstrap & migrations
The service runs goose migrations (mirrored for Postgres and SQLite) and seeds baseline
records on startup before serving traffic.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Health & API docs
A lightweight `/health` endpoint reports liveness, and Swagger/OpenAPI publishes the
generated API contract with an interactive UI at `/swagger/`.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Web app
A SvelteKit (Svelte 5) frontend in `web/`; it is early-stage — currently an Atomic-Design
component library built in Storybook, not yet consuming the API — and most application
logic lives in the Go backend.

→ Details: [Web app](%5B020%5D_WEB_APP.md)
