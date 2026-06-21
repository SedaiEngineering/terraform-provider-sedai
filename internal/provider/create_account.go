package provider

import (
	"context"
	"errors"
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/credentials"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &createAccount{}
	_ resource.ResourceWithImportState = &createAccount{}
)

// CreateAccount is a helper function to simplify the provider implementation.
func CreateAccount() resource.Resource {
	return &createAccount{}
}

// createAccount is the resource implementation.
type createAccount struct{}

type accountModel struct {
	ID                     basetypes.StringValue `tfsdk:"id"`
	Name                   basetypes.StringValue `tfsdk:"name"`
	CloudProvider          basetypes.StringValue `tfsdk:"cloud_provider"`
	IntegrationType        basetypes.StringValue `tfsdk:"integration_type"`
	Role                   basetypes.StringValue `tfsdk:"role"`
	ExternalId             basetypes.StringValue `tfsdk:"external_id"`
	AccessKey              basetypes.StringValue `tfsdk:"access_key"`
	SecretKey              basetypes.StringValue `tfsdk:"secret_key"`
	ClusterProvider        basetypes.StringValue `tfsdk:"cluster_provider"`
	ClusterURL             basetypes.StringValue `tfsdk:"cluster_url"`
	ProjectId              basetypes.StringValue `tfsdk:"project_id"`
	Zone                   basetypes.StringValue `tfsdk:"zone"`
	Region                 basetypes.StringValue `tfsdk:"region"`
	IsZonalCluster         basetypes.BoolValue   `tfsdk:"is_zonal_cluster"`
	ServiceAccountJson     basetypes.StringValue `tfsdk:"service_account_json"`
	CACertificate          basetypes.StringValue `tfsdk:"ca_certificate"`
	AgentApiKey            basetypes.StringValue `tfsdk:"agent_api_key"`
	KubeInstallCmd         basetypes.StringValue `tfsdk:"kube_install_cmd"`
	HelmInstallCmd         basetypes.StringValue `tfsdk:"helm_install_cmd"`
	CreateSecretKubectlCmd basetypes.StringValue `tfsdk:"create_secret_kubectl_cmd"`
	TenantId                    basetypes.StringValue `tfsdk:"tenant_id"`
	SubscriptionId              basetypes.StringValue `tfsdk:"subscription_id"`
	ClientId                    basetypes.StringValue `tfsdk:"client_id"`
	ClientSecret                basetypes.StringValue `tfsdk:"client_secret"`
	UserSelectedManagedServices basetypes.ListValue   `tfsdk:"user_selected_managed_services"`
}

// Metadata returns the resource type name.
func (r *createAccount) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

