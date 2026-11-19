---
paths:
  - "web/**/*.svelte"
  - "web/**/*.ts"
  - "web/**/*.css"
---

# SvelteKit frontend (`web/`)

The frontend is `web/`, **not** `apps/web`. Run scripts from inside `web/`, or use the Makefile
`web-*` targets (`make web-dev`, `make web-build`, `make web-install`, `make web-env`,
`make web-lint`), which point at `web/` correctly.

## Stack

| Concern | Choice | Notes |
| --- | --- | --- |
| Framework | SvelteKit 2 + Svelte 5 | Runes mode **forced** for all non-`node_modules` files (`svelte.config.js`) |
| Build | Vite 8 | `@tailwindcss/vite` + `sveltekit()` |
| Language | TypeScript 6 | `strict`, `moduleResolution: bundler`; `<script lang="ts">` everywhere |
| Styling | Tailwind CSS 4, CSS-first | No `tailwind.config.js`; theme declared in CSS via `@theme` |
| Icons | `@iconify/svelte` | String ids (`heroicons:check`), always through the `Icon` atom |
| Catalog | Storybook 10 | addons: svelte-csf, docs, a11y, vitest, Chromatic |
| Tests | Vitest 4, browser mode (Playwright) | Driven **by the stories**; there is no `test` npm script |
| Lint | ESLint 10 flat + Prettier 3 | `prettier-plugin-svelte`, `prettier-plugin-tailwindcss` |

## Commands (from `web/`)

`npm run dev` · `npm run check` (**the type gate** — `svelte-kit sync` + `svelte-check`) ·
`npm run lint` (`prettier --check` then `eslint`) · `npm run format` · `npm run storybook` (:6006) ·
`npm run build`. Adding a story adds test coverage; there is no standalone unit-test runner.

The dev server for previews is the `web` config in `.claude/launch.json` (port 5199).

## Data access

- One service module per feature in `src/lib/services/*` over the `src/lib/api` fetch client:
  `apiGet` / `apiSend` for JSON, `apiUpload` for multipart imports.
- Every service has a **mock branch** (fixtures in `src/lib/data/fixtures`) and a live branch.
  `src/lib/api/config.ts`: `useMocks` is true whenever `VITE_API_URL` is empty, so the app runs
  standalone on fixtures. `make web-env` seeds `web/.env`; set `VITE_USE_MOCKS=false` and
  `VITE_API_URL=http://localhost:6060` to hit the live Go API.
- Add a mock branch for every new service call — leaving one live-only breaks fixture mode.

## Session and account

- The API derives the account from the session cookie, so **no request sends an `account_id`** —
  do not reintroduce one, the backend rejects unknown fields.
- `accountStore` (`src/lib/stores/account.svelte.ts`) resolves the signed-in account from
  `GET /auth/me`; use it for `isAdmin`, the displayed email, and `signOut()`. `GET /accounts` is
  admin-only and is not part of the bootstrap path.
- Every request sends `credentials: 'include'`. A 401 from any request routes to `/login` via
  `setUnauthorizedHandler`; the root layout awaits the session and guards every route. `/login` is
  the one route that renders without a session.

## Errors

Handle errors explicitly — never silently swallow. Surface user-relevant failures in the UI
(the `alert` molecule / `toast` store), and keep messages specific enough to act on.

## Theming

`src/app.css` is the canonical theme for both the app and Storybook: `@theme` tokens
(`--font-sans` Noto Sans, `--font-display` / `--font-heading` EB Garamond), the full `taupe` scale
(50–900), and a base layer setting `bg-taupe-100` and serif headings. Webfonts in `static/fonts/`.
The app loads it through `src/routes/+layout.svelte` → `src/routes/layout.css` → `src/app.css`;
Storybook imports `src/app.css` directly. **Do not create a second theme** in `layout.css` or in
story files. Prettier's `tailwindStylesheet` (class sorting) points at `src/routes/layout.css`.

## Known formatting drift

Most files under `src/lib/components/atoms/` are authored with **2-space** indentation while
Prettier is configured for **tabs**, so `prettier --check` flags them and `npm run lint` is **not
green out of the box**. `npm run check` is the gate that is green. Match a file's existing style
when editing it; do **not** mass-reformat the component tree as a drive-by — fix formatting only as
its own intentional change.
