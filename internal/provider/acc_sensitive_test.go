package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSECURITY verifies that sensitive credential fields are never stored in
// Terraform state (or are stored as empty strings), and do not appear in plan
// output — preventing secret leakage through state backends and plan logs.
//
// The Terraform Plugin Framework marks fields Sensitive: true which causes them
// to be redacted in plan display. The state-level check here verifies the value
// stored is empty (write-only semantics) not the actual credential.
func TestAccSECURITY(t *testing.T) {

	// SEC-001: AWS access_key is not stored in state.
	t.Run("SEC-001", func(t *testing.T) {
		name := "sec-001-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSKeys(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						// access_key is Sensitive — stored as empty after apply (write-only).
						resource.TestCheckResourceAttr("sedai_account.test", "access_key", ""),
					),
				},
			},
		})
	})

	// SEC-002: AWS secret_key is not stored in state.
	t.Run("SEC-002", func(t *testing.T) {
		name := "sec-002-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSKeys(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account.test", "secret_key", ""),
					),
				},
			},
		})
	})

	// SEC-003: IAM role external_id is not stored in state.
	t.Run("SEC-003", func(t *testing.T) {
		name := "sec-003-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSRole(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						// external_id is write-only — must be empty in state.
						resource.TestCheckResourceAttr("sedai_account.test", "external_id", ""),
					),
				},
			},
		})
	})

	// SEC-004: Azure client_secret is not stored in state.
	t.Run("SEC-004", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("AZURE_* env vars not set")
		}
		name := "sec-004-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecAzureConfig(name, tenantId, subscriptionId, clientId, clientSecret),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "client_secret", ""),
					),
				},
			},
		})
	})

	// SEC-005: GCP service_account_json is not stored in state.
	t.Run("SEC-005", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP_PROJECT_ID or GCP_SERVICE_ACCOUNT_JSON not set")
		}
		name := "sec-005-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecGCPConfig(name, projectId, saJson),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "service_account_json", ""),
					),
				},
			},
		})
	})

	// SEC-006: Datadog api_key is not stored in state.
	t.Run("SEC-006", func(t *testing.T) {
		name := "sec-006-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecDatadogConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_datadog_monitoring_provider.test", "id"),
						resource.TestCheckResourceAttr("sedai_datadog_monitoring_provider.test", "api_key", ""),
					),
				},
			},
		})
	})

	// SEC-007: Datadog application_key is not stored in state.
	t.Run("SEC-007", func(t *testing.T) {
		name := "sec-007-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecDatadogConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_datadog_monitoring_provider.test", "application_key", ""),
					),
				},
			},
		})
	})

	// SEC-008: New Relic api_key is not stored in state.
	t.Run("SEC-008", func(t *testing.T) {
		name := "sec-008-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecNewRelicConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_newrelic_monitoring_provider.test", "id"),
						resource.TestCheckResourceAttr("sedai_newrelic_monitoring_provider.test", "api_key", ""),
					),
				},
			},
		})
	})

	// SEC-009: Federated Prometheus client_secret is not stored in state.
	t.Run("SEC-009", func(t *testing.T) {
		name := "sec-009-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccSecFPClientCredsConfig(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_federated_prometheus_monitoring_provider.test", "id"),
						resource.TestCheckResourceAttr("sedai_federated_prometheus_monitoring_provider.test", "client_secret", ""),
					),
				},
			},
		})
	})

	// SEC-010: After import, sensitive fields remain empty — backend never returns them.
	t.Run("SEC-010", func(t *testing.T) {
		name := "sec-010-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AWSKeys(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
				},
				{
					ResourceName:      "sedai_account.test",
					ImportState:       true,
					ImportStateVerify: true,
					// Sensitive write-only fields are empty after import —
					// the backend doesn't return them and that is expected behaviour.
					ImportStateVerifyIgnore: []string{
						"role",
						"external_id",
						"access_key",
						"secret_key",
						"client_secret",
						"service_account_json",
					},
				},
				// After import, re-plan must show 0 changes (no spurious drift from empty sensitives).
				{
					Config:             testAccAccountConfig_AWSKeys(name),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})
}

// --- HCL config builders ---

func testAccSecAzureConfig(name, tenantId, subscriptionId, clientId, clientSecret string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AZURE"
  integration_type = "AGENTLESS"
  tenant_id        = %[2]q
  subscription_id  = %[3]q
  client_id        = %[4]q
  client_secret    = %[5]q
}
`, name, tenantId, subscriptionId, clientId, clientSecret)
}

func testAccSecGCPConfig(name, projectId, serviceAccountJson string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name                 = %[1]q
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = %[2]q
  service_account_json = %[3]q
}
`, name, projectId, serviceAccountJson)
}

func testAccSecDatadogConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_datadog_monitoring_provider" "test" {
  account_id      = sedai_account.test.id
  name            = %[1]q
  api_key         = "test-api-key-placeholder"
  application_key = "test-app-key-placeholder"
}
`, name)
}

func testAccSecNewRelicConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_newrelic_monitoring_provider" "test" {
  account_id = sedai_account.test.id
  name       = %[1]q
  api_key    = "test-nr-api-key-placeholder"
}
`, name)
}

func testAccSecFPClientCredsConfig(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "ROLE"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}

resource "sedai_federated_prometheus_monitoring_provider" "test" {
  account_id       = sedai_account.test.id
  name             = %[1]q
  endpoint         = "https://prometheus.example.com"
  integration_type = "CLIENT_CREDENTIALS"
  client_id        = "test-client-id"
  client_secret    = "test-client-secret-placeholder"
  token_endpoint   = "https://auth.example.com/token"
}
`, name)
}
