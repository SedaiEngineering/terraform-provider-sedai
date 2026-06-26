package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGRP covers all sedai_group acceptance and validator tests.
// Subtests are named GRP-NNN to match the testing manifest IDs.
//
// Critical P0 drift tests: GRP-026 through GRP-031 verify that resource
// counts (lambda_count, ec2_count, etc.) stored on the backend do NOT
// appear as spurious plan changes — the Diligent bug reported in jun18.md.
func TestAccGRP(t *testing.T) {

	// ── Happy-path CRUD ───────────────────────────────────────────────────

	t.Run("GRP-001", func(t *testing.T) {
		name := "grp-basic-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttr("sedai_group.test", "name", name),
					),
				},
			},
		})
	})

	t.Run("GRP-002", func(t *testing.T) {
		name := "grp-defaults-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "enabled", "true"),
						resource.TestCheckResourceAttr("sedai_group.test", "auto_refresh", "true"),
					),
				},
				{
					Config:             testAccGroupConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GRP-003", func(t *testing.T) {
		name := "grp-disabled-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithEnabled(name, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "enabled", "false"),
					),
				},
			},
		})
	})

	t.Run("GRP-004", func(t *testing.T) {
		name := "grp-norefresh-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithAutoRefresh(name, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "auto_refresh", "false"),
					),
				},
			},
		})
	})

	t.Run("GRP-005", func(t *testing.T) {
		name := "grp-restype-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithResourceTypes(name, []string{"AWS_LAMBDA", "AWS_EC2"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "resource_types.#", "2"),
					),
				},
			},
		})
	})

	t.Run("GRP-006", func(t *testing.T) {
		name := "grp-tags-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithTags(name, "env", "production"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "tags.0.key", "env"),
						resource.TestCheckResourceAttr("sedai_group.test", "tags.0.exact.0", "production"),
					),
				},
			},
		})
	})

	t.Run("GRP-007", func(t *testing.T) {
		name := "grp-tags-regex-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithTagsRegex(name, "Name", "platform-*"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "tags.0.key", "Name"),
						resource.TestCheckResourceAttr("sedai_group.test", "tags.0.regex.0", "platform-*"),
					),
				},
			},
		})
	})

	t.Run("GRP-008", func(t *testing.T) {
		name := "grp-regions-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithRegions(name, []string{"us-east-1", "us-west-2"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "regions.#", "2"),
					),
				},
			},
		})
	})

	t.Run("GRP-009", func(t *testing.T) {
		name := "grp-ns-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithNamespaces(name, []string{"default", "kube-system"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "namespaces.#", "2"),
					),
				},
			},
		})
	})

	// ── Update tests ──────────────────────────────────────────────────────

	t.Run("GRP-011", func(t *testing.T) {
		name := "grp-upd-name-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
				},
				{
					Config: testAccGroupConfig(name + "-v2"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "name", name+"-v2"),
					),
				},
			},
		})
	})

	t.Run("GRP-012", func(t *testing.T) {
		name := "grp-toggle-en-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithEnabled(name, true),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "enabled", "true"),
					),
				},
				{
					Config: testAccGRPConfigWithEnabled(name, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "enabled", "false"),
					),
				},
				{
					Config:             testAccGRPConfigWithEnabled(name, false),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GRP-013", func(t *testing.T) {
		name := "grp-add-restype-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithResourceTypes(name, []string{"AWS_LAMBDA"}),
				},
				{
					Config: testAccGRPConfigWithResourceTypes(name, []string{"AWS_LAMBDA", "AWS_EC2"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "resource_types.#", "2"),
					),
				},
				{
					Config:             testAccGRPConfigWithResourceTypes(name, []string{"AWS_LAMBDA", "AWS_EC2"}),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("GRP-014", func(t *testing.T) {
		name := "grp-upd-tags-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithTags(name, "env", "staging"),
				},
				{
					Config: testAccGRPConfigWithTags(name, "env", "production"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "tags.0.exact.0", "production"),
					),
				},
			},
		})
	})

	// ── Hierarchy tests ───────────────────────────────────────────────────

	t.Run("GRP-015", func(t *testing.T) {
		parent := "grp-parent-" + randString(5)
		child := "grp-child-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithParent(parent, child),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "parent_group_id"),
					),
				},
				{
					Config:             testAccGRPConfigWithParent(parent, child),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── Import tests ──────────────────────────────────────────────────────

	t.Run("GRP-021", func(t *testing.T) {
		name := "grp-import-" + randString(6)
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

	t.Run("GRP-022", func(t *testing.T) {
		name := "grp-import-nodrift-" + randString(5)
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
					ImportStateVerify: false,
				},
				{
					Config:             testAccGroupConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── P0 resource count drift tests (Diligent jun18.md) ────────────────
	//
	// The backend stores discovery counts (lambda_count, ec2_count, etc.) as
	// mutable fields. When these were tracked in state they caused spurious
	// diffs on every re-plan because Sedai increments them as it discovers
	// resources. The fix: do NOT store count fields in Terraform state.
	// Tests GRP-026 to GRP-031 guard this regression.

	// GRP-026: Apply a basic group, then re-plan. The plan must be empty even
	// though the backend has incremented resource-discovery counts since apply.
	t.Run("GRP-026", func(t *testing.T) {
		name := "grp-count-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
					),
				},
				{
					Config:             testAccGroupConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GRP-027: Group with resource_types — no drift after multiple re-plans.
	// resource_types was another field whose normalization could produce drift.
	t.Run("GRP-027", func(t *testing.T) {
		name := "grp-restype-drift-" + randString(5)
		cfg := testAccGRPConfigWithResourceTypes(name, []string{"AWS_LAMBDA", "AWS_EC2"})
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// GRP-028: Group with tags — tags must not drift.
	t.Run("GRP-028", func(t *testing.T) {
		name := "grp-tags-drift-" + randString(5)
		cfg := testAccGRPConfigWithTags(name, "env", "production")
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// GRP-029: Group with regions — no drift.
	t.Run("GRP-029", func(t *testing.T) {
		name := "grp-region-drift-" + randString(5)
		cfg := testAccGRPConfigWithRegions(name, []string{"us-east-1", "eu-west-1"})
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: cfg},
				{Config: cfg, PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// GRP-030: auto_refresh defaults to true — must not drift when omitted.
	// Diligent saw "~ auto_refresh = true -> null" because a failed apply left
	// state inconsistent and the next plan tried to "fix" it.
	t.Run("GRP-030", func(t *testing.T) {
		name := "grp-autorefresh-drift-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "auto_refresh", "true"),
					),
				},
				{
					Config:             testAccGroupConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GRP-031: enabled defaults to true — must not drift when omitted.
	t.Run("GRP-031", func(t *testing.T) {
		name := "grp-enabled-drift-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "enabled", "true"),
					),
				},
				{
					Config:             testAccGroupConfig(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── KUBERNETES_DAEMONSET normalization ────────────────────────────────

	// GRP-034: The HCL accepts `KUBERNETES_DAEMONSET` (correctly spelled) while
	// the backend stores the misspelled `KUBERNETES_DEAMONSET`. The provider
	// must translate transparently — the user sees only the correct spelling
	// and plans must be drift-free.
	t.Run("GRP-034", func(t *testing.T) {
		name := "grp-kube-ds-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGRPConfigWithResourceTypes(name, []string{"KUBERNETES_DAEMONSET"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "resource_types.0", "KUBERNETES_DAEMONSET"),
					),
				},
				{
					Config:             testAccGRPConfigWithResourceTypes(name, []string{"KUBERNETES_DAEMONSET"}),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// GRP-035: KUBERNETES_DAEMONSET alongside other types — both must be
	// idempotent after apply.
	t.Run("GRP-035", func(t *testing.T) {
		name := "grp-kube-ds-multi-" + randString(4)
		cfg := testAccGRPConfigWithResourceTypes(name, []string{"KUBERNETES_DEPLOYMENT", "KUBERNETES_DAEMONSET", "KUBERNETES_STATEFULSET"})
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: cfg,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group.test", "resource_types.#", "3"),
					),
				},
				{
					Config:             cfg,
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ── Validator tests (run without TF_ACC) ──────────────────────────────

	// GRP-040: Invalid resource_type value fires groupResourceTypeListValidator.
	t.Run("GRP-040", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGRPInvalid_BadResourceType(),
					ExpectError: regexpMustCompile(`(?i)(value must be one of|INVALID_TYPE|resource_types)`),
				},
			},
		})
	})

	// GRP-041: Empty name should fail with a required attribute error.
	t.Run("GRP-041", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGRPInvalid_EmptyName(),
					ExpectError: regexpMustCompile(`(?i)(name|argument|required)`),
				},
			},
		})
	})
}

// ── GRP config builders ───────────────────────────────────────────────────────

func testAccGRPConfigWithEnabled(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name    = %q
  enabled = %v
}
`, name, enabled)
}

func testAccGRPConfigWithAutoRefresh(name string, ar bool) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name         = %q
  auto_refresh = %v
}
`, name, ar)
}

func testAccGRPConfigWithResourceTypes(name string, types []string) string {
	quoted := make([]string, len(types))
	for i, t := range types {
		quoted[i] = fmt.Sprintf("%q", t)
	}
	typesStr := ""
	for i, q := range quoted {
		if i > 0 {
			typesStr += ", "
		}
		typesStr += q
	}
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name           = %q
  resource_types = [%s]
}
`, name, typesStr)
}

func testAccGRPConfigWithTags(name, key, exactVal string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
  tags {
    key   = %q
    exact = [%q]
  }
}
`, name, key, exactVal)
}

func testAccGRPConfigWithTagsRegex(name, key, regexVal string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %q
  tags {
    key   = %q
    regex = [%q]
  }
}
`, name, key, regexVal)
}

func testAccGRPConfigWithRegions(name string, regions []string) string {
	regStr := ""
	for i, r := range regions {
		if i > 0 {
			regStr += ", "
		}
		regStr += fmt.Sprintf("%q", r)
	}
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name    = %q
  regions = [%s]
}
`, name, regStr)
}

func testAccGRPConfigWithNamespaces(name string, namespaces []string) string {
	nsStr := ""
	for i, ns := range namespaces {
		if i > 0 {
			nsStr += ", "
		}
		nsStr += fmt.Sprintf("%q", ns)
	}
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name       = %q
  namespaces = [%s]
}
`, name, nsStr)
}

func testAccGRPConfigWithParent(parentName, childName string) string {
	return fmt.Sprintf(`
resource "sedai_group" "parent" {
  name = %q
}

resource "sedai_group" "test" {
  name            = %q
  parent_group_id = sedai_group.parent.id
}
`, parentName, childName)
}

// ── Invalid configs ───────────────────────────────────────────────────────────

func testAccGRPInvalid_BadResourceType() string {
	return `
resource "sedai_group" "test" {
  name           = "validator-test"
  resource_types = ["INVALID_TYPE"]
}
`
}

func testAccGRPInvalid_EmptyName() string {
	return `
resource "sedai_group" "test" {
  name = ""
}
`
}
