locals {
  services = [
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "serviceusage.googleapis.com",
    "sqladmin.googleapis.com",
    "sts.googleapis.com",
  ]
}

resource "google_project_service" "enabled" {
  for_each = toset(local.services)

  service            = each.value
  disable_on_destroy = false
}

resource "google_artifact_registry_repository" "ta11y" {
  repository_id = "ta11y"
  location      = var.region
  format        = "DOCKER"
  description   = "ta11y API images"

  # Keep the five newest images and drop anything older than 30 days, so storage stays near zero.
  cleanup_policy_dry_run = false
  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions {
      keep_count = 5
    }
  }
  cleanup_policies {
    id     = "delete-old"
    action = "DELETE"
    condition {
      older_than = "2592000s"
    }
  }

  depends_on = [google_project_service.enabled]
}

# Runs the API. Only what the service reads: its database and its secrets.
resource "google_service_account" "api" {
  account_id   = "ta11y-api"
  display_name = "ta11y API (Cloud Run runtime)"
}

resource "google_project_iam_member" "api_cloudsql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.api.email}"
}