// Schema defines the schema for the resource.
func (r *createAccount) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Sedai account for a cloud provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Sedai account ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Account name.",
			},
			"cloud_provider": schema.StringAttribute{
				Required:    true,
				Description: "Cloud provider. Valid values: `AWS`, `AZURE`, `GCP`, `KUBERNETES`.",
			},
			"integration_type": schema.StringAttribute{
				Required:    true,
				Description: "Integration type. Valid values: `AGENTLESS`, `AGENT_BASED`.",
			},
			"cluster_provider": schema.StringAttribute{
				Optional:    true,
				Description: "Kubernetes cluster provider. Required when `cloud_provider = \"KUBERNETES\"`. Valid values: `AWS`, `GCP`, `AZURE`, `SELF_MANAGED`.",
			},
			"cluster_url": schema.StringAttribute{
				Optional:    true,
				Description: "Cluster API server URL. Required for agentless Kubernetes.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "GCP project ID. Required for GCP and Kubernetes GCP accounts.",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "GCP zone. Used for zonal GKE clusters.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Cluster region. Used for Kubernetes accounts.",
			},
			"is_zonal_cluster": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the GKE cluster is zonal (vs regional).",
			},
			"service_account_json": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "GCP service account JSON key. Required for GCP and Kubernetes GCP accounts.",
			},
			"ca_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "CA certificate for cluster TLS verification (Kubernetes agentless).",
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Description: "IAM role ARN for role-based authentication (AWS / Kubernetes AWS).",
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Description: "External ID for the IAM role (AWS / Kubernetes AWS).",
			},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "AWS access key for static credential authentication.",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "AWS secret key for static credential authentication.",
			},
			"agent_api_key": schema.StringAttribute{
				Computed:    true,
				Description: "Agent API key. Populated only for `AGENT_BASED` integration.",
			},
			"kube_install_cmd": schema.StringAttribute{
				Computed:    true,
				Description: "kubectl command to install the Sedai agent. Populated only for `AGENT_BASED` integration.",
			},
			"helm_install_cmd": schema.StringAttribute{
				Computed:    true,
				Description: "Helm command to install the Sedai agent. Populated only for `AGENT_BASED` integration.",
			},
			"create_secret_kubectl_cmd": schema.StringAttribute{
				Computed:    true,
				Description: "kubectl command to create the agent secret. Populated only for `AGENT_BASED` integration.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "Azure Active Directory tenant ID. Required for Azure accounts.",
			},
			"subscription_id": schema.StringAttribute{
				Optional:    true,
				Description: "Azure subscription ID. Required for Azure accounts.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Azure service principal client ID. Required for Azure accounts.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Azure service principal client secret. Required for Azure accounts.",
			},
			"user_selected_managed_services": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Cloud services to enable. AWS values: `LAMBDA`, `EC2`, `ECS`, `EBS`, `EFS`, `S3`, `RDS`, `DYNAMO_DB`, `DATABRICKS`. Azure values: `VM`, `AZURE_DISK`, `AZURE_BLOB`, `DATABRICKS`. GCP values: `GCE`, `DATAFLOW`, `GCP_DISK`, `CLOUD_STORAGE`, `BIG_QUERY`, `DATABRICKS`.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *createAccount) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createAccountRequest := createAccountRequest(plan)
	response, err := account.CreateAccount(createAccountRequest)
	if err != nil {
		// POST failed — verify if the backend created it anyway.
		// Handles EOF-during-POST where the server processed the request but the
		// response was lost in transit. See LIMITATIONS.md for known edge cases.
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			existing, searchErr := account.SearchAccountsByName(plan.Name.ValueString())
			if searchErr == nil && len(existing) > 0 {
				resp.Diagnostics.AddWarning(
					"Account created despite connection error",
					"Account '"+plan.Name.ValueString()+"' was found on the backend after a failed POST — "+
						"the response was likely lost in transit. Using existing ID: "+existing[0].ID,
				)
				plan.ID = basetypes.NewStringValue(existing[0].ID)
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}
		resp.Diagnostics.AddError("Unable to create account", err.Error())
		return
	}

	// Write the ID to state immediately — before any post-create work.
	// If anything below fails, the account exists on the backend and Terraform
	// knows its ID, so the next apply will Update rather than Create again.
	plan.ID = basetypes.NewStringValue(response["accountId"].(string))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.IntegrationType.ValueString() == "AGENT_BASED" {
		agentInstallationCommand, err := account.GetAgentInstallationCommand(plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to get agent installation command", err.Error())
			return
		}
		plan.AgentApiKey = basetypes.NewStringValue(agentInstallationCommand["apiKey"].(string))
		plan.KubeInstallCmd = basetypes.NewStringValue(agentInstallationCommand["kubeInstallCmd"].(string))
		plan.HelmInstallCmd = basetypes.NewStringValue(agentInstallationCommand["helmInstallCmd"].(string))
		if agentInstallationCommand["createSecretKubectlCmd"] != nil {
			plan.CreateSecretKubectlCmd = basetypes.NewStringValue(agentInstallationCommand["createSecretKubectlCmd"].(string))
		} else {
			plan.CreateSecretKubectlCmd = basetypes.NewStringValue("")
		}
	} else {
		plan.AgentApiKey = basetypes.NewStringValue("")
		plan.KubeInstallCmd = basetypes.NewStringValue("")
		plan.HelmInstallCmd = basetypes.NewStringValue("")
		plan.CreateSecretKubectlCmd = basetypes.NewStringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Poll until the account is queryable on the backend before returning.
	// The Sedai API is async — returning immediately lets dependent resources
	// (MPs, groups) fire before the account is ready, causing race conditions.
	// Soft timeout: proceed after 30s regardless so a slow backend doesn't block forever.
	accountId := plan.ID.ValueString()
	for i := 0; i < 15; i++ {
		fetched, err := account.SearchAccountsById(accountId)
		if err == nil && fetched != nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *createAccount) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetchedAccount, err := account.SearchAccountsById(state.ID.ValueString())
	if err != nil {
		var notFound *impl.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error fetching sedai account",
			"Could not fetch account with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	if fetchedAccount == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = basetypes.NewStringValue(fetchedAccount.ID)
	state.Name = basetypes.NewStringValue(fetchedAccount.Name)
	state.CloudProvider = basetypes.NewStringValue(fetchedAccount.AccountDetails.CloudProvider)
	state.IntegrationType = basetypes.NewStringValue(fetchedAccount.AccountDetails.IntegrationType)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *createAccount) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountModel
	var state accountModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	createAccountRequest := createAccountRequest(plan)
	response, err := account.CreateAccount(createAccountRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create account", err.Error())
		return
	}

	if plan.IntegrationType.ValueString() == "AGENT_BASED" {
		agentInstallationCommand, err := account.GetAgentInstallationCommand(plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to get agent installation command", err.Error())
			return
		}
		plan.AgentApiKey = basetypes.NewStringValue(agentInstallationCommand["apiKey"].(string))
		plan.KubeInstallCmd = basetypes.NewStringValue(agentInstallationCommand["kubeInstallCmd"].(string))
		plan.HelmInstallCmd = basetypes.NewStringValue(agentInstallationCommand["helmInstallCmd"].(string))
		if agentInstallationCommand["createSecretKubectlCmd"] != nil {
			plan.CreateSecretKubectlCmd = basetypes.NewStringValue(agentInstallationCommand["createSecretKubectlCmd"].(string))
		} else {
			plan.CreateSecretKubectlCmd = basetypes.NewStringValue("")
		}
	} else {
		plan.AgentApiKey = basetypes.NewStringValue("")
		plan.KubeInstallCmd = basetypes.NewStringValue("")
		plan.HelmInstallCmd = basetypes.NewStringValue("")
		plan.CreateSecretKubectlCmd = basetypes.NewStringValue("")
	}

	plan.ID = basetypes.NewStringValue(response["accountId"].(string))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *createAccount) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := account.DeleteAccount(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete account", err.Error())
		return
	}
}

