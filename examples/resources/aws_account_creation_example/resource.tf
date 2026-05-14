# AWS account — agentless, role-based
resource "sedai_create_account" "aws_account_role_based" {
  name                           = "my-aws-account"
  cloud_provider                 = "AWS"
  integration_type               = "AGENTLESS"
  role                           = "arn:aws:iam::123456789:role/SedaiRole"
  external_id                    = "ext-123456789"
  user_selected_managed_services = ["EC2", "ECS", "RDS", "S3"]
}

# AWS account — agentless, key-based
resource "sedai_create_account" "aws_account_key_based" {
  name             = "my-aws-account"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  access_key       = "access_key"
  secret_key       = "secret_key"
}
