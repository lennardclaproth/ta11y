---
description: Implement a feature change with the repo's full discipline — rules, feature doc, changelog ID, validation
argument-hint: "<what to build> [optional: existing feature ID]"
allowed-tools: Bash, Read, Edit, Write, Grep, Glob
---

Feature work: $ARGUMENTS

Before writing code:

1. Read `apps/api/docs/features/FEATURES.md` and find the entry this change belongs to.
2. Determine the feature ID:
   - Extending existing behaviour → reuse that feature's `[NNN]` ID.
   - Genuinely new capability → take the next ID in the `CHANGELOG.md` sequence.
   - Can't map it confidently → **stop and ask** rather than guessing.
3. Search for the nearby established pattern in the packages you're about to touch, and follow it
   instead of inventing a parallel one. The area rules in `.claude/rules/` load automatically —
   follow the one that matches.

While implementing:

- Minimal, focused diff. No unrelated refactors, renames, or modernization.
- Backend: keep the `Commands`/`Queries` boundary, keep handlers thin, reuse `internal/money`,
  `internal/date`, `internal/sorting`, put feature errors in that package's `errors.go`.
- Frontend: respect the Atomic Design tiers and the atom boundary rule, and add the mock branch
  alongside every live service call.
- Ask before adding a dependency, changing an existing migration, or touching Docker/CI/infra.

To finish:

1. Add or update tests for the behaviour that changed.
2. Update the feature's entry in `FEATURES.md` if behaviour changed.
3. Add the `CHANGELOG.md` entry under `Unreleased` with the right Keep a Changelog category and the
   `[NNN]` ID in the text.
4. Run the validation from `/verify` and report what passed, failed, or was skipped.
