package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPRI(t *testing.T) {
	t.Run("PRI-001", func(t *testing.T) {
		name := "pri-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Single(name, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_priority.test", "id"),
						resource.TestCheckResourceAttr("sedai_group_priority.test", "group_priorities.#", "1"),
						resource.TestCheckResourceAttr("sedai_group_priority.test", "group_priorities.0.priority", "1"),
					),
				},
			},
		})
	})

	t.Run("PRI-002", func(t *testing.T) {
		prefix := "pri-multi-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Multiple(prefix, []int{1, 2, 3}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_priority.test", "group_priorities.#", "3"),
					),
				},
			},
		})
	})

	t.Run("PRI-003", func(t *testing.T) {
		prefix := "pri-hi-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Multiple(prefix, []int{1, 2, 3}),
					Check: resource.ComposeTestCheckFunc(
						// Priority 1 = highest priority on backend (0-based 0)
						resource.TestCheckResourceAttrSet("sedai_group_priority.test", "group_priorities.0.group_id"),
					),
				},
			},
		})
	})

	t.Run("PRI-004", func(t *testing.T) {
		prefix := "pri-reorder-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Multiple(prefix, []int{1, 2, 3}),
				},
				{
					Config: testAccGroupPriorityConfig_Multiple(prefix, []int{3, 1, 2}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_group_priority.test", "group_priorities.#", "3"),
					),
				},
			},
		})
	})

	// PRI-005: Priority=0 is invalid (1-based)
	t.Run("PRI-005", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccGroupPriorityConfig_InvalidZero(),
					ExpectError: regexpMustCompile(`priority`),
				},
			},
		})
	})

	// PRI-006: Priority=1 is the minimum valid value
	t.Run("PRI-006", func(t *testing.T) {
		name := "pri-valid-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Single(name, 1),
					Check:  resource.TestCheckResourceAttr("sedai_group_priority.test", "group_priorities.0.priority", "1"),
				},
			},
		})
	})

	// PRI-007: Import by composite ID (comma-separated group IDs)
	t.Run("PRI-007", func(t *testing.T) {
		prefix := "pri-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Multiple(prefix, []int{1, 2}),
				},
				{
					ResourceName:      "sedai_group_priority.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	// PRI-008: Destroy is a no-op (no API call expected to fail)
	t.Run("PRI-008", func(t *testing.T) {
		name := "pri-noop-del-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Single(name, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group_priority.test", "id"),
					),
				},
				// Terraform will call destroy as part of cleanup — should succeed without API errors
			},
		})
	})

	// PRI-012: Group referenced no longer exists → resource removed from state gracefully
	t.Run("PRI-012", func(t *testing.T) {
		name := "pri-gone-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccGroupPriorityConfig_Single(name, 1),
				},
				{
					// Config with no group_priority resource — simulates removal
					Config: testAccGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckNoResourceAttr("sedai_group_priority.test", "id"),
					),
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func testAccGroupPriorityConfig_Single(name string, priority int) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_priority" "test" {
  group_priorities {
    group_id = sedai_group.test.id
    priority = %[2]d
  }
}
`, name, priority)
}

func testAccGroupPriorityConfig_Multiple(prefix string, priorities []int) string {
	blocks := ""
	for i, p := range priorities {
		blocks += fmt.Sprintf(`
  group_priorities {
    group_id = sedai_group.g%d.id
    priority = %d
  }`, i, p)
	}

	groups := ""
	for i := range priorities {
		groups += fmt.Sprintf(`
resource "sedai_group" "g%d" {
  name = "%s-%d"
}
`, i, prefix, i)
	}

	return groups + `
resource "sedai_group_priority" "test" {` + blocks + `
}
`
}

func testAccGroupPriorityConfig_InvalidZero() string {
	return `
resource "sedai_group_priority" "test" {
  group_priorities {
    group_id = "some-group-id"
    priority = 0
  }
}
`
}
