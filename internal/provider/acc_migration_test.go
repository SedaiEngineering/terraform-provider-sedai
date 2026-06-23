package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMIGR covers v1→v2 state migration scenarios for all major resource types.
// These use the ImportState + PlanOnly pattern because no real v1 state files are
// available; ImportState exercises the same Read code path as a migration would.
// MIGR-009 through MIGR-016 cover the remaining gaps not addressed in MIGR-001–008.
func TestAccMIGR(t *testing.T) {

	t.Run("MIGR-009", func(t *testing.T) {
		// Account re-imported from ID then verified stable (Azure cloud provider)
		name := "migr-az-acct-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccMigrAzureAccountConfig(name)},
				{
					ResourceName:      "sedai_account.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{Config: testAccMigrAzureAccountConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-010", func(t *testing.T) {
		// Group with no settings re-imported — verifies resource count stability
		name := "migr-grp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupConfig(name)},
				{
					ResourceName:      "sedai_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{Config: testAccGroupConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-011", func(t *testing.T) {
		// group_settings re-imported — verifies sedai_sync_enabled null vs false stability
		name := "migr-gset-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT")},
				{
					ResourceName:      "sedai_group_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				// After import, a re-plan must show 0 changes (the P0 drift bug)
				{Config: testAccGroupWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-012", func(t *testing.T) {
		// CloudWatch monitoring provider re-imported, sensitive fields excluded from verify
		name := "migr-cw-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccMigrCloudWatchConfig(name)},
				{
					ResourceName:            "sedai_cloudwatch_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"access_key", "secret_key"},
				},
				{Config: testAccMigrCloudWatchConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-013", func(t *testing.T) {
		// Datadog monitoring provider re-imported — api_key and application_key excluded
		name := "migr-dd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccMigrDatadogConfig(name)},
				{
					ResourceName:            "sedai_datadog_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"api_key", "application_key"},
				},
				{Config: testAccMigrDatadogConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-014", func(t *testing.T) {
		// VictoriaMetrics monitoring provider re-imported — bearer_token excluded
		name := "migr-vm-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccMigrVMConfig(name)},
				{
					ResourceName:            "sedai_victoria_metrics_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"bearer_token", "client_secret"},
				},
				{Config: testAccMigrVMConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-015", func(t *testing.T) {
		// Full stack (account + group + settings + MP) imported piecemeal and verified stable.
		// Simulates a Diligent migration where all resources were created outside Terraform.
		name := "migr-full-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Create the full stack
				{Config: testAccMigrFullStackConfig(name)},
				// Import each resource in turn
				{
					ResourceName:      "sedai_account.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					ResourceName:      "sedai_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					ResourceName:      "sedai_group_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					ResourceName:            "sedai_cloudwatch_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"access_key", "secret_key"},
				},
				// After all imports, the full config still plans clean
				{Config: testAccMigrFullStackConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MIGR-016", func(t *testing.T) {
		// account_settings re-imported then verified stable (covers the third P0 Diligent bug:
		// resource counts in state must remain consistent across reads).
		name := "migr-aset-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccMigrAccountSettingsConfig(name)},
				{
					ResourceName:      "sedai_account_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{Config: testAccMigrAccountSettingsConfig(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})
}

// ---- HCL builders for MIGR tests ----

func testAccMigrAzureAccountConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AZURE"
  integration_type = "ROLE"
  subscription_id  = "00000000-0000-0000-0000-000000000001"
  tenant_id        = "00000000-0000-0000-0000-000000000002"
  client_id        = "00000000-0000-0000-0000-000000000003"
  client_secret    = "azure-test-secret"
}
`, name)
}

func testAccMigrCloudWatchConfig(name string) string {
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
  use_account_credentials = true
}
`, name)
}

func testAccMigrDatadogConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_datadog_monitoring_provider" "test" {
  account_id      = sedai_account.test.id
  name            = %[1]q
  api_key         = "dd-api-key-migr"
  application_key = "dd-app-key-migr"
}
`, name)
}

func testAccMigrVMConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AWS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_victoria_metrics_monitoring_provider" "test" {
  account_id   = sedai_account.test.id
  name         = %[1]q
  endpoint     = "http://victoria-metrics.example.com:8428"
  bearer_token = "eyJhbGciOiJSUzI1NiJ9.migr-token"
}
`, name)
}

func testAccMigrFullStackConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}

resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
  depends_on              = [sedai_account.test]
}
`, name)
}

func testAccMigrAccountSettingsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_account_settings" "test" {
  account_id        = sedai_account.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name)
}
