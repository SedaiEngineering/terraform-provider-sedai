package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIMPORT(t *testing.T) {
	// IMPORT-001: Import sedai_account
	t.Run("IMPORT-001", func(t *testing.T) {
		name := "imp-acct-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"role", "external_id", "access_key", "secret_key", "service_account_json"},
				},
			},
		})
	})

	// IMPORT-002: Import sedai_group
	t.Run("IMPORT-002", func(t *testing.T) {
		name := "imp-grp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
				},
				{
					ResourceName:      "sedai_group.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// IMPORT-003: Import sedai_group_settings
	t.Run("IMPORT-003", func(t *testing.T) {
		name := "imp-gset-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					ResourceName:      "sedai_group_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// IMPORT-004: Import sedai_account_settings
	t.Run("IMPORT-004", func(t *testing.T) {
		name := "imp-aset-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					ResourceName:      "sedai_account_settings.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// IMPORT-005: Import sedai_group_priority
	t.Run("IMPORT-005", func(t *testing.T) {
		prefix := "imp-pri-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Single(prefix, 1),
				},
				{
					ResourceName:      "sedai_group_priority.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// IMPORT-006: Import sedai_cloudwatch_monitoring_provider
	t.Run("IMPORT-006", func(t *testing.T) {
		name := "imp-cw-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccCloudWatchConfig_AccountCreds(name),
				},
				{
					ResourceName:            "sedai_cloudwatch_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"access_key", "secret_key"},
				},
			},
		})
	})

	// IMPORT-007: Import sedai_datadog_monitoring_provider
	t.Run("IMPORT-007", func(t *testing.T) {
		name := "imp-dd-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDatadogConfig(name),
				},
				{
					ResourceName:            "sedai_datadog_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"api_key", "application_key"},
				},
			},
		})
	})

	// IMPORT-008: Import sedai_newrelic_monitoring_provider
	t.Run("IMPORT-008", func(t *testing.T) {
		name := "imp-nr-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccNewRelicConfig(name),
				},
				{
					ResourceName:            "sedai_newrelic_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"api_key"},
				},
			},
		})
	})

	// IMPORT-009: Import sedai_fp_monitoring_provider
	t.Run("IMPORT-009", func(t *testing.T) {
		name := "imp-fp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFPMonitoringConfig_NoAuth(name),
				},
				{
					ResourceName:      "sedai_fp_monitoring_provider.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// IMPORT-010: Import sedai_gke_monitoring_provider
	t.Run("IMPORT-010", func(t *testing.T) {
		name := "imp-gke-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGKEMonitoringConfig(name),
				},
				{
					ResourceName:            "sedai_gke_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"service_account_json"},
				},
			},
		})
	})

	// IMPORT-011: Import non-existent ID → clear error
	t.Run("IMPORT-011", func(t *testing.T) {
		name := "imp-gone-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
				},
				{
					ResourceName:  "sedai_account.test",
					ImportState:   true,
					ImportStateId: "non-existent-id-zzz",
					ExpectError:   regexpMustCompile(`(not found|404|does not exist)`),
				},
			},
		})
	})

	// IMPORT-012: After import, plan shows 0 changes
	t.Run("IMPORT-012", func(t *testing.T) {
		name := "imp-nodrift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"role", "external_id"},
				},
				{
					Config:             testAccAccountConfig_AWSRole(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// IMPORT-014: Full import-then-manage workflow
	t.Run("IMPORT-014", func(t *testing.T) {
		name := "imp-manage-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupWithSettingsConfig(name, "DATA_PILOT", "DATA_PILOT"),
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
				// Now manage: change mode
				{
					Config: testAccGroupWithSettingsConfig(name, "CO_PILOT", "DATA_PILOT"),
					Check:  resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "CO_PILOT"),
				},
				{
					Config:             testAccGroupWithSettingsConfig(name, "CO_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// testAccGroupWithSettingsConfigForImport is a variant that names the resources
// "test" to match import address expectations above.
func testAccGroupWithSettingsConfigForImport(name, avail, optim string) string {
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
