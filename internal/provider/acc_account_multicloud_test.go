package provider_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// Azure
// ---------------------------------------------------------------------------

func TestAccACCT_Azure(t *testing.T) {
	t.Run("ACCT-050", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("AZURE_TENANT_ID, AZURE_SUBSCRIPTION_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET required")
		}
		name := "az-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "name", name),
						resource.TestCheckResourceAttr("sedai_account.test", "cloud_provider", "AZURE"),
					),
				},
			},
		})
	})

	t.Run("ACCT-051", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
				},
				{
					Config:             testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-052", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-replace-" + randString(6)
		altSub := subscriptionId + "-alt"
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
				},
				{
					Config: testAccAccountConfig_Azure(name, tenantId, altSub, clientId, clientSecret),
					// subscription_id change triggers RequiresReplace → new resource
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
					),
				},
			},
		})
	})

	t.Run("ACCT-053", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-imp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"client_secret"},
				},
			},
		})
	})

	t.Run("ACCT-054", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-imp2-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"client_secret"},
				},
				{
					Config:             testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-055", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-ms-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_AzureWithManagedServices(name, tenantId, subscriptionId, clientId, clientSecret, []string{"AZURE_VM", "AZURE_SQL"}),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
					),
				},
				{
					Config:             testAccAccountConfig_AzureWithManagedServices(name, tenantId, subscriptionId, clientId, clientSecret, []string{"AZURE_VM", "AZURE_SQL"}),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ACCT-056 to ACCT-058: Azure field validation is enforced in Create (not at
	// plan time), so these require TF_ACC=1 and a real backend.
	t.Run("ACCT-056", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AzureMissingField("missing-tenant", "", "sub-123", "client-123", "secret-123"),
					ExpectError: regexpMustCompile(`(?i)(tenant_id|Invalid|Error)`),
				},
			},
		})
	})

	t.Run("ACCT-057", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AzureMissingField("missing-sub", "tenant-123", "", "client-123", "secret-123"),
					ExpectError: regexpMustCompile(`(?i)(subscription_id|Invalid|Error)`),
				},
			},
		})
	})

	t.Run("ACCT-058", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_AzureMissingField("missing-client", "tenant-123", "sub-123", "", "secret-123"),
					ExpectError: regexpMustCompile(`(?i)(client_id|Invalid|Error)`),
				},
			},
		})
	})

	t.Run("ACCT-059", func(t *testing.T) {
		tenantId := os.Getenv("AZURE_TENANT_ID")
		subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if tenantId == "" || subscriptionId == "" || clientId == "" || clientSecret == "" {
			t.Skip("Azure credentials not set")
		}
		name := "az-del-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret),
				},
				// test cleanup destroys the resource; idempotent if backend succeeds
			},
		})
	})
}

// ---------------------------------------------------------------------------
// GCP
// ---------------------------------------------------------------------------

func TestAccACCT_GCP(t *testing.T) {
	t.Run("ACCT-060", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP_PROJECT_ID and GCP_SERVICE_ACCOUNT_JSON required")
		}
		name := "gcp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCP(name, projectId, saJson),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cloud_provider", "GCP"),
					),
				},
			},
		})
	})

	t.Run("ACCT-061", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCP(name, projectId, saJson),
				},
				{
					Config:             testAccAccountConfig_GCP(name, projectId, saJson),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-062", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-ms-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCPWithManagedServices(name, projectId, saJson, `["GCP_CLOUD_SQL"]`),
				},
				{
					Config: testAccAccountConfig_GCPWithManagedServices(name, projectId, saJson, `["GCP_CLOUD_SQL","GCP_CLOUD_RUN"]`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account.test", "user_selected_managed_services.#", "2"),
					),
				},
			},
		})
	})

	t.Run("ACCT-063", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-imp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCP(name, projectId, saJson),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"service_account_json"},
				},
			},
		})
	})

	t.Run("ACCT-064", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-imp2-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCP(name, projectId, saJson),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"service_account_json"},
				},
				{
					Config:             testAccAccountConfig_GCP(name, projectId, saJson),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-065", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-ms2-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCPWithManagedServices(name, projectId, saJson, `["GCP_CLOUD_SQL","GCP_CLOUD_RUN"]`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "user_selected_managed_services.#", "2"),
					),
				},
				{
					Config:             testAccAccountConfig_GCPWithManagedServices(name, projectId, saJson, `["GCP_CLOUD_SQL","GCP_CLOUD_RUN"]`),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	// ACCT-066/067: GCP field validation is enforced in Create (not plan time).
	// Requires TF_ACC=1 and a real backend.
	t.Run("ACCT-066", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_GCPMissingProjectId("missing-pid", "fake-sa-json"),
					ExpectError: regexpMustCompile(`(?i)(project_id|Invalid|Error)`),
				},
			},
		})
	})

	t.Run("ACCT-067", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_GCPMissingServiceAccountJson("missing-sa", "proj-123"),
					ExpectError: regexpMustCompile(`(?i)(service_account_json|Invalid|Error)`),
				},
			},
		})
	})

	t.Run("ACCT-068", func(t *testing.T) {
		projectId := os.Getenv("GCP_PROJECT_ID")
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		if projectId == "" || saJson == "" {
			t.Skip("GCP credentials not set")
		}
		name := "gcp-del-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_GCP(name, projectId, saJson),
				},
			},
		})
	})
}

