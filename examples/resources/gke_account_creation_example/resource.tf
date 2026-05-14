# Kubernetes (GKE) — agentless
resource "sedai_create_account" "gke_account_agentless" {
  name                 = "my-gke-account"
  cloud_provider       = "KUBERNETES"
  integration_type     = "AGENTLESS"
  cluster_provider     = "GCP"
  service_account_json = "service_account_json_content"
  project_id           = "project_id"
  zone                 = "us-central1-a"
  cluster_url          = "cluster_url"
}

# Kubernetes (GKE) — agent-based
resource "sedai_create_account" "gke_account_agent_based" {
  name             = "my-gke-account"
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
  cluster_provider = "GCP"
}
