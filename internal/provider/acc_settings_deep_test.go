package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ─── GSET deep-nesting tests ─────────────────────────────────────────────────
// Each sub-block test follows the same two-step pattern:
//   1. Apply config with the block populated
//   2. PlanOnly: true, ExpectNonEmptyPlan: false  →  0 planned changes

func TestAccGSET_DeepNesting(t *testing.T) {

	// ── kube_app_settings ────────────────────────────────────────────────────

	t.Run("GSET-036", func(t *testing.T) {
		name := "gset-kube-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithKube(name, "2", "8", "1.0", "268435456")},
				{
					Config:             testAccGroupSettingsConfig_WithKube(name, "2", "8", "1.0", "268435456"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-037", func(t *testing.T) {
		name := "gset-kube-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithKube(name, "1", "4", "0.5", "134217728")},
				{Config: testAccGroupSettingsConfig_WithKube(name, "2", "10", "1.0", "268435456")},
				{
					Config:             testAccGroupSettingsConfig_WithKube(name, "2", "10", "1.0", "268435456"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-038", func(t *testing.T) {
		name := "gset-kube-rm-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithKube(name, "2", "8", "1.0", "268435456")},
				// Remove the kube_app_settings block — unmanaged after removal
				{Config: testAccGroupSettingsConfig_NoSubblocks(name)},
				{
					Config:             testAccGroupSettingsConfig_NoSubblocks(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-039", func(t *testing.T) {
		name := "gset-kube-partial-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				// Only manage horizontal_scaling_min/max — other kube fields unmanaged
				{Config: testAccGroupSettingsConfig_KubePartial(name, "1", "5")},
				{
					Config:             testAccGroupSettingsConfig_KubePartial(name, "1", "5"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── bucket_settings ───────────────────────────────────────────────────────

	t.Run("GSET-040", func(t *testing.T) {
		name := "gset-bucket-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithBucket(name, "DATA_PILOT")},
				{
					Config:             testAccGroupSettingsConfig_WithBucket(name, "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-041", func(t *testing.T) {
		name := "gset-bucket-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithBucket(name, "DATA_PILOT")},
				{Config: testAccGroupSettingsConfig_WithBucket(name, "CO_PILOT")},
				{
					Config:             testAccGroupSettingsConfig_WithBucket(name, "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// bucket_settings: AUTO optimization_mode → plan-time validator rejects
	t.Run("GSET-042", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_BucketAutoInvalid(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|invalid|not supported|bucket)`),
				},
			},
		})
	})

	// ── serverless_settings (Lambda) ─────────────────────────────────────────

	t.Run("GSET-043", func(t *testing.T) {
		name := "gset-lambda-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST", "AUTO")},
				{
					Config:             testAccGroupSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST", "AUTO"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-044", func(t *testing.T) {
		name := "gset-lambda-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST", "AUTO")},
				{Config: testAccGroupSettingsConfig_WithServerless(name, "DATA_PILOT", "DATA_PILOT", "DURATION", "OFF")},
				{
					Config:             testAccGroupSettingsConfig_WithServerless(name, "DATA_PILOT", "DATA_PILOT", "DURATION", "OFF"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// serverless_settings: CO_PILOT optimization_mode → validator rejects (no CO_PILOT for Lambda)
	t.Run("GSET-045", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_ServerlessCopilotInvalid(),
					ExpectError: regexp.MustCompile(`(?i)(CO_PILOT|invalid|not supported|serverless|lambda)`),
				},
			},
		})
	})

	// serverless_settings + top-level CO_PILOT optimization_mode → provider-level conflict
	t.Run("GSET-046", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_ServerlessTopLevelConflict(),
					ExpectError: regexp.MustCompile(`(?i)(CO_PILOT|serverless|conflict|invalid)`),
				},
			},
		})
	})

	// ── ecs_app_settings ──────────────────────────────────────────────────────

	t.Run("GSET-047", func(t *testing.T) {
		name := "gset-ecs-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithECS(name, "1", "4", "256", "512")},
				{
					Config:             testAccGroupSettingsConfig_WithECS(name, "1", "4", "256", "512"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-048", func(t *testing.T) {
		name := "gset-ecs-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithECS(name, "1", "4", "256", "512")},
				{Config: testAccGroupSettingsConfig_WithECS(name, "2", "10", "512", "1024")},
				{
					Config:             testAccGroupSettingsConfig_WithECS(name, "2", "10", "512", "1024"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── volume_settings (EBS) ─────────────────────────────────────────────────

	t.Run("GSET-049", func(t *testing.T) {
		name := "gset-vol-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_WithVolume(name, "DATA_PILOT", "CO_PILOT")},
				{
					Config:             testAccGroupSettingsConfig_WithVolume(name, "DATA_PILOT", "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// volume_settings: AUTO availability_mode → validator rejects
	t.Run("GSET-050", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_VolumeAutoInvalid(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|invalid|not supported|volume)`),
				},
			},
		})
	})

	// volume_settings + top-level AUTO → provider-level conflict
	t.Run("GSET-051", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_VolumeTopLevelConflict(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|volume|conflict|invalid)`),
				},
			},
		})
	})

	// ── app_settings ──────────────────────────────────────────────────────────

	// GSET-060: SED-21393 regression — AUTO mode with NO app_settings → 0 drift
	t.Run("GSET-060", func(t *testing.T) {
		name := "gset-auto-noapp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_AutoModeNoApp(name)},
				{
					Config:             testAccGroupSettingsConfig_AutoModeNoApp(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-061: CO_PILOT mode + app_settings block → 0 drift
	t.Run("GSET-061", func(t *testing.T) {
		name := "gset-copilot-app-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_CoPilotWithApp(name)},
				{
					Config:             testAccGroupSettingsConfig_CoPilotWithApp(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-062: AUTO mode + app_settings block → provider rejects at apply time
	t.Run("GSET-062", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_AutoModeWithApp(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|app_settings|invalid|not supported|conflict)`),
				},
			},
		})
	})

	// GSET-063: app_settings AUTO availability_mode → plan-time validator rejects
	t.Run("GSET-063", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupSettingsConfig_AppAutoInvalid(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|invalid|not supported|app)`),
				},
			},
		})
	})

	// ── Multiple blocks simultaneously ────────────────────────────────────────

	// GSET-070: kube + serverless + bucket blocks together → 0 drift
	t.Run("GSET-070", func(t *testing.T) {
		name := "gset-multi-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGroupSettingsConfig_MultiBlock(name)},
				{
					Config:             testAccGroupSettingsConfig_MultiBlock(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// ─── ASET deep-nesting tests ─────────────────────────────────────────────────
// Mirror GSET tests but for sedai_account_settings (uses account_id).
// Each test creates a real account and configures its settings.

func TestAccASET_DeepNesting(t *testing.T) {

	t.Run("ASET-021", func(t *testing.T) {
		name := "aset-kube-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithKube(name, "2", "8")},
				{
					Config:             testAccAccountSettingsConfig_WithKube(name, "2", "8"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-022", func(t *testing.T) {
		name := "aset-kube-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithKube(name, "1", "4")},
				{Config: testAccAccountSettingsConfig_WithKube(name, "3", "12")},
				{
					Config:             testAccAccountSettingsConfig_WithKube(name, "3", "12"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-023", func(t *testing.T) {
		name := "aset-kube-rm-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithKube(name, "2", "8")},
				{Config: testAccAccountSettingsConfig_NoSubblocks(name)},
				{
					Config:             testAccAccountSettingsConfig_NoSubblocks(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-024", func(t *testing.T) {
		name := "aset-lambda-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST")},
				{
					Config:             testAccAccountSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-025", func(t *testing.T) {
		name := "aset-lambda-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithServerless(name, "DATA_PILOT", "AUTO", "COST")},
				{Config: testAccAccountSettingsConfig_WithServerless(name, "AUTO", "DATA_PILOT", "DURATION")},
				{
					Config:             testAccAccountSettingsConfig_WithServerless(name, "AUTO", "DATA_PILOT", "DURATION"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-026", func(t *testing.T) {
		name := "aset-ecs-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithECS(name, "1", "4", "256", "512")},
				{
					Config:             testAccAccountSettingsConfig_WithECS(name, "1", "4", "256", "512"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-027", func(t *testing.T) {
		name := "aset-bucket-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithBucket(name, "DATA_PILOT")},
				{
					Config:             testAccAccountSettingsConfig_WithBucket(name, "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ASET-028", func(t *testing.T) {
		name := "aset-volume-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccAccountSettingsConfig_WithVolume(name, "DATA_PILOT", "CO_PILOT")},
				{
					Config:             testAccAccountSettingsConfig_WithVolume(name, "DATA_PILOT", "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ASET-036: SED-21393 regression — AUTO mode with app_settings → error
	t.Run("ASET-036", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountSettingsConfig_AutoWithApp(),
					ExpectError: regexp.MustCompile(`(?i)(AUTO|app_settings|invalid|not supported|conflict)`),
				},
			},
		})
	})

	// ASET-037: CO_PILOT + serverless_settings → error
	t.Run("ASET-037", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountSettingsConfig_CopilotServerlessConflict(),
					ExpectError: regexp.MustCompile(`(?i)(CO_PILOT|serverless|conflict|invalid)`),
				},
			},
		})
	})
}

// ─── RSET deep-nesting tests ─────────────────────────────────────────────────
// sedai_resource_settings has no nested blocks — just top-level mode fields.
// Requires SEDAI_TEST_RESOURCE_ID (a backend-discovered resource the test tenant has).

func TestAccRSET_DeepNesting(t *testing.T) {
	resourceId := os.Getenv("SEDAI_TEST_RESOURCE_ID")
	if resourceId == "" {
		t.Skip("SEDAI_TEST_RESOURCE_ID not set — skipping resource settings tests")
	}

	t.Run("RSET-011", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_resource_settings.test", "availability_mode", "DATA_PILOT"),
						resource.TestCheckResourceAttr("sedai_resource_settings.test", "optimization_mode", "DATA_PILOT"),
						resource.TestCheckResourceAttrSet("sedai_resource_settings.test", "resource_type"),
					),
				},
				{
					Config:             testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("RSET-012", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT")},
				{Config: testAccResourceSettingsConfig(resourceId, "CO_PILOT", "DATA_PILOT")},
				{
					Config:             testAccResourceSettingsConfig(resourceId, "CO_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("RSET-013", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT")},
				{Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "CO_PILOT")},
				{
					Config:             testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("RSET-014", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccResourceSettingsConfig_WithSync(resourceId, "DATA_PILOT", "DATA_PILOT", "true")},
				{
					Config:             testAccResourceSettingsConfig_WithSync(resourceId, "DATA_PILOT", "DATA_PILOT", "true"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{Config: testAccResourceSettingsConfig_WithSync(resourceId, "DATA_PILOT", "DATA_PILOT", "false")},
				{
					Config:             testAccResourceSettingsConfig_WithSync(resourceId, "DATA_PILOT", "DATA_PILOT", "false"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("RSET-015", func(t *testing.T) {
		// resource_type is Computed — verify it survives an update without forcing replace
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT"),
					Check:  resource.TestCheckResourceAttrSet("sedai_resource_settings.test", "resource_type"),
				},
				{
					Config: testAccResourceSettingsConfig(resourceId, "CO_PILOT", "DATA_PILOT"),
					Check:  resource.TestCheckResourceAttrSet("sedai_resource_settings.test", "resource_type"),
				},
				{
					Config:             testAccResourceSettingsConfig(resourceId, "CO_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("RSET-016", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT")},
				{
					ResourceName:            "sedai_resource_settings.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"sedai_sync_enabled"},
				},
			},
		})
	})

	t.Run("RSET-017", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT")},
				{
					ResourceName:            "sedai_resource_settings.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"sedai_sync_enabled"},
				},
				{
					Config:             testAccResourceSettingsConfig(resourceId, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// ─── HCL config builders ─────────────────────────────────────────────────────

func testAccGroupSettingsConfig_NoSubblocks(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name)
}

func testAccGroupSettingsConfig_WithKube(name, minReplicas, maxReplicas, minCPUCores, minMemoryBytes string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  kube_app_settings {
    horizontal_scaling_enabled    = true
    horizontal_scaling_min_replicas = %[2]s
    horizontal_scaling_max_replicas = %[3]s
    vertical_scaling_enabled      = true
    vertical_scaling_min_cpu_cores = %[4]s
    vertical_scaling_min_memory_bytes = %[5]s
  }
}
`, name, minReplicas, maxReplicas, minCPUCores, minMemoryBytes)
}

func testAccGroupSettingsConfig_KubePartial(name, minReplicas, maxReplicas string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  kube_app_settings {
    horizontal_scaling_min_replicas = %[2]s
    horizontal_scaling_max_replicas = %[3]s
  }
}
`, name, minReplicas, maxReplicas)
}

func testAccGroupSettingsConfig_WithBucket(name, optMode string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  bucket_settings {
    optimization_mode = %[2]q
    sedai_sync_enabled = false
  }
}
`, name, optMode)
}

func testAccGroupSettingsConfig_BucketAutoInvalid() string {
	// Uses hardcoded group_id so no sedai_group API call is needed.
	// validateTopLevelModeConflicts fires in Create before any API call.
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "AUTO"
  bucket_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_WithServerless(name, availMode, optMode, optFocus, concurrencyMode string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  serverless_settings {
    availability_mode    = %[2]q
    optimization_mode    = %[3]q
    optimization_focus   = %[4]q
    concurrency_mode     = %[5]q
    max_cost_change_pct  = 10
    max_latency_change_pct = 5
  }
}
`, name, availMode, optMode, optFocus, concurrencyMode)
}

func testAccGroupSettingsConfig_ServerlessCopilotInvalid() string {
	return `
resource "sedai_group" "test" {
  name = "serverless-copilot-invalid"
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  serverless_settings {
    optimization_mode = "CO_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_ServerlessTopLevelConflict() string {
	// Uses hardcoded group_id so no sedai_group API call is needed.
	// validateTopLevelModeConflicts fires in Create before any API call.
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "CO_PILOT"
  serverless_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_WithECS(name, minReplicas, maxReplicas, minCPU, minMemory string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  ecs_app_settings {
    horizontal_scaling_enabled    = true
    horizontal_scaling_min_replicas = %[2]s
    horizontal_scaling_max_replicas = %[3]s
    vertical_scaling_enabled      = true
    vertical_scaling_min_cpu      = %[4]s
    vertical_scaling_min_memory   = %[5]s
  }
}
`, name, minReplicas, maxReplicas, minCPU, minMemory)
}

func testAccGroupSettingsConfig_WithVolume(name, availMode, optMode string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  volume_settings {
    availability_mode  = %[2]q
    optimization_mode  = %[3]q
    sedai_sync_enabled = false
  }
}
`, name, availMode, optMode)
}

func testAccGroupSettingsConfig_VolumeAutoInvalid() string {
	return `
resource "sedai_group" "test" {
  name = "vol-auto-invalid"
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  volume_settings {
    availability_mode = "AUTO"
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_VolumeTopLevelConflict() string {
	// Uses hardcoded group_id so no sedai_group API call is needed.
	// validateTopLevelModeConflicts fires in Create before any API call.
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "AUTO"
  optimization_mode = "DATA_PILOT"
  volume_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_AutoModeNoApp(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "AUTO"
  optimization_mode = "AUTO"
}
`, name)
}

func testAccGroupSettingsConfig_CoPilotWithApp(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "CO_PILOT"
  optimization_mode = "CO_PILOT"
  app_settings {
    availability_mode             = "CO_PILOT"
    optimization_mode             = "CO_PILOT"
    is_prod                       = false
    horizontal_scaling_enabled    = true
    horizontal_scaling_min_replicas = 1
    horizontal_scaling_max_replicas = 5
  }
}
`, name)
}

func testAccGroupSettingsConfig_AutoModeWithApp() string {
	// Uses hardcoded group_id so no sedai_group API call is needed.
	// validateTopLevelModeConflicts fires in Create before any API call.
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "AUTO"
  optimization_mode = "AUTO"
  app_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGroupSettingsConfig_AppAutoInvalid() string {
	return `
resource "sedai_group" "test" {
  name = "app-auto-invalid"
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "CO_PILOT"
  optimization_mode = "CO_PILOT"
  app_settings {
    availability_mode = "AUTO"
  }
}
`
}

func testAccGroupSettingsConfig_MultiBlock(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}
resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  kube_app_settings {
    horizontal_scaling_min_replicas = 1
    horizontal_scaling_max_replicas = 5
  }
  bucket_settings {
    optimization_mode  = "DATA_PILOT"
    sedai_sync_enabled = false
  }
  volume_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "CO_PILOT"
  }
}
`, name)
}

// ── account_settings config builders ─────────────────────────────────────────

func testAccAccountSettingsConfig_WithKube(name, minReplicas, maxReplicas string) string {
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
  kube_app_settings {
    horizontal_scaling_min_replicas = %[2]s
    horizontal_scaling_max_replicas = %[3]s
  }
}
`, name, minReplicas, maxReplicas)
}

func testAccAccountSettingsConfig_NoSubblocks(name string) string {
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

func testAccAccountSettingsConfig_WithServerless(name, availMode, optMode, optFocus string) string {
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
  serverless_settings {
    availability_mode  = %[2]q
    optimization_mode  = %[3]q
    optimization_focus = %[4]q
  }
}
`, name, availMode, optMode, optFocus)
}

