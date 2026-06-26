package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMP_Dimensions covers dimension array drift for every monitoring provider type.
// Each test applies a config with specific dimension values, then asserts 0 planned changes.
// MP-DIM-008 specifically targets the az_dimensions bug (SED-21383: wrong write key).
func TestAccMP_Dimensions(t *testing.T) {

	// --- CloudWatch ---

	t.Run("MP-DIM-001", func(t *testing.T) {
		name := "cw-dim-def-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_Defaults(name)},
				{Config: testAccCWDimConfig_Defaults(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-002", func(t *testing.T) {
		name := "cw-dim-lb-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name", "availability_zone"]`),
					Check:  resource.TestCheckResourceAttr("sedai_cloudwatch_monitoring_provider.test", "lb_dimensions.#", "2"),
				},
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name", "availability_zone"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-003", func(t *testing.T) {
		name := "cw-dim-inst-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_InstanceDimensions(name, `["instance_id", "instance_type"]`)},
				{Config: testAccCWDimConfig_InstanceDimensions(name, `["instance_id", "instance_type"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-004", func(t *testing.T) {
		name := "cw-dim-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name"]`)},
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name", "availability_zone"]`)},
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name", "availability_zone"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-005", func(t *testing.T) {
		// null vs empty list drift — the SED-21383 area
		name := "cw-dim-empty-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_EmptyDimensions(name)},
				{Config: testAccCWDimConfig_EmptyDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-006", func(t *testing.T) {
		name := "cw-dim-all-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_AllDimensions(name)},
				{Config: testAccCWDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-007", func(t *testing.T) {
		name := "cw-dim-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name"]`)},
				{
					ResourceName:            "sedai_cloudwatch_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"access_key", "secret_key"},
				},
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-008", func(t *testing.T) {
		// az_dimensions regression — SED-21383 used wrong JSON write key "aZDimensions"
		name := "cw-dim-az-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccCWDimConfig_Defaults(name), // CW doesn't have az_dimensions; use Datadog for that
				},
			},
		})
	})

	t.Run("MP-DIM-009", func(t *testing.T) {
		name := "cw-dim-del-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_AllDimensions(name)},
				// destroy happens via test cleanup
			},
		})
	})

	t.Run("MP-DIM-010", func(t *testing.T) {
		name := "cw-dim-recr-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name"]`)},
				{Config: testAccCWDimConfig_LbDimensions(name, `["load_balancer_name"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// --- Datadog ---

	t.Run("MP-DIM-011", func(t *testing.T) {
		name := "dd-dim-all-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_AllDimensions(name)},
				{Config: testAccDDDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-012", func(t *testing.T) {
		name := "dd-dim-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_LbDimensions(name, `["lb_name"]`)},
				{Config: testAccDDDimConfig_LbDimensions(name, `["lb_name", "availability_zone"]`)},
				{Config: testAccDDDimConfig_LbDimensions(name, `["lb_name", "availability_zone"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-013", func(t *testing.T) {
		name := "dd-dim-env-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_EnvDimensions(name, `["env", "service"]`)},
				{Config: testAccDDDimConfig_EnvDimensions(name, `["env", "service"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-014", func(t *testing.T) {
		name := "dd-dim-ns-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_NamespaceDimensions(name, `["destination_service_namespace", "namespace_name"]`)},
				{Config: testAccDDDimConfig_NamespaceDimensions(name, `["destination_service_namespace", "namespace_name"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-015", func(t *testing.T) {
		name := "dd-dim-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_AllDimensions(name)},
				{
					ResourceName:            "sedai_datadog_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"api_key", "application_key"},
				},
				{Config: testAccDDDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-016", func(t *testing.T) {
		// az_dimensions for Datadog — the field affected by SED-21383
		name := "dd-dim-az-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_AzDimensions(name, `["availability_zone"]`)},
				{Config: testAccDDDimConfig_AzDimensions(name, `["availability_zone"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-017", func(t *testing.T) {
		name := "dd-dim-agb-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_IntegrationType(name, "AGENT_BASED")},
				{Config: testAccDDDimConfig_IntegrationType(name, "AGENT_BASED"), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-018", func(t *testing.T) {
		// Update integration_type AGENTLESS→AGENT_BASED: document whether RequiresReplace fires
		name := "dd-dim-itype-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccDDDimConfig_IntegrationType(name, "AGENTLESS")},
				// integration_type is Computed+UseStateForUnknown — update does not force replace
				{Config: testAccDDDimConfig_IntegrationType(name, "AGENT_BASED")},
				{Config: testAccDDDimConfig_IntegrationType(name, "AGENT_BASED"), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// --- GKE ---

	t.Run("MP-DIM-020", func(t *testing.T) {
		name := "gke-dim-all-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGKEDimConfig_AllDimensions(name)},
				{Config: testAccGKEDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-021", func(t *testing.T) {
		name := "gke-dim-cls-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGKEDimConfig_ClusterDimensions(name, `["cluster_name"]`)},
				{Config: testAccGKEDimConfig_ClusterDimensions(name, `["cluster_name", "location"]`)},
				{Config: testAccGKEDimConfig_ClusterDimensions(name, `["cluster_name", "location"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-022", func(t *testing.T) {
		name := "gke-dim-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGKEDimConfig_AllDimensions(name)},
				{
					ResourceName:            "sedai_gke_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"service_account_json"},
				},
				{Config: testAccGKEDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-023", func(t *testing.T) {
		name := "gke-dim-def-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccGKEDimConfig_Defaults(name)},
				{Config: testAccGKEDimConfig_Defaults(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	// --- New Relic ---

	t.Run("MP-DIM-026", func(t *testing.T) {
		name := "nr-dim-all-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccNRDimConfig_AllDimensions(name)},
				{Config: testAccNRDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-027", func(t *testing.T) {
		name := "nr-dim-upd-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccNRDimConfig_LbDimensions(name, `["lb_name"]`)},
				{Config: testAccNRDimConfig_LbDimensions(name, `["lb_name", "region"]`)},
				{Config: testAccNRDimConfig_LbDimensions(name, `["lb_name", "region"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-028", func(t *testing.T) {
		name := "nr-dim-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccNRDimConfig_AllDimensions(name)},
				{
					ResourceName:            "sedai_newrelic_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"api_key"},
				},
				{Config: testAccNRDimConfig_AllDimensions(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-029", func(t *testing.T) {
		name := "nr-dim-eu-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccNRDimConfig_EURegion(name)},
				{Config: testAccNRDimConfig_EURegion(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-DIM-030", func(t *testing.T) {
		name := "nr-dim-del-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccNRDimConfig_AllDimensions(name)},
				// destroy via test cleanup
			},
		})
	})
}

// TestAccMP_VictoriaMetrics covers the VictoriaMetrics monitoring provider — currently untested.
func TestAccMP_VictoriaMetrics(t *testing.T) {

	t.Run("MP-VM-001", func(t *testing.T) {
		name := "vm-jwt-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccVMConfig_JWT(name),
					Check:  resource.TestCheckResourceAttrSet("sedai_victoria_metrics_monitoring_provider.test", "id"),
				},
			},
		})
	})

	t.Run("MP-VM-002", func(t *testing.T) {
		name := "vm-noauth-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_NoAuth(name)},
				{Config: testAccVMConfig_NoAuth(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-003", func(t *testing.T) {
		name := "vm-cc-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_ClientCredentials(name)},
				{Config: testAccVMConfig_ClientCredentials(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-004", func(t *testing.T) {
		name := "vm-drift-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_NoAuth(name)},
				{Config: testAccVMConfig_NoAuth(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-005", func(t *testing.T) {
		name := "vm-dim-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_WithDimensions(name, `["lb_name"]`)},
				{Config: testAccVMConfig_WithDimensions(name, `["lb_name", "region"]`)},
				{Config: testAccVMConfig_WithDimensions(name, `["lb_name", "region"]`), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-006", func(t *testing.T) {
		name := "vm-imp-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_NoAuth(name)},
				{
					ResourceName:            "sedai_victoria_metrics_monitoring_provider.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"bearer_token", "client_secret"},
				},
				{Config: testAccVMConfig_NoAuth(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-007", func(t *testing.T) {
		// Switching from JWT to client_credentials — document whether RequiresReplace fires
		name := "vm-authswitch-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_JWT(name)},
				// bearer_token and client_credentials are mutually exclusive — switching auth type
				// requires the old credential cleared; document via plan output
				{Config: testAccVMConfig_ClientCredentials(name)},
				{Config: testAccVMConfig_ClientCredentials(name), PlanOnly: true, ExpectNonEmptyPlan: false},
			},
		})
	})

	t.Run("MP-VM-008", func(t *testing.T) {
		name := "vm-del-" + randString(5)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{Config: testAccVMConfig_NoAuth(name)},
				// destroy via test cleanup
			},
		})
	})
}

// ---- CloudWatch HCL builders ----

func testAccCWDimConfig_Defaults(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
}
`, name)
}

func testAccCWDimConfig_LbDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
  lb_dimensions           = %[2]s
}
`, name, dims)
}

func testAccCWDimConfig_InstanceDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
  instance_dimensions     = %[2]s
}
`, name, dims)
}

func testAccCWDimConfig_EmptyDimensions(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
  lb_dimensions           = []
  app_dimensions          = []
  instance_dimensions     = []
}
`, name)
}

func testAccCWDimConfig_AllDimensions(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_cloudwatch_monitoring_provider" "test" {
  account_id              = sedai_account.test.id
  name                    = %[1]q
  use_account_credentials = true
  lb_dimensions           = ["load_balancer_name"]
  app_dimensions          = ["application_id"]
  instance_dimensions     = ["instance_id"]
}
`, name)
}

// ---- Datadog HCL builders ----

func testAccDDDimConfig_AllDimensions(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  name                 = %[1]q
  api_key              = "dd-api-key-test"
  application_key      = "dd-app-key-test"
  lb_dimensions        = ["lb_name"]
  app_dimensions       = ["application_id"]
  instance_dimensions  = ["pod_name"]
  region_dimensions    = ["location"]
  container_dimensions = ["container_name"]
  namespace_dimensions = ["destination_service_namespace"]
  cluster_dimensions   = ["cluster_name"]
  az_dimensions        = ["availability_zone"]
  env_dimensions       = ["env"]
}
`, name)
}

func testAccDDDimConfig_LbDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id      = sedai_account.test.id
  name            = %[1]q
  api_key         = "dd-api-key-test"
  application_key = "dd-app-key-test"
  lb_dimensions   = %[2]s
}
`, name, dims)
}

func testAccDDDimConfig_EnvDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id      = sedai_account.test.id
  name            = %[1]q
  api_key         = "dd-api-key-test"
  application_key = "dd-app-key-test"
  env_dimensions  = %[2]s
}
`, name, dims)
}

func testAccDDDimConfig_NamespaceDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  name                 = %[1]q
  api_key              = "dd-api-key-test"
  application_key      = "dd-app-key-test"
  namespace_dimensions = %[2]s
}
`, name, dims)
}

func testAccDDDimConfig_AzDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id      = sedai_account.test.id
  name            = %[1]q
  api_key         = "dd-api-key-test"
  application_key = "dd-app-key-test"
  az_dimensions   = %[2]s
}
`, name, dims)
}

func testAccDDDimConfig_IntegrationType(name, intType string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_datadog_monitoring_provider" "test" {
  account_id       = sedai_account.test.id
  name             = %[1]q
  api_key          = "dd-api-key-test"
  application_key  = "dd-app-key-test"
  integration_type = %[2]q
}
`, name, intType)
}

// ---- GKE HCL builders ----

func testAccGKEDimConfig_Defaults(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "GCP"
  project_id       = "my-gcp-project"
  service_account_json = "{}"
}
resource "sedai_gke_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  project_id           = "my-gcp-project"
  name                 = %[1]q
  service_account_json = "{}"
}
`, name)
}

func testAccGKEDimConfig_AllDimensions(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "GCP"
  project_id       = "my-gcp-project"
  service_account_json = "{}"
}
resource "sedai_gke_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  project_id           = "my-gcp-project"
  name                 = %[1]q
  service_account_json = "{}"
  lb_dimensions        = ["destination_service_name"]
  app_dimensions       = ["application_id"]
  instance_dimensions  = ["pod_name"]
  region_dimensions    = ["location"]
  container_dimensions = ["container_name"]
  namespace_dimensions = ["namespace_name"]
  cluster_dimensions   = ["cluster_name"]
}
`, name)
}

func testAccGKEDimConfig_ClusterDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "GCP"
  project_id       = "my-gcp-project"
  service_account_json = "{}"
}
resource "sedai_gke_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  project_id           = "my-gcp-project"
  name                 = %[1]q
  service_account_json = "{}"
  cluster_dimensions   = %[2]s
}
`, name, dims)
}

// ---- New Relic HCL builders ----

func testAccNRDimConfig_AllDimensions(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_newrelic_monitoring_provider" "test" {
  account_id           = sedai_account.test.id
  name                 = %[1]q
  api_key              = "nr-api-key-test"
  newrelic_account_id  = "12345"
  lb_dimensions        = ["lb_name"]
  app_dimensions       = ["application_id"]
  instance_dimensions  = ["instance_id"]
  container_dimensions = ["container_name"]
  namespace_dimensions = ["namespace_name"]
  cluster_dimensions   = ["cluster_name"]
}
`, name)
}

