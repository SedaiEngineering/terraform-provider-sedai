package provider_test

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-sedai/internal/provider"
)

// testAccProtoV6ProviderFactories wires up the sedai provider for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"sedai": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// regexpMustCompile compiles a regexp pattern and panics on failure.
func regexpMustCompile(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// keep the resource import used so the compile-time check below works
var _ = resource.TestCase{}

// testAccPreCheck aborts the test when the environment variables required for
// acceptance tests are missing.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("SEDAI_BASE_URL") == "" {
		t.Fatal("SEDAI_BASE_URL must be set for acceptance tests")
	}
	if os.Getenv("SEDAI_API_TOKEN") == "" {
		t.Fatal("SEDAI_API_TOKEN must be set for acceptance tests")
	}
}

// testAccSystemPreCheck skips the test unless all of: SEDAI_BASE_URL,
// SEDAI_API_TOKEN, and TF_SYSTEM_TESTS=1 are set. Uses t.Skip (not t.Fatal)
// so it is safe to call directly in a test body before resource.Test.
func testAccSystemPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("SEDAI_BASE_URL") == "" || os.Getenv("SEDAI_API_TOKEN") == "" {
		t.Skip("SEDAI_BASE_URL and SEDAI_API_TOKEN must be set for system tests")
	}
	if os.Getenv("TF_SYSTEM_TESTS") != "1" {
		t.Skip("TF_SYSTEM_TESTS=1 required for system tests")
	}
}

// randString returns a random lowercase alphanumeric string of length n.
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// --- Shared HCL config builders ---

// testAccProviderConfig returns the provider configuration block.
// In acceptance tests the provider reads SEDAI_BASE_URL / SEDAI_API_TOKEN
// from the environment so no explicit values are needed here.
func testAccProviderConfig() string {
	return `
provider "sedai" {}
`
}

// testAccAccountConfig_AWSRole creates a single AWS account using IAM role authentication.
func testAccAccountConfig_AWSRole(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
`, name)
}

// testAccGroupConfig creates a single group with the given name.
func testAccGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
`, name)
}

// testAccGroupWithSettingsConfig creates a group and its settings block.
func testAccGroupWithSettingsConfig(name, avail, optim string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = %[2]q
  optimization_mode = %[3]q
}
`, name, avail, optim)
}

// testAccGroupSettingsConfigNoSync returns group + settings without sedai_sync_enabled.
// Used to verify the P0 false→null drift bug is fixed.
func testAccGroupSettingsConfigNoSync(name, avail, optim string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = %[2]q
  optimization_mode = %[3]q
  # sedai_sync_enabled intentionally omitted — must not produce false->null drift
}
`, name, avail, optim)
}

// testAccGroupSettingsConfigWithSync returns group + settings with explicit sedai_sync_enabled.
func testAccGroupSettingsConfigWithSync(name, avail, optim, syncEnabled string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_settings" "test" {
  group_id           = sedai_group.test.id
  availability_mode  = %[2]q
  optimization_mode  = %[3]q
  sedai_sync_enabled = %[4]s
}
`, name, avail, optim, syncEnabled)
}

// testAccGroupSettingsConfig returns a standalone group_settings block pointing at groupRef.
func testAccGroupSettingsConfig(groupRef, avail, optim string) string {
	return fmt.Sprintf(`
resource "sedai_group_settings" "test" {
  group_id          = %[1]s
  availability_mode = %[2]q
  optimization_mode = %[3]q
}
`, groupRef, avail, optim)
}

// testAccCloudWatchConfig returns account + CloudWatch monitoring provider named "test".
func testAccCloudWatchConfig(accountName string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  integration_type        = "ROLE"
  use_account_credentials = true
}
`, accountName)
}

// testAccFullStackConfig generates a full Diligent-style stack: N accounts, each
// with a group, group_settings, and a CloudWatch monitoring provider.
// Used by SCALE-002, SCALE-003, DEP-010 system tests.
func testAccFullStackConfig(prefix string, count int) string {
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString(fmt.Sprintf(`
resource "sedai_account" "acct%[1]d" {
  name             = "%[2]s-acct-%[1]d"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::%[3]012d:role/SedaiRole"
  external_id      = "sedai-ext-%[1]d"
}

resource "sedai_group" "grp%[1]d" {
  name       = "%[2]s-grp-%[1]d"
  depends_on = [sedai_account.acct%[1]d]
}

resource "sedai_group_settings" "gset%[1]d" {
  group_id          = sedai_group.grp%[1]d.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}

resource "sedai_cloudwatch_monitoring_provider" "cw%[1]d" {
  account_id              = sedai_account.acct%[1]d.id
  name                    = "%[2]s-cw-%[1]d"
  integration_type        = "ROLE"
  use_account_credentials = true
  depends_on              = [sedai_account.acct%[1]d]
}
`, i, prefix, 100000000000+int64(i)))
	}
	return sb.String()
}

