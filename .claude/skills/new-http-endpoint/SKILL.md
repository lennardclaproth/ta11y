---
name: new-http-endpoint
description: Add an HTTP endpoint to the Go API — feature-local DTOs, httpx codec helpers, validation, Swagger annotations, and route registration in the right auth tier. Use when adding or changing a route under apps/api/transport/http/handlers.
---

# Add an HTTP endpoint

Model the new handler on `apps/api/transport/http/handlers/assets/asset_create.go`. Read it first;
it shows every convention below in one file.

## 1. Decide where it lives

`apps/api/transport/http/handlers/<feature>/<verb>_<noun>.go` — one exported constructor per file.
The feature package must already own the behaviour; if it doesn't, the feature package needs a
`Commands` or `Queries` method first (see the `new-feature-package` skill).

## 2. Feature-local DTOs

Declare request and response types in the handler package — not in a shared `api` package:

```go
type CreateAssetRequest struct {
    ClassID uuid.UUID `json:"class_id"`
    Name    string    `json:"name"`
}

type CreateAssetResponse struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
}
```

Never add an `account_id` field. The account comes from the session.

## 3. Transport validation

Add an unexported `isValid() (bool, map[string]string)` method on the request type returning a
field-keyed problem map. Validate shapes here: required, length, decimals via `money.ParsePrice`,
`YYYY-MM-DD` dates. Domain invariants stay in the feature package.

## 4. The handler

```go
func CreateAsset(log logging.Logger, commands assets.Commands) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        accountID, ok := httpx.AccountID(w, r)
        if !ok {
            return
        }
        req, err := httpx.JSONDecode[CreateAssetRequest](r)
        if err != nil {
            if httpx.WriteDecodeError(w, err) {
                return
            }
            httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
            return
        }
        if isValid, problems := req.isValid(); !isValid {
            httpx.JSONEncode(w, http.StatusBadRequest, problems)
            return
        }
        // map to feature inputs, call Commands/Queries, map known errors to status codes
        httpx.JSONEncode(w, http.StatusCreated, CreateAssetResponse{ /* ... */ })
    })
}
```

- `httpx` is the `transport/http` package. Use `JSONDecode`, `JSONEncode`, `DecodeQuery`,
  `DecodeMultipartFile`, `WriteDecodeError`, `AccountID`.
- Error bodies are `map[string]string{"error": "..."}`; validation bodies are the problem map.
- Log unexpected failures with `log.Error(r.Context(), "...", err)` and return 500. Map the
  feature's sentinel errors from its `errors.go` to 404/409 rather than leaking them.
- No business logic, orchestration, or domain decisions in the handler.

## 5. Swagger annotations

Put them on the constructor, above the `func`: `@Summary`, `@Description`, `@Tags`, `@Accept`,
`@Produce`, `@Param`, `@Success`, every `@Failure` you actually return, and `@Router`.

## 6. Register the route

In `registerRoutes` in `apps/api/cmd/ta11y/main.go`, using Go 1.22 method+pattern
syntax and the right tier:

- `protected(...)` — the default; requires a session.
- `public(...)` — health, Swagger, sign-in only.
- `adminOnly(...)` — account administration, market-data curation, provider credentials, EOD imports.

## 7. Finish

`make swagger`, then `make build` and `make test`. Add a handler test next to the file when the
mapping logic is non-trivial (see `asset_worth_test.go`). Update `FEATURES.md` and `CHANGELOG.md`
if this changes a documented feature.
