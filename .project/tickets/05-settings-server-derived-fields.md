# 05 — `autonomous_action_without_traffic` is server-derived from `is_prod` (kube/ecs) → guaranteed drift

**Severity:** High
**Component:** `kube_app_settings.go`, `ecs_app_settings.go`; SDK settings apply path
**Breaking change:** Possibly (removing/computing an attribute)
**Backend ref:** `TopologySettingsV2.updateSettingAutoActionWithoutTrafficWRToIsProd` (`sedai-core`, lines 120–128)

> **Scope verified (this review):** the backend method overrides `autonomousActionWithoutTraffic`
> for **ECS and Kube only** (lines 121–127). It does **not** touch `container` or any other
> resource type. So `container_app_settings` is *not* server-overridden and must **keep** the
> field — do not include it in this fix or you introduce an unnecessary breaking change.
> Confirmed: there is no third branch in the method.

## Problem

On every account/group settings POST, the backend runs (before persisting):

```java
// updateSettingAutoActionWithoutTrafficWRToIsProd(CompositeResourceSettingDetail)
ecs.getAutonomousActionWithoutTraffic().setEnabled(!ecsIsProd);
kube.getAutonomousActionWithoutTraffic().setEnabled(!kubeIsProd);
```

For **ECS and Kubernetes**, `autonomousActionWithoutTraffic` is forced to
`!isProd`, **ignoring whatever the client sent**. Our TF settings blocks expose
both `is_prod` and `autonomous_action_without_traffic` as independent,
user-settable fields.

## Impact / UX

A user who writes `is_prod = true` + `autonomous_action_without_traffic = true`
gets `false` stored on the backend. The next `terraform plan` shows a perpetual
diff (`true -> false` proposed, then re-proposed forever) — the field can never
converge. The provider is letting the user set something the server silently
overrides: a confusing, never-clean plan.

## Defensive fix (pick one; A preferred)

- **A (recommended):** Remove `autonomous_action_without_traffic` from the
  `kube_app_settings` / `ecs_app_settings` blocks **only** (container is not
  affected — see scope note above). It is not independently controllable for
  these types. Document that it is derived as `!is_prod`. Cleanest UX — the schema
  stops offering an illusion of control.
- **B:** Keep the field but add a plan-time cross-field validator requiring
  `autonomous_action_without_traffic == !is_prod` for kube/ecs, with a clear
  message. Prevents the drift but adds friction.
- **C:** Mark it `Computed` (server-owned) and refresh it from the backend on
  Read so state always matches. Removes drift but the user can't set it.

No non-kube/ecs type is overridden by the backend method (verified), so all other
types keep the field as-is.

## Related defensive note (NPE hardening)

The same backend method dereferences `settings.getEcsSettings().getIsProd()` and
`...getAutonomousActionWithoutTraffic()` **before** the payload null-check, so a
partial `CompositeResourceSettingDetail` POST → `NullPointerException` → HTTP 500.
The SDK's read-modify-write currently always sends fully-populated sections, which
avoids this — keep it that way and add a test asserting the kube/ecs sections
(and their `isProd`/`autonomousActionWithoutTraffic` slots) are always present in
the POST blob.

## Acceptance criteria

- [ ] `is_prod = true` with `autonomous_action_without_traffic = true` no longer
      produces a perpetual diff for kube/ecs.
- [ ] The relationship is documented in `docs/resources/{group,account}_settings.md`.
- [ ] Test asserts the settings POST blob never omits kube/ecs sections.

## Effort
~Half a day + product confirmation on A vs B vs C.
