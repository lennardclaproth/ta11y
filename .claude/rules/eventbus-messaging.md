---
paths:
  - "apps/api/internal/eventbus/**/*.go"
  - "apps/api/transport/messaging/**/*.go"
---

# Event bus and handlers

`internal/eventbus` is the canonical bus (in-memory implementation in `internal/eventbus/memory`).
There is no `internal/bus` and no `internal/messaging` — do not reintroduce either.

- Publish from a feature's `Commands` (e.g. `account.Commands.Create` publishes `AccountCreated`).
- Subscribe with typed handlers.
- Keep exported event payload contracts explicit and documented.
- Handlers live in `transport/messaging/handlers/<feature>` (`assets`, `importer`, `portfolio`) and
  stay **thin**: react to the event and delegate to feature collaborators. No business rules, no
  direct SQL.
- Startup wiring fans one event out to several features (`AccountCreated` → portfolio, cashflow,
  assets, and the importer account projection). When you add a subscriber, register it in that
  wiring rather than chaining publishes from inside another handler.
