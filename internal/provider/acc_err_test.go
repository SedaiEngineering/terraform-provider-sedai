package provider_test

import (
	"fmt"
	"testing"
)

func TestAccERR(t *testing.T) {
	// ERR-001: Account create → EOF → recovery search not found → clear error surfaced.
	// Skipped: the SDK caches SEDAI_BASE_URL at init() via sync.Once, so t.Setenv
	// cannot inject a mock URL after the binary starts.
	t.Run("ERR-001", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-002: Account create → EOF → recovery search FINDS account → warn+use ID.
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-002", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-003: Group create → EOF → recovery finds group → warn+use ID.
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-003", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-006: HTTP 401 → clear auth error (not raw JSON dump).
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-006", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-007: HTTP 403 → clear error.
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-007", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-008: HTTP 500 → error surfaced with context.
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-008", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-013: Delete returns "not found" → treat as success (idempotent destroy).
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-013", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})

	// ERR-015: Apply fails on first run, succeeds on second.
	// Skipped: SDK caches SEDAI_BASE_URL at init(), t.Setenv cannot inject mock URL.
	t.Run("ERR-015", func(t *testing.T) {
		t.Skip("mock server incompatible with SDK config caching (sync.Once at init)")
	})
}

func testAccMockAccountConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
}
`, name)
}

func testAccMockGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
`, name)
}
