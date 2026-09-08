# CLAUDE.md

Repo-wide guidance. Area-specific conventions live in [.claude/rules/](.claude/rules/) and load
automatically when you touch matching files — Go backend, HTTP transport, storage, migrations,
event bus, Atomic Design, SvelteKit. Read the matching rule before changing an area; don't
duplicate its content here.

## Repository shape

Monorepo, two apps:

- `apps/api` — Go backend (module `github.com/lennardclaproth/my-finances-tracker`, single-module
  `go.work`). Almost all logic lives here. Single entrypoint `cmd/my-finances-tracker/main.go`.
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
