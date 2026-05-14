# Sedai Terraform Provider — Resource Reference

## AWS

### `sedai_create_account`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Account name |
| `cloud_provider` | string | yes | Must be `"AWS"` |
| `integration_type` | string | yes | `"AGENTLESS"` or `"AGENT_BASED"` |
| `role` | string | no | IAM role ARN for role-based auth |
| `external_id` | string | no | External ID for the IAM role |
| `access_key` | string | no | AWS access key (static credentials) |
| `secret_key` | string | no | AWS secret key (static credentials) |
| `user_selected_managed_services` | list(string) | no | AWS services to enable (see below) |

**Credential priority:** `role` → `access_key`/`secret_key` → env-supplied

**`user_selected_managed_services` valid values:**

| Value | Description |
|-------|-------------|
| `"LAMBDA"` | Lambda functions |
| `"EC2"` | EC2 virtual machines |
| `"ECS"` | ECS containers |
| `"EKS"` | EKS Kubernetes |
| `"EBS"` | EBS storage |
| `"EFS"` | EFS storage |
| `"S3"` | S3 storage |
| `"RDS"` | RDS databases |
| `"DYNAMO_DB"` | DynamoDB |
| `"DATABRICKS"` | Databricks on AWS |

```hcl
resource "sedai_create_account" "aws" {
  name                           = "my-aws-account"
  cloud_provider                 = "AWS"
  integration_type               = "AGENTLESS"
  role                           = "arn:aws:iam::123456789:role/SedaiRole"
  external_id                    = "ext-123456789"
  user_selected_managed_services = ["EC2", "ECS", "RDS", "S3"]
}
```

---

### `sedai_create_cloudwatch_monitoring_provider`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `account_id` | string | yes | Sedai account ID (from `sedai_create_account`) |
| `use_account_credentials` | bool | no | Reuse the account's IAM role (default: `true`) |
| `role` | string | no | Override role ARN for explicit credentials |
| `external_id` | string | no | External ID for the override role |
| `access_key` | string | no | Static AWS access key override |
| `secret_key` | string | no | Static AWS secret key override |
| `lb_dimensions` | list(string) | no | LB metric dimensions (default: `["load_balancer_name"]`) |
| `app_dimensions` | list(string) | no | App metric dimensions (default: `["application_id"]`) |
| `instance_dimensions` | list(string) | no | Instance metric dimensions (default: `["instance_id"]`) |

**Note:** When `use_account_credentials = true` (default), the server reuses the account's IAM role — no need to repeat credentials here.

```hcl
resource "sedai_create_cloudwatch_monitoring_provider" "aws" {
  account_id              = sedai_create_account.aws.id
  use_account_credentials = true

  depends_on = [sedai_create_account.aws]
}
```

---

## Azure

### `sedai_create_account`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Account name |
| `cloud_provider` | string | yes | Must be `"AZURE"` |
| `integration_type` | string | yes | `"AGENTLESS"` or `"AGENT_BASED"` |
| `tenant_id` | string | yes | Azure Active Directory tenant ID |
| `subscription_id` | string | yes | Azure subscription ID |
| `client_id` | string | yes | Service principal client ID |
| `client_secret` | string | yes | Service principal client secret (sensitive) |
| `user_selected_managed_services` | list(string) | no | Azure services to enable (see below) |

**`user_selected_managed_services` valid values:**

| Value | Description |
|-------|-------------|
| `"VM"` | Azure Virtual Machines |
| `"AKS"` | AKS Kubernetes |
| `"AZURE_DISK"` | Azure Disk storage |
| `"AZURE_BLOB"` | Azure Blob storage |
| `"DATABRICKS"` | Databricks on Azure |

```hcl
resource "sedai_create_account" "azure" {
  name                           = "my-azure-account"
  cloud_provider                 = "AZURE"
  integration_type               = "AGENTLESS"
  tenant_id                      = var.tenant_id
  subscription_id                = var.subscription_id
  client_id                      = var.client_id
  client_secret                  = var.client_secret
  user_selected_managed_services = ["VM", "AKS", "AZURE_DISK", "AZURE_BLOB"]
}
```

---

### `sedai_create_azure_monitoring_provider`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `account_id` | string | yes | Sedai account ID (from `sedai_create_account`) |
| `lb_dimensions` | list(string) | no | LB metric dimensions (default: `["load_balancer_name"]`) |
| `app_dimensions` | list(string) | no | App metric dimensions (default: `["application_id"]`) |
| `instance_dimensions` | list(string) | no | Instance metric dimensions (default: `["instance_id"]`) |
| `region_dimensions` | list(string) | no | Region metric dimensions |
| `container_dimensions` | list(string) | no | Container metric dimensions |
| `namespace_dimensions` | list(string) | no | Namespace metric dimensions |
| `cluster_dimensions` | list(string) | no | Cluster metric dimensions |

