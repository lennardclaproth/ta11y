---
description: Run the validation that matches what changed, then report honestly what passed, failed, and was skipped
argument-hint: "[optional: extra scope, e.g. 'also integration tests']"
allowed-tools: Bash, Read, Grep, Glob
---

Validate the current working tree. $ARGUMENTS

1. Run `git status --porcelain` and `git diff --name-only HEAD` to see what actually changed.
2. Pick the checks that match, and skip the ones that don't:
   - Go files under `apps/api` → `make build`, then `make test`, then `make lint`.
   - HTTP handler or DTO changes (`apps/api/transport/http/**`) → also `make swagger`, then check
     whether `apps/api/docs` changed and mention it.
   - Migrations → confirm the Postgres and SQLite files are a matching pair with identical
     timestamps, and that no existing migration was modified (`git diff --name-only HEAD -- apps/api/migrations`).
   - Files under `web/` → `npm run check` from `web/`. Do **not** run `npm run lint` as a gate; it
     is not green out of the box because of the known tabs-vs-spaces drift.
3. If something fails, fix it and re-run that check — don't move on and report it as a caveat.
4. Finish with a short report: one line per check with its real outcome, and an explicit list of
   anything you skipped and why. Never imply a command ran when it didn't.