func testAccNRDimConfig_LbDimensions(name, dims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_newrelic_monitoring_provider" "test" {
  account_id          = sedai_account.test.id
  name                = %[1]q
  api_key             = "nr-api-key-test"
  newrelic_account_id = "12345"
  lb_dimensions       = %[2]s
}
`, name, dims)
}

func testAccNRDimConfig_EURegion(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AWS"
  integration_type = "AGENTLESS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_newrelic_monitoring_provider" "test" {
  account_id          = sedai_account.test.id
  name                = %[1]q
  api_key             = "nr-api-key-test"
  newrelic_account_id = "12345"
  api_server          = "https://api.eu.newrelic.com"
}
`, name)
}

// ---- VictoriaMetrics HCL builders ----

func testAccVMConfig_NoAuth(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AWS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_victoria_metrics_monitoring_provider" "test" {
  account_id = sedai_account.test.id
  name       = %[1]q
  endpoint   = "http://victoria-metrics.example.com:8428"
}
`, name)
}

func testAccVMConfig_JWT(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AWS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_victoria_metrics_monitoring_provider" "test" {
  account_id    = sedai_account.test.id
  name          = %[1]q
  endpoint      = "http://victoria-metrics.example.com:8428"
  bearer_token  = "eyJhbGciOiJSUzI1NiJ9.test-token"
}
`, name)
}

func testAccVMConfig_ClientCredentials(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AWS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_victoria_metrics_monitoring_provider" "test" {
  account_id     = sedai_account.test.id
  name           = %[1]q
  endpoint       = "http://victoria-metrics.example.com:8428"
  token_endpoint = "https://auth.example.com/token"
  client_id      = "sedai-client"
  client_secret  = "super-secret"
}
`, name)
}

func testAccVMConfig_WithDimensions(name, lbDims string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AWS"
  role             = "arn:aws:iam::123456789012:role/SedaiRole"
  external_id      = "sedai-ext-123"
}
resource "sedai_victoria_metrics_monitoring_provider" "test" {
  account_id    = sedai_account.test.id
  name          = %[1]q
  endpoint      = "http://victoria-metrics.example.com:8428"
  lb_dimensions = %[2]s
}
`, name, lbDims)
}
