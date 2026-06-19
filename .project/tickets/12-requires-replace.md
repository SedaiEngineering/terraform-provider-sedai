# 12 — `RequiresReplace` on immutable identity fields

**Severity:** High
**Component:** all `create_*_monitoring_provider.go`, `create_account.go`
**Breaking change:** Behavioral (changing these now forces replace instead of update)

## Problem

Conceptually immutable identity attributes are not marked `RequiresReplace`:

- `account_id` on **every** monitoring provider — the MP belongs to one account;
  re-pointing it is a different resource.
- `cloud_provider` on `sedai_account` — cannot meaningfully change for an existing
  account.

Today, changing one of these re-calls the create/upsert endpoint with undefined
rebinding behavior rather than a clean replace.

Note: the settings resources already do this correctly — `account_id` /
`group_id` / `resource_id` use `RequiresReplace` (see `account_settings.go`,
`group_settings.go`, `resource_settings.go`). This ticket extends the same
discipline to the resources that lack it.

## Why it's wrong

`terraform-provider-aws` consistently marks immutable-after-create fields with
`RequiresReplace()` (framework) / `ForceNew` (SDKv2). It makes the plan honest:
the user sees `-/+ destroy and then create replacement` instead of a misleading
in-place update.

## Desired behavior

- Add `stringplanmodifier.RequiresReplace()` to:
  - `account_id` on all 7 monitoring providers.
  - `cloud_provider` on `sedai_account` (confirm with product that it's truly
    immutable; if a cloud can be changed in place, leave it).
- Audit other attributes for immutability (e.g. is `integration_type` mutable on
  an account?). Mark accordingly.

## Acceptance criteria

- [ ] Changing `account_id` on a monitoring provider produces a replace plan.
- [ ] Changing `cloud_provider` produces a replace plan (if confirmed immutable).
- [ ] Acceptance tests assert the replace behavior.
- [ ] Docs note which attributes force replacement.

## Effort
~Half a day + product confirmation on immutability.
