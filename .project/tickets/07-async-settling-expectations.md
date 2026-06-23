# 07 — Group settling: stale read-after-write on priority, and async settings propagation

**Severity:** High (priority drift); Medium (settings propagation docs)
**Component:** `group_priority.go`, `group_settings.go`, docs; backend `SedaiGroupServiceV2.updateGroupsPriority`
**Breaking change:** No
**Backend refs (all verified):** `getSedaiGroupSetting` (1659–1688, `@Cacheable`), `updateGroupsPriority` (1400–1489, no `@CacheEvict`), `updateSedaiGroupResources` (734–736, sets `settings_refreshed=false`), `UpdateGroupResourcesMinion.runMinion` (82–86)

## Verified finding 1 — priority write does NOT evict the read cache (read-after-write staleness)

The priority read path is cached:
- `sedai_group_priority` Read → `GetGroupById` → `GET /settingsV2/topology/configs/group` →
  `getGroupLevelDefaultResourceSetting` → `resourceSettingsService.getSedaiGroupSetting(groupId)`,
  which is **`@Cacheable(SEDAI_GROUP_RESOURCE_SETTING_CACHE, key="#groupId")`** and reads the
  **base** `sedai_settings_group_attributes` table via `findById` (NOT a materialized view).

The priority write path does **not** evict that cache:
- `POST /api/sedaigroup/updatePriority` → `SedaiGroupApi.updateGroupsPriority` →
  `SedaiGroupServiceV2.updateGroupsPriority` (1400–1489). It `saveAll`s the new priorities and
  calls `refreshPriorityView()` + `updateSettingsRefreshed(false)`, but carries **no
  `@CacheEvict`**.
- Its sibling write methods all DO evict `SEDAI_GROUP_RESOURCE_SETTING_CACHE`:
  `createOrUpdateGroupSetting` (key-evict), `enableorDisableGroup` (allEntries),
  `setPriorityToGroups` (allEntries). `updateGroupsPriority` is the lone exception.

**Effect:** after a `terraform apply` writes new priorities, a subsequent `sedai_group_priority`
Read returns the **cached pre-update priority** → Terraform sees the old value → "priority reverts"
drift on the next plan.

**TTL is verified, and it is long.** `cacheConfig.yaml:110–112` sets
`sedaiGroupResourceSettingCache: cacheExpiryMins: 720` → **12 hours** (applied as Hazelcast
`timeToLiveSeconds` + `maxIdleSeconds`, and Redis TTL). The cache also has a near-cache with
`invalidateOnChange(true)`, but that only fires on *map* changes — and `updateGroupsPriority` writes
straight to the DB repo, never touching the cache — so no invalidation event is emitted. The stale
entry persists for the full **12h** window (or until an unrelated `enableorDisableGroup` /
`setPriorityToGroups` / `createOrUpdateGroupSetting` runs — those evict allEntries/key — or a node
restart). The drift is therefore **deterministic and long-lived**: a normal `plan → apply → plan`
loop populates the cache on the first plan, the apply changes priority without evicting, and the
next plan reads stale for up to 12 hours.

> Note: the materialized view `mv_sedai_settings_group_attributes` is NOT the culprit — the read
> uses the base table, and `updateGroupsPriority` refreshes the MV synchronously anyway. The
> earlier MV-staleness hypothesis was checked and refuted.

### Fix
- **Backend (preferred):** add `@CacheEvict(SEDAI_GROUP_RESOURCE_SETTING_CACHE)` to
  `updateGroupsPriority`, matching its sibling write methods.
- **Provider-side mitigation if the backend fix lags:** in `sedai_group_priority`, treat the write
  response (`GroupPriorityUpdateStatus.requestedPriority`, which reflects the accepted value) as
  authoritative for state rather than relying solely on an immediate cached Read.

## Verified finding 2 — settings application to new members is asynchronous

When a group's membership changes, `updateSedaiGroupResources` sets
`sedai_settings_group_attributes.settings_refreshed = false` (734–736); the
`UpdateGroupResourcesMinion` later calls `refreshAndPopulateSettings()` when membership changed or
settings are stale (82–86). So the **stored** group setting (what Terraform reads back via
`GET configs/group`) is updated synchronously, but its **propagation to individual resources'
effective settings** happens on the next minion cycle (and only if the minion is enabled).

This is not a Terraform drift issue (TF reads the stored value), but it is a user expectation issue:
the *effect* of an apply lands shortly after, not instantly.

## Defensive actions

- [ ] Backend: add the missing `@CacheEvict(SEDAI_GROUP_RESOURCE_SETTING_CACHE, allEntries=true)`
      to `updateGroupsPriority`, matching its sibling write methods. (The 12h TTL rules out
      "harmless by expiry" — this must be fixed, not waited out.)
- [ ] `sedai_group_priority`: source post-write state from the update response, not an immediately
      cached Read, so an apply is stable even if the cache is warm.
- [ ] Docs: add a "how groups settle" note to `group.md` / `group_settings.md` /
      `group_priority.md` — an apply records desired state; membership and per-resource settings
      converge shortly after via Sedai's background workers.
- [ ] Confirm no resource introduces a minion-dependent readiness poll.

## Acceptance criteria

- [ ] `sedai_group_priority` is stable across repeated plans immediately after an apply (no
      cache-lag diff), verified on a live tenant.
- [ ] Docs set the eventual-consistency expectation for all three group resources.
- [ ] No provider code blocks waiting on membership/settings materialization.

## Effort
~Half a day (backend 1-line evict + provider read-back hardening + live verification).
