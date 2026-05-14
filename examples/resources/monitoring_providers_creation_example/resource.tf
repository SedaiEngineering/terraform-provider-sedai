# GKE account for monitoring provider
resource "sedai_create_account" "gke_account" {
  name                 = "my-gke-account"
  cloud_provider       = "KUBERNETES"
  integration_type     = "AGENTLESS"
  cluster_provider     = "GCP"
  service_account_json = "service_account_json_content"
  project_id           = "project_id"
  cluster_url          = "cluster_url"
}

# GKE monitoring provider
resource "sedai_create_gke_monitoring_provider" "gke_monitoring_provider" {
  account_id           = sedai_create_account.gke_account.id
  service_account_json = "service_account_json_content"
  project_id           = "project_id"

  depends_on = [sedai_create_account.gke_account]
}

# EKS account for monitoring providers
resource "sedai_create_account" "eks_account" {
  name             = "my-eks-account"
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
  cluster_provider = "AWS"
}

# CloudWatch monitoring provider — account credentials
resource "sedai_create_cloudwatch_monitoring_provider" "cloudwatch_mp" {
  account_id              = sedai_create_account.eks_account.id
  use_account_credentials = true

  depends_on = [sedai_create_account.eks_account]
}

# Federated Prometheus — JWT
resource "sedai_create_federated_prometheus_monitoring_provider" "fp_jwt" {
  account_id       = sedai_create_account.eks_account.id
  integration_type = sedai_create_account.eks_account.integration_type
  endpoint         = "endpoint"
  bearer_token     = "bearer_token"
}

# Federated Prometheus — client credentials
resource "sedai_create_federated_prometheus_monitoring_provider" "fp_client_creds" {
  account_id       = sedai_create_account.eks_account.id
  integration_type = sedai_create_account.eks_account.integration_type
  endpoint         = "endpoint"
  token_endpoint   = "token_endpoint"
  client_id        = "client_id"
  client_secret    = "client_secret"
}

# Federated Prometheus — no auth
resource "sedai_create_federated_prometheus_monitoring_provider" "fp_no_auth" {
  account_id       = sedai_create_account.eks_account.id
  integration_type = sedai_create_account.eks_account.integration_type
  endpoint         = "endpoint"
}

# Datadog monitoring provider
resource "sedai_create_datadog_monitoring_provider" "datadog_mp" {
  account_id       = sedai_create_account.eks_account.id
  integration_type = sedai_create_account.eks_account.integration_type
  api_key          = "api_key"
  application_key  = "application_key"
}

# New Relic monitoring provider
resource "sedai_create_newrelic_monitoring_provider" "newrelic_mp" {
  account_id          = sedai_create_account.eks_account.id
  integration_type    = sedai_create_account.eks_account.integration_type
  api_key             = "api_key"
  api_server          = "api_server"
  newrelic_account_id = "newrelic_account_id"
}
