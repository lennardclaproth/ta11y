# CLAUDE.md

Repo-wide guidance. Area-specific conventions live in [.claude/rules/](.claude/rules/) and load
automatically when you touch matching files — Go backend, HTTP transport, storage, migrations,
event bus, Atomic Design, SvelteKit. Read the matching rule before changing an area; don't
duplicate its content here.

## Repository shape

Monorepo, two apps:

- `apps/api` — Go backend (module `github.com/lennardclaproth/ta11y`, single-module
  `go.work`). Almost all logic lives here. Single entrypoint `cmd/ta11y/main.go`.
- `web` — SvelteKit 2 / Svelte 5 frontend. The directory is **`web/`, not `apps/web`**.

## Commands

The root `Makefile` is the source of truth (`make help`); [README.md](README.md) has the
end-to-end run guide. Don't invent an alternative when a target exists.

- `make build` / `make run` — build/run the API (port from `config.yaml`).
- `make dev` — API with `air` hot reload. `make install-tools` installs `swag`, `goose`, `air`, `golangci-lint`.
- `make test` — unit (`go test ./apps/api/...`). `make test-integration` — `integration`-tagged suite. `make test-all` — both.
- `make lint` — `go fmt` + `go vet` + `golangci-lint` (cd's into `apps/api`).
- `make swagger` — regenerate OpenAPI into `apps/api/docs`. Never hand-edit generated files there.
- `make db-up` / `make db-down` — local Postgres via `deploy/docker/compose.dev.yaml`
  (`make db-up-all` adds Elasticsearch/Kibana/APM). The API auto-creates the DB and migrates on start.
- `make migrate-up` / `make migrate-down` / `make migrate-create name=<n>` — goose (needs `DATABASE_URL`).
- `make web-install` / `make web-dev` (:5199) / `make web-build` / `make web-env` / `make web-lint`.

Single Go test: `go test -run TestName ./internal/<pkg>/...` from `apps/api`.
Frontend gate: `npm run check` from `web/`.

**Local config:** the server reads `apps/api/config.yaml` relative to cwd — run it from `apps/api`
— plus `apps/api/.env` (`make env` seeds it). The frontend reads `web/.env` (`VITE_API_URL`,
`VITE_USE_MOCKS`); with no `VITE_API_URL` it runs on mock fixtures. Provider API keys
(`MARKETSTACK_API_KEY`, `ALPHA_VANTAGE_API_KEY`) come from env only.

## Change discipline

- Minimal, focused diffs. Don't refactor, rename, or modernize unrelated code. Prefer editing an
  existing file over adding a new abstraction unless there's clear benefit.
- Clean up only what the current change introduced, plus any scratch files you created.
- **Requires explicit approval:** adding/removing/upgrading dependencies; modifying existing
  migrations; touching Docker, CI, or infra files; deleting documentation.
- For complex multi-step work keep a concise `tmp/plan.md` and delete it when done.

## Validation

Run `make test` after code changes; `make lint` and `make build` when finalizing; `make swagger`
for API-contract changes. State plainly which validations you ran and which you skipped — never
imply a command was run when it wasn't, and don't call work complete without the relevant checks.

## Documentation

- Feature docs are consolidated in `apps/api/docs/features/FEATURES.md` — one short overview of
  every feature. The legacy per-feature `NNN_NAME.md` files are gone; don't recreate them. Read it
  before changing a documented feature and update its entry in the same task when behaviour changes.
- Update the root `CHANGELOG.md` under `Unreleased` with Keep a Changelog categories
  (`Added`/`Changed`/`Fixed`) and the `[NNN]` feature ID. New IDs come from the changelog sequence
  (highest today: `030`). Skip the changelog for refactor-only or tooling-only changes.
- If a change can't be mapped confidently to an existing feature ID, stop and ask.

## Gotchas

- A bare `go build ./...` from a linked git worktree needs `GOWORK=off` (the repo `go.work`
  resolves to the main checkout). `make build`/`make run`/the VS Code launch config are unaffected.
- `npm run lint` in `web/` is **not** green out of the box (tabs-vs-spaces drift under
  `components/atoms/`). `npm run check` is the gate. Don't mass-reformat as a drive-by.
- The API derives the account from the session cookie; no request carries an `account_id`.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **ta11y** (5362 symbols, 14736 relationships, 377 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact before editing.** Use `impact({target: "symbolName", direction: "upstream"})` or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .`; report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- MUST warn on HIGH/CRITICAL `risk` pre-edit; never use `riskSharedAxes` to waive a HIGH/CRITICAL `risk` warning. Compare File/symbol: MCP File omits axes; Graph-RAG expands File.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- **MUST use `query({search_query: "concept"})` for concepts/flows, `context({name: "symbolName"})` for a named symbol, or `impact` for blast radius, on read-only callers, dependencies, imports, or execution flow.** Graph first; text search only for empty/`UNKNOWN`/literals.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/ta11y/context` | Codebase overview, check index freshness |
| `gitnexus://repo/ta11y/clusters` | All functional areas |
| `gitnexus://repo/ta11y/processes` | All execution flows |
| `gitnexus://repo/ta11y/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
