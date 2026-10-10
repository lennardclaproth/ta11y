# [020] Web App

> Detail for the **Web app** entry in [FEATURES.md](FEATURES.md). Read that first for what the
> app does; this covers how it is put together and what is still missing.

## Overview

`web/` (not `apps/web`) is a SvelteKit 2 / Svelte 5 app in forced runes mode, built with Vite
and Tailwind 4, catalogued in Storybook. It talks to the Go API over REST plus a per-account
WebSocket, and runs standalone against fixtures when no API URL is configured.

## Component architecture

Strict Atomic Design under `web/src/lib/components/<tier>/`, importing only from the same or a
lower tier. Populated today: **atoms** (23), **molecules** (24), **organisms** (16),
**templates** (2 — `app-shell`, `page-content`). The `pages/` tier is not created yet; route
components live in `src/routes/` instead.

### Notifications

One molecule reports everything: `molecules/notice-band` (`NoticeBand`), a ruled band in normal
document flow. `organisms/notice-band-region` renders the `toast` store as bands under the
masthead; the store keeps at most three notices and drops the oldest for a newer one, so nothing
counts down unseen. Its API, timers and intents are otherwise unchanged. A dialog
or drawer covers that region with its own scrim, so an overlay that stays open reports its own
outcomes in a band inside it (`ProviderCatalogueDrawer`). A problem that stays is rendered
by the component that owns the content (`DataTable` on a failed load, with an optional `onRetry`);
a refused save is rendered inside its dialog, drawer or form, and never also as a toast. Per-field
errors stay in `FormField`, carrying the same icon as the band.

The rules are enforced by convention, not tooling — see
[.claude/rules/atomic-design.md](../../../../.claude/rules/atomic-design.md) for the tier
boundaries, the four-file component anatomy, and the shared prop vocabulary.

## Route map

| Route | Purpose |
| --- | --- |
| `/` | Redirects to `/cashflow` (`+page.ts`) |
| `/login` | The one route that renders without a session |
| `/cashflow` | Analytics, transactions table, tagging, manual entry |
| `/portfolio` | Positions, snapshots, manual transactions, rebuild |
| `/assets` | Asset classes and worth tracking |
| `/admin/listings` | Listing CRUD + provider catalogue drawer |
| `/admin/dailies` | End-of-day prices per listing + manual price upload |
| `/admin/credentials` | Provider API keys |

`/admin/*` sits behind a layout that renders nothing until the session resolves, then shows an
unavailable notice to non-admins. That hides screens; the API enforces the `admin` flag itself
and answers 403 regardless.

## Where controls live

Every page follows one arrangement, so a control's place says what it affects:

| Region | Component | Holds |
| --- | --- | --- |
| Top bar | `organisms/top-navbar` | The wordmark and the destinations. Nothing that changes the page. |
| Your overview | `organisms/account-overview` | Account, net worth with its change, the one period, sign out. Inline from `lg`, behind one entry below it. |
| Page heading | `templates/page-content` (`title`) | The page's own `h1`, on the page rather than in the bar. |
| Chart section | `molecules/analytics-card` (`actions`) | Actions that recompute the chart, such as rebuilding the portfolio. |
| Ledger header | `organisms/ledger-toolbar` | Everything that acts on the rows: searching, row filters, adding, importing, selection actions. |

The period is one app-wide value in `stores/period.svelte.ts`, with eight presets (`1M … Max`)
plus a hand-picked range, chosen in the account overview and read by whichever page is open. It
survives navigating between pages; a URL that carries `from`/`to` seeds it once, before anything
else has set it. Admin pages have no charts and no period and simply never read it.

Net worth comes from `GET /assets/snapshots` through `stores/net-worth.svelte.ts`, which fetches
the series once. The displayed worth is the latest snapshot and the change is the first against
the last snapshot inside the period — no new calculation. Fewer than two snapshots in the period
is reported in words rather than as a zero, and the oldest snapshot is what gives `Max` a real
start date.

## Data access

One service module per feature in `src/lib/services/*` over the `src/lib/api` fetch client
(`apiGet` / `apiSend` / `apiUpload`). **Every service has a mock branch** backed by fixtures in
`src/lib/data/fixtures`, selected by `useMocks` when `VITE_API_URL` is empty — so the app runs
with no backend. A service added without its mock branch breaks fixture mode.

`connectRealtime` subscribes to `/ws/accounts/{id}` and debounces a refresh callback. It no-ops
on the server and in mock mode. Note that EOD imports raise no event: they carry no account
scope, so the hub skips them.

State lives in four runes stores: `account.svelte.ts` (session, `isAdmin`, sign-out),
`toast.svelte.ts`, `period.svelte.ts` (the one app-wide period) and `net-worth.svelte.ts` (the
total-worth snapshots behind the account overview).

## Code map

| Path | Responsibility |
| --- | --- |
| `src/routes/*` | Pages, root layout, admin guard layout |
| `src/lib/components/*` | Atomic Design tiers |
| `src/lib/services/*` | One module per feature, each with a mock branch |
| `src/lib/api/*` | Fetch client, config, money helpers, shared DTOs |
| `src/lib/data/fixtures/*` | Fixture data backing mock mode |
| `src/lib/stores/*` | `accountStore`, `toast`, `periodStore`, `netWorthStore` |
| `src/app.css` | Canonical theme, loaded by both the app and Storybook |
| `.storybook/*` | Storybook config; stories double as the Vitest browser test corpus |

## Gaps / not implemented

- **No `pages/` Atomic tier** — route components carry page composition directly.
- **No admin accounts screen.** `GET`/`POST /accounts` are admin-only and have no UI, so the
  `admin` flag can only be set on the bootstrapped account.
- **The period is not persisted.** It is one in-memory value, so it survives navigating between
  pages but not a reload, which starts at year-to-date again.
- **Net worth is only as fresh as the snapshots.** They are rebuilt when a portfolio rebuild
  completes, so between rebuilds the overview reports the last computed figure.
- **No standalone test runner.** Stories are the test corpus under Vitest browser mode; there
  is no `test` script and no non-story frontend tests.
- **`npm run lint` is not green out of the box** — files under `components/atoms/` use 2-space
  indent while Prettier is configured for tabs. `npm run check` is the gate that is green.
