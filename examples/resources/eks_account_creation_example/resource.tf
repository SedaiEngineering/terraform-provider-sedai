# Agentless integration using aws role
resource "sedai_create_account" "eks_account_agentless_role_based" {
    name = "eks_account_agentless_role_based"
    role = "role_arn"
    external_id = "external_id"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENTLESS"
    cluster_provider = "AWS"
    cluster_url = "cluster_url"
    region = "us-east-1"
}

# Agentless integration using aws access key and secret key
resource "sedai_create_account" "eks_account_agentless_key_based" {
    name = "eks_account_agentless_key_based"
    access_key = "access_key"
    secret_key = "secret_key"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENTLESS"
    cluster_provider = "AWS"
    cluster_url = "cluster_url"
    region = "us-east-1"
}

# Agent based integration using aws role
resource "sedai_create_account" "eks_account_agent_based" {
    name = "eks_account_agent_based_role_based"
    cloud_provider = "KUBERNETES"
    integration_type = "AGENT_BASED"
    cluster_provider = "AWS"
}