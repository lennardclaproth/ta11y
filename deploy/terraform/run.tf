# ta11y on Cloud Run: the API serving the web app at / and itself under /api, from one image.
# Scale-to-zero (min 0) and a single instance at most (max 1): the event bus, the websocket hub,
# temporary uploads and start-up recovery are all per process.
resource "google_cloud_run_v2_service" "api" {
  name     = "ta11y-api"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  # Public without granting run.invoker to allUsers, which domain-restricted sharing
  # (iam.allowedPolicyMemberDomains) forbids. The API authenticates every request itself
  # (session cookie, default-deny routing).
  invoker_iam_disabled = true

  deletion_protection = true

  template {
    service_account = google_service_account.api.email
    # Websockets live as long as the request; the hub closes idle ones after five minutes.
    timeout                          = "3600s"
    max_instance_request_concurrency = 80

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.ta11y.connection_name]
      }
    }

    containers {
      image = var.api_image

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        # Instance-based billing: CPU stays allocated after a response, so imports and rebuilds
        # running on the event bus finish before the idle instance is stopped.
        cpu_idle          = false
        startup_cpu_boost = true
      }

      volume_mounts {
        name       = "cloudsql"
        mount_path = "/cloudsql"
      }

      dynamic "env" {
        for_each = {
          AUTH_FRONTEND_URL          = "https://${var.domain}"
          AUTH_ALLOWED_EMAILS        = join(",", var.allowed_emails)
          AUTH_BOOTSTRAP_ADMIN_EMAIL = var.bootstrap_admin_email
          GOOGLE_REDIRECT_URL        = "https://${var.domain}/api/auth/google/callback"
          GOOGLE_CLIENT_ID           = var.google_client_id
          MARKETDATA_MAX_SEED_PAGES  = tostring(var.max_seed_pages)
          ELASTIC_APM_ACTIVE         = "false"
        }
        content {
          name  = env.key
          value = env.value
        }
      }

      dynamic "env" {
        for_each = local.secret_env
        content {
          name = env.key
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.ta11y[env.value].secret_id
              version = "latest"
            }
          }
        }
      }

      startup_probe {
        http_get {
          path = "/api/health"
        }
        initial_delay_seconds = 0
        period_seconds        = 2
        failure_threshold     = 30
      }
    }
  }

  # CI ships every new image (`gcloud run services update --image`); Terraform owns the rest.
  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }

  depends_on = [
    google_project_service.enabled,
    google_secret_manager_secret_iam_member.api_access,
    google_secret_manager_secret_version.ta11y,
    google_project_iam_member.api_cloudsql_client,
  ]
}

# The custom domain, with a Google-managed certificate. The domain must be verified in Search
# Console for the identity running Terraform first.
resource "google_cloud_run_domain_mapping" "api" {
  name     = var.domain
  location = var.region

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.api.name
  }
}
