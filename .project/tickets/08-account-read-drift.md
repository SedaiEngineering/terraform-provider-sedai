# 08 — `sedai_account` Read refreshes only 4 of 24 attributes (no drift detection)

**Severity:** Critical
**Component:** `internal/provider/create_account.go` (Read, ~lines 242–273)
**Breaking change:** No

## Problem

`sedai_account.Read` only re-sets `id`, `name`, `cloud_provider`, and
`integration_type` from the backend. The other ~20 attributes are never
refreshed: `cluster_provider`, `cluster_url`, `project_id`, `zone`, `region`,
`is_zonal_cluster`, `ca_certificate`, `role`, `external_id`, `tenant_id`,
`subscription_id`, `client_id`, `user_selected_managed_services`, and the
secrets.

## Why it's wrong

Drift detection is a core promise of Terraform: `plan` is supposed to surface
out-of-band changes. `terraform-provider-aws` and `-github` re-set the **full**
model from the API in Read (`fwflex.Flatten` / `d.Set` per attribute). Today an
operator who changes an account's region, managed services, or cluster config in
the Sedai UI will see **no drift** in `terraform plan`.

## Desired behavior

- In Read, refresh the full account model from `account.SearchAccountsById`
  (or a richer fetch if needed), mapping every config attribute back into state.
- **Write-only secrets** (`secret_key`, `client_secret`, `service_account_json`)
  are the standard carve-out — they aren't returned by the API; preserve prior
  state for them (document this, as the monitoring providers already do for their
  credentials).
## Hard prerequisite — SDK model is the real blocker (verified this review)

This is **not** a provider-only change. The Go SDK model is the bottleneck:

- `account.SearchAccountsById` hits `GET /api/site/accounts/{id}` and runs
  `config.TransformJSON(resp["result"], &Account{})`.
- The SDK `Account` struct is `{ID, Name, AccountDetails}`, and `AccountDetails` contains
  **only** `cloudProvider` + `integrationType` (`sdk/sedai/account/models.go`). Those are
  exactly the 4 fields Read sets today.
- So even if the backend GET returns the full account JSON, **`TransformJSON` silently
  drops every field the struct doesn't declare.** Adding `d.Set` calls in the provider
  Read will not help until the struct grows the fields.

**Sequencing:** (1) confirm what `GET /api/site/accounts/{id}` actually returns on the wire;
(2) expand the SDK `Account`/`AccountDetails` model to deserialize the refreshable fields;
(3) only then wire them into provider Read. Steps 1–2 are an `sedai-sdk-go` change and must
land first. Refresh whatever subset the backend returns; document the rest as non-refreshable.

## Acceptance criteria

- [ ] Changing a refreshable account attribute out-of-band produces a drift diff
      on the next `terraform plan`.
- [ ] Write-only secret fields do not produce phantom diffs.
- [ ] Any attribute that genuinely cannot be refreshed (backend limitation) is
      documented as such in `docs/resources/account.md`.

## Effort
~1 day (gated on what the backend GET returns; may spawn an SDK ticket).
