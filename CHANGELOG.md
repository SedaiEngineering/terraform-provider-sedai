## 2.0.0

### MIGRATION GUIDE — What to change in your existing Terraform templates

If you are upgrading from a previous version, review the following checklist.
Each item describes exactly what to find in your HCL and what to change.

---

**1. Remove count references from `sedai_group` outputs or other resources**

If you reference count attributes on the group resource, switch to the data source:

```hcl
# BEFORE (no longer works):
output "lambda_count" {
  value = sedai_group.my_group.lambda_count
}

# AFTER:
data "sedai_group" "my_group" { name = "my-group" }
output "lambda_count" {
  value = data.sedai_group.my_group.lambda_count
}
```

---

**2. Remove `autonomous_action_without_traffic` from kube/ecs settings blocks**

```hcl
# BEFORE (causes plan error after upgrade):
resource "sedai_group_settings" "prod" {
  ...
  kube_app_settings {
    is_prod                         = true
    autonomous_action_without_traffic = false   # ← REMOVE THIS LINE
  }
  ecs_app_settings {
    autonomous_action_without_traffic = true    # ← REMOVE THIS LINE
  }
}

# AFTER:
resource "sedai_group_settings" "prod" {
  ...
  kube_app_settings {
    is_prod = true
  }
}
```

---

**3. Add `availability_mode` and `optimization_mode` to `sedai_resource_settings`**

These are now Required (were Optional):

```hcl
# BEFORE (causes plan error after upgrade):
resource "sedai_resource_settings" "my_lambda" {
  resource_id        = "abc-123"
  sedai_sync_enabled = true
}

# AFTER:
resource "sedai_resource_settings" "my_lambda" {
  resource_id       = "abc-123"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  sedai_sync_enabled = true
}
```

---

**4. Check `resource_types` on `sedai_group` resources**

If you use resource types that were removed from the accepted list, update them.
The full valid list is in the BREAKING CHANGES section below. Common ones to check:

```hcl
# Types that were REMOVED and will now fail at plan:
# GCP_VM_INSTANCE  → use GCP_VM instead
# GCP_SNAPSHOT, GCP_BACKEND_SERVICE, GCP_DATAFLOW_*, GCP_BIGQUERY_* → remove
# AZURE_BLOB, AZURE_SNAPSHOT → remove (0 discovery emit sites)

# Example fix:
resource "sedai_group" "prod" {
  resource_types = [
    "GCP_VM",        # was GCP_VM_INSTANCE — corrected
    "GCP_DISK",
    "AZURE_VM",
    "AZURE_TAGS",    # new — captures tag-discovered VMs
  ]
}
```

---

**5. No changes needed for (already compatible):**
- `sedai_account` — all existing attributes still work
- `sedai_group` definition fields (name, cloud_account_ids, regions, namespaces etc.)
- `sedai_group_settings` and `sedai_account_settings` — modes still Required, same values
- `sedai_cloudwatch_monitoring_provider` — no HCL changes needed
- All monitoring providers — no HCL changes needed
- `sedai_group_priority` — no HCL changes needed

---

### BREAKING CHANGES

* **`sedai_group` — resource counts removed.** `lambda_count`, `ec2_count`, `ecs_count`,
  `kube_count`, `s3_count`, `ebs_count`, `azure_vm_count`, `azure_lb_count`,
  `streaming_count`, and `resource_counts` no longer exist on the managed resource.
  These caused perpetual plan noise as cloud inventory changed. Use `data.sedai_group`
  to read counts.

* **`sedai_group` — `resource_types` validator updated.** The accepted list is now the
  authoritative set verified against sedai-core discovery code:
  - AWS compute: `AWS_EC2`, `AWS_TAGS`, `AWS_LB`, `AWS_ASG`
  - AWS serverless/container: `AWS_LAMBDA`, `AWS_ECS`
  - AWS database: `AWS_RDS`, `AWS_DYNAMODB`
  - AWS storage: `AWS_EBS`, `AWS_S3`
  - Azure compute: `AZURE_VM`, `AZURE_TAGS`, `AZURE_LB`, `AZURE_VMSS`
  - Azure storage: `AZURE_DISK`, `AZURE_STORAGE_BUCKET`
  - GCP compute: `GCP_VM`, `GCP_TAGS`, `GCP_LB`
  - GCP storage: `GCP_DISK`, `GCP_BUCKET`
  - Kubernetes: `KUBERNETES_DEPLOYMENT`, `KUBERNETES_STATEFULSET`, `KUBERNETES_DAEMONSET`, `KUBERNETES_CRONJOB`

  Note: `AWS_TAGS`, `AZURE_TAGS`, `GCP_TAGS` represent VMs discovered via tag-based
  policies (same physical resource type as `AWS_EC2`/`AZURE_VM`/`GCP_VM`, different
  discovery path). To capture all VMs, include both the direct type and the tag type.

  Note: `GCP_VM` is the correct wire value. `GCP_VM_INSTANCE` has been removed.

