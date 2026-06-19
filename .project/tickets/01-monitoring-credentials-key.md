# 01 — Monitoring-provider credentials sent under the wrong JSON key (silent credential loss)

**Severity:** Critical
**Component:** `sedai-sdk-go` `sdk/sedai/monitoringProvider/monitoringProvider.go` (GKE, Datadog, NewRelic, Federated Prometheus, Victoria Metrics)
**Breaking change:** No (fixes a silent failure)
**Reverses:** `/tmp/tf.md` A7 — see "Backend verification" below. **Do NOT apply A7 as written.**

## Problem

Five monitoring providers send credentials under the JSON key
`details.credentialsDetails` (**plural**). The backend model field is
`credentialsDetail` (**singular**). Because the backend deserializes with
`FAIL_ON_UNKNOWN_PROPERTIES = false`, the plural key is **silently ignored** —
the provider is created/updated with **null credentials and no error returned**.

Confirmed Go SDK call sites (grep):
- CloudWatch — `monitoringProvider.go:68` → `details["credentialsDetail"]` **(singular — CORRECT)**
- GKE — `:183` → `"credentialsDetails"` ❌
- Datadog — `:334` → `"credentialsDetails"` ❌
- NewRelic — `:442` → `"credentialsDetails"` ❌
- Federated Prometheus — `:566` → `"credentialsDetails"` ❌
- Victoria Metrics — `:810` → `"credentialsDetails"` ❌

## Backend verification (why A7 is backwards)

- `org.sedai.edison.models.monitoring.MonitoringProviderDetails` (`sedai-models`,
  line 53): `private CredentialsDetail credentialsDetail;` — JSON property
  `credentialsDetail` (singular). Polymorphic via
  `@JsonTypeInfo(... property = "monitoringProvider")`.
- `getCredentials()` (line 142): `if (credentialsDetail != null) return credentialsDetail.getCredentials();`
- `DefaultMonitoringProviderService.createOrUpdateSedaiMonitoringProviderWithValidation`
  reads credentials via `model.getDetails().getCredentials()` → so a null
  `credentialsDetail` means null credentials.
- `FAIL_ON_UNKNOWN_PROPERTIES` is disabled across sedai-core (house style;
  e.g. `SedaiTokenGeneratorService`, `DefaultShadowResourceSavingsService`,
  `StreamingAfoParser`), so the unknown plural property is dropped, not rejected.

`tf.md` A7 recommended pluralizing CloudWatch to match the others — that would
have **broken the only correct provider**. The backend model is the source of
truth here (Python-SDK parity not required for the conclusion).

## Impact / UX

`terraform apply` reports success; the monitoring provider exists in the Sedai UI;
metric discovery then silently collects nothing because credentials were never
stored. This is the worst possible experience — a green apply that doesn't work,
with no signal to the user.

## Fix

1. Change `"credentialsDetails"` → `"credentialsDetail"` in the GKE, Datadog,
   NewRelic, Federated Prometheus, and Victoria Metrics builders. **Leave
   CloudWatch (`:68`) unchanged.**
2. **Defensive TF layer:** credentials are write-only and not returned by Read,
   so the provider can't diff them. After Create, surface any backend-side
   credential/connection failure if the API exposes one (e.g. provider status or
   a discovery/test-connection result), and add a unit test asserting the
   serialized payload contains `credentialsDetail` (singular) for every provider.
3. Add an SDK regression test that round-trips each provider's request through a
   stub asserting the `details.credentialsDetail.credentials` path is populated.

## Acceptance criteria

- [ ] All monitoring providers serialize credentials under
      `details.credentialsDetail`.
- [ ] CloudWatch unchanged and still working.
- [ ] A created GKE/Datadog/NewRelic/FP/VM provider has non-null credentials
      server-side (verified against a test tenant).
- [ ] SDK test asserts the key name for every provider, preventing regression.

## Effort
~Half a day (1-line per provider + tests + one live verification).
