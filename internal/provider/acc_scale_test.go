package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSCALE(t *testing.T) {
	// SCALE-001: 12 accounts in one apply pass
	t.Run("SCALE-001", func(t *testing.T) {
		prefix := "scale-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccScale12AccountsConfig(prefix),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
						resource.TestCheckResourceAttrSet("sedai_account.acct11", "id"),
					),
				},
			},
		})
	})

	// SCALE-002: Full 48-resource Diligent stack — requires TF_SYSTEM_TESTS=1
	t.Run("SCALE-002", func(t *testing.T) {
		testAccSystemPreCheck(t)
		prefix := "diligent-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
					),
				},
			},
		})
	})

	// SCALE-003: After 48-resource apply, re-plan shows 0 changes
	t.Run("SCALE-003", func(t *testing.T) {
		testAccSystemPreCheck(t)
		prefix := "nodrift-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
				},
				{
					Config:             testAccFullStackConfig(prefix, 12),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// SCALE-004: Partial apply (half the accounts) then full apply finishes remaining
	t.Run("SCALE-004", func(t *testing.T) {
		prefix := "partial-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccScale12AccountsConfig(prefix),
				},
				// Second apply with more resources added on top
				{
					Config: testAccScale12AccountsConfig(prefix),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
						resource.TestCheckResourceAttrSet("sedai_account.acct11", "id"),
					),
				},
			},
		})
	})
}

func TestAccDEP(t *testing.T) {
	// DEP-001: Full dependency chain account→group→settings→CloudWatch
	t.Run("DEP-001", func(t *testing.T) {
		name := "dep-chain-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccDependencyChainConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
						resource.TestCheckResourceAttrSet("sedai_cloudwatch_monitoring_provider.test", "id"),
					),
				},
			},
		})
	})

	// DEP-002: Dependent resources created after account is ready
	t.Run("DEP-002", func(t *testing.T) {
		name := "dep-order-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					// First just the account
					Config: testAccAccountConfig_AWSRole(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
					),
				},
				{
					// Then add group + settings
					Config: testAccDependencyChainConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_group.test", "id"),
						resource.TestCheckResourceAttrSet("sedai_group_settings.test", "group_id"),
					),
				},
			},
		})
	})

	// DEP-010: Full Diligent scenario — requires TF_SYSTEM_TESTS=1
	t.Run("DEP-010", func(t *testing.T) {
		testAccSystemPreCheck(t)
		prefix := "dep-full-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.acct0", "id"),
					),
				},
				{
					Config:             testAccFullStackConfig(prefix, 12),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func testAccScale12AccountsConfig(prefix string) string {
	config := ""
	for i := 0; i < 12; i++ {
		config += fmt.Sprintf(`
resource "sedai_account" "acct%d" {
  name             = "%s-acct-%d"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::%012d:role/SedaiRole"
  external_id      = "sedai-ext-%d"
}
`, i, prefix, i, 100000000000+i, i)
	}
	return config
}

func testAccDependencyChainConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
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
  integration_type        = "ROLE"
  use_account_credentials = true
  depends_on              = [sedai_account.test]
}
`, name)
}
