# Azure account — agentless
resource "sedai_create_account" "azure_account" {
  name                           = "my-azure-account"
  cloud_provider                 = "AZURE"
  integration_type               = "AGENTLESS"
  tenant_id                      = "tenant_id"
  subscription_id                = "subscription_id"
  client_id                      = "client_id"
  client_secret                  = "client_secret"
  user_selected_managed_services = ["VM", "AZURE_DISK", "AZURE_BLOB"]
}
