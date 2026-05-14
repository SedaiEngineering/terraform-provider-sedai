# Azure Monitor — uses account credentials automatically
resource "sedai_create_azure_monitoring_provider" "azure_mp" {
  account_id = sedai_create_account.azure_account.id

  depends_on = [sedai_create_account.azure_account]
}
