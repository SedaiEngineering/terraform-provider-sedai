# Interface-design tickets

Issues found in an interface-design review of the Sedai Terraform provider,
benchmarked against `terraform-provider-aws` (same plugin framework) and
`terraform-provider-github`. Scope: the **user-facing Terraform contract**
(schema, plan/apply behavior, drift) — not the credential/transport bugs
already tracked elsewhere (see `/tmp/tf.md` A1–A7, SED-21371).

> **File numbers are the work order: `01` is the most critical, `15` gates the rest.**
> Tickets cross-reference each other by these numbers, so a number is also a stable ID —
> if priorities change again, re-rank in the table here rather than renumbering files
> (renumbering means rewriting every cross-reference, as this pass did).

## How "criticality" is ranked here

Ordered by operational blast radius, worst first:

1. **Silent failures** — `apply` reports success but the result is broken/empty, with
   **no signal to the user**. The worst possible UX: a green apply that lies.
2. **Unconvergeable drift** — `plan` never reaches zero diff; the field can never settle.
3. **Visible apply failures / missing drift detection** — loud failures, or out-of-band
   changes Terraform can't see.
4. **Consistency / behavioral honesty** — surprising-but-survivable schema inconsistencies.

Effort and the owning repo are tracked separately (a quick win can still be lower
criticality). **Read the cross-repo note before planning a release.**

## ⚠️ Cross-repo reality — a provider-only release does NOT ship this backlog

The critical path runs through **three repos**. Plan releases accordingly:

| Repo | Tickets that need changes here |
|---|---|
| `terraform-provider-sedai` (this repo) | 02, 03 (guard), 04, 05, 06, 07 (mitigation), 08*, 11, 12, 13*, 14, 15 |
| `sedai-sdk-go` | 01, 02 (helper), 03 (AZ key), 09, 13, **08 (model expansion — prerequisite)** |
| `sedai-core` (backend) | 01 (status surfacing, optional), 05 (NPE hardening), 07 (`@CacheEvict`) |

`*` = primary work is gated on another repo (see the ticket).

## Master backlog — ordered by criticality

| Pri | ID | Ticket | Severity | Repo(s) | Breaking? | Blocked on |
|----|----|--------|----------|---------|-----------|-----------|
| 1 | 01 | [MP credentials sent under wrong JSON key (silent loss)](01-monitoring-credentials-key.md) | Critical | sdk | No | — |
| 2 | 02 | [Created group has no members until a minion runs](02-group-create-no-membership.md) | High | provider + sdk | No | — |
| 3 | 03 | [`*_dimensions` contract: inverted guard + `aZDimensions` write key](03-dimensions-contract.md) | High | provider + sdk | Possibly | — |
| 4 | 04 | [Remove volatile counts from the `sedai_group` resource](04-group-counts-in-state.md) | Critical | provider | No | — |
| 5 | 05 | [`autonomous_action_without_traffic` server-derived from `is_prod` (kube/ecs)](05-settings-server-derived-fields.md) | High | provider | Possibly | — |
| 6 | 06 | [`UseStateForUnknown` on computed MP attrs; `id` Computed-only](06-mp-computed-plan-modifiers.md) | Critical | provider | No | — |
| 7 | 07 | [Priority read-after-write 12h cache drift; async settling](07-async-settling-expectations.md) | High | provider + core | No | core `@CacheEvict` for full fix |
| 8 | 08 | [`sedai_account` Read refreshes only 4/24 attributes](08-account-read-drift.md) | Critical | provider + sdk | No | **SDK `Account` model expansion** |
| 9 | 09 | [Bind group ID from create response (drop TOCTOU search-by-name)](09-group-id-from-create-response.md) | Medium-High | sdk | No | — |
| 10 | 10 | [Validate core enums at plan time](10-account-enum-validation.md) | Critical | provider | No | — |
| 11 | 11 | [Declare cross-field constraints on account / CloudWatch / FP](11-cross-field-validation.md) | High | provider | No | 10 (managed-services union validator) |
| 12 | 12 | [`RequiresReplace` on immutable identity fields](12-requires-replace.md) | High | provider | Behavioral | product confirm immutability |
| 13 | 13 | [Top-level settings mode fanned out to AUTO-incompatible sections](13-settings-mode-fanout.md) | Medium | sdk | No | live backend check (AUTO in app section) |
| 14 | 14 | [Make settings-mode requiredness consistent](14-settings-mode-consistency.md) | Medium | provider | Possibly | product decision (A vs B) |
| — | 15 | [Testing strategy & customer-issue reproductions](15-testing-strategy.md) | High (**gates all**) | all + CI | No | — |

