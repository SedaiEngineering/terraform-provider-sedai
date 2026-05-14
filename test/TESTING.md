# Local Testing Guide — Sedai Terraform Provider

## Prerequisites

- Go 1.22+
- Terraform CLI installed
- Access to a running Sedai instance
- Cloud credentials (AWS role/keys, Azure service principal, or GCP service account)

---

## Build & Install

Build the provider binary:

```bash
cd terraform-provider-sedai
mkdir -p bin
go build -o bin/terraform-provider-sedai .
```

Then add a `dev_overrides` block to `~/.terraformrc` pointing at the `bin/` directory:

```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "hashicorp.com/io/sedai" = "/Users/harikrishnana/Sedai/sedai-terraform/terraform-provider-sedai/bin"
  }
  direct {}
}
```

> With `dev_overrides` active, `terraform init` will print a warning that the provider is overridden — this is expected. You do **not** need a `version` constraint in your test configs.

### Using local sedai-sdk-go changes

If you've made changes to `sedai-sdk-go/` locally, add a `replace` directive in `terraform-provider-sedai/go.mod`:

```go
replace github.com/SedaiEngineering/sedai-sdk-go => ../sedai-sdk-go
```

Then run `go mod tidy` before building.

---

## Environment Variables

```bash
export SEDAI_BASE_URL="https://your-sedai-instance.com"
export SEDAI_API_TOKEN="your-api-token"
```

---

## Test Configurations

Each config below creates an account and its monitoring provider together. Terraform handles the creation order automatically via the `account_id` reference.

### AWS — Account + CloudWatch

```hcl
# test/aws/main.tf
terraform {
  required_providers {
    sedai = { source = "hashicorp.com/io/sedai" }
  }
}

resource "sedai_create_account" "aws" {
  name             = "test-aws-account"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789:role/SedaiRole"
  external_id      = "your-external-id"
}

resource "sedai_create_cloudwatch_monitoring_provider" "aws" {
  account_id              = sedai_create_account.aws.id
  use_account_credentials = true
  lb_dimensions           = ["load_balancer_name"]
  app_dimensions          = ["application_id"]
  instance_dimensions     = ["instance_id"]

  depends_on = [sedai_create_account.aws]
}

output "account_id" { value = sedai_create_account.aws.id }
output "mp_id"      { value = sedai_create_cloudwatch_monitoring_provider.aws.id }
```

### Azure — Account + Azure Monitor

```hcl
# test/azure/main.tf
terraform {
  required_providers {
    sedai = { source = "hashicorp.com/io/sedai" }
  }
}

variable "tenant_id"       {}
variable "subscription_id" {}
variable "client_id"       {}
variable "client_secret"   { sensitive = true }

resource "sedai_create_account" "azure" {
  name             = "test-azure-account"
  cloud_provider   = "AZURE"
  integration_type = "AGENTLESS"
  tenant_id        = var.tenant_id
  subscription_id  = var.subscription_id
  client_id        = var.client_id
  client_secret    = var.client_secret
}

resource "sedai_create_azure_monitoring_provider" "azure" {
  account_id    = sedai_create_account.azure.id
  client_id     = var.client_id
  client_secret = var.client_secret

  depends_on = [sedai_create_account.azure]
}

output "account_id" { value = sedai_create_account.azure.id }
output "mp_id"      { value = sedai_create_azure_monitoring_provider.azure.id }
```

### GCP — Account + GKE Monitoring (baseline, already implemented)

```hcl
# test/gcp/main.tf
terraform {
  required_providers {
    sedai = { source = "hashicorp.com/io/sedai" }
  }
}

variable "project_id" {}

resource "sedai_create_account" "gcp" {
  name                 = "test-gcp-account"
  cloud_provider       = "KUBERNETES"
  cluster_provider     = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = var.project_id
  zone                 = "us-central1-a"
  service_account_json = file("sa.json")
}

resource "sedai_create_gke_monitoring_provider" "gcp" {
  account_id           = sedai_create_account.gcp.id
  project_id           = var.project_id
  service_account_json = file("sa.json")

  depends_on = [sedai_create_account.gcp]
}

output "account_id" { value = sedai_create_account.gcp.id }
output "mp_id"      { value = sedai_create_gke_monitoring_provider.gcp.id }
```

---

## Run a Test

```bash
cd test/aws        # or test/azure, test/gcp

terraform init
terraform plan     # preview — no changes made
terraform apply    # creates account, then monitoring provider
terraform output   # verify both IDs are populated
terraform destroy  # teardown (MP deleted first, then account)
```

For Azure, pass variables via a `.tfvars` file:

```bash
# test/azure/terraform.tfvars
tenant_id       = "your-tenant-id"
subscription_id = "your-subscription-id"
client_id       = "your-client-id"
client_secret   = "your-client-secret"
```

```bash
terraform apply -var-file="terraform.tfvars"
```

---

## Unit & Acceptance Tests

```bash
# Unit tests (no live API)
cd terraform-provider-sedai
go test ./internal/provider/... -v

# Acceptance tests (hits real Sedai API)
TF_ACC=1 go test ./internal/provider/... -v -run TestAcc
```
