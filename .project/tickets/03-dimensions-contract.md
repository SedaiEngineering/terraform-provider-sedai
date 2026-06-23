# 03 — Fix the `*_dimensions` contract (Optional+Computed + inverted IsNull)

**Severity:** High
**Component:** **two repos** — inverted-guard fix in the provider
(`create_{gke,fp,datadog,newrelic,vm}_monitoring_provider.go`); AZ-key fix in the SDK
(`sedai-sdk-go monitoringProvider.go`). Contract decision applies to all providers with dimensions.
**Breaking change:** Possibly (depends on chosen contract)
**Related:** `/tmp/tf.md` A3, B2. Credentials-key issue split out to ticket 01 (and `tf.md` A7 is reversed there — backend uses singular `credentialsDetail`).

## Problem

Two compounding issues with the `*_dimensions` list attributes
(`lb_dimensions`, `app_dimensions`, `instance_dimensions`, …):

1. **Muddled type.** They are modeled `Optional + Computed`. For user-supplied
   filter lists this makes it nearly impossible to ever set the list back to
   empty, and invites perpetual diffs (the server-returned value fights the
   config).
2. **Broken write path.** The send logic uses inverted null checks —
   `if plan.LbDimensions.IsNull()` then convert (GKE ~lines 293–312; FP ~lines
   360–395; same class of bug in Datadog/NewRelic per tf.md A3). Net effect:
   user-set dimensions are dropped and omitted dimensions get clobbered to `[]`.
   Also `azDimensions` has a read/write key mismatch (`azDimensions` vs
   `aZDimensions`, tf.md B2) → always parses back empty → perpetual drift.

## Why it's wrong

A list filter is either user-managed (so: `Optional`, not Computed, and the write
path must send exactly what the user set) or server-derived (so: `Computed`-only).
The current both-at-once design plus the inverted guard means dimensions are
simultaneously broken and noisy.

## Desired behavior

1. **Decide the contract.** If dimensions are user-supplied filters → make them
   `Optional` only (drop `Computed`). If they are server-derived defaults the user
   can override → keep `Optional + Computed` **with** `UseStateForUnknown()`
   (ticket 06) and ensure Read round-trips them faithfully.
2. **Fix the inverted write guard** (provider repo) so the correct branch sends the
   user's values (`!IsNull()`). Confirmed inverted in GKE (`create_gke_*:293–312`)
   and FP (`create_fp_*:390`); apply to Datadog, NewRelic, VM per tf.md A3.
   **Azure is already correct** (`create_azure_*:225` uses `!IsNull()`) — exclude it
   from this fix.
3. **Fix the `azDimensions` key mismatch — direction now resolved.** The canonical
   backend field is **`azDimensions`** (`SedaiMonitoringProvider.azDimensions`,
   getter `getAzDimensions()` → JSON `azDimensions`); **no `aZDimensions` exists
   anywhere in `sedai-models`**. So the **read path is already correct** and the bug
   is the **write key in the SDK**: `monitoringProvider.go:349` (Datadog), `:579` and
   `:823` (FP) send `"aZDimensions"`. Change those three to `"azDimensions"`. Because
   the backend uses `FAIL_ON_UNKNOWN_PROPERTIES=false`, the mistyped key is silently
   dropped today → AZ dimensions never persist → perpetual drift.

## Acceptance criteria

- [ ] Setting `lb_dimensions = ["x"]` actually sends `["x"]` and round-trips.
- [ ] Omitting a dimension list does not clobber the backend to `[]`.
- [ ] AZ dimensions persist without perpetual drift.
- [ ] Repeated `plan` on unchanged dimensions shows no diff.
- [ ] Acceptance tests per provider cover set / unset / change.

## Effort
~1 day (touches several files; needs backend key verification).
