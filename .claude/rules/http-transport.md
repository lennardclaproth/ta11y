---
paths:
  - "apps/api/transport/http/**/*.go"
  - "apps/api/cmd/ta11y/*.go"
---

# HTTP transport (`apps/api/transport/http`)

Lives at module path `.../transport/http` — **not** under `internal/`. Imported as `httpx`.
Handlers are organized per feature under `transport/http/handlers/<feature>`.

## Handler shape

- A constructor returns `http.Handler`, implemented as `return http.HandlerFunc(func(w, r) { … })` — see `transport/http/handlers/assets/asset_create.go`.
- Define **feature-local** request/response DTOs in that handler package, in the style of
  `transport/http/handlers/assets`. Use the shared `api` package only outside the refactored
  transport tree, or where a shared contract genuinely must stay shared.
- Use the codec helpers: `httpx.JSONDecode`, `httpx.JSONEncode`, `httpx.DecodeQuery`,
  `httpx.DecodeMultipartFile`, `httpx.WriteDecodeError`.
- Never use the legacy `internal/http.Endpoint` wrapper or import the removed `internal/http`.
- Carry the Swagger annotations on the handler constructor.

## Keep handlers thin

Allowed: decode, validate transport-specific input, map to feature inputs, map known feature
errors to HTTP status codes, encode. Not allowed: business rules, orchestration policy, or domain
decisions — those belong in the feature package. Domain invariants stay in the feature; only
transport-shaped validation lives here.

## Sessions, not `account_id`

The account for every request comes from the session. Read it with
`httpx.AccountID(w, r)` — it writes the error response and returns `false` when absent. Never
reintroduce an `account_id` field in a request DTO, query string, or multipart form; request
decoding rejects unknown fields.

## Routing and auth tiers

`Router` is a thin slice over `http.ServeMux`; patterns use Go 1.22 method+pattern syntax
(`"POST /accounts"`) via `HandleWithMiddleware`. Route registration in
`cmd/ta11y/main.go` goes through one of three helpers:

- `protected` — **the default**. A route added without thought requires a session.
- `public` — health, Swagger, and the sign-in endpoints only.
- `adminOnly` — wraps `apphttp.RequireAdmin`: account administration, market-data curation,
  provider credentials, EOD imports.

Adding a route means choosing a tier deliberately; default to `protected`.

`Server` wraps the mux with Elastic APM, request-identifier, origin-check and CORS middleware
(`WithCORS`, from `server.cors_allowed_origins`, applied outermost so OPTIONS preflight is
answered before routing) and does graceful shutdown.

## Swagger

Keep annotations in sync whenever an endpoint's contract changes, then regenerate with
`make swagger`. The output under `apps/api/docs` (`swagger.json`, `swagger.yaml`, `docs.go`) is
**generated** — never hand-edit it.