// ---------------------------------------------------------------------------
// Kubernetes AGENTLESS
// ---------------------------------------------------------------------------

func TestAccACCT_KubernetesAgentless(t *testing.T) {
	t.Run("ACCT-070", func(t *testing.T) {
		role := os.Getenv("K8S_ROLE")
		extId := os.Getenv("K8S_EXTERNAL_ID")
		if role == "" || extId == "" {
			t.Skip("K8S_ROLE and K8S_EXTERNAL_ID required")
		}
		name := "k8s-al-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentless(name, "AWS", "", role, extId),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cloud_provider", "KUBERNETES"),
						resource.TestCheckResourceAttr("sedai_account.test", "integration_type", "AGENTLESS"),
					),
				},
			},
		})
	})

	t.Run("ACCT-071", func(t *testing.T) {
		clusterUrl := os.Getenv("K8S_CLUSTER_URL")
		caCert := os.Getenv("K8S_CA_CERT")
		role := os.Getenv("K8S_ROLE")
		extId := os.Getenv("K8S_EXTERNAL_ID")
		if clusterUrl == "" || role == "" || extId == "" {
			t.Skip("K8S_CLUSTER_URL, K8S_ROLE, K8S_EXTERNAL_ID required")
		}
		name := "k8s-al-url-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentlessWithUrl(name, "AWS", clusterUrl, caCert, role, extId),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cluster_url", clusterUrl),
					),
				},
			},
		})
	})

	t.Run("ACCT-072", func(t *testing.T) {
		role := os.Getenv("K8S_ROLE")
		extId := os.Getenv("K8S_EXTERNAL_ID")
		if role == "" || extId == "" {
			t.Skip("K8S_ROLE and K8S_EXTERNAL_ID required")
		}
		name := "k8s-al-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentless(name, "AWS", "", role, extId),
				},
				{
					Config:             testAccAccountConfig_K8sAgentless(name, "AWS", "", role, extId),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-073", func(t *testing.T) {
		role := os.Getenv("K8S_ROLE")
		extId := os.Getenv("K8S_EXTERNAL_ID")
		if role == "" || extId == "" {
			t.Skip("K8S_ROLE and K8S_EXTERNAL_ID required")
		}
		name := "k8s-al-imp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentless(name, "AWS", "", role, extId),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"role", "external_id", "ca_certificate"},
				},
			},
		})
	})

	t.Run("ACCT-074", func(t *testing.T) {
		clientId := os.Getenv("AZURE_CLIENT_ID")
		clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
		if clientId == "" || clientSecret == "" {
			t.Skip("AZURE_CLIENT_ID and AZURE_CLIENT_SECRET required for K8s/AZURE cluster provider")
		}
		name := "k8s-al-az-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentlessAzureCluster(name, clientId, clientSecret),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cluster_provider", "AZURE"),
					),
				},
			},
		})
	})

	t.Run("ACCT-075", func(t *testing.T) {
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		projectId := os.Getenv("GCP_PROJECT_ID")
		if saJson == "" || projectId == "" {
			t.Skip("GCP_SERVICE_ACCOUNT_JSON and GCP_PROJECT_ID required for K8s/GCP cluster provider")
		}
		name := "k8s-al-gcp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentlessGCPCluster(name, projectId, saJson),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cluster_provider", "GCP"),
					),
				},
			},
		})
	})

	t.Run("ACCT-076", func(t *testing.T) {
		name := "k8s-al-sm-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentlessSelfManaged(name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "cluster_provider", "SELF_MANAGED"),
					),
				},
			},
		})
	})

	t.Run("ACCT-077", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_K8sMissingClusterProvider("missing-cp"),
					ExpectError: regexpMustCompile(`(?i)cluster.?provider`),
				},
			},
		})
	})
}

