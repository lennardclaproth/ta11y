# The one always-on piece: the smallest shared-core Postgres instance.
resource "google_sql_database_instance" "ta11y" {
  name             = "ta11y"
  database_version = "POSTGRES_16"
  region           = var.region

  settings {
    tier              = var.db_tier
    edition           = "ENTERPRISE"
    availability_type = "ZONAL"
    disk_type         = "PD_SSD"
    disk_size         = 10
    disk_autoresize   = true

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = false
      backup_retention_settings {
        retained_backups = 7
      }
    }

    # Public IP with no authorised networks: reachable only through the Cloud SQL connector
    # (the Cloud Run socket mount), which authenticates with IAM. No VPC connector to pay for.
    ip_configuration {
      ipv4_enabled = true
      ssl_mode     = "ENCRYPTED_ONLY"
    }
  }

  deletion_protection = true

  depends_on = [google_project_service.enabled]
}

resource "google_sql_database" "ta11y" {
  name     = "ta11y"
  instance = google_sql_database_instance.ta11y.name
}

resource "random_password" "db" {
  length  = 32
  special = false
}

resource "google_sql_user" "api" {
  name     = "ta11y"
  instance = google_sql_database_instance.ta11y.name
  password = random_password.db.result
}

locals {
  # lib/pq reads `host` as the Unix socket directory Cloud Run mounts at /cloudsql.
  database_url = format(
    "postgres://%s:%s@/%s?host=/cloudsql/%s&sslmode=disable",
    google_sql_user.api.name,
    random_password.db.result,
    google_sql_database.ta11y.name,
    google_sql_database_instance.ta11y.connection_name,
  )
}
