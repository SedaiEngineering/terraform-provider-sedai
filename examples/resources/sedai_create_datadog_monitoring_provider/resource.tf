# Datadog monitoring provider
resource "sedai_create_datadog_monitoring_provider" "datadog_mp" {
  account_id       = sedai_create_account.eks_account.id
  integration_type = sedai_create_account.eks_account.integration_type
  api_key          = "api_key"
  application_key  = "application_key"
}
