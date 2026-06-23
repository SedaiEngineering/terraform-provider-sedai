package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccASET(t *testing.T) {
	t.Run("ASET-001", func(t *testing.T) {
		name := "aset-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account_settings.test", "availability_mode", "DATA_PILOT"),
						resource.TestCheckResourceAttr("sedai_account_settings.test", "optimization_mode", "DATA_PILOT"),
					),
				},
			},
		})
	})

	t.Run("ASET-002", func(t *testing.T) {
		name := "aset-copilot-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "CO_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account_settings.test", "availability_mode", "CO_PILOT"),
					),
				},
			},
		})
	})

	t.Run("ASET-003", func(t *testing.T) {
		name := "aset-auto-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "AUTO", "AUTO"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account_settings.test", "availability_mode", "AUTO"),
						resource.TestCheckResourceAttr("sedai_account_settings.test", "optimization_mode", "AUTO"),
					),
				},
			},
		})
	})

	// ASET-004: sedai_sync_enabled omitted → defaults false, no drift on re-plan
	t.Run("ASET-004", func(t *testing.T) {
		name := "aset-nosync-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account_settings.test", "sedai_sync_enabled", "false"),
					),
				},
				{
					Config:             testAccAccountWithSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ASET-005: sedai_sync_enabled=true persists
	t.Run("ASET-005", func(t *testing.T) {
		name := "aset-sync-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfigSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account_settings.test", "sedai_sync_enabled", "true"),
					),
				},
				{
					Config:             testAccAccountWithSettingsConfigSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ASET-006: toggle sync false→true→false
	t.Run("ASET-006", func(t *testing.T) {
		name := "aset-tog-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfigSync(name, "DATA_PILOT", "DATA_PILOT", "false"),
					Check:  resource.TestCheckResourceAttr("sedai_account_settings.test", "sedai_sync_enabled", "false"),
				},
				{
					Config: testAccAccountWithSettingsConfigSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					Check:  resource.TestCheckResourceAttr("sedai_account_settings.test", "sedai_sync_enabled", "true"),
				},
				{
					Config: testAccAccountWithSettingsConfigSync(name, "DATA_PILOT", "DATA_PILOT", "false"),
					Check:  resource.TestCheckResourceAttr("sedai_account_settings.test", "sedai_sync_enabled", "false"),
				},
			},
		})
	})

	// ASET-016: AUTO mode + app_settings block is INVALID — same validator as group_settings.
	// validateTopLevelModeConflicts rejects AUTO combined with app_settings; fires in
	// Create before any API call, so resource.UnitTest works without a backend.
	t.Run("ASET-016", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountSettingsConfig_InvalidAutoWithAppSettings(),
					ExpectError: regexpMustCompile(`(?i)(AUTO.*app_settings|app_settings.*AUTO|availability_mode.*AUTO)`),
				},
			},
		})
	})

	// ASET-017: Invalid availability_mode value rejected by plan-time settingsConfigModeValidator.
	t.Run("ASET-017", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountSettingsConfig_InvalidMode(),
					ExpectError: regexpMustCompile(`(availability_mode|optimization_mode)`),
				},
			},
		})
	})

	// ASET-018: Changing account_id forces replacement
	t.Run("ASET-018", func(t *testing.T) {
		name1 := "aset-r1-" + randString(5)
		name2 := "aset-r2-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name1, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config: testAccAccountWithSettingsConfigTwoAccounts(name1, name2, "DATA_PILOT", "DATA_PILOT"),
					// Switching account_id on the settings resource requires replace
					ExpectNonEmptyPlan: true,
				},
			},
		})
	})

	// ASET-019: Partial spec — only mode fields, no settings blocks → no drift
	t.Run("ASET-019", func(t *testing.T) {
		name := "aset-partial-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "CO_PILOT", "DATA_PILOT"),
				},
				{
					Config:             testAccAccountWithSettingsConfig(name, "CO_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ASET-020: Account update does not affect settings
	t.Run("ASET-020", func(t *testing.T) {
		name := "aset-survive-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config: testAccAccountWithSettingsConfigAndManagedServices(name, "DATA_PILOT", "DATA_PILOT", `["LAMBDA"]`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account_settings.test", "availability_mode", "DATA_PILOT"),
					),
				},
			},
		})
	})
}

func testAccAccountWithSettingsConfig(name, avail, optim string) string {
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
  availability_mode = %[2]q
  optimization_mode = %[3]q
}
`, name, avail, optim)
}

func testAccAccountWithSettingsConfigNoSync(name, avail, optim string) string {
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
  availability_mode = %[2]q
  optimization_mode = %[3]q
  # sedai_sync_enabled intentionally omitted — must not drift
}
`, name, avail, optim)
}

func testAccAccountWithSettingsConfigSync(name, avail, optim, syncEnabled string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_account_settings" "test" {
  account_id         = sedai_account.test.id
  availability_mode  = %[2]q
  optimization_mode  = %[3]q
  sedai_sync_enabled = %[4]s
}
`, name, avail, optim, syncEnabled)
}

func testAccAccountWithSettingsConfigTwoAccounts(name1, name2, avail, optim string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_account" "test2" {
  name             = %[2]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-456"
}

resource "sedai_account_settings" "test" {
  account_id        = sedai_account.test2.id
  availability_mode = %[3]q
  optimization_mode = %[4]q
}
`, name1, name2, avail, optim)
}

func testAccAccountWithSettingsConfigAndManagedServices(name, avail, optim, services string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
  managed_services = %[4]s
}

resource "sedai_account_settings" "test" {
  account_id        = sedai_account.test.id
  availability_mode = %[2]q
  optimization_mode = %[3]q
}
`, name, avail, optim, services)
}

func testAccAccountSettingsConfig_InvalidAutoWithAppSettings() string {
	// Uses hardcoded account_id so no sedai_account API call is needed.
	// validateTopLevelModeConflicts fires in Create before any API call.
	return `
resource "sedai_account_settings" "test" {
  account_id        = "validator-test-account"
  availability_mode = "AUTO"
  optimization_mode = "DATA_PILOT"
  app_settings {
    availability_mode = "DATA_PILOT"
  }
}
`
}

func testAccAccountSettingsConfig_InvalidMode() string {
	// Uses hardcoded account_id so no sedai_account API call is needed.
	// settingsConfigModeValidator fires at plan time before any API call.
	return `
resource "sedai_account_settings" "test" {
  account_id        = "validator-test-account"
  availability_mode = "INVALID_MODE"
  optimization_mode = "DATA_PILOT"
}
`
}
