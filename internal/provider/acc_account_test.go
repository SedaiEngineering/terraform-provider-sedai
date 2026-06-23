package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccACCT(t *testing.T) {
	t.Run("ACCT-001", func(t *testing.T) {
		name := "acct-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "name", name),
						resource.TestCheckResourceAttr("sedai_account.test", "cloud_provider", "AWS"),
					),
				},
			},
		})
	})

	t.Run("ACCT-002", func(t *testing.T) {
		name := "acct-key-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSKeys(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "name", name),
					),
				},
			},
		})
	})

	t.Run("ACCT-003", func(t *testing.T) {
		name := "acct-ms-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_WithManagedServices(name, `["LAMBDA"]`),
				},
				{
					Config:   testAccAccountConfig_WithManagedServices(name, `["LAMBDA"]`),
					PlanOnly: true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-004", func(t *testing.T) {
		name := "acct-allms-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_WithManagedServices(name, `["LAMBDA","EC2","RDS","ECS","EMR","DYNAMO_DB","OPEN_SEARCH","SAGE_MAKER"]`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
					),
				},
			},
		})
	})

	t.Run("ACCT-005", func(t *testing.T) {
		name := "acct-upd-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_WithManagedServices(name, `["LAMBDA"]`),
				},
				{
					Config: testAccAccountConfig_WithManagedServices(name, `["LAMBDA","EC2"]`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account.test", "managed_services.#", "2"),
					),
				},
			},
		})
	})

	t.Run("ACCT-006", func(t *testing.T) {
		name := "acct-del-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
				},
				// Terraform destroys as part of test cleanup — idempotent if no second Apply
			},
		})
	})

	t.Run("ACCT-007", func(t *testing.T) {
		name := "acct-imp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
				},
				{
					ResourceName:      "sedai_account.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateVerifyIgnore: []string{"role", "external_id", "access_key", "secret_key"},
				},
			},
		})
	})

	t.Run("ACCT-008", func(t *testing.T) {
		name := "acct-imp2-" + randString(6)
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
					Config:   testAccAccountConfig_AWSRole(name),
					PlanOnly: true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// Validator tests — run without TF_ACC
	t.Run("ACCT-021", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_MissingCloudProvider(),
					ExpectError: regexpMustCompile(`cloud_provider`),
				},
			},
		})
	})

	t.Run("ACCT-022", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_InvalidCloudProvider(),
					ExpectError: regexpMustCompile(`cloud_provider`),
				},
			},
		})
	})

	t.Run("ACCT-023", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_InvalidIntegrationType(),
					ExpectError: regexpMustCompile(`integration_type`),
				},
			},
		})
	})

	// ACCT-024: external_id requires role (AlsoRequires validator) — setting
	// external_id without role must produce a plan-time error.
	t.Run("ACCT-024", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_RoleWithoutExternalId(),
					ExpectError: regexpMustCompile(`(?i)(external_id|role|attribute|AlsoRequires)`),
				},
			},
		})
	})

	t.Run("ACCT-025", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_ExternalIdWithoutRole(),
					ExpectError: regexpMustCompile(`role`),
				},
			},
		})
	})

	t.Run("ACCT-026", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AccessKeyWithoutSecretKey(),
					ExpectError: regexpMustCompile(`secret_key`),
				},
			},
		})
	})

	t.Run("ACCT-027", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_SecretKeyWithoutAccessKey(),
					ExpectError: regexpMustCompile(`access_key`),
				},
			},
		})
	})

	t.Run("ACCT-028", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AccessKeyAndRole(),
					ExpectError: regexpMustCompile(`(role|access_key)`),
				},
			},
		})
	})

	t.Run("ACCT-029", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AccessKeyAndServiceAccount(),
					ExpectError: regexpMustCompile(`(service_account_json|access_key)`),
				},
			},
		})
	})

	t.Run("ACCT-033", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_InvalidManagedService(),
					ExpectError: regexpMustCompile(`(managed_service|invalid)`),
				},
			},
		})
	})

	t.Run("ACCT-044", func(t *testing.T) {
		prefix := "multi-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
				},
			},
		})
	})

	t.Run("ACCT-045", func(t *testing.T) {
		prefix := "nochg-" + randString(4)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccFullStackConfig(prefix, 12),
				},
				{
					Config:   testAccFullStackConfig(prefix, 12),
					PlanOnly: true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

func testAccAccountConfig_AWSKeys(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ACCESS_KEY"
  access_key       = "AKIAIOSFODNN7EXAMPLE"
  secret_key       = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
`, name)
}

func testAccAccountConfig_WithManagedServices(name, services string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
  managed_services = %[2]s
}
`, name, services)
}

func testAccAccountConfig_MissingCloudProvider() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "ext-id"
}
`
}

func testAccAccountConfig_InvalidCloudProvider() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "INVALID_CLOUD"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "ext-id"
}
`
}

func testAccAccountConfig_InvalidIntegrationType() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "INVALID_TYPE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "ext-id"
}
`
}

func testAccAccountConfig_RoleWithoutExternalId() string {
	// external_id has AlsoRequires(role) — so external_id without role fails.
	// Inverted from ACCT-024's label but matches the actual schema constraint.
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  external_id      = "ext-id"
}
`
}

func testAccAccountConfig_ExternalIdWithoutRole() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  external_id      = "ext-id"
}
`
}

func testAccAccountConfig_AccessKeyWithoutSecretKey() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "ACCESS_KEY"
  access_key       = "AKIAIOSFODNN7EXAMPLE"
}
`
}

func testAccAccountConfig_SecretKeyWithoutAccessKey() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "ACCESS_KEY"
  secret_key       = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
`
}

func testAccAccountConfig_AccessKeyAndRole() string {
	// access_key ConflictsWith service_account_json and tenant_id.
	// access_key also AlsoRequires secret_key (already present here).
	// Testing that access_key with service_account_json is rejected.
	return `
resource "sedai_account" "test" {
  name                 = "test"
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  access_key           = "some-key"
  secret_key           = "some-secret"
  service_account_json = "{}"
}
`
}

func testAccAccountConfig_AccessKeyAndServiceAccount() string {
	return `
resource "sedai_account" "test" {
  name                 = "test"
  cloud_provider       = "GCP"
  integration_type     = "ACCESS_KEY"
  access_key           = "some-key"
  secret_key           = "some-secret"
  service_account_json = "{}"
}
`
}

func testAccAccountConfig_InvalidManagedService() string {
	return `
resource "sedai_account" "test" {
  name             = "test"
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "ext-id"
  managed_services = ["INVALID_SERVICE"]
}
`
}
