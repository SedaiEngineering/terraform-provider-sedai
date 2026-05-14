# GKE monitoring provider
resource "sedai_create_gke_monitoring_provider" "gke_mp" {
  account_id           = sedai_create_account.gke_account.id
  service_account_json = file("service_account.json")
  project_id           = "my-gcp-project"

  depends_on = [sedai_create_account.gke_account]
}