* **`sedai_group_settings` / `sedai_account_settings` — `autonomous_action_without_traffic`
  removed** from `kube_app_settings` and `ecs_app_settings` blocks. This field is
  server-derived from `is_prod` for Kube/ECS workloads and cannot be set independently.
  Remove it from any existing HCL configs.

* **`sedai_resource_settings` — `availability_mode` and `optimization_mode` are now
  Required** (were Optional). Any `sedai_resource_settings` block missing either mode
  must be updated to include them.

---

### BUG FIXES

**Stability / reliability (addresses Diligent parallel-apply failures)**

* Fixed HTTP connection pool starvation under parallel apply — `MaxIdleConnsPerHost`
  now set to 10 to match Terraform's default parallelism.
* Added retry with exponential backoff and jitter — transient 429/5xx failures now
  recover automatically without user intervention.
* POST requests no longer retried — prevents duplicate resource creation when a
  backend response is lost in transit (EOF).
* Account ID now written to state immediately after create, before any post-create
  work — prevents state corruption when agent command fetch fails.
* Verify-after-failure on create for all resources — if a POST response is lost,
  the provider polls the backend to confirm creation before erroring.
* Monitoring provider graceful delete — exporter deregistration failures no longer
  surface as errors; the provider verifies the resource is gone and succeeds.
* Unchecked type assertions on API responses replaced with safe extraction — prevents
  process panics when backend returns null fields under parallel load.
* `stringsToList` now returns empty list instead of null for empty slices — fixes
  perpetual `null → []` drift on groups with empty `namespaces`/`regions`/`resource_ids`.

**Credentials / security**

* Fixed monitoring provider credentials silently dropped for GKE, Datadog, NewRelic,
  Federated Prometheus, and Victoria Metrics — wrong JSON key (`credentialsDetails`
  vs `credentialsDetail`) caused credentials to be ignored by the backend.
* API keys and secrets for Datadog, NewRelic, FP, and Victoria Metrics now marked
  `Sensitive` — no longer printed in plan/apply output.

**Plan stability / drift**

* `sedai_sync_enabled` no longer drifts — `false → null` on every plan when the
  attribute is omitted from HCL.
* `auto_refresh` no longer drifts — same fix.
* Group monitoring provider `name` and `integration_type` no longer show
  `(known after apply)` on unchanged plans — `UseStateForUnknown` added.
* Group membership now materializes immediately after create — `auto_refresh=false`
  groups were previously empty indefinitely after create.
* Group ID now read from create response directly — eliminates TOCTOU race condition
  under parallel applies that could bind wrong IDs to wrong groups.
* `azDimensions` key corrected in SDK for Datadog, FP, Victoria Metrics — AZ
  dimensions were silently dropped, causing perpetual drift.
* Inverted `IsNull()` guard fixed in GKE, Datadog, NewRelic, FP dimension builders —
  user-set dimensions were dropped, null dimensions were sent.
* `autonomous_action_without_traffic` removed from kube/ecs — backend overrides this
  from `is_prod`, was causing perpetual drift when both fields were set.
* Group priority 12h cache drift mitigated at provider layer — state now sourced
  from update response. **Full fix requires `@CacheEvict` on
  `SedaiGroupServiceV2.updateGroupsPriority` in sedai-core.**

**Schema correctness**

* `cloud_provider`, `integration_type`, `role`, `external_id` on `sedai_account` now
  have `RequiresReplace` — changing identity fields forces destroy+recreate instead of
  ambiguous in-place update.
* `account_id` on all 8 monitoring providers now has `RequiresReplace`.
* `group_id` in `sedai_group_priority` blocks now has `RequiresReplace`.
* Plan-time enum validators added to `sedai_account` — `cloud_provider`,
  `integration_type`, `cluster_provider`, and `user_selected_managed_services` now
  fail at plan with clear messages instead of mid-apply with partial state.
* Cross-field constraints added: CloudWatch `use_account_credentials` conflicts with
  explicit credentials; FP/VM bearer token conflicts with OAuth2 client credentials;
  Azure/GCP/AWS credential fields conflict with each other on accounts.
* Settings mode fan-out guard — `availability_mode = AUTO` with `app_settings`,
  `bucket_settings`, or `volume_settings` blocks now fails at plan. `optimization_mode
  = CO_PILOT` with `serverless_settings` now fails at plan.
* Account Read now refreshes `subscriptionId`, `tenantId` (Azure), `projectId` (GCP),
  `clusterProvider`, `clusterUrl`, and `userSelectedManagedServices` — previously
  only 4 of 24 attributes were refreshed.

**Drift detection**

* Out-of-band resource deletion now triggers graceful recreate on next plan — previously
  caused a hard error requiring manual `terraform state rm`.

### FEATURES

* New data sources: `data.sedai_group_settings`, `data.sedai_account_settings`,
  `data.sedai_resource_settings`, `data.sedai_group_priority`.

---

## 0.1.0 (Unreleased)

FEATURES:
