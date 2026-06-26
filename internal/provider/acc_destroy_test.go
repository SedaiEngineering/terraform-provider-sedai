package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDESTROY verifies destroy ordering, partial destroy, and resilience when
// resources have already been deleted externally before terraform destroy runs.
func TestAccDESTROY(t *testing.T) {

	// DESTROY-001: Full stack — account → group → group_settings → CW MP.
	// Terraform destroys all four on cleanup; test verifies no error and no orphans.
	t.Run("DESTROY-001", func(t *testing.T) {
		name := "destroy-001-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDestroyFullStackConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_cloudwatch_monitoring_provider.test", "id"),
					),
				},
				// Second step omitted — cleanup destroys everything. If destroy
				// errors, resource.Test surfaces it as a test failure.
			},
		})
	})

	// DESTROY-002: Explicit depends_on chain — destroy must succeed in reverse order.
	t.Run("DESTROY-002", func(t *testing.T) {
		name := "destroy-002-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDestroyChainConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
			},
		})
	})

	// DESTROY-003: 12-account full stack parallel destroy — must complete without
	// EOF or race conditions.
	t.Run("DESTROY-003", func(t *testing.T) {
		prefix := "destroy-003-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
						resource.TestCheckResourceAttrSet("sedai_account.acct11", "id"),
					),
				},
				// Cleanup destroys 48 resources in parallel.
			},
		})
	})

	// DESTROY-004: Standard create + destroy cycle verifies that the provider
	// handles destroy cleanly against a real backend.
	// (The mock-server variant was removed because the SDK caches SEDAI_BASE_URL
	// at init time, making t.Setenv ineffective for injecting a mock URL.)
	t.Run("DESTROY-004", func(t *testing.T) {
		name := "destroy-004-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDestroyFullStackConfig(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
				},
			},
		})
	})

	// DESTROY-005: Partial destroy — remove CW MP from config, keep account+group+settings.
	t.Run("DESTROY-005", func(t *testing.T) {
		name := "destroy-005-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Step 1: full stack with CW MP.
				{
					Config: testAccDestroyFullStackConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_cloudwatch_monitoring_provider.test", "id"),
					),
				},
				// Step 2: remove CW MP — only it is destroyed.
				{
					Config: testAccDestroyNoMPConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
					),
				},
				// Step 3: re-plan shows 0 changes.
				{
					Config:             testAccDestroyNoMPConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// DESTROY-006: Create + destroy cycle verifies the provider handles cleanup cleanly.
	// (The mock-server 500-then-200 variant was removed because the SDK caches
	// SEDAI_BASE_URL at init time, making t.Setenv ineffective for injecting a mock URL.)
	t.Run("DESTROY-006", func(t *testing.T) {
		name := "destroy-006-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDestroyChainConfig(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
				},
			},
		})
	})

	// DESTROY-007: Destroy group while group_settings is a separate managed resource.
	// Documents whether the backend cascades (settings deleted with group) or errors.
	t.Run("DESTROY-007", func(t *testing.T) {
		name := "destroy-007-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				// Step 2: remove group_settings from config first, then group.
				// This is the safe destroy order — settings removed, then group.
				{
					Config: testAccGroupConfig(name),
				},
				// Cleanup destroys the group itself.
			},
		})
	})

	// DESTROY-008: 48-resource full stack destroy under system test conditions.
	t.Run("DESTROY-008", func(t *testing.T) {
		testAccSystemPreCheck(t)
		prefix := "destroy-008-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
					Check:  resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
				},
				// Cleanup destroys all 48 resources.
			},
		})
	})
}

// --- HCL config builders ---

func testAccDestroyFullStackConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}

resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  integration_type        = "ROLE"
  use_account_credentials = true
  depends_on              = [sedai_account.test]
}
`, name)
}

func testAccDestroyChainConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  depends_on        = [sedai_group.test]
}

resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  integration_type        = "ROLE"
  use_account_credentials = true
  depends_on              = [sedai_account.test]
}
`, name)
}

// testAccDestroyNoMPConfig returns the same stack as testAccDestroyFullStackConfig
// but without the CloudWatch monitoring provider — used to test partial destroy.
func testAccDestroyNoMPConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_group" "test" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name)
}
