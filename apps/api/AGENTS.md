# AGENTS.md — `apps/api`

The backend conventions that used to live here now live in the repo's Claude Code rules, so they
load automatically when the matching files are edited instead of relying on an agent remembering to
read this file. Read the ones relevant to your change:

| Rule | Applies to |
| --- | --- |
| [.claude/rules/go-backend.md](../../.claude/rules/go-backend.md) | `apps/api/**/*.go` — package-by-feature, `Commands`/`Queries` boundary, shared packages, logging, comments, tests |
| [.claude/rules/http-transport.md](../../.claude/rules/http-transport.md) | `transport/http/**` — handler shape, DTOs, codec helpers, auth tiers, Swagger |
| [.claude/rules/storage.md](../../.claude/rules/storage.md) | `internal/storage/**` — sqlx stores, `WithTx`/`GetExecutor`, `qualifyTable` |
| [.claude/rules/migrations.md](../../.claude/rules/migrations.md) | `migrations/**` — mirrored Postgres + SQLite, never edit an existing migration |
| [.claude/rules/eventbus-messaging.md](../../.claude/rules/eventbus-messaging.md) | `internal/eventbus/**`, `transport/messaging/**` |

Repo-wide rules (commands, change discipline, validation, changelog) are in the root
[CLAUDE.md](../../CLAUDE.md).
