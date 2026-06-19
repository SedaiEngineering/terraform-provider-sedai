# 14 — Make settings-mode requiredness consistent across the three settings resources

**Severity:** Medium
**Component:** `resource_settings.go`, `group_settings.go`, `account_settings.go`
**Breaking change:** Possibly (if `resource_settings` modes become Required)

## Problem

The three settings resources disagree on whether the core modes are required:

- `sedai_account_settings`: `availability_mode`, `optimization_mode` are
  **Required**.
- `sedai_group_settings`: **Required**.
- `sedai_resource_settings`: `availability_mode`, `optimization_mode`,
  `sedai_sync_enabled` are all **Optional**.

So a `sedai_resource_settings` block with everything null is a legal but no-op
resource, and the contract differs from its siblings for no documented reason.

## Why it's wrong

Consistency across sibling resources is a usability standard — users expect the
same attribute to behave the same way. The current split is surprising and makes
the partial-spec intent ambiguous (is omission "leave backend untouched" or
"unset"?).

## Desired behavior

Pick one contract and apply it uniformly, with the choice documented:

- **Option A (consistency):** make `resource_settings` modes `Required` like the
  others. Simplest mental model; breaking for anyone relying on omission.
- **Option B (partial-spec):** make all three settings resources' modes
  `Optional`, document omission as "leave the backend's current value untouched",
  and ensure Read/Update honor that field-level partial spec everywhere.

Recommend **Option B** — it matches the per-resource-type block design the
settings layer already uses (nil pointer = unmanaged) and is the more powerful
contract. Confirm with product.

## Acceptance criteria

- [ ] All three settings resources treat `availability_mode` /
      `optimization_mode` / `sedai_sync_enabled` identically.
- [ ] Behavior of an omitted mode is documented in each resource's docs.
- [ ] Tests cover the omitted-field path for the chosen contract.

## Effort
~Half a day + product decision on A vs B.
