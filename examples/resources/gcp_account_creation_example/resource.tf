# GCP account — without BigQuery (GKEMONITORING auto-created)
resource "sedai_create_account" "gcp_account" {
  name                 = "my-gcp-account"
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = "project_id"
  service_account_json = "service_account_json_content"
}

# GCP account — with BigQuery (GKEMONITORING + BQMONITORING auto-created)
resource "sedai_create_account" "gcp_account_with_bq" {
  name                           = "my-gcp-account"
  cloud_provider                 = "GCP"
  integration_type               = "AGENTLESS"
  project_id                     = "project_id"
  service_account_json           = "service_account_json_content"
  user_selected_managed_services = ["BIG_QUERY"]
}
