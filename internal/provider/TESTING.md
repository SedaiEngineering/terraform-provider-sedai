# Acceptance Test Guide

## Prerequisites

- Go 1.24+
- A dedicated Sedai test tenant (do not use production)
- API credentials for that tenant

---

## 1. Backend setup

You need a running Sedai instance and an API token. Point the tests at it with two environment variables:

```bash
export SEDAI_BASE_URL=https://your-sedai-host    # no trailing slash
export SEDAI_API_TOKEN=your-api-token
```

The test tenant should be empty or disposable — acceptance tests create and destroy real resources. Tests clean up after themselves, but interrupted runs may leave orphans.

---

## 2. Cloud credentials (optional, per cloud)

Multi-cloud account tests skip automatically when credentials are absent. Set only what you have:

```bash
# Azure
export AZURE_TENANT_ID=...
export AZURE_SUBSCRIPTION_ID=...
export AZURE_CLIENT_ID=...
export AZURE_CLIENT_SECRET=...

# GCP
export GCP_PROJECT_ID=...
export GCP_SERVICE_ACCOUNT_JSON='{ "type": "service_account", ... }'

# Kubernetes (agentless, AWS cluster)
export K8S_CLUSTER_URL=https://...
export K8S_CA_CERT=...          # base64-encoded PEM
export K8S_ROLE=arn:aws:iam::123456789012:role/SedaiRole
export K8S_EXTERNAL_ID=...

# Kubernetes (agent-based)
export K8S_AGENT_CLUSTER_URL=https://...

# resource_settings tests
export SEDAI_TEST_RESOURCE_ID=an-existing-resource-id-in-the-test-tenant
```

---

## 3. Running tests

### Unit tests (no backend needed)
```bash
make test-unit
```

### Acceptance tests (requires SEDAI_BASE_URL + SEDAI_API_TOKEN)
```bash
TF_ACC=1 make test-acc
```

### Run a single test by manifest ID
```bash
make test-id ID=GSET-006
make test-id ID=ACCT-050
```

### System tests (large stacks — opt-in)
```bash
TF_ACC=1 TF_SYSTEM_TESTS=1 make test-system
```

### Generate an HTML report
```bash
TF_ACC=1 make test-acc-json      # writes test-results.json
make test-report                  # writes test-report.html
```

---

## 4. Test categories

| Prefix | What it covers |
|--------|---------------|
| ACCT | AWS, Azure, GCP, Kubernetes account lifecycle |
| GRP | Group create/update/delete/import |
| GSET | Group settings, all sub-blocks, drift |
| ASET | Account settings, drift |
| RSET | Resource settings |
| PRI | Group priority ordering |
| MP / MP-DIM / MP-VM | Monitoring providers and dimension arrays |
| DS | Data sources |
| DRIFT | Cross-resource drift scenarios |
| DEP | Dependency ordering |
| SCALE | 12-account and 48-resource stacks |
| DESTROY | Full stack and partial destroy ordering |
| RECOVERY | Interrupted apply re-apply (Diligent scenario) |
| ERR | EOF recovery, HTTP errors (mock server) |
| IMPORT | Import round-trips for all resource types |
| MIGR | v1 → v2 state migration via import |
| SEC | Sensitive fields not present in state |

---

## 5. Tips

- Tests run in parallel by default (`-parallel=4`). Lower it if the backend rate-limits: edit the `test-acc` target in `GNUmakefile`.
- A failed mid-run will leave resources in the tenant. Re-running the same test will attempt to recreate them — the provider handles duplicate-name detection and reuses existing resources where possible.
- Mock-server tests (`ERR-*`, `RECOVERY-*`, `DESTROY-004/006`) run against a local httptest server, not your backend, so they always work regardless of backend state.