func (r *createAccount) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createCredentials(plan accountModel) interface{} {
	if plan.CloudProvider.ValueString() == "AZURE" {
		return credentials.NewAzureClientCredentials(plan.ClientId.ValueString(), plan.ClientSecret.ValueString())
	}

	if plan.CloudProvider.ValueString() == "GCP" {
		return credentials.NewGCPServiceAccountJsonCredentials(plan.ServiceAccountJson.ValueString())
	}

	// Check if role and external id are provided.
	if plan.Role.ValueString() != "" && plan.ExternalId.ValueString() != "" {
		return credentials.NewAwsRoleCredentials(plan.Role.ValueString(), plan.ExternalId.ValueString())
	}

	// Check if cloud provider is KUBERNETES and cluster provider is GCP
	if plan.CloudProvider.ValueString() == "KUBERNETES" {
		if plan.ClusterProvider.ValueString() == "AWS" {
			if plan.IntegrationType.ValueString() == "AGENT_BASED" {
				return credentials.NewAwsCredentials()
			} else {
				if plan.Role.ValueString() != "" {
					return credentials.NewAwsRoleCredentials(plan.Role.ValueString(), plan.ExternalId.ValueString())
				} else {
					return credentials.NewAwsKeyCredentials(plan.AccessKey.ValueString(), plan.SecretKey.ValueString())
				}
			}
		}
		if plan.ClusterProvider.ValueString() == "GCP" {
			return credentials.NewGCPServiceAccountJsonCredentials(plan.ServiceAccountJson.ValueString())
		}
	}

	return nil
}

func createAccountRequest(plan accountModel) account.CreateAccountRequest {
	credential := createCredentials(plan)
	createAccountRequest := account.CreateAccountRequest{
		Name:            plan.Name.ValueString(),
		CloudProvider:   plan.CloudProvider.ValueString(),
		Credentials:     credential,
		IntegrationType: plan.IntegrationType.ValueString(),
	}

	// in case of update
	if plan.ID.ValueString() != "" {
		createAccountRequest.ID = plan.ID.ValueString()
	}

	if plan.ClusterProvider.ValueString() != "" {
		createAccountRequest.ClusterProvider = plan.ClusterProvider.ValueString()
	}
	if plan.ClusterURL.ValueString() != "" {
		createAccountRequest.ClusterUrl = plan.ClusterURL.ValueString()
	}
	if plan.ProjectId.ValueString() != "" {
		createAccountRequest.ProjectId = plan.ProjectId.ValueString()
	}
	if plan.Zone.ValueString() != "" {
		createAccountRequest.Zone = plan.Zone.ValueString()
	}
	if plan.Region.ValueString() != "" {
		createAccountRequest.Region = plan.Region.ValueString()
	}
	if plan.IsZonalCluster.ValueBool() {
		createAccountRequest.IsZonalCluster = true
	}
	if plan.CACertificate.ValueString() != "" {
		createAccountRequest.CaCertificate = plan.CACertificate.ValueString()
	}
	if plan.TenantId.ValueString() != "" {
		createAccountRequest.TenantId = plan.TenantId.ValueString()
	}
	if plan.SubscriptionId.ValueString() != "" {
		createAccountRequest.SubscriptionId = plan.SubscriptionId.ValueString()
	}
	if !plan.UserSelectedManagedServices.IsNull() && !plan.UserSelectedManagedServices.IsUnknown() {
		createAccountRequest.UserSelectedManagedServices = convertFromBaseTypes(plan.UserSelectedManagedServices)
	}

	return createAccountRequest
}

