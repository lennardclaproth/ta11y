# `.claude/` — shared AI development config

Everything here except `settings.local.json` and `.cache/` is committed, so the whole team (and any
fresh clone or worktree) gets the same behaviour.

| Path | What it does | When it loads |
| --- | --- | --- |
| `../CLAUDE.md` | Repo-wide instructions: shape, commands, change discipline, validation, changelog rules | Every session |
| `rules/*.md` | Area conventions, each scoped with `paths:` frontmatter | Only when Claude reads a matching file |
| `settings.json` | Permission allow/deny/ask lists and hook registrations | Every session, client-enforced |
| `settings.local.json` | Your personal overrides (gitignored) | Every session, wins over `settings.json` |
| `commands/*.md` | Slash commands — prompt templates you invoke by name | Only when you type `/name` |
| `skills/*/SKILL.md` | Multi-step procedures Claude picks up when the task matches the description | On demand |
| `agents/*.md` | Review subagents with their own context window | When spawned |
| `hooks/*.mjs` | Shell handlers on lifecycle events — the enforcement layer | On their event |
| `launch.json` | Dev server definitions for the browser preview | `preview_start` |

The distinction that matters: **`CLAUDE.md` and `rules/` are context — advice Claude usually
follows. `settings.json` and `hooks/` are enforcement — the client applies them regardless.** Put a
rule in prose when judgement is needed, and in settings or a hook when it must hold every time.

## Rules

Path-scoped, so a session that only touches Go never loads the frontend conventions:

| Rule | `paths:` |
| --- | --- |
| `go-backend.md` | `apps/api/**/*.go` |
| `http-transport.md` | `apps/api/transport/http/**`, the entrypoint |
| `storage.md` | `apps/api/internal/storage/**` |
| `migrations.md` | `apps/api/migrations/**` |
| `eventbus-messaging.md` | `apps/api/internal/eventbus/**`, `apps/api/transport/messaging/**` |
| `atomic-design.md` | `web/src/lib/components/**` |
| `svelte-frontend.md` | `web/**/*.svelte`, `.ts`, `.css` |

A rule with no `paths:` would load every session — keep them scoped, and put anything genuinely
repo-wide in the root `CLAUDE.md` instead. Run `/context` in a session to see what actually loaded.

## Permissions

`settings.json` pre-approves the read-only and validation commands (`make build|test|lint|swagger`,
`go build|vet|test`, `npm run check|lint`, read-only `git`) so routine work stops prompting.

It also denies what shouldn't be touched by hand:

- reading any `.env` (provider API keys live there)
- editing the generated Swagger output under `apps/api/docs` — regenerate with `make swagger`

and asks first for migrations, `deploy/`, `.github/`, and dependency manifests.

Override any of it for yourself in `settings.local.json`; that file is gitignored.

## Hooks

| Hook | Event | What it does |
| --- | --- | --- |
| `session-context.mjs` | `SessionStart` | Reports branch, uncommitted count, the next `[NNN]` changelog feature ID, and whether a `tmp/plan.md` is unfinished |
| `gowork-guard.mjs` | `PreToolUse` (`Bash`, `PowerShell`) | In a linked git worktree, prefixes `GOWORK=off` on a bare `go build/vet/test/list/run` so `go.work` doesn't resolve to the main checkout |
| `gofmt-file.mjs` | `PostToolUse` (`Edit`, `Write`) | Runs `gofmt` on the one `.go` file just written, and records it |
| `verify-changed.mjs` | `Stop` | If the session edited Go files, runs `go build ./...` and `go vet ./...` in `apps/api`; a failure blocks the turn from ending (exit 2) so it gets fixed rather than reported as done. Gives up after two blocks so a stuck build can't trap the session |

They're Node ESM scripts (`node` is already required by `web/`), so they behave the same on Windows,
macOS, Linux and CI. Each reads the hook JSON payload on stdin. All of them fail open: a missing Go
toolchain or an unreadable payload exits 0 rather than breaking the session.

`.cache/` holds the per-session list of touched files that `gofmt-file` writes and
`verify-changed` consumes. It's gitignored and safe to delete.

## Commands

- `/verify` — run the validation matching what changed, then report honestly what passed, failed, or was skipped
- `/feature <what>` — feature work with the full discipline: feature ID, rules, tests, `FEATURES.md`, `CHANGELOG.md`
- `/migration <name>` — create a mirrored goose pair (Postgres + SQLite)
- `/fe-check` — run the frontend type gate and fix what it reports

## Skills

`new-http-endpoint`, `new-component`, `new-feature-package` — invoked by name or picked up
automatically when a task matches. They cost no context until used, which is why the long
step-by-step procedures live here rather than in `CLAUDE.md`.

## Agents

`go-reviewer` and `svelte-reviewer` review a diff against this repo's layering rules in their own
context window. Ask for them by name ("have the go-reviewer look at this").
