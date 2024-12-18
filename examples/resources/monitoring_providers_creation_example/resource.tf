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
  integration_type = sedai_create_account.gke_account.integration_type
}

# EKS account for monitoring providers
resource "sedai_create_account" "eks_account" {
    name = "eks_account"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENT_BASED"
    cluster_provider = "AWS"
}

# Federated prometheus monitoring provider with JWT
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_jwt" {
    account_id = sedai_create_account.eks_account.id
    integration_type = sedai_create_account.eks_account.integration_type
    endpoint = "endpoint"
    bearer_token = "bearer_token"
}

# Federated prometheus monitoring provider with client credentials
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_client_creds" {
    account_id = sedai_create_account.eks_account.id
    integration_type = sedai_create_account.eks_account.integration_type
    token_endpoint = "token_endpoint"
    client_id = "client_id"
    client_secret = "client_secret"
}

# Federated prometheus monitoring provider with no auth
resource "sedai_create_federated_prometheus_monitoring_provider" "federate_prometheus_monitoring_provider_no_auth" {
    account_id = sedai_create_account.eks_account.id
    integration_type = sedai_create_account.eks_account.integration_type
    endpoint = "endpoint"
}

# Datadog monitoring provider
resource "sedai_create_datadog_monitoring_provider" "datadog_monitoring_provider" {
    account_id = sedai_create_account.eks_account.id
    integration_type = sedai_create_account.eks_account.integration_type
    api_key = "api_key"
    application_key = "application_key"
}

# New relic monitoring provider
resource "sedai_create_newrelic_monitoring_provider" "newrelic_monitoring_provider" {
    account_id = sedai_create_account.eks_account.id
    integration_type = sedai_create_account.eks_account.integration_type
    api_key = "api_key"
    api_server = "api_server"
    newrelic_account_id = "newrelic_account_id"
}
