package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccDRIFT contains cross-resource idempotency tests. Every test applies
// a configuration, then runs a plan-only step and asserts zero changes.
//
// These tests directly guard the Diligent P0 issue: their CDKTF apply of 48
// resources was followed by "~ sedai_sync_enabled = false -> null" and
// spurious resource-count diffs on every subsequent plan.
//
// Run: TF_ACC=1 SEDAI_BASE_URL=... SEDAI_API_TOKEN=... go test ./... -run TestAccDRIFT
func TestAccDRIFT(t *testing.T) {

	// DRIFT-001: Single group — no drift.
	t.Run("DRIFT-001", func(t *testing.T) {
		name := "drift-grp-" + randString(6)
		cfg := testAccGroupConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-002: Group + group_settings, sedai_sync_enabled omitted — no drift.
	// This is the exact scenario from fb-jun-17: every re-plan showed
	// "~ sedai_sync_enabled = false -> null".
	t.Run("DRIFT-002", func(t *testing.T) {
		name := "drift-gset-" + randString(6)
		cfg := testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
				// Third plan to confirm the idempotency is stable, not one-shot.
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-003: Group with resource_types including KUBERNETES_DAEMONSET — no drift.
	// The provider normalises KUBERNETES_DAEMONSET ↔ KUBERNETES_DEAMONSET; this
	// test verifies the round-trip is transparent.
	t.Run("DRIFT-003", func(t *testing.T) {
		name := "drift-kube-ds-" + randString(5)
		cfg := testAccDRIFTGroupWithKubeDaemonsetConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-004: Group with tags — no drift.
	t.Run("DRIFT-004", func(t *testing.T) {
		name := "drift-tags-" + randString(5)
		cfg := testAccDRIFTGroupWithTagsConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-005: Group settings with kube_app_settings block — no drift.
	// Partial kube spec must not expand into more attributes on re-plan.
	t.Run("DRIFT-005", func(t *testing.T) {
		name := "drift-kube-blk-" + randString(5)
		cfg := testAccDRIFTGroupSettingsKubeConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-006: Group settings with serverless_settings block — no drift.
	t.Run("DRIFT-006", func(t *testing.T) {
		name := "drift-serverless-" + randString(5)
		cfg := testAccDRIFTGroupSettingsServerlessConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-007: Mode flip DATA_PILOT→CO_PILOT→DATA_PILOT leaves state clean.
	t.Run("DRIFT-007", func(t *testing.T) {
		name := "drift-mode-flip-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupSettingsConfigNoSync(name, "DATA_PILOT", "DATA_PILOT"),
				},
				{
					Config: testAccGroupSettingsConfigNoSync(name, "CO_PILOT", "CO_PILOT"),
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

	// DRIFT-008: 3 groups + settings (scale-down of Diligent pattern) — no drift.
	// Used to catch batch-level drift that single-resource tests miss.
	t.Run("DRIFT-008", func(t *testing.T) {
		prefix := "drift-3g-" + randString(4)
		cfg := testAccDRIFTThreeGroupsConfig(prefix)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-009: sedai_sync_enabled toggle (false→true→false) is idempotent.
	t.Run("DRIFT-009", func(t *testing.T) {
		name := "drift-sync-toggle-" + randString(5)
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

	// DRIFT-010: Group with multiple resource_types, including KUBERNETES_DAEMONSET.
	// Verifies type-list normalization is idempotent across all Kubernetes types.
	t.Run("DRIFT-010", func(t *testing.T) {
		name := "drift-k8s-types-" + randString(4)
		cfg := testAccDRIFTGroupAllKubeTypesConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-011: enabled=false group — no drift on omitted-field defaults.
	t.Run("DRIFT-011", func(t *testing.T) {
		name := "drift-disabled-" + randString(5)
		cfg := testAccDRIFTDisabledGroupConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-012: group settings with bucket block — no drift.
	t.Run("DRIFT-012", func(t *testing.T) {
		name := "drift-bucket-" + randString(5)
		cfg := testAccDRIFTGroupSettingsBucketConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-013: group settings with volume block — no drift.
	t.Run("DRIFT-013", func(t *testing.T) {
		name := "drift-volume-" + randString(5)
		cfg := testAccDRIFTGroupSettingsVolumeConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-014: group settings with multiple nested blocks — no drift.
	t.Run("DRIFT-014", func(t *testing.T) {
		name := "drift-multi-blk-" + randString(4)
		cfg := testAccDRIFTGroupSettingsMultiBlockConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-015: Group with regions — no drift.
	t.Run("DRIFT-015", func(t *testing.T) {
		name := "drift-regions-" + randString(5)
		cfg := testAccDRIFTGroupWithRegionsConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-016: Group with namespaces — no drift.
	t.Run("DRIFT-016", func(t *testing.T) {
		name := "drift-ns-" + randString(5)
		cfg := testAccDRIFTGroupWithNamespacesConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-017: Parent → child group hierarchy — no drift in either resource.
	t.Run("DRIFT-017", func(t *testing.T) {
		parent := "drift-parent-" + randString(4)
		child := "drift-child-" + randString(4)
		cfg := testAccDRIFTGroupParentChildConfig(parent, child)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-018: Group settings AUTO mode — no drift.
	t.Run("DRIFT-018", func(t *testing.T) {
		name := "drift-auto-" + randString(5)
		cfg := testAccGroupSettingsConfigNoSync(name, "AUTO", "AUTO")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-019: Mixed mode (AUTO avail / DATA_PILOT optim) — no drift.
	t.Run("DRIFT-019", func(t *testing.T) {
		name := "drift-mixed-" + randString(5)
		cfg := testAccGroupSettingsConfigNoSync(name, "AUTO", "DATA_PILOT")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// DRIFT-020: Complete group settings with kube + sedai_sync_enabled=false —
	// no drift on any of the nested fields.
	t.Run("DRIFT-020", func(t *testing.T) {
		name := "drift-full-gset-" + randString(4)
		cfg := testAccDRIFTFullGroupSettingsConfig(name)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})
}

// ── DRIFT config builders ─────────────────────────────────────────────────────

func testAccDRIFTGroupWithKubeDaemonsetConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name           = %q
  resource_types = ["KUBERNETES_DAEMONSET", "KUBERNETES_DEPLOYMENT"]
}
`, name)
}

func testAccDRIFTGroupWithTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
  tags {
    key   = "env"
    exact = ["production", "staging"]
  }
  tags {
    key   = "team"
    regex = ["platform-*"]
  }
}
`, name)
}

func testAccDRIFTGroupSettingsKubeConfig(name string) string {
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
    horizontal_scaling_min_replicas = 1
    horizontal_scaling_max_replicas = 10
    vertical_scaling_enabled        = true
  }
}
`, name)
}

func testAccDRIFTGroupSettingsServerlessConfig(name string) string {
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
    concurrency_mode  = "OFF"
  }
}
`, name)
}

func testAccDRIFTThreeGroupsConfig(prefix string) string {
	var sb strings.Builder
	sb.WriteString(testAccProviderConfig())
	for i := 0; i < 3; i++ {
		sb.WriteString(fmt.Sprintf(`
resource "sedai_group" "g%d" {
  name = "%s-%d"
}

resource "sedai_group_settings" "s%d" {
  group_id          = sedai_group.g%d.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}
`, i, prefix, i, i, i))
	}
	return sb.String()
}

func testAccDRIFTGroupAllKubeTypesConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name           = %q
  resource_types = [
    "KUBERNETES_DEPLOYMENT",
    "KUBERNETES_STATEFULSET",
    "KUBERNETES_DAEMONSET",
    "KUBERNETES_CRONJOB",
  ]
}
`, name)
}

func testAccDRIFTDisabledGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name         = %q
  enabled      = false
  auto_refresh = false
}
`, name)
}

func testAccDRIFTGroupSettingsBucketConfig(name string) string {
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

func testAccDRIFTGroupSettingsVolumeConfig(name string) string {
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

func testAccDRIFTGroupSettingsMultiBlockConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id           = sedai_group.test.id
  availability_mode  = "DATA_PILOT"
  optimization_mode  = "DATA_PILOT"
  sedai_sync_enabled = false

  kube_app_settings {
    horizontal_scaling_enabled = true
    vertical_scaling_enabled   = true
  }

  serverless_settings {
    optimization_mode = "AUTO"
    concurrency_mode  = "OFF"
  }

  bucket_settings {
    optimization_mode = "DATA_PILOT"
  }
}
`, name)
}

func testAccDRIFTGroupWithRegionsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name    = %q
  regions = ["us-east-1", "us-west-2", "eu-west-1"]
}
`, name)
}

func testAccDRIFTGroupWithNamespacesConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name       = %q
  namespaces = ["default", "kube-system", "monitoring"]
}
`, name)
}

func testAccDRIFTGroupParentChildConfig(parent, child string) string {
	return fmt.Sprintf(`
resource "sedai_group" "parent" {
  name = %q
}

resource "sedai_group" "test" {
  name            = %q
  parent_group_id = sedai_group.parent.id
}
`, parent, child)
}

func testAccDRIFTFullGroupSettingsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
}

resource "sedai_group_settings" "test" {
  group_id           = sedai_group.test.id
  availability_mode  = "DATA_PILOT"
  optimization_mode  = "DATA_PILOT"
  sedai_sync_enabled = false

  kube_app_settings {
    horizontal_scaling_enabled      = true
    horizontal_scaling_min_replicas = 1
    horizontal_scaling_max_replicas = 10
    vertical_scaling_enabled        = true
    predictive_scaling_enabled      = false
  }
}
`, name)
}