// ---------------------------------------------------------------------------
// Kubernetes AGENT_BASED
// ---------------------------------------------------------------------------

func TestAccACCT_KubernetesAgentBased(t *testing.T) {
	t.Run("ACCT-080", func(t *testing.T) {
		clusterUrl := os.Getenv("K8S_AGENT_CLUSTER_URL")
		if clusterUrl == "" {
			t.Skip("K8S_AGENT_CLUSTER_URL required for agent-based K8s tests")
		}
		name := "k8s-ab-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentBased(name, "AWS"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("sedai_account.test", "id"),
						resource.TestCheckResourceAttr("sedai_account.test", "integration_type", "AGENT_BASED"),
						resource.TestCheckResourceAttrSet("sedai_account.test", "agent_api_key"),
						resource.TestCheckResourceAttrSet("sedai_account.test", "kube_install_cmd"),
						resource.TestCheckResourceAttrSet("sedai_account.test", "helm_install_cmd"),
					),
				},
			},
		})
	})

	t.Run("ACCT-081", func(t *testing.T) {
		clusterUrl := os.Getenv("K8S_AGENT_CLUSTER_URL")
		if clusterUrl == "" {
			t.Skip("K8S_AGENT_CLUSTER_URL required")
		}
		name := "k8s-ab-drift-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentBased(name, "AWS"),
				},
				{
					Config:             testAccAccountConfig_K8sAgentBased(name, "AWS"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: false,
				},
			},
		})
	})

	t.Run("ACCT-082", func(t *testing.T) {
		clusterUrl := os.Getenv("K8S_AGENT_CLUSTER_URL")
		if clusterUrl == "" {
			t.Skip("K8S_AGENT_CLUSTER_URL required")
		}
		name := "k8s-ab-imp-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentBased(name, "AWS"),
				},
				{
					ResourceName:            "sedai_account.test",
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"agent_api_key", "kube_install_cmd", "helm_install_cmd", "create_secret_kubectl_cmd"},
				},
			},
		})
	})

	t.Run("ACCT-083", func(t *testing.T) {
		clusterUrl := os.Getenv("K8S_AGENT_CLUSTER_URL")
		if clusterUrl == "" {
			t.Skip("K8S_AGENT_CLUSTER_URL required")
		}
		name := "k8s-ab-del-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentBased(name, "AWS"),
				},
			},
		})
	})

	t.Run("ACCT-084", func(t *testing.T) {
		saJson := os.Getenv("GCP_SERVICE_ACCOUNT_JSON")
		clusterUrl := os.Getenv("K8S_AGENT_CLUSTER_URL")
		if saJson == "" || clusterUrl == "" {
			t.Skip("GCP_SERVICE_ACCOUNT_JSON and K8S_AGENT_CLUSTER_URL required for GKE agent-based")
		}
		name := "k8s-ab-gke-" + randString(6)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfig_K8sAgentBased(name, "GCP"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("sedai_account.test", "cluster_provider", "GCP"),
						resource.TestCheckResourceAttrSet("sedai_account.test", "agent_api_key"),
					),
				},
			},
		})
	})

	t.Run("ACCT-085", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      testAccAccountConfig_K8sAgentBasedMissingClusterProvider("missing-cp-ab"),
					ExpectError: regexpMustCompile(`(?i)cluster.?provider`),
				},
			},
		})
	})
}

// ---------------------------------------------------------------------------
// HCL config builders
// ---------------------------------------------------------------------------

