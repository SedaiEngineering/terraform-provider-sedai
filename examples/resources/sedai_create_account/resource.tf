# AWS — role-based authentication
resource "sedai_create_account" "aws_account" {
  name             = "my-aws-account"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789:role/SedaiRole"
  external_id      = "sedai-external-id"
  user_selected_managed_services = ["EC2", "RDS", "S3"]
}

# AWS — access key authentication
resource "sedai_create_account" "aws_account_key" {
  name             = "my-aws-account-key"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  access_key       = "AKIAIOSFODNN7EXAMPLE"
  secret_key       = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  user_selected_managed_services = ["LAMBDA", "EC2", "ECS", "RDS", "S3"]
}

# Azure
resource "sedai_create_account" "azure_account" {
  name             = "my-azure-account"
  cloud_provider   = "AZURE"
  integration_type = "AGENTLESS"
  tenant_id        = "tenant-id"
  subscription_id  = "subscription-id"
  client_id        = "client-id"
  client_secret    = "client-secret"
  user_selected_managed_services = ["VM", "AZURE_DISK", "AZURE_BLOB"]
}

# GCP
resource "sedai_create_account" "gcp_account" {
  name                 = "my-gcp-account"
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = "my-gcp-project"
  service_account_json = file("service_account.json")
  user_selected_managed_services = ["GCE", "DATAFLOW", "BIG_QUERY"]
}

# Kubernetes (EKS) — agent-based
resource "sedai_create_account" "eks_account" {
  name             = "my-eks-account"
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
  cluster_provider = "AWS"
}

# Kubernetes (AKS) — agent-based
resource "sedai_create_account" "aks_account_agent" {
  name             = "my-aks-account"
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
  cluster_provider = "AZURE"
}

# Kubernetes (AKS) — agentless
resource "sedai_create_account" "aks_account_agentless" {
  name             = "my-aks-account-agentless"
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AZURE"
  tenant_id        = "tenant-id"
  subscription_id  = "subscription-id"
  client_id        = "client-id"
  client_secret    = "client-secret"
  cluster_url      = "https://my-aks-endpoint"
}

# Kubernetes (GKE) — agentless
resource "sedai_create_account" "gke_account" {
  name                 = "my-gke-account"
  cloud_provider       = "KUBERNETES"
  integration_type     = "AGENTLESS"
  cluster_provider     = "GCP"
  service_account_json = file("service_account.json")
  project_id           = "my-gcp-project"
  cluster_url          = "https://my-cluster-endpoint"
}
