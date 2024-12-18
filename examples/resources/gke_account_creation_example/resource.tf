# Agentless based integration
resource "sedai_create_account" "gke_account_agentless" {
    name = "gke_account_agentless"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENTLESS"
    cluster_provider = "GCP"
    service_account_json = "service_account_json_content"
    region = "us-east-1"
    zone = "us-east-1a"
    project_id = "project_id"
    cluster_url = "cluster_url"
}

# Agent based integration
resource "sedai_create_account" "gke_account_agent_based" {
    name = "gke_account_agent_based"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENT_BASED"
    cluster_provider = "GCP"
}