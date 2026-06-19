# 13 — Top-level settings mode is fanned out to every section, including AUTO-incompatible ones

**Severity:** Medium
**Component:** SDK `account/settings.go`, `groups/settings.go` (UpdateAccountSettings / UpdateGroupSettings); `app_settings.go`
**Breaking change:** No
**Backend ref:** availability/optimization `configMode` enum = `ResourceConfigMode {AUTO, MANUAL, OFF, DATA_PILOT, CO_PILOT}` (`sedai-models` `ResourceConfigMode.java`)

## Problem

`UpdateAccountSettings` / `UpdateGroupSettings` apply the single top-level
`availability_mode` / `optimization_mode` uniformly to **every** resource-type
section in the settings blob:

```go
for _, sectionRaw := range s.Raw {
    config.MutatePathIfExists(section, s.AvailabilityMode, "availability", "configMode")
    config.MutatePathIfExists(section, s.OptimizationMode, "optimization", "setting", "configMode")
    ...
}
```

The top-level `availability_mode` is validated with `settingsConfigModeValidator`
(allows `AUTO`). But the `app_settings` block has a dedicated
`noAutoConfigModeValidator` precisely because the app type cannot be `AUTO`. So a
settings-wide `availability_mode = "AUTO"` gets written into the app section too,
contradicting the per-block rule.

## Why it matters (UX)

The provider enforces "app can't be AUTO" at the block level but then bypasses it
via the top-level fan-out. Depending on backend behavior, this either silently
coerces (→ drift) or is accepted into an invalid state. Either way the user gets
inconsistent, surprising results.

## Verification needed

Confirm backend behavior when `AUTO` is written to the app section's
`availability.configMode`: rejected, coerced, or stored. (`ResourceConfigMode`
*accepts* `AUTO` as a value, so it likely won't 400 — the constraint is
semantic, applied per resource type.) This determines whether the fix is
"don't fan out to AUTO-incompatible sections" or "validate at plan time."

## Defensive fix

- Scope the top-level mode fan-out so AUTO-incompatible sections (app, and any
  others product identifies) are skipped or receive a valid fallback, **or**
- add a plan-time validator: if `availability_mode = "AUTO"`, reject configs that
  also manage an app-type section, with a clear message.
- Keep `DATA_PILOT` / `CO_PILOT` / `AUTO` as the TF-exposed subset (verified
  valid); do not expose `OFF` / `MANUAL`.

## Acceptance criteria

- [ ] Backend behavior for app + `AUTO` is documented (from a live check).
- [ ] A settings-wide `AUTO` either skips AUTO-incompatible sections or fails at
      plan with a clear message — never silently drifts.
- [ ] Test covers the account/group settings fan-out for an app section.

## Effort
~Half a day incl. one live backend check.
