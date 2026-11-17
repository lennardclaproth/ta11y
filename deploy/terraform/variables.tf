variable "project_id" {
  description = "GCP project that hosts ta11y."
  type        = string
}

variable "region" {
  description = "Region for Cloud Run, Cloud SQL and Artifact Registry. Must support Cloud Run domain mappings."
  type        = string
  default     = "europe-west1"
}

variable "domain" {
  description = "Host ta11y is served from, e.g. ta11y.example.com. The API serves the web app at / and itself under /api."
  type        = string
}

variable "allowed_emails" {
  description = "Addresses allowed to sign in. Required outside development."
  type        = list(string)
}

variable "bootstrap_admin_email" {
  description = "Address that claims the seeded admin account on first sign-in."
  type        = string
  default     = ""
}

variable "google_client_id" {
  description = "OAuth client id of the Google sign-in client."
  type        = string
}

variable "google_client_secret" {
  description = "OAuth client secret of the Google sign-in client. Stored in Secret Manager (and in state)."
  type        = string
  sensitive   = true
}

variable "marketstack_api_key" {
  description = "Optional MarketStack API key(s), comma-separated. Empty skips the secret."
  type        = string
  default     = ""
  sensitive   = true
}

variable "alpha_vantage_api_key" {
  description = "Optional Alpha Vantage API key(s), comma-separated. Empty skips the secret."
  type        = string
  default     = ""
  sensitive   = true
}

variable "github_repository" {
  description = "owner/name of the GitHub repository allowed to deploy through Workload Identity Federation."
  type        = string
  default     = "lennardclaproth/ta11y"
}

variable "db_tier" {
  description = "Cloud SQL machine tier. db-f1-micro is the cheapest shared-core tier."
  type        = string
  default     = "db-f1-micro"
}

variable "api_image" {
  description = "Image for the first deployment of the Cloud Run service. CI replaces it on every deploy, and Terraform ignores later changes."
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "max_seed_pages" {
  description = "Catalogue seed page cap (MARKETDATA_MAX_SEED_PAGES)."
  type        = number
  default     = 10
}
