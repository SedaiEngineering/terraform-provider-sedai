# 15 — Testing strategy & customer-issue reproductions (cross-cutting)

**Severity:** High (gates every other ticket)
**Component:** whole provider + SDK; CI
**Breaking change:** No

## Goal

Nothing in this backlog ships without tests, and **every customer-reported issue
must have a reproduction test** that fails on the current code and passes after the
fix — so we can prove the fix and prevent regressions. This ticket defines the test
layers, the customer-issue→test matrix, and the CI wiring. Each other ticket's
Definition of Done references the relevant layer here.

## Test layers

1. **Unit (Go, no network)** — schema correctness, validators (`OneOf`, ranges,
   cross-field), `conversions.go` pointer/nullable helpers, plan modifiers. Fast,
   run on every PR.
2. **SDK contract / serialization (Go, mocked HTTP via `gock`/`httptest`)** —
   assert the *exact JSON shape* the SDK sends against the **backend model field
   names** (our source of truth from the sedai-core trace), and that responses
   parse correctly. This layer is what catches the credentials-key class of bug
   (ticket 01). Golden-payload tests per resource. Run on every PR.
3. **Acceptance (`TF_ACC=1`, real tenant via `SEDAI_BASE_URL`/`SEDAI_API_TOKEN`)**
   — full CRUD lifecycle, `terraform import`, **drift detection**, **plan
   idempotency** (apply → plan shows no diff), and parallelism. Gated behind env;
   run nightly + pre-release against a dedicated test tenant.
4. **Regression reproductions** — one named test per customer-reported issue
   (below), each asserting the specific symptom is gone.

> The plan-idempotency assertion (apply, then `terraform plan` must report
> `0 to add, 0 to change, 0 to destroy`) is the single highest-value check — it
> catches phantom diffs (C1), counts churn (C2), MP `(known after apply)` churn
> (06), dimensions drift (03), and priority cache drift (07) in one place. Add it
> to every resource's acceptance test.

## Customer-issue → reproduction test matrix

| Source | Issue | Reproduction test (must fail before fix) |
|---|---|---|
| tf.md C1 | Phantom `update in-place` on a no-op stack | Acceptance: apply a multi-resource stack twice; second `plan` is empty. |
| tf.md C2 | `*_count` / `resource_counts` churn in plan | Acceptance: group with `auto_refresh=true`; repeated `plan` shows no diff; counts present only on `data.sedai_group` (ticket 04). |
| tf.md C3 / SED-21371 | ~80% apply failure, no retry | SDK unit: mock server returns `EOF`/503 then 200 → SDK retries and succeeds. Acceptance: apply with `-parallelism=10` across many resources completes. |
| tf2.md Part 1 | Parallel-apply hang / state corruption (48 res) | Acceptance: parallel create of accounts+groups+settings+MPs; all tracked in state; ID written before readiness poll; `import` recovers an out-of-band-created resource; 404 on Read → clean removal. |
| tf2.md Critical | Secrets in plaintext | Unit/acceptance: `plan`/state output redacts every `Sensitive` attribute. |
| 01 | MP credentials dropped (plural key) | SDK contract: each MP payload contains `details.credentialsDetail` (singular); live create yields non-null credentials. |
| 05 | `autonomous_action_without_traffic` server-overridden | Acceptance: kube/ecs settings with `is_prod=true`; repeated `plan` shows no perpetual diff. |
| 02 | Created group has no members | Acceptance: create group; after apply, membership populated for discovered inventory (incl. `auto_refresh=false`). |
| 09 | TOCTOU / wrong group ID | SDK + acceptance: concurrent creates of distinct names never cross-bind IDs; ID sourced from create response. |
| 07 | Priority read-after-write drift (12h cache) | Acceptance: set priority, apply, immediate `plan` is empty (requires backend `@CacheEvict` fix). |
| 08 | Account drift not detected | Acceptance: change an account attr out-of-band; `plan` shows drift. |

## Per-ticket DoD (applies to tickets 01–14)

Each ticket is "done" only when it includes:
- [ ] Unit tests for any schema/validator/conversion change.
- [ ] SDK contract test for any request/response shape change (asserted against the
      backend model field names).
- [ ] An acceptance test covering create/read/update/delete/import as relevant,
      **including the plan-idempotency assertion**.
- [ ] For tickets tied to a customer issue, the named regression test from the
      matrix above.

## CI wiring

- [ ] Run layers 1–2 on every PR (`go test ./...`, no network).
- [ ] Run layers 3–4 nightly and pre-release against the test tenant; document the
      required env (`SEDAI_BASE_URL`, `SEDAI_API_TOKEN`) and how to provision a
      throwaway tenant/account for them.
- [ ] Track coverage; new code must not lower it. Surface the customer-issue
      regression suite as a named CI job so its status is visible per release.

## Acceptance criteria

- [ ] Every customer-reported issue has a named, passing regression test.
- [ ] `go test ./...` (unit + contract) is green and wired into PR CI.
- [ ] Acceptance suite runs against a tenant and is green pre-release.
- [ ] Plan-idempotency assertion exists for every resource.

## Effort
Ongoing; ~2–3 days to stand up the harness (contract-test scaffold + acceptance
test framework + CI), then per-ticket increments.
