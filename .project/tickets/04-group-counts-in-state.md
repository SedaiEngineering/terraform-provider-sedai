# 04 — Remove volatile resource counts from the `sedai_group` resource

**Severity:** Critical
**Component:** `internal/provider/create_group.go`
**Breaking change:** No (removing computed-only attributes)
**Related:** `/tmp/tf.md` C1/C2

## Problem

`sedai_group` stores live cloud-inventory counts as `Computed` attributes on the
**managed resource**: `lambda_count`, `ec2_count`, `ecs_count`, `kube_count`,
`s3_count`, `ebs_count`, `azure_vm_count`, `azure_lb_count`, `streaming_count`,
plus the dynamic `resource_counts` map.

- Model + schema: `create_group.go` (~lines 51–62 and ~148–161)
- Refreshed from the backend on every read: `refreshCountsFromBackend` /
  `populateCounts` (~lines 433–467)
- Written on Create (~line 226)

These attributes have **no `UseStateForUnknown` plan modifier** (contrast `id`,
which has one).

## Why it's wrong

This violates the core Terraform contract: state holds *desired config + server
identity*, never runtime telemetry. Neither `terraform-provider-aws` nor
`terraform-provider-github` ever stores live inventory on a managed resource.

Concrete effects:
1. Any unrelated change marks every count `(known after apply)` — the wall of
   diffs in the customer screenshot. One real change detonates all counts.
2. Cloud inventory drifts between runs → perpetual `update in-place` with **zero**
   config change. Actively incompatible with `auto_refresh = true`, which changes
   membership over time by design.
3. `refreshCountsFromBackend` zeroes counts on a transient read failure → writes
   `0`s into state → next plan shows `0 -> <real>`.

## Backend mechanism (verified via `sedai-core`)

Counts are not live cloud state — they are read from a **materialized membership
table `group_resources`** via `getResourceTypeCountPerGroup()`, and that table is
only as fresh as the last materialization:

- `group_resources` is (re)computed by `SedaiGroupServiceV2.updateSedaiGroupResources(groupId)`
  (`SedaiGroupServiceV2.java:587–738`), which **checksums** the resolved resource
  set and only rewrites on change.
- **Group create does NOT materialize membership.** `createGroup`
  (`SedaiGroupServiceV2.java:1701–1725`) saves the definition + inits settings,
  but never calls `updateSedaiGroupResources`. So a freshly-created group has
  **zero** `group_resources` rows → counts read back as 0.
- Membership is later filled by `UpdateGroupResourcesMinion` — a Quartz
  cluster-singleton that only refreshes groups with **`autoRefresh=true`**
  (`UpdateGroupResourcesMinion.java:68`), on its own schedule, and is skipped
  entirely if its DB config flags are all false.

Net: counts are **0 at create → N after the minion runs** (guaranteed `0 -> N`
drift), and for `auto_refresh = true` they keep drifting as cloud inventory
changes. There is no point in time at which storing them on the managed resource
is stable. This is the mechanistic proof behind the "scary plan."

## Desired behavior

- Remove all `*_count` attributes and `resource_counts` from the **`sedai_group`
  resource** schema and model.
- Remove `refreshCountsFromBackend` calls from Create/Update/Read.
- Keep these counts **only** on the `data.sedai_group` data source, where volatile
  read-only numbers are appropriate (they already exist there).

## Acceptance criteria

- [ ] `sedai_group` resource schema no longer contains any count attribute.
- [ ] A `terraform plan` on an unchanged group with `auto_refresh = true` shows
      **no** diff across repeated runs while cloud inventory changes.
- [ ] `data.sedai_group` still exposes counts unchanged.
- [ ] Docs (`docs/resources/group.md`) regenerated; counts documented only on the
      data source.
- [ ] State-migration / upgrade note added to CHANGELOG (computed attrs dropping
      is non-breaking, but call it out).

## Effort
~Half a day incl. docs + a plan-stability check.
