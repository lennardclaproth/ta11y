locals {
  # Secret name => value. Optional provider keys are left out when empty, so no empty secret
  # version is ever created and the service gets no variable for them.
  secret_values = merge(
    {
      "ta11y-database-url"         = local.database_url
      "ta11y-google-client-secret" = var.google_client_secret
    },
    var.marketstack_api_key == "" ? {} : { "ta11y-marketstack-api-key" = var.marketstack_api_key },
    var.alpha_vantage_api_key == "" ? {} : { "ta11y-alpha-vantage-api-key" = var.alpha_vantage_api_key },
  )

  # Environment variable => secret name, for the Cloud Run service.
  secret_env = merge(
    {
      DATABASE_URL         = "ta11y-database-url"
      GOOGLE_CLIENT_SECRET = "ta11y-google-client-secret"
    },
    var.marketstack_api_key == "" ? {} : { MARKETSTACK_API_KEY = "ta11y-marketstack-api-key" },
    var.alpha_vantage_api_key == "" ? {} : { ALPHA_VANTAGE_API_KEY = "ta11y-alpha-vantage-api-key" },
  )
}

resource "google_secret_manager_secret" "ta11y" {
  for_each = nonsensitive(toset(keys(local.secret_values)))

  secret_id = each.value
  replication {
    auto {}
  }

  depends_on = [google_project_service.enabled]
}

resource "google_secret_manager_secret_version" "ta11y" {
  for_each = nonsensitive(toset(keys(local.secret_values)))

  secret      = google_secret_manager_secret.ta11y[each.value].id
  secret_data = local.secret_values[each.value]
}

resource "google_secret_manager_secret_iam_member" "api_access" {
  for_each = google_secret_manager_secret.ta11y

  secret_id = each.value.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.api.email}"
}
