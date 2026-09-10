# [020] Web App

> Detail for the **Web app** entry in [FEATURES.md](FEATURES.md). Read that first for what the
> app does; this covers how it is put together and what is still missing.

## Overview

`web/` (not `apps/web`) is a SvelteKit 2 / Svelte 5 app in forced runes mode, built with Vite
and Tailwind 4, catalogued in Storybook. It talks to the Go API over REST plus a per-account
WebSocket, and runs standalone against fixtures when no API URL is configured.

## Component architecture

Strict Atomic Design under `web/src/lib/components/<tier>/`, importing only from the same or a
lower tier. Populated today: **atoms** (23), **molecules** (24), **organisms** (15),
**templates** (2 — `app-shell`, `page-content`). The `pages/` tier is not created yet; route
components live in `src/routes/` instead.

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

## Data access

One service module per feature in `src/lib/services/*` over the `src/lib/api` fetch client
(`apiGet` / `apiSend` / `apiUpload`). **Every service has a mock branch** backed by fixtures in
`src/lib/data/fixtures`, selected by `useMocks` when `VITE_API_URL` is empty — so the app runs
with no backend. A service added without its mock branch breaks fixture mode.

`connectRealtime` subscribes to `/ws/accounts/{id}` and debounces a refresh callback. It no-ops
on the server and in mock mode. Note that EOD imports raise no event: they carry no account
scope, so the hub skips them.

State lives in two runes stores: `account.svelte.ts` (session, `isAdmin`, sign-out) and
`toast.svelte.ts`.

## Code map

| Path | Responsibility |
| --- | --- |
| `src/routes/*` | Pages, root layout, admin guard layout |
| `src/lib/components/*` | Atomic Design tiers |
| `src/lib/services/*` | One module per feature, each with a mock branch |
| `src/lib/api/*` | Fetch client, config, money helpers, shared DTOs |
| `src/lib/data/fixtures/*` | Fixture data backing mock mode |
| `src/lib/stores/*` | `accountStore`, `toast` |
| `src/app.css` | Canonical theme, loaded by both the app and Storybook |
| `.storybook/*` | Storybook config; stories double as the Vitest browser test corpus |

## Gaps / not implemented

- **No `pages/` Atomic tier** — route components carry page composition directly.
- **No admin accounts screen.** `GET`/`POST /accounts` are admin-only and have no UI, so the
  `admin` flag can only be set on the bootstrapped account.
- **No cashflow or portfolio CSV import UI.** `importCashflow` and `importPortfolio` exist in
  the service layer with no caller; only the EOD upload is wired.
- **No standalone test runner.** Stories are the test corpus under Vitest browser mode; there
  is no `test` script and no non-story frontend tests.
- **`npm run lint` is not green out of the box** — files under `components/atoms/` use 2-space
  indent while Prettier is configured for tabs. `npm run check` is the gate that is green.
