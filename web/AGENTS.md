# AGENTS.md — `web/` (frontend)

The frontend conventions that used to live here now live in the repo's Claude Code rules, so they
load automatically when the matching files are edited. Read the ones relevant to your change:

| Rule | Applies to |
| --- | --- |
| [.claude/rules/atomic-design.md](../.claude/rules/atomic-design.md) | `web/src/lib/components/**` — the tier taxonomy, the atom boundary rule, the four-file component anatomy, stories, the shared `intent/variant/size/shape` vocabulary |
| [.claude/rules/svelte-frontend.md](../.claude/rules/svelte-frontend.md) | `web/**` — stack, commands, services + mock branches, session/account handling, theming, the known formatting drift |

Repo-wide rules (commands, change discipline, validation, changelog) are in the root
[CLAUDE.md](../CLAUDE.md).
