# Values for the GitHub repository variables the deploy workflow reads.
output "github_variables" {
  description = "Set these as GitHub Actions repository variables (Settings → Secrets and variables → Actions → Variables)."
  value = {
    GCP_PROJECT_ID    = var.project_id
    GCP_REGION        = var.region
    GCP_WIF_PROVIDER  = google_iam_workload_identity_pool_provider.github.name
    GCP_DEPLOY_SA     = google_service_account.deploy.email
    GCP_IMAGE_REPO    = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.ta11y.repository_id}"
    CLOUD_RUN_SERVICE = google_cloud_run_v2_service.api.name
  }
}

output "dns_records" {
  description = "DNS records to create for the domain mapping."
  value       = google_cloud_run_domain_mapping.api.status[0].resource_records
}

output "run_url" {
  description = "Default Cloud Run URL (works before DNS is in place)."
  value       = google_cloud_run_v2_service.api.uri
}
