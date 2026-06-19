# 02 — A created group has no members until a minion runs (and `auto_refresh=false` groups may never populate)

**Severity:** High
**Component:** `create_group.go` (Create); SDK `groups.CreateGroup`
**Breaking change:** No
**Backend refs:** `SedaiGroupServiceV2.createGroup` (1701–1725), `UpdateGroupResourcesMinion.runMinion` (52–97), `SedaiGroupApi.updateGroupResources` (88–107)

## Problem

`sedai_group` Create persists the group definition and initializes its settings,
but **does not materialize group membership** (`group_resources`):

- Backend `createGroup` saves the definition + calls `initSettingForGroup`, and
  **never** calls `updateSedaiGroupResources`.
- The SDK's `CreateGroup` likewise does **not** fire
  `/api/sedaigroup/updategroupresources/{id}` — only `UpdateGroup` does
  (`groups.go:338`).
- The background `UpdateGroupResourcesMinion` refreshes membership **only for
  groups with `autoRefresh = true`** (`UpdateGroupResourcesMinion.java:68`), on
  its own Quartz schedule, and is skipped entirely when its DB config flags are
  off.

## Impact / UX

- `terraform apply` reports success and the group exists, but it **contains no
  resources and drives no optimization** until membership is materialized.
- For `auto_refresh = true`: membership appears only after the next minion cycle
  (eventual, minutes).
- For `auto_refresh = false`: membership **can remain empty indefinitely**. Create
  doesn't materialize it; the periodic `UpdateGroupResourcesMinion` skips it
  (auto_refresh gate, verified at line 68); and the only event-driven refreshers
  are narrowly scoped — `SedaiResourceSettingsService.onEvent` on a cloud-update
  refreshes just the *account's* dynamic group (`getGroupIdForAccount`, one group),
  and `handleGroupResourceAndSettingsRefresh` only refreshes an explicitly-passed
  set of group IDs. None of these blanket-refresh arbitrary manual groups. So a
  static group typically gets members only when a later `terraform apply` edits its
  definition (which fires `updategroupresources`). A user who creates one and never
  edits it can have a permanently empty, no-op group. Silent and confusing.

## Defensive fix

- In `sedai_group` Create, after the group is created, fire the membership
  recompute (the same `updategroupresources/{id}` call `UpdateGroup` already
  makes — expose `refreshGroupResources` or an SDK `MaterializeGroupResources`
  helper and call it on Create too). Best-effort, with errors surfaced as a
  warning (not a hard failure), since it depends on cloud discovery having run.
- Document clearly: membership reflects **already-discovered** cloud inventory at
  apply time; resources discovered later are picked up by the minion only when
  `auto_refresh = true`. Set the expectation that a brand-new account's inventory
  may not be discovered yet, so membership can legitimately start empty.

## Acceptance criteria

- [ ] After `terraform apply` of a `sedai_group`, `group_resources` is populated
      for inventory already known to Sedai (verified on a test tenant).
- [ ] An `auto_refresh = false` group has members immediately after create **for
      inventory already discovered by Sedai**, not only after a later edit. (If the
      account's inventory has not been discovered yet, an empty membership is correct
      and expected — the test must seed discovered inventory first, or assert the
      recompute was *called*, not that it returned members.)
- [ ] Recompute failures degrade to a warning, never a failed apply.
- [ ] Docs explain the discovery/eventual-consistency dependency.

## Effort
~Half a day (SDK helper + Create wiring + a live check).
