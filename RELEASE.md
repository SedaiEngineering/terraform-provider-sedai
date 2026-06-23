# Release Process

## Overview

Two repos ship together. Always release the SDK first, then update the provider to reference it, then release the provider.

```
sedai-sdk-go  →  terraform-provider-sedai
```

---

## 1. Release the SDK

1. Merge changes into the release branch on `SedaiEngineering/sedai-sdk-go`
2. On GitHub: **Releases → Draft a new release**
3. Tag: `v1.x.x` — minor bump for new features, patch for fixes
4. Check **"Set as a pre-release"** if beta; uncheck **"Set as the latest release"**
5. Publish

---

## 2. Update the SDK reference in the provider

Make sure the local `replace` directive in `go.mod` is commented out:
```
// replace github.com/SedaiEngineering/sedai-sdk-go => ../sedai-sdk-go
```

Then update the dependency:
```bash
GIT_SSH_COMMAND="ssh -i ~/.ssh/github" \
GONOSUMDB=github.com/SedaiEngineering/* \
GOPRIVATE=github.com/SedaiEngineering/* \
go get github.com/SedaiEngineering/sedai-sdk-go@v1.x.x && go mod tidy
```

Commit:
```bash
git add go.mod go.sum
git commit -m "chore: bump sedai-sdk-go to v1.x.x"
```
> [!NOTE] 
> Use a go.work file for local dependencies instead of a `replace` directive
---

## 3. Release the provider

1. Merge changes into the release branch on `SedaiEngineering/terraform-provider-sedai`
2. On GitHub: **Releases → Draft a new release**
3. Tag: `v2.0.0-beta.1` (beta) or `v2.0.0` (stable)
4. For beta: check **"Set as a pre-release"**, uncheck **"Set as the latest release"**
5. Publish — goreleaser runs automatically and publishes to the Terraform Registry

The registry picks up the new version within ~10 minutes. Pre-release versions are published but not served as latest — users must explicitly pin the version:
```hcl
required_providers {
  sedai = {
    source  = "SedaiEngineering/sedai"
    version = "2.0.0-beta.1"
  }
}
```

---

## Versioning reference

| Release type | SDK | Provider |
|---|---|---|
| Beta | `v1.3.0` | `v2.0.0-beta.1` |
| Stable GA | `v1.3.0` | `v2.0.0` |
| Patch | `v1.3.1` | `v2.0.1` |
