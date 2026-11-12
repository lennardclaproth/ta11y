terraform {
  required_version = ">= 1.6"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  # State holds the generated database password and the secret values passed in as variables,
  # so the bucket must be private with versioning on. Configure it at init time:
  #   terraform init -backend-config="bucket=<state-bucket>"
  backend "gcs" {
    prefix = "ta11y"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
