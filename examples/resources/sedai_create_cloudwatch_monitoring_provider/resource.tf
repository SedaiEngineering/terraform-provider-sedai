# CloudWatch — using account credentials (default)
resource "sedai_create_cloudwatch_monitoring_provider" "cloudwatch_account_creds" {
  account_id              = sedai_create_account.aws_account.id
  use_account_credentials = true

  depends_on = [sedai_create_account.aws_account]
}

# CloudWatch — explicit role override
resource "sedai_create_cloudwatch_monitoring_provider" "cloudwatch_explicit_role" {
  account_id              = sedai_create_account.aws_account.id
  use_account_credentials = false
  role                    = "arn:aws:iam::123456789:role/SedaiMonitoringRole"
  external_id             = "ext-123456789"

  depends_on = [sedai_create_account.aws_account]
}
