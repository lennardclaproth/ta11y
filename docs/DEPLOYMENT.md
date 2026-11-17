# Deploying ta11y to Google Cloud

How production is set up, why it is set up that way, how to do it again from an empty project,
and the commands you need afterwards. The design and its limits are summarised in
[[036] GCP deployment](../apps/api/docs/features/%5B036%5D_GCP_DEPLOYMENT.md); this is the
hands-on guide.

Placeholders used below — fill in once:

| Placeholder | Meaning | Ours |
| --- | --- | --- |
| `<project-id>` | GCP project **ID** (not the name or number) | see `gcloud config get-value project` |
| `<region>` | Region for everything regional | `europe-west1` (Belgium) |
| `<domain>` | Host the app is served from | `ta11y.finance` |
| `<state-bucket>` | Terraform state bucket | `<project-id>-tfstate` |

Commands are written one per line so they work in Command Prompt, PowerShell and Git Bash alike.
Where shells differ (environment variables), all three forms are given.

---

## 1. What runs where

```
browser ── https://<domain> ── Cloud Run service "ta11y-api"
                                 /       → the web app (static SvelteKit build)
                                 /api/*  → the Go API
                    Cloud Run ── /cloudsql socket ──▶ Cloud SQL Postgres "ta11y"
                    Cloud Run ── env from ──────────▶ Secret Manager
GitHub Actions (push to main) ── builds image ──▶ Artifact Registry ──▶ new Cloud Run revision
```

| Piece | What it is | Cost at personal use |
| --- | --- | --- |
| Cloud Run `ta11y-api` | One container: the API, which also serves the web app | ~free tier; scales to zero |
| Cloud SQL `ta11y` | Postgres 16, `db-f1-micro`, 10 GB SSD, daily backups | ~$8–11/month, always on |
| Secret Manager | `DATABASE_URL`, Google client secret, optional provider keys | free tier |
| Artifact Registry `ta11y` | Container images; keeps the 5 newest | free tier (~33 MB per image) |
| Domain mapping | `<domain>` → Cloud Run, with a Google-managed HTTPS certificate | free |
| Workload Identity Federation | Lets GitHub Actions deploy without any stored key | free |

