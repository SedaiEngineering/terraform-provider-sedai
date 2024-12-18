# GKE account for monitoring provider
resource "sedai_create_account" "gke_account" {
  name = "gke_account"
  cloud_provider = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "GCP"
  service_account_json = "service_account_json_content"
  region = "us-east-1"
  project_id = "project_id"
  cluster_url = "cluster_url"
}

# Google cloud monitoring
resource "sedai_create_gke_monitoring_provider" "gke_monitoring_provider" {
  account_id = sedai_create_account.gke_account.id
  service_account_json = "service_account_json_content"
  project_id = "project_id"
  integration_type = "AGENTLESS"
}

# EKS account for monitoring providers
resource "sedai_create_account" "eks_account" {
    name = "eks_account"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENT_BASED"
    cluster_provider = "AWS"
}

# Federated prometheus monitoring provider agentless with JWT
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_jwt" {
    account_id = sedai_create_account.eks_account.id
    integration_type = "AGENTLESS"
    endpoint = "endpoint"
    bearer_token = "bearer_token"
}


# Federated prometheus monitoring provider agentless with client credentials
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_client_creds" {
    account_id = sedai_create_account.eks_account.id
    integration_type = "AGENTLESS"
    endpoint = "endpoint"
    token_endpoint = "token_endpoint"
    client_id = "client_id"
    client_secret = "client_secret"
}

# Federated prometheus monitoring provider agentless with no auth
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_no_auth" {
    account_id = sedai_create_account.eks_account.id
    integration_type = "AGENTLESS"
    endpoint = "endpoint"
}