Why the order departs from the raw severity labels: the labels are kept per ticket,
but a *silent* failure (01, 02, 03) outranks a *visible* one even when both are
"Critical/High". 04 (volatile counts) and 06 (plan modifiers) stay near the top because
they erase ~80% of day-2 `plan` noise (the customer's "scary plan") and are a few hours
each. 08 (account Read) is labeled Critical but sits at Pri 8 because it is **blocked on
an SDK model change** and is a missing feature (drift detection) rather than active
corruption.

## ✅ Review corrections (this pass — verified against source)

Findings from auditing each ticket's assumptions against `internal/provider`,
`sedai-sdk-go`, and `sedai-core`/`sedai-models`. The tickets have been updated.

- **09 — the misleading clue is in the SDK, not the ticket.** Backend confirmed:
  `createSedaiGroup` returns `BaseResponse<String>` with `result = groupId`
  (`SedaiGroupApi.java:564`, `createGroup` returns `String` at `:1701`). But the SDK's
  own comment (`groups.go:276`) falsely says *"returns only `{status:"OK"}`"* — it must
  be **deleted** as part of the fix, or the implementer will conclude 09 is impossible.
  `result` is present because the SDK already sends `Accept: application/json`.
- **08 — hard SDK blocker, not "likely".** The Go SDK `AccountDetails` deserializes only
  `cloudProvider` + `integrationType`; `TransformJSON` drops every undeclared field. 08
  **cannot** be done in the provider until the SDK `Account` model is expanded (and the
  backend GET confirmed to return the fields). Re-sequenced as a prerequisite.
- **03 — fix direction resolved + cross-repo.** Canonical backend field is **`azDimensions`**
  (no `aZDimensions` exists in `sedai-models`). The **read path is correct**; the bug is the
  **write key in the SDK** (`monitoringProvider.go:349/579/823` send `aZDimensions`). The
  inverted-guard bug is separate and lives in the provider; **Azure's write path is already
  correct** and is excluded from that fix.
- **05 — scope tightened.** The backend method overrides `autonomousActionWithoutTraffic`
  for **ECS and Kube only**; `container` is not touched and must keep the field. Listing
  `container_app_settings.go` was an overreach (would cause an unnecessary breaking change).
- **02 — AC made conditional.** Membership materializes only from **already-discovered**
  inventory; "members immediately after create" is true only when inventory was discovered
  first. AC and the regression test now reflect that.

### From the backend trace (account-settings + cloudwatch flows, verified against `sedai-core`)

- Backend credentials field is `MonitoringProviderDetails.credentialsDetail` (singular);
  `FAIL_ON_UNKNOWN_PROPERTIES=false` ⇒ the plural key is silently dropped. CloudWatch
  (`monitoringProvider.go:68`, singular) is correct; GKE/Datadog/NewRelic/FP/VM are not (ticket 01).
- `GET /monitoringProviders/{id}` wraps a single result in a 0/1-element list — the SDK's
  `result[0]` is **correct** (not a bug).
- Settings read-modify-write round-trips correctly: GET `attributes.setting` ≡ POST
  `CompositeResourceSettingDetail` field shape.
- `ResourceConfigMode = {AUTO, MANUAL, OFF, DATA_PILOT, CO_PILOT}` — the TF settings-mode
  validator (DATA_PILOT/CO_PILOT/AUTO) is a correct subset.

### How the group pipeline actually works (verified in `sedai-core`)

- Membership lives in the `group_resources` table, materialized by `updateSedaiGroupResources`
  (checksum-gated). **Group create does NOT materialize it**; `createGroup` only saves the
  definition + inits settings (active=false).
- `UpdateGroupResourcesMinion` (Quartz cluster-singleton) refreshes membership **only for
  `auto_refresh=true`** groups, then re-applies settings when membership changed or
  `settings_refreshed=false`.
- Counts come from `group_resources` (stale between minion runs) — the mechanistic proof for ticket 04.
- `active`/`priority` live in `sedai_settings_group_attributes`, separate from the group
  definition. The priority **read** (`getSedaiGroupSetting`) uses the base table and is
  `@Cacheable`; the priority **write** (`updateGroupsPriority`) refreshes the materialized
  view but — unlike its sibling write methods — does **not** evict that read cache →
  read-after-write staleness (ticket 07). The cache TTL is **verified at 720 min / 12h**
  (`cacheConfig.yaml:111`), so the drift is deterministic and long-lived. The earlier "MV
  staleness" hypothesis was checked and **refuted**.
- The backend `createGroup` **returns the new groupId** and enforces name uniqueness
  atomically — the SDK ignores both and its inline comment denies the return value (ticket 09).

## Testing — applies to every ticket

| # | Ticket | Severity |
|---|--------|----------|
| 15 | [Testing strategy & customer-issue reproductions](15-testing-strategy.md) | High (gates all) |

**Definition of Done for every ticket (01–14):** no fix merges without tests —
unit (schema/validators/conversions), SDK contract tests asserting the exact JSON
shape against the backend model field names, and an acceptance test covering the
CRUD/import lifecycle **including the plan-idempotency assertion** (apply → next
plan is empty). Every customer-reported issue (tf.md C1/C2/C3, tf2.md parallel-apply
incident, secrets-in-plaintext) must have a named regression test that fails before
the fix and passes after. See ticket 15 for the full test matrix and CI wiring.

## Explicitly NOT doing (rejected recommendations)

- **Renaming monitoring providers** to `sedai_monitoring_provider_<tech>` — current
  tech-first naming is idiomatic (cf. Vault `vault_aws_secret_backend`; AWS
  `aws_cloudwatch_log_group`). Breaking change for cosmetic gain.
- **Lowercasing enum values** + translation layer — `terraform-provider-aws`
  passes raw API enum casing through (validated). Casing-match is acceptable;
  the real fix is plan-time validation (ticket 10), keeping current casing.
- **`tf.md` A7 as written** (pluralize CloudWatch's `credentialsDetail`) — the
  backend trace proves the singular key is correct. Applying A7 would break the
  one working provider. See ticket 01 for the correct (reversed) fix.
