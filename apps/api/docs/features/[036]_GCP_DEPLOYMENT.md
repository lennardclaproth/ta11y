# [036] GCP Deployment (Cloud Run, scale to zero)

One container on Cloud Run runs the API and serves the built web app; Cloud SQL Postgres holds
the data. Everything but the database scales to zero. Infrastructure is Terraform in
`deploy/terraform`; GitHub Actions ships images.

## Shape

```
browser ── https://<domain> ── Cloud Run "ta11y-api" (min 0, max 1, CPU always allocated)
                                 /       → web app (web/build, SPA fallback to index.html)
                                 /api/*  → API, prefix stripped before middleware and routes
Cloud Run ── /cloudsql socket ──▶ Cloud SQL Postgres 16 (db-f1-micro, public IP, no networks)
Cloud Run ── secrets ──▶ Secret Manager (DATABASE_URL, Google client secret, provider keys)
GitHub Actions ── WIF ──▶ Artifact Registry + `gcloud run services update --image`
```

One origin means no CORS, a host-only session cookie, and websockets on the same host
(`wss://<domain>/api/ws/accounts/{id}`).

## Processing rules

- **Single instance.** `max_instance_count = 1` is load-bearing: the in-memory event bus, the
  websocket hub, import uploads in `/tmp` and start-up recovery ([002], [005]) are per process.
- **Instance-based billing** (`cpu_idle = false`): background work on the bus keeps CPU after the
  response; the instance stops after Cloud Run's idle window and the service scales to zero.
- **Shutdown**: Cloud Run's SIGTERM gives 10s — HTTP shutdown (≤5s) then bus drain (≤4s) ([017]).
- **Serving the app.** `server.web_dir` (`WEB_DIR`) switches the server to site mode
  (`transport/http/site.go`): existing files are served as-is, `/_app/immutable/*` cached for a
  year, everything else gets `index.html` with `no-cache`. Path-scoped cookies carry the prefix
  (`CookieSettings.PathPrefix`), so the sign-in flow cookie is set on `/api/auth`.
- **Frontend build** is an SPA (`adapter-static`, `ssr = false`) built with `VITE_API_URL=/api`, so
  the image does not depend on the domain.
- **Migrations** still run at boot; `EnsureDBExists` finds the Terraform-created database.
- **Config** is `apps/api/config.cloudrun.yaml` in the image plus env from `run.tf` ([018]).

## Code map

| Path | Role |
| --- | --- |
| `Dockerfile`, `.dockerignore` | Web build (node) + API build (Go) → distroless image |
| `apps/api/config.cloudrun.yaml` | Non-secret production config baked into the image |
| `apps/api/transport/http/site.go` | SPA handler; `server.go` `WithSite` mounts API under `/api` |
| `deploy/terraform/*.tf` | APIs, Artifact Registry, Cloud SQL, secrets, Cloud Run, domain mapping, WIF |
| `.github/workflows/deploy.yml` | Build, push, update the Cloud Run image on push to `main` |

## First deploy

Step-by-step setup (project, state bucket, OAuth client, Terraform, DNS at GoDaddy, GitHub
variables), day-to-day operations and gcloud commands: [docs/DEPLOYMENT.md](../../../../docs/DEPLOYMENT.md).

## Gaps / not implemented

- No horizontal scaling: needs a shared bus, shared upload storage and cross-instance notify.
- Terraform state contains the database password and the secrets passed as variables.
- Domain mappings exist only in some regions (default `europe-west1`).
- Cold start on the first request after idle (~1–3s); the first page load waits for it.
- No staging environment; `main` deploys straight to production.
- The organisation policy forbids `allUsers` bindings, so the service is public through
  `invoker_iam_disabled` instead; a policy enforcing `run.managed.requireInvokerIam` would block it.
