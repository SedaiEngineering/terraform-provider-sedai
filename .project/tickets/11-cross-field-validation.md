# 11 — Declare cross-field constraints (account / CloudWatch / Federated Prometheus)

**Severity:** High
**Component:** `create_account.go`, `create_cloudwatch_monitoring_provider.go`, `create_fp_monitoring_provider.go`
**Breaking change:** No (rejects combinations that already fail at apply)

## Problem

`sedai_account` is a single 24-attribute resource where ~16 attributes are
conditionally required depending on `cloud_provider`, with **zero declarative
enforcement**:

- Azure → `tenant_id`, `subscription_id`, `client_id`, `client_secret`
- GCP → `service_account_json`, `project_id`
- AWS → (`role` + `external_id`) **or** (`access_key` + `secret_key`)
- KUBERNETES → `cluster_provider` (+ provider-specific creds)

Invalid combinations are only caught at runtime by the SDK. Same pattern in:
- **CloudWatch:** `use_account_credentials` vs explicit `access_key`/`secret_key`/
  `role` — a mutually-exclusive choice with no guardrail.
- **Federated Prometheus / Victoria Metrics:** three auth modes (bearer token /
  client credentials / none) — a textbook `ExactlyOneOf` group, currently
  resolved by priority-order `if` logic in the resource.

## Why it's wrong

Both reference providers declare these with framework validators
(`stringvalidator.ConflictsWith`, `AlsoRequires`, `ExactlyOneOf`,
`objectvalidator`/`schemavalidator`), giving users **plan-time** errors and
self-documenting schema. The flat Sedai schema is hard to use correctly and
offers no feedback until apply.

## Desired behavior

- Account: add cross-field validators expressing the per-cloud required/forbidden
  sets. Use `terraform-plugin-framework-validators` (`schemavalidator.*` with
  path expressions; consider `*.AtLeastOneOf` / `*.Conflicts` keyed off
  `cloud_provider`). Where casing matters, also resolve managed-services
  cloud-correctness from ticket 10.
- CloudWatch: `ConflictsWith` between `use_account_credentials = true` and the
  explicit credential attributes; `AlsoRequires` to pair `access_key`+`secret_key`
  and `role`+`external_id`.
- FP/VM: `ExactlyOneOf` (or at least `Conflicts`) across the three auth modes.

## Acceptance criteria

- [ ] An Azure account missing `tenant_id` fails at **plan** with a clear message.
- [ ] CloudWatch with `use_account_credentials = true` **and** `access_key` set
      fails at plan.
- [ ] FP with both `bearer_token` and `client_id`/`client_secret` set fails at
      plan.
- [ ] Validator tests cover each cloud/auth combination.

## Effort
~1–1.5 days (the account matrix is the bulk).