func testAccAccountConfig_Azure(name, tenantId, subscriptionId, clientId, clientSecret string) string {
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

func testAccAccountConfig_AzureWithManagedServices(name, tenantId, subscriptionId, clientId, clientSecret string, services []string) string {
	quoted := make([]string, len(services))
	for i, s := range services {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AZURE"
  integration_type = "AGENTLESS"
  tenant_id        = %[2]q
  subscription_id  = %[3]q
  client_id        = %[4]q
  client_secret    = %[5]q
  user_selected_managed_services = [%[6]s]
}
`, name, tenantId, subscriptionId, clientId, clientSecret, strings.Join(quoted, ", "))
}

// testAccAccountConfig_AzureMissingField returns an Azure config with one credential field left empty,
// used for validator tests. The provider should reject this at plan time.
func testAccAccountConfig_AzureMissingField(name, tenantId, subscriptionId, clientId, clientSecret string) string {
	cfg := fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "AZURE"
  integration_type = "AGENTLESS"
`, name)
	if tenantId != "" {
		cfg += fmt.Sprintf("  tenant_id       = %q\n", tenantId)
	}
	if subscriptionId != "" {
		cfg += fmt.Sprintf("  subscription_id = %q\n", subscriptionId)
	}
	if clientId != "" {
		cfg += fmt.Sprintf("  client_id       = %q\n", clientId)
	}
	if clientSecret != "" {
		cfg += fmt.Sprintf("  client_secret   = %q\n", clientSecret)
	}
	cfg += "}\n"
	return cfg
}

func testAccAccountConfig_GCP(name, projectId, serviceAccountJson string) string {
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

func testAccAccountConfig_GCPWithManagedServices(name, projectId, serviceAccountJson, services string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name                 = %[1]q
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  project_id           = %[2]q
  service_account_json = %[3]q
  user_selected_managed_services     = %[4]s
}
`, name, projectId, serviceAccountJson, services)
}

func testAccAccountConfig_GCPMissingProjectId(name, serviceAccountJson string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name                 = %[1]q
  cloud_provider       = "GCP"
  integration_type     = "AGENTLESS"
  service_account_json = %[2]q
}
`, name, serviceAccountJson)
}

func testAccAccountConfig_GCPMissingServiceAccountJson(name, projectId string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "GCP"
  integration_type = "AGENTLESS"
  project_id       = %[2]q
}
`, name, projectId)
}

func testAccAccountConfig_K8sAgentless(name, clusterProvider, clusterUrl, role, externalId string) string {
	cfg := fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = %[2]q
`, name, clusterProvider)
	if clusterUrl != "" {
		cfg += fmt.Sprintf("  cluster_url = %q\n", clusterUrl)
	}
	if role != "" {
		cfg += fmt.Sprintf("  role        = %q\n", role)
	}
	if externalId != "" {
		cfg += fmt.Sprintf("  external_id = %q\n", externalId)
	}
	cfg += "}\n"
	return cfg
}

func testAccAccountConfig_K8sAgentlessWithUrl(name, clusterProvider, clusterUrl, caCert, role, externalId string) string {
	cfg := fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = %[2]q
  cluster_url      = %[3]q
  role             = %[4]q
  external_id      = %[5]q
`, name, clusterProvider, clusterUrl, role, externalId)
	if caCert != "" {
		cfg += fmt.Sprintf("  ca_certificate = %q\n", caCert)
	}
	cfg += "}\n"
	return cfg
}

func testAccAccountConfig_K8sAgentlessAzureCluster(name, clientId, clientSecret string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "AZURE"
  client_id        = %[2]q
  client_secret    = %[3]q
}
`, name, clientId, clientSecret)
}

func testAccAccountConfig_K8sAgentlessGCPCluster(name, projectId, serviceAccountJson string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name                 = %[1]q
  cloud_provider       = "KUBERNETES"
  integration_type     = "AGENTLESS"
  cluster_provider     = "GCP"
  project_id           = %[2]q
  service_account_json = %[3]q
}
`, name, projectId, serviceAccountJson)
}

func testAccAccountConfig_K8sAgentlessSelfManaged(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
  cluster_provider = "SELF_MANAGED"
}
`, name)
}

func testAccAccountConfig_K8sMissingClusterProvider(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENTLESS"
}
`, name)
}

func testAccAccountConfig_K8sAgentBased(name, clusterProvider string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
  cluster_provider = %[2]q
}
`, name, clusterProvider)
}

func testAccAccountConfig_K8sAgentBasedMissingClusterProvider(name string) string {
	return fmt.Sprintf(`
resource "sedai_account" "test" {
  name             = %[1]q
  cloud_provider   = "KUBERNETES"
  integration_type = "AGENT_BASED"
}
`, name)
}
