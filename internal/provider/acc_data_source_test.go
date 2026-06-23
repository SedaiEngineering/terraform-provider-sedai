package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDS(t *testing.T) {
	// DS-001: data.sedai_account lookup by name
	t.Run("DS-001", func(t *testing.T) {
		name := "ds-acct-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDataSourceAccountConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sedai_account.lookup", "id"),
						resource.TestCheckResourceAttr("data.sedai_account.lookup", "name", name),
					),
				},
			},
		})
	})

	// DS-002: data.sedai_account not found → error
	t.Run("DS-002", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccDataSourceAccountConfig_NotFound(),
					ExpectError: regexpMustCompile(`(not found|no account)`),
				},
			},
		})
	})

	// DS-003: data.sedai_account multiple matches → error (if backend supports it)
	t.Run("DS-003", func(t *testing.T) {
		prefix := "ds-multi-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccDataSourceAccountConfig_DuplicateName(prefix),
					ExpectError: regexpMustCompile(`(multiple|ambiguous|more than one)`),
				},
			},
		})
	})

	// DS-005: data.sedai_groups returns list of groups
	t.Run("DS-005", func(t *testing.T) {
		name := "ds-grps-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDataSourceGroupsConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sedai_groups.all", "groups.#"),
					),
				},
			},
		})
	})

	// DS-006: data.sedai_group lookup by name
	t.Run("DS-006", func(t *testing.T) {
		name := "ds-grp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDataSourceGroupConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sedai_group.lookup", "id"),
						resource.TestCheckResourceAttr("data.sedai_group.lookup", "name", name),
					),
				},
			},
		})
	})

	// DS-007: data.sedai_group not found → error
	t.Run("DS-007", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccDataSourceGroupConfig_NotFound(),
					ExpectError: regexpMustCompile(`(not found|no group)`),
				},
			},
		})
	})

	// DS-010: data.sedai_group_settings returns all blocks populated
	t.Run("DS-010", func(t *testing.T) {
		name := "ds-gset-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDataSourceGroupSettingsConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sedai_group_settings.lookup", "availability_mode"),
						resource.TestCheckResourceAttrSet("data.sedai_group_settings.lookup", "optimization_mode"),
					),
				},
			},
		})
	})

	// DS-011: data.sedai_account_settings all fields
	t.Run("DS-011", func(t *testing.T) {
		name := "ds-aset-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDataSourceAccountSettingsConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sedai_account_settings.lookup", "availability_mode"),
						resource.TestCheckResourceAttrSet("data.sedai_account_settings.lookup", "optimization_mode"),
					),
				},
			},
		})
	})
}

func testAccDataSourceAccountConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

data "sedai_account" "lookup" {
  name       = %[1]q
  depends_on = [sedai_account.test]
}
`, name)
}

func testAccDataSourceAccountConfig_NotFound() string {
	return `
data "sedai_account" "lookup" {
  name = "this-account-definitely-does-not-exist-zzzzzz"
}
`
}

func testAccDataSourceAccountConfig_DuplicateName(prefix string) string {
	return fmt.Sprintf(`
resource "sedai_account" "a1" {
  name             = "%[1]s-same"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::111111111111:role/SedaiRole"
  external_id      = "ext-1"
}

resource "sedai_account" "a2" {
  name             = "%[1]s-same"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::222222222222:role/SedaiRole"
  external_id      = "ext-2"
}

data "sedai_account" "lookup" {
  name       = "%[1]s-same"
  depends_on = [sedai_account.a1, sedai_account.a2]
}
`, prefix)
}

func testAccDataSourceGroupsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

data "sedai_groups" "all" {
  depends_on = [sedai_group.test]
}
`, name)
}

func testAccDataSourceGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

data "sedai_group" "lookup" {
  name       = %[1]q
  depends_on = [sedai_group.test]
}
`, name)
}

func testAccDataSourceGroupConfig_NotFound() string {
	return `
data "sedai_group" "lookup" {
  name = "this-group-definitely-does-not-exist-zzzzzz"
}
`
}

func testAccDataSourceGroupSettingsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_group" "test" {
  name = %[1]q
}

resource "sedai_group_settings" "test" {
  group_id          = sedai_group.test.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
}

data "sedai_group_settings" "lookup" {
  group_id   = sedai_group.test.id
  depends_on = [sedai_group_settings.test]
}
`, name)
}

func testAccDataSourceAccountSettingsConfig(name string) string {
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

data "sedai_account_settings" "lookup" {
  account_id = sedai_account.test.id
  depends_on = [sedai_account_settings.test]
}
`, name)
}