**Note:** Always uses account credentials (`useAccountCredentials: true`). No credentials needed on the MP resource itself.

```hcl
resource "sedai_create_azure_monitoring_provider" "azure" {
  account_id = sedai_create_account.azure.id

  depends_on = [sedai_create_account.azure]
}
```

---

## GCP

### `sedai_create_account`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Account name |
| `cloud_provider` | string | yes | Must be `"GCP"` |
| `integration_type` | string | yes | `"AGENTLESS"` or `"AGENT_BASED"` |
| `project_id` | string | yes | GCP project ID |
| `service_account_json` | string | yes | GCP service account JSON key (sensitive) |
| `user_selected_managed_services` | list(string) | no | GCP services to enable (see below) |

**`user_selected_managed_services` valid values:**

| Value | Description |
|-------|-------------|
| `"GCE"` | Google Compute Engine VMs |
| `"DATAFLOW"` | Google Dataflow streaming jobs |
| `"GCP_DISK"` | GCP persistent disks |
| `"CLOUD_STORAGE"` | Google Cloud Storage buckets |
| `"BIG_QUERY"` | BigQuery — also triggers auto-creation of `BQMONITORING` provider |
| `"DATABRICKS"` | Databricks on GCP |

**Auto-created monitoring providers:**
- `GKEMONITORING` — always auto-created on account creation
- `BQMONITORING` — auto-created only when `"BIG_QUERY"` is in `user_selected_managed_services`

```hcl
# Without BigQuery (GKEMONITORING only)
resource "sedai_create_account" "gcp" {
  name                 = "my-gcp-account"
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = var.project_id
  service_account_json = var.service_account_json
}

# With BigQuery (GKEMONITORING + BQMONITORING)
resource "sedai_create_account" "gcp" {
  name                           = "my-gcp-account"
  cloud_provider                 = "GCP"
  integration_type               = "AGENTLESS"
  project_id                     = var.project_id
  service_account_json           = var.service_account_json
  user_selected_managed_services = ["BIG_QUERY"]
}
```

> GCP monitoring providers are auto-created by Sedai on account creation. No separate `sedai_create_*_monitoring_provider` resource is needed for GCP.

---

### `sedai_create_gke_monitoring_provider`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `account_id` | string | yes | Sedai account ID |
| `project_id` | string | yes | GCP project ID |
| `service_account_json` | string | yes | GCP service account JSON (sensitive) |
| `integration_type` | string | no | `"AGENTLESS"` or `"AGENT_BASED"` (default: `"AGENT_BASED"`) |
| `lb_dimensions` | list(string) | no | Default: `["destination_service_name"]` |
| `app_dimensions` | list(string) | no | Default: `["application_id"]` |
| `instance_dimensions` | list(string) | no | Default: `["pod_name"]` |
| `region_dimensions` | list(string) | no | Default: `["location"]` |
| `container_dimensions` | list(string) | no | Default: `["container_name"]` |
| `namespace_dimensions` | list(string) | no | Default: `["destination_service_namespace", "namespace_name"]` |
| `cluster_dimensions` | list(string) | no | Default: `["cluster_name"]` |

```hcl
resource "sedai_create_gke_monitoring_provider" "gcp" {
  account_id           = sedai_create_account.gcp.id
  project_id           = var.project_id
  service_account_json = var.service_account_json

  depends_on = [sedai_create_account.gcp]
}
```

---

## Kubernetes

### `sedai_create_account`

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | yes | Account name |
| `cloud_provider` | string | yes | Must be `"KUBERNETES"` |
| `integration_type` | string | yes | `"AGENTLESS"` or `"AGENT_BASED"` |
| `cluster_provider` | string | yes | `"AWS"`, `"GCP"`, `"AZURE"`, or `"SELF_MANAGED"` |
| `cluster_url` | string | no | Cluster API URL (required for agentless) |
| `ca_certificate` | string | no | CA certificate for cluster TLS |
| `region` | string | no | Cluster region |
| `role` | string | no | IAM role ARN (AWS cluster provider) |
| `external_id` | string | no | External ID for IAM role |
| `access_key` | string | no | AWS access key (AWS cluster provider) |
| `secret_key` | string | no | AWS secret key (AWS cluster provider) |
| `project_id` | string | no | GCP project ID (GCP cluster provider) |
| `zone` | string | no | GCP zone (GCP cluster provider) |
| `is_zonal_cluster` | bool | no | Whether the GKE cluster is zonal |
| `service_account_json` | string | no | GCP service account JSON (GCP cluster provider) |