func testAccAccountSettingsConfig_WithECS(name, minReplicas, maxReplicas, minCPU, minMemory string) string {
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
  ecs_app_settings {
    horizontal_scaling_min_replicas = %[2]s
    horizontal_scaling_max_replicas = %[3]s
    vertical_scaling_min_cpu        = %[4]s
    vertical_scaling_min_memory     = %[5]s
  }
}
`, name, minReplicas, maxReplicas, minCPU, minMemory)
}

func testAccAccountSettingsConfig_WithBucket(name, optMode string) string {
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
  bucket_settings {
    optimization_mode = %[2]q
  }
}
`, name, optMode)
}

func testAccAccountSettingsConfig_WithVolume(name, availMode, optMode string) string {
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
  volume_settings {
    availability_mode = %[2]q
    optimization_mode = %[3]q
  }
}
`, name, availMode, optMode)
}

func testAccAccountSettingsConfig_AutoWithApp() string {
	return `
resource "sedai_account" "test" {
  name             = "aset-auto-app-invalid"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_account_settings" "test" {
  account_id        = sedai_account.test.id
  availability_mode = "AUTO"
  optimization_mode = "AUTO"
  app_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccAccountSettingsConfig_CopilotServerlessConflict() string {
	return `
resource "sedai_account" "test" {
  name             = "aset-copilot-serverless"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_account_settings" "test" {
  account_id        = sedai_account.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "CO_PILOT"
  serverless_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`
}

// ── resource_settings config builders ────────────────────────────────────────

func testAccResourceSettingsConfig(resourceId, availMode, optMode string) string {
	return fmt.Sprintf(`
resource "sedai_resource_settings" "test" {
  resource_id       = %[1]q
  availability_mode = %[2]q
  optimization_mode = %[3]q
}
`, resourceId, availMode, optMode)
}

func testAccResourceSettingsConfig_WithSync(resourceId, availMode, optMode, syncEnabled string) string {
	return fmt.Sprintf(`
resource "sedai_resource_settings" "test" {
  resource_id        = %[1]q
  availability_mode  = %[2]q
  optimization_mode  = %[3]q
  sedai_sync_enabled = %[4]s
}
`, resourceId, availMode, optMode, syncEnabled)
}
