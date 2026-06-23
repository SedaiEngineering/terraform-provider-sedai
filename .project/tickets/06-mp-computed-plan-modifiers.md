# 06 — `UseStateForUnknown` on computed monitoring-provider attrs; make `id` Computed-only

**Severity:** Critical
**Component:** all 7 `internal/provider/create_*_monitoring_provider.go`
**Breaking change:** No

## Problem

Every monitoring-provider resource exposes `id`, `name`, `integration_type`, and
all `*_dimensions` lists as **`Optional + Computed` with no plan modifier**.
Contrast `sedai_account.id` and `sedai_group.id`, which correctly have
`stringplanmodifier.UseStateForUnknown()`.

Two issues:
1. Missing `UseStateForUnknown()` on persistent computed attributes.
2. `id` is marked `Optional` — it is backend-assigned and should be
   `Computed`-only.

## Why it's wrong

The `terraform-provider-aws` convention is emphatic: **every** `Computed`
attribute that should persist gets `UseStateForUnknown()`, precisely to avoid
spurious `(known after apply)` churn. And per aws, "ID is never Optional" —
marking `id` Optional lets users write `id = "..."`, which is incorrect.

Effect today: every monitoring-provider `plan` shows avoidable churn on
`name`/`integration_type`/dimensions even when nothing changed.

## Desired behavior

For each of the 7 monitoring-provider resources:
- Add `UseStateForUnknown()` (string/list plan modifier as appropriate) to every
  persistent computed attribute: `id`, `name`, `integration_type`, and any
  `*_dimensions` that remain `Computed` (see ticket 03 for the dimensions
  contract decision).
- Change `id` from `Optional + Computed` to `Computed` only.

Apply the same `id` fix to any other resource where `id` is `Optional + Computed`
(e.g. `sedai_account` — verify).

## Acceptance criteria

- [ ] Repeated `terraform plan` on an unchanged monitoring provider shows no diff.
- [ ] `id` is `Computed`-only on all monitoring providers (and audited elsewhere).
- [ ] Writing `id = "x"` in config produces a schema error.
- [ ] Docs regenerated.

## Effort
~Half a day across 7 files (mechanical, repetitive).
