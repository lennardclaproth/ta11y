---
paths:
  - "apps/api/**/*.go"
---

# Go backend conventions (`apps/api`)

Go module `github.com/lennardclaproth/my-finances-tracker`, single-module `go.work`.
Single entrypoint: `cmd/my-finances-tracker/main.go`. The 2026 restructure has landed —
`go build ./...` and `go vet ./...` are green, `make build` / `make run` / `make test` work.
Do not treat the codebase as mid-refactor and do not leave known breakage behind.

## Package-by-feature

Feature packages under `internal/`: `account`, `auth`, `cashflow`, `portfolio`, `assets`,
`marketdata`, `vendor`, `importer`, `files`. Supporting: `storage`, `eventbus`, `notify`,
`bootstrap`, `logging`, `observability`, `config`. Shared utilities: `money`, `date`, `sorting`.

Keep feature packages coherent — new behaviour goes in the package that owns the concept.

## Application boundary = `Commands` + `Queries`

- `Commands` = writes: mutate state, publish events. `Queries` = reads.
- They depend on **small feature-owned interfaces** (`creator`, `queryStore`, …) that
  `internal/storage` implements directly.
- Do **not** add adapter interfaces, reflection bridges, or duplicate domain types to avoid
  touching the real boundary. Change the boundary instead.
- Do not add a local transport-side interface when a concrete `Commands`, `Queries`, or feature
  service already expresses the dependency.
- Other collaborators (builders, syncers, processors, services) are fine where that is the
  established nearby pattern.
- `bootstrap` saves through the owning feature package (`account`, `vendor`, `marketdata`),
  never through new storage-level startup wiring.

## Reuse before writing

Before adding a helper, utility, abstraction, validator, mapper, error type, logger wrapper, or
reusable type, search `apps/api` for the established equivalent. In particular:

- date parsing / ranges / formatting → `internal/date`
- currency, prices, decimal parsing and formatting → `internal/money`
- sort fields and directions → `internal/sorting`

Feature-level errors belong in that package's `errors.go` when one exists.

## Logging

Use the `internal/logging` slog wrapper with a deliberate level (debug/info/warn/error).
Never log credentials, session tokens, provider API keys, or email addresses beyond what an
existing call site already logs. Preserve the surrounding logging style; don't make broad
observability changes unless asked.

## Comments and docs

Document exported types and functions. Document unexported ones when purpose, invariants, or
behaviour are not obvious. No comments that restate the code, and **never numbered step comments**.
Keep comments true when behaviour changes.

## Tests

Add or update tests when behaviour changes; unit tests first, integration tests
(`apps/api/test`, `integration` build tag) when they are the right fit. Never edit a test purely
to make it pass. Run a single test with `go test -run TestName ./internal/<pkg>/...` from `apps/api`.

## Worktrees

A bare `go build ./...` from a linked git worktree needs `GOWORK=off` (the repo `go.work` resolves
to the main checkout). `make build` / `make run` / the VS Code launch config are unaffected.