Everything is defined in [`deploy/terraform`](../deploy/terraform). The image is defined by the
root [`Dockerfile`](../Dockerfile) and shipped by
[`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml).

## 2. Why it is built this way

- **Scale to zero, but at most one instance.** The event bus, the websocket hub, uploaded CSVs
  (`/tmp`) and start-up recovery all live inside the process. Two instances would split them, so
  `max_instance_count = 1`. Zero instances when idle keeps Cloud Run free.
- **CPU always allocated** (`cpu_idle = false`). Imports and rebuilds run on the event bus after
  the HTTP response; without this, Cloud Run would freeze them. The instance still stops after
  ~15 idle minutes.
- **One origin for app and API.** The API serves the web build at `/` and itself under `/api`.
  No CORS, the session cookie stays first-party, websockets work on the same host, and there is
  one thing to deploy.
- **The database is the only always-on part.** Cloud SQL cannot scale to zero; it is the main
  cost. It is reached through Cloud Run's built-in Cloud SQL socket, so it needs no VPC and no
  open network.
- **No keys anywhere.** GitHub proves to Google that a run comes from `main` of this repository
  and gets short-lived credentials. The only secrets live in Secret Manager.
- **Public without `allUsers`.** The project's organisation forbids granting anything to
  `allUsers` (domain-restricted sharing), so the service sets `invoker_iam_disabled = true`
  instead. The API checks the session on every request itself.
- **Terraform owns the infrastructure, CI owns the image.** Terraform ignores the image after the
  first apply; every push to `main` builds and rolls out a new one.

## 3. One-time setup

### 3.1 Tools and login

Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install),
[Terraform](https://developer.hashicorp.com/terraform/install) ≥ 1.6, Docker Desktop and the
[GitHub CLI](https://cli.github.com/).

```bash
gcloud auth login
```

Terraform does **not** use that login; it uses *Application Default Credentials*:

```bash
gcloud auth application-default login
```

On the consent page **tick every checkbox** (or *Select all*). Leaving "See, edit, configure,
and delete your Google Cloud data" unticked fails with *"cloud-platform scope is required but
not consented"*.

```bash
gcloud config set project <project-id>
```

```bash
gcloud auth application-default set-quota-project <project-id>
```

### 3.2 Project and billing

Create the project in the console and link a billing account. Note its **ID**.

Pick the region now. Every resource lives there, and moving later means recreating the database
and migrating its data. The region must support Cloud Run domain mappings.

### 3.3 Terraform state bucket

Terraform keeps its state here. The state contains the database password and the client secret,
so the bucket is private and versioned.

```bash
gcloud storage buckets create gs://<state-bucket> --location=<region> --uniform-bucket-level-access --public-access-prevention
```

```bash
gcloud storage buckets update gs://<state-bucket> --versioning
```

### 3.4 Verify the domain

Google only maps domains you have proven to own, with the same account that runs Terraform.

```bash
gcloud domains verify <domain>
```

Search Console opens. Choose *Domain name provider*; it gives a `google-site-verification=…`
value. At GoDaddy: **Domain Portfolio → `<domain>` → DNS → Add New Record**, type `TXT`, name
`@`, value the verification string. Click **Verify**.

### 3.5 Google sign-in client

Terraform cannot create this; do it in the console under **APIs & Services**.

1. **OAuth consent screen:** External; add your address as a test user; add `<domain>` under
   *Authorized domains*.
2. **Credentials → Create credentials → OAuth client ID → Web application.**
   Authorised redirect URI: `https://<domain>/api/auth/google/callback` (exactly).
3. Keep the **client ID** (goes in `terraform.tfvars`) and the **client secret** (`GOCSPX-…`,
   goes in an environment variable).

### 3.6 Terraform variables

In `deploy/terraform`, copy `terraform.tfvars.example` to `terraform.tfvars` (git-ignored) and
fill in `project_id`, `region`, `domain`, `allowed_emails`, `bootstrap_admin_email` and
`google_client_id`.

**Delete the `google_client_secret` line.** Pass the secret through the environment instead, so it
never touches disk. A value in `terraform.tfvars` would win over the environment variable.

| Shell | Command |
| --- | --- |
| Command Prompt | `set TF_VAR_google_client_secret=GOCSPX-...` |
| PowerShell | `$env:TF_VAR_google_client_secret = 'GOCSPX-...'` |
| Git Bash | `export TF_VAR_google_client_secret='GOCSPX-...'` |

The variable only lives in that terminal window — run Terraform from the same one.

### 3.7 Create the infrastructure

All Terraform commands run from `deploy/terraform` (elsewhere: *"No configuration files"*).

```bash
terraform init -backend-config="bucket=<state-bucket>"
```

**First, the APIs and the image repository.** A new project has the needed Google APIs switched
off, and the Cloud Run service needs an image that lives in a repository Terraform creates — so
these go first:

```bash
terraform apply -target=google_project_service.enabled -target=google_artifact_registry_repository.ta11y
```

Terraform warns that `-target` is for exceptional cases; that is expected. If it fails with *"API
has not been used in project … or it is disabled"*, wait a minute and run it again.

**Then everything else.** Two ways to get the first image:

- **Let CI build it (what we did).** Apply without an image; the service starts on Google's
  "Congratulations | Cloud Run" placeholder until the first deploy from `main` replaces it:

  ```bash
  terraform apply
  ```

- **Build it locally** (needs Docker Desktop running; run from the repository root):

  ```bash
  gcloud auth configure-docker <region>-docker.pkg.dev
  ```

  ```bash
  docker build -t <region>-docker.pkg.dev/<project-id>/ta11y/ta11y:initial .
  ```

  ```bash
  docker push <region>-docker.pkg.dev/<project-id>/ta11y/ta11y:initial
  ```

  ```bash
  terraform apply -var="api_image=<region>-docker.pkg.dev/<project-id>/ta11y/ta11y:initial"
  ```

Creating Cloud SQL takes 5–15 minutes; Terraform prints *"Still creating…"* meanwhile. An
interrupted or failed apply is safe to re-run: Terraform only creates what is missing.

### 3.8 DNS at GoDaddy

```bash
terraform output dns_records
```

**Bare domain** (`ta11y.finance`): four `A` and four `AAAA` records, all with name `@`. They are
normally:

| Type | Name | Value |
| --- | --- | --- |
| A | `@` | `216.239.32.21`, `216.239.34.21`, `216.239.36.21`, `216.239.38.21` (one record each) |
| AAAA | `@` | `2001:4860:4802:32::15`, `…:34::15`, `…:36::15`, `…:38::15` (one record each) |

**Subdomain** (e.g. `test.ta11y.finance`): one `CNAME`, name `test`, value `ghs.googlehosted.com`.

At GoDaddy, also:

- **Delete every other `A @` record.** GoDaddy adds a *Parked* record, and a GoDaddy website
  (Website Builder / Airo) adds its own (`76.223.105.230`, `13.248.243.5`). If any remain, some
  visitors land on GoDaddy's page and Google cannot issue the certificate. If they cannot be
  deleted or come back, disconnect the domain from the GoDaddy website first (*My Products* →
  the website → Settings → Domain).
- **Turn off domain forwarding** for the domain.
- **CAA records:** with none (our case), nothing to do. If any exist, add `CAA @ 0 issue
  "pki.goog"` and `CAA @ 0 issue "letsencrypt.org"`.

Then wait: DNS minutes to an hour, the certificate 15–60 minutes after that.

```bash
nslookup <domain> 8.8.8.8
```

```bash
gcloud beta run domain-mappings describe --domain <domain> --region <region>
```

Ready when the conditions show `CertificateProvisioned: True` and `Ready: True`. If your own PC
still shows the old page, `ipconfig /flushdns`.

### 3.9 Connect GitHub

The workflow needs **variables**, not secrets (nothing here is sensitive; Workload Identity
Federation replaces keys).

```bash
terraform output github_variables
```

Set all of them at once from PowerShell, in `deploy/terraform`:

```powershell
(terraform output -json github_variables | ConvertFrom-Json).PSObject.Properties | ForEach-Object { gh variable set $_.Name --body $_.Value --repo lennardclaproth/ta11y }
```

Or by hand: GitHub → repository → **Settings → Secrets and variables → Actions → Variables** tab
(not *Secrets* — the workflow reads `vars.*`).

```bash
gh variable list --repo lennardclaproth/ta11y
```

### 3.10 First deploy and check

Merge to `main`. The *Deploy* workflow builds the image, pushes it and points the service at it.
Only runs from `refs/heads/main` can deploy.

- `https://<domain>/api/health` → `{"status":"ok"}`
- `https://<domain>/` → sign-in page; sign in with `bootstrap_admin_email` first, so that account
  claims the seeded admin account.

## 4. Day to day

| Change | How |
| --- | --- |
| Code (API or web) | Merge to `main`. Nothing else. |
| Infrastructure, env vars, secrets | Edit `deploy/terraform`, then `terraform plan` and `terraform apply` from that folder (with `TF_VAR_google_client_secret` set). |
| Rotate the Google client secret | New secret in the console, then `terraform apply` with the new `TF_VAR_google_client_secret`; the next revision picks it up. |
| Roll back | Point traffic at an earlier revision (below). Only the 5 newest images are kept. |

## 5. Handy gcloud commands

Assumes `gcloud config set project <project-id>`. Add `--region <region>` where shown.

### Account and project

```bash
gcloud config list
```

```bash
gcloud auth list
```

```bash
gcloud projects describe <project-id> --format="value(parent.type,parent.id)"
```

### Cloud Run

```bash
gcloud run services describe ta11y-api --region <region>
```

```bash
gcloud run services describe ta11y-api --region <region> --format="value(status.url)"
```

```bash
gcloud run revisions list --service ta11y-api --region <region>
```

Roll back to an earlier revision (send all traffic to it):

```bash
gcloud run services update-traffic ta11y-api --region <region> --to-revisions <revision-name>=100
```

Back to the newest revision:

```bash
gcloud run services update-traffic ta11y-api --region <region> --to-latest
```

Deploy a specific image by hand (what CI does):

```bash
gcloud run services update ta11y-api --region <region> --image <region>-docker.pkg.dev/<project-id>/ta11y/ta11y:<tag>
```

### Logs

Recent application logs (JSON from the API's structured logger):

```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ta11y-api" --limit 50 --freshness 1h --format="value(timestamp,jsonPayload.msg,jsonPayload.error)"
```

Errors only:

```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=ta11y-api AND severity>=ERROR" --limit 20 --freshness 1d
```

Live tail:

```bash
gcloud beta run services logs tail ta11y-api --region <region>
```

In the console: **Cloud Run → ta11y-api → Logs**.

### Cloud SQL

```bash
gcloud sql instances list
```

```bash
gcloud sql instances describe ta11y
```

```bash
gcloud sql backups list --instance ta11y
```

On-demand backup (e.g. before a risky migration):

```bash
gcloud sql backups create --instance ta11y
```

Stop and start the database (stopped = storage cost only, **app is down**):

```bash
gcloud sql instances patch ta11y --activation-policy=NEVER
```

```bash
gcloud sql instances patch ta11y --activation-policy=ALWAYS
```

To browse the data: console → **SQL → ta11y → Cloud SQL Studio**, database `ta11y`, user
`ta11y`, password from the connection string:

```bash
gcloud secrets versions access latest --secret=ta11y-database-url
```

That prints `postgres://ta11y:<password>@/ta11y?...` — don't paste it anywhere.

### Secrets

```bash
gcloud secrets list
```

```bash
gcloud secrets versions list ta11y-google-client-secret
```

Secret values are managed by Terraform; change them through `terraform apply`. A version added
by hand becomes `latest` and is what Cloud Run reads, but Terraform no longer knows the value in
use.

### Images

```bash
gcloud artifacts docker images list <region>-docker.pkg.dev/<project-id>/ta11y --include-tags
```

### Domain and certificate

```bash
gcloud beta run domain-mappings list --region <region>
```

```bash
gcloud beta run domain-mappings describe --domain <domain> --region <region>
```

### Costs

Console → **Billing → Reports**, grouped by service. Set a budget alert under **Billing → Budgets
& alerts** (e.g. $15/month) so an unexpected cost shows up by email.

## 6. Troubleshooting (what we ran into)

| Symptom | Cause | Fix |
| --- | --- | --- |
| `'export' is not recognized` | Command Prompt, not bash | Use `set NAME=value` (see 3.6) |
| `Error: No configuration files` | Terraform run outside `deploy/terraform` | `cd` into `deploy/terraform` |
| `could not find default credentials` | No Application Default Credentials | `gcloud auth application-default login` |
| `cloud-platform scope is required but not consented` | A consent checkbox left unticked | Log in again, tick every box |
| `API has not been used in project … or it is disabled` | API still switching on | Wait a minute, re-run the apply |
| `One or more users named in the policy do not belong to a permitted customer` | Org policy forbids `allUsers` | Already solved: `invoker_iam_disabled = true` in `run.tf` |
| Domain shows a GoDaddy page | Extra GoDaddy `A @` records | Delete them (3.8) |
| Certificate stays pending | Wrong/extra DNS records, or a blocking CAA record | Check `nslookup <domain> 8.8.8.8`, then 3.8 |
| `gcloud sql instances list` is empty after an apply | Only the targeted apply ran | Run the full `terraform apply` |
| Sign-in fails with a redirect error | Redirect URI mismatch | OAuth client URI must be exactly `https://<domain>/api/auth/google/callback` |
