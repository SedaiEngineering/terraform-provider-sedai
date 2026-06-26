package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGSET covers all sedai_group_settings acceptance and validator tests.
// Subtests are named GSET-NNN to match the testing manifest IDs.
//
// Critical P0 regression tests: GSET-006 through GSET-011 guard the
// sedai_sync_enabled false->null drift bug reported by Diligent (fb-jun-17).
func TestAccGSET(t *testing.T) {

	// ── Happy-path CRUD ───────────────────────────────────────────────────

	t.Run("GSET-001", func(t *testing.T) {
		name := "gset-basic-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "DATA_PILOT"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "DATA_PILOT"),
					),
				},
			},
		})
	})

	t.Run("GSET-002", func(t *testing.T) {
		name := "gset-copilot-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "CO_PILOT", "CO_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "CO_PILOT"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "CO_PILOT"),
					),
				},
			},
		})
	})

	t.Run("GSET-003", func(t *testing.T) {
		name := "gset-auto-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "AUTO", "AUTO"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "AUTO"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "AUTO"),
					),
				},
			},
		})
	})

	t.Run("GSET-004", func(t *testing.T) {
		name := "gset-sync-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "true"),
					),
				},
			},
		})
	})

	t.Run("GSET-005", func(t *testing.T) {
		name := "gset-syncoff-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "false"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "false"),
					),
				},
			},
		})
	})

	// ── P0 drift regression: sedai_sync_enabled false→null (Diligent jun-17) ─

	// GSET-006: Omitting sedai_sync_enabled must NOT produce false->null drift
	// on the next plan. This was the exact bug Diligent reported — every
	// re-plan showed a spurious diff that required manual apply to "fix",
	// then reappeared on the next plan. The fix is Default: StaticBool(false).
	t.Run("GSET-006", func(t *testing.T) {
		name := "gset-nosync-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "false"),
					),
				},
				// Re-plan with identical config: must show ZERO changes.
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-007: Same regression check with explicit sedai_sync_enabled=false.
	// Both omitted and explicit false must be idempotent.
	t.Run("GSET-007", func(t *testing.T) {
		name := "gset-explfalse-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "false"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "false"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "false"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-008: Apply omitting sedai_sync_enabled, then switch to explicit true.
	// Verifies the default does not prevent toggling on.
	t.Run("GSET-008", func(t *testing.T) {
		name := "gset-toggle-sync-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "false"),
					),
				},
				{
					Config: testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "true"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-009: Apply with explicit true, then drop sedai_sync_enabled entirely.
	// The attribute must revert to false without drift on the subsequent plan.
	t.Run("GSET-009", func(t *testing.T) {
		name := "gset-drop-sync-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigWithSync(name, "DATA_PILOT", "DATA_PILOT", "true"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "true"),
					),
				},
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "sedai_sync_enabled", "false"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-010: Mode update from DATA_PILOT to CO_PILOT must be drift-free.
	t.Run("GSET-010", func(t *testing.T) {
		name := "gset-mode-upd-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config: testAccGroupSettingsConfigNoSync(name, "CO_PILOT", "CO_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "CO_PILOT"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "CO_PILOT"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "CO_PILOT", "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-011: 12 groups each with settings but no sedai_sync_enabled — zero
	// drift after apply. This is the exact Diligent 12-group-per-account pattern.
	t.Run("GSET-011", func(t *testing.T) {
		prefix := "gset-12g-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSET12GroupsNoSyncConfig(prefix),
				},
				{
					Config:             testAccGSET12GroupsNoSyncConfig(prefix),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── Nested block tests ────────────────────────────────────────────────

	t.Run("GSET-012", func(t *testing.T) {
		name := "gset-kube-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETKubeBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.0.horizontal_scaling_enabled", "true"),
					),
				},
				{
					Config:             testAccGSETKubeBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-013", func(t *testing.T) {
		name := "gset-serverless-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETServerlessBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				{
					Config:             testAccGSETServerlessBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-014", func(t *testing.T) {
		name := "gset-bucket-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETBucketBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				{
					Config:             testAccGSETBucketBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-015", func(t *testing.T) {
		name := "gset-volume-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETVolumeBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				{
					Config:             testAccGSETVolumeBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-016", func(t *testing.T) {
		name := "gset-app-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETAppBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				{
					Config:             testAccGSETAppBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-017: Add kube block on second apply, then remove it. Both steps
	// must be drift-free after their apply.
	t.Run("GSET-017", func(t *testing.T) {
		name := "gset-add-kube-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config: testAccGSETKubeBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.0.horizontal_scaling_enabled", "true"),
					),
				},
				{
					Config:             testAccGSETKubeBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-018: Import existing group settings via group_id.
	t.Run("GSET-018", func(t *testing.T) {
		name := "gset-import-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					ResourceName:            "sedai_group_settings.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"sedai_sync_enabled"},
				},
			},
		})
	})

	// GSET-019: After import, re-plan is empty.
	t.Run("GSET-019", func(t *testing.T) {
		name := "gset-import-nodrift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					ResourceName:      "sedai_group_settings.test",
					ImportState:       true,
					ImportStateVerify: false,
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-020: Mixed availability_mode / optimization_mode (DATA_PILOT / CO_PILOT).
	t.Run("GSET-020", func(t *testing.T) {
		name := "gset-mixed-modes-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "CO_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "DATA_PILOT"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "CO_PILOT"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "CO_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-021: Partial kube_app_settings spec — only boolean flags, no integers.
	// The unset integer fields must not appear in subsequent plan diffs.
	t.Run("GSET-021", func(t *testing.T) {
		name := "gset-kube-partial-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETKubePartialConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.0.horizontal_scaling_enabled", "true"),
					),
				},
				{
					Config:             testAccGSETKubePartialConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GSET-022: group_id change forces replacement (RequiresReplace modifier).
	t.Run("GSET-022", func(t *testing.T) {
		name1 := "gset-grp-a-" + randString(4)
		name2 := "gset-grp-b-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETTwoGroupsFirstSettings(name1, name2),
				},
				{
					Config: testAccGSETTwoGroupsSecondSettings(name1, name2),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
			},
		})
	})

	// GSET-023: Deleting group_settings resets modes to DATA_PILOT (soft delete).
	t.Run("GSET-023", func(t *testing.T) {
		name := "gset-delete-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "CO_PILOT", "CO_PILOT"),
				},
				// Destroy happens automatically at end of test.
				// Verify the group resource itself is NOT destroyed.
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
					),
				},
			},
		})
	})

	// ── Mode-conflict validator tests (run without TF_ACC) ────────────────

	// GSET-024: TOP-LEVEL availability_mode=AUTO + app_settings → error.
	// Caught by validateTopLevelModeConflicts in Create (before any API call).
	t.Run("GSET-024", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_AutoAvailWithAppSettings(),
					ExpectError: regexpMustCompile(`AUTO`),
				},
			},
		})
	})

	// GSET-025: TOP-LEVEL optimization_mode=AUTO + app_settings → error.
	t.Run("GSET-025", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_AutoOptimWithAppSettings(),
					ExpectError: regexpMustCompile(`AUTO`),
				},
			},
		})
	})

	// GSET-026: TOP-LEVEL optimization_mode=AUTO + bucket_settings → error.
	t.Run("GSET-026", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_AutoOptimWithBucketSettings(),
					ExpectError: regexpMustCompile(`AUTO`),
				},
			},
		})
	})

	// GSET-027: TOP-LEVEL availability_mode=AUTO + volume_settings → error.
	t.Run("GSET-027", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_AutoAvailWithVolumeSettings(),
					ExpectError: regexpMustCompile(`AUTO`),
				},
			},
		})
	})

	// GSET-028: TOP-LEVEL optimization_mode=CO_PILOT + serverless_settings → error.
	t.Run("GSET-028", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_CoPilotOptimWithServerless(),
					ExpectError: regexpMustCompile(`CO_PILOT`),
				},
			},
		})
	})

	// GSET-029: app_settings.availability_mode=AUTO fires noAutoConfigModeValidator
	// at plan time (before any Create call).
	t.Run("GSET-029", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_AppBlockAutoMode(),
					ExpectError: regexpMustCompile(`(?i)(AUTO|value must be one of)`),
				},
			},
		})
	})

	// GSET-030: serverless_settings.optimization_mode=CO_PILOT fires
	// noCopilotConfigModeValidator at plan time.
	t.Run("GSET-030", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_ServerlessCoPilot(),
					ExpectError: regexpMustCompile(`(?i)(CO_PILOT|value must be one of)`),
				},
			},
		})
	})

	// GSET-031: availability_mode with invalid value fires settingsConfigModeValidator.
	t.Run("GSET-031", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_BadAvailMode(),
					ExpectError: regexpMustCompile(`(?i)(value must be one of|SPEED|availability_mode)`),
				},
			},
		})
	})

	// GSET-032: optimization_mode with invalid value.
	t.Run("GSET-032", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGSETInvalid_BadOptimMode(),
					ExpectError: regexpMustCompile(`(?i)(value must be one of|TURBO|optimization_mode)`),
				},
			},
		})
	})

	// ── Mixed-mode and combined block tests ───────────────────────────────

	t.Run("GSET-033", func(t *testing.T) {
		name := "gset-avail-auto-optim-dp-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "AUTO", "DATA_PILOT"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "availability_mode", "AUTO"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "optimization_mode", "DATA_PILOT"),
					),
				},
				{
					Config:             testAccGroupSettingsConfigNoSync(name, "AUTO", "DATA_PILOT"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-034", func(t *testing.T) {
		name := "gset-kube-full-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETKubeFullConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.0.horizontal_scaling_min_replicas", "2"),
						resource.TestCheckResourceAttr("sedai_group_settings.test", "kube_app_settings.0.horizontal_scaling_max_replicas", "20"),
					),
				},
				{
					Config:             testAccGSETKubeFullConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GSET-035", func(t *testing.T) {
		name := "gset-container-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGSETContainerAppBlockConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
				{
					Config:             testAccGSETContainerAppBlockConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// ── GSET config builders ──────────────────────────────────────────────────────

func testAccGSET12GroupsNoSyncConfig(prefix string) string {
	var sb strings.Builder
	sb.WriteString(testAccProviderConfig())
	for i := 0; i < 12; i++ {
		sb.WriteString(fmt.Sprintf(`
resource "sedai_group" "grp%d" {
  name = "%s-g%d"
}

resource "sedai_group_settings" "gset%d" {
  group_id          = sedai_group.grp%d.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, i, prefix, i, i, i))
	}
	return sb.String()
}

func testAccGSETKubeBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  kube_app_settings {
    horizontal_scaling_enabled = true
    vertical_scaling_enabled   = true
  }
}
`, name)
}

func testAccGSETKubePartialConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  kube_app_settings {
    horizontal_scaling_enabled = true
    vertical_scaling_enabled   = false
  }
}
`, name)
}

func testAccGSETKubeFullConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  kube_app_settings {
    horizontal_scaling_enabled      = true
    horizontal_scaling_min_replicas = 2
    horizontal_scaling_max_replicas = 20
    vertical_scaling_enabled        = true
  }
}
`, name)
}

func testAccGSETServerlessBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  serverless_settings {
    optimization_mode = "AUTO"
    concurrency_mode  = "AUTO"
  }
}
`, name)
}

func testAccGSETBucketBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "CO_PILOT"

  bucket_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`, name)
}

func testAccGSETVolumeBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  volume_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "CO_PILOT"
  }
}
`, name)
}

func testAccGSETAppBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  app_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "CO_PILOT"
  }
}
`, name)
}

func testAccGSETContainerAppBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  container_app_settings {
    availability_mode = "DATA_PILOT"
    optimization_mode = "DATA_PILOT"
  }
}
`, name)
}

func testAccGSETTwoGroupsFirstSettings(name1, name2 string) string {
	return fmt.Sprintf(`
resource "sedai_group" "grpa" {
  name = %q
}

resource "sedai_group" "grpb" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.grpa.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name1, name2)
}

func testAccGSETTwoGroupsSecondSettings(name1, name2 string) string {
	return fmt.Sprintf(`
resource "sedai_group" "grpa" {
  name = %q
}

resource "sedai_group" "grpb" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.grpb.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, name1, name2)
}

// ── Invalid configs for validator tests ──────────────────────────────────────
//
// GSET-024 to GSET-028 test validateTopLevelModeConflicts which fires in
// Create() before any API call. To avoid the sedai_group resource attempting
// an API call first, these configs use a hardcoded group_id literal so only
// sedai_group_settings.test is in the plan.
//
// GSET-029 to GSET-032 test schema-level attribute validators which fire at
// plan time — no API call is ever made.

func testAccGSETInvalid_AutoAvailWithAppSettings() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "AUTO"
  optimization_mode = "DATA_PILOT"

  app_settings {
    availability_mode = "DATA_PILOT"
  }
}
`
}

func testAccGSETInvalid_AutoOptimWithAppSettings() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "AUTO"

  app_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGSETInvalid_AutoOptimWithBucketSettings() string {
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

func testAccGSETInvalid_AutoAvailWithVolumeSettings() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "AUTO"
  optimization_mode = "DATA_PILOT"

  volume_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`
}

func testAccGSETInvalid_CoPilotOptimWithServerless() string {
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

// testAccGSETInvalid_AppBlockAutoMode: app_settings.availability_mode=AUTO
// fires noAutoConfigModeValidator at plan time (schema validator, no API call).
func testAccGSETInvalid_AppBlockAutoMode() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  app_settings {
    availability_mode = "AUTO"
  }
}
`
}

// testAccGSETInvalid_ServerlessCoPilot: serverless_settings.optimization_mode=CO_PILOT
// fires noCopilotConfigModeValidator at plan time (schema validator, no API call).
func testAccGSETInvalid_ServerlessCoPilot() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  serverless_settings {
    optimization_mode = "CO_PILOT"
  }
}
`
}

func testAccGSETInvalid_BadAvailMode() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "SPEED"
  optimization_mode = "DATA_PILOT"
}
`
}

func testAccGSETInvalid_BadOptimMode() string {
	return `
resource "sedai_group_settings" "test" {
  group_id          = "validator-test-group"
  availability_mode = "DATA_PILOT"
  optimization_mode = "TURBO"
}
`
}
