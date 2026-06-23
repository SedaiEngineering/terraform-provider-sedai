# 10 — Validate core enums at plan time

**Severity:** Critical
**Component:** `internal/provider/create_account.go`, `internal/provider/validators.go`
**Breaking change:** No (rejects input that already fails at apply)

## Problem

Several enum-valued attributes on `sedai_account` have **no validators** — valid
values live only in the description text, so validity is checked by the SDK at
**apply** time (or not at all):

- `cloud_provider` (~line 82) — `AWS`, `AZURE`, `GCP`, `KUBERNETES`
- `integration_type` (~line 86) — `AGENTLESS`, `AGENT_BASED`
- `cluster_provider` (~line 90) — `AWS`, `GCP`, `AZURE`, `SELF_MANAGED`
- `user_selected_managed_services` (~line 173) — per-cloud service enum list

This is internally inconsistent: `availability_mode` / `optimization_mode` use
`settingsConfigModeValidator`, and `resource_types` uses
`groupResourceTypeListValidator`. The account enums were simply missed.

## Why it's wrong

Both reference providers validate **every** enum at plan time
(`stringvalidator.OneOf` / `StringEnumType`). Without it, a typo — or, ironically,
the lowercase `"aws"` some style guides ask for — passes `plan` and fails at
`apply`, frequently **after** dependent resources (monitoring providers, groups)
are already mid-create, leaving partial state.

## Desired behavior

- Add `stringvalidator.OneOf(...)` to `cloud_provider`, `integration_type`,
  `cluster_provider`.
- Add a list validator
  (`listvalidator.ValueStringsAre(stringvalidator.OneOf(...))`) to
  `user_selected_managed_services`. Note the valid set is cloud-specific; either
  validate the union here and enforce cloud-correctness in cross-field validation
  (ticket 11), or build a per-cloud validator.
- **Keep the existing SCREAMING_SNAKE casing** — do not add a lowercase
  translation layer (see README "NOT doing").
- Audit every other resource for un-validated enum attributes and close the gaps.

## Acceptance criteria

- [ ] `terraform plan` with `cloud_provider = "aws"` (wrong case) fails at plan
      with a clear "expected one of …" message.
- [ ] All enum attributes across all resources have plan-time validators.
- [ ] No casing change to accepted values.
- [ ] Validator unit tests added/updated in `validators_test.go`.

## Effort
~Half a day.
