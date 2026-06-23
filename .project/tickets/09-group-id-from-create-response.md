# 09 — Bind the group ID from the create response instead of search-by-name (TOCTOU + redundant round-trips)

**Severity:** Medium-High
**Component:** SDK `groups.CreateGroup`
**Breaking change:** No
**Related:** `/tmp/tf.md` B5
**Backend refs:** `SedaiGroupApi.createSedaiGroup` (549–579) → `createGroup` returns the new `groupId`; `GroupNameAlreadyExistsException` → `ApiStatus.VALIDATION_ERROR`

## Problem

The backend create endpoint **returns the new group ID** (`BaseResponse<String>`
with `result = groupId`) and already enforces name uniqueness atomically inside a
`@Transactional` method (throws `GroupNameAlreadyExistsException`). The SDK
ignores both:

1. It runs a **client-side `SearchGroupsByName` pre-check** before create — a
   TOCTOU race: two concurrent creates of the same name can both pass, and search
   errors are swallowed.
2. After create it **discards the returned ID** and runs `SearchGroupsByName`
   *again* to find the new group — a second full `GetAllGroups` pull, and
   non-deterministic if duplicate names exist (`created[0]`).

## Impact / UX

- Under `terraform apply -parallelism>1`, a resource can bind to the wrong
  group's ID, or a create can spuriously "succeed but not be found."
- Two extra `GetAllGroups` round-trips per create (the list endpoint is already a
  full-table pull) — adds load and EOF surface (ties into `/tmp/tf.md` C3 /
  SED-21371).

## Backend confirmation (this review)

Verified against `sedai-core`:
- `SedaiGroupApi.createSedaiGroup` (549–579) returns `BaseResponse<String>` and sets
  `response.setResult(groupServiceV2.createGroup(groupName, definition))` (line 564).
- `SedaiGroupServiceV2.createGroup` (1701) is declared `public String createGroup(...)` —
  it returns the new **groupId** string. So `resp["result"]` **is** the group ID.
- Duplicate name → `GroupNameAlreadyExistsException` → `ApiStatus.VALIDATION_ERROR` (line 570).

> ⚠️ **The SDK's own comment is wrong and must be deleted.** `groups.go:276` currently says
> *"The create endpoint returns only `{status: "OK"}` — query to find the new ID."* That is
> false per the backend (above). Whoever implements this will otherwise trust the comment and
> conclude the fix is impossible. The `result` field IS present **as long as the `Accept:
> application/json` header is sent** — which the SDK already does (`groups.go:264`); keep it.

## Defensive fix

- Read the group ID directly from the create response (`resp["result"]`).
- **Delete the stale `groups.go:276` comment** claiming the response carries no ID.
- Drop the post-create `SearchGroupsByName` (`groups.go:277`).
- Drop (or keep only as a fast-fail nicety) the client-side name pre-check
  (`groups.go:242`); rely on the backend's uniqueness check and map `VALIDATION_ERROR`
  to a clear "group already exists" error.

## Acceptance criteria

- [ ] `CreateGroup` returns the ID from the create response with no follow-up
      search.
- [ ] Concurrent creates of distinct names never cross-bind IDs.
- [ ] A duplicate-name create returns a clear, deterministic error sourced from
      the backend.
- [ ] Net round-trips per create reduced by two.

## Effort
~Half a day incl. a concurrency test.
