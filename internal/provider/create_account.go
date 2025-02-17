package provider

import (
	"context"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/credentials"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &createAccount{}
)

// CreateAccount is a helper function to simplify the provider implementation.
func CreateAccount() resource.Resource {
	return &createAccount{}
}

// createAccount is the resource implementation.
type createAccount struct{}

type accountModel struct {
	ID                     basetypes.StringValue `tfsdk:"id"`
	Name                   string                `tfsdk:"name"`
	CloudProvider          string                `tfsdk:"cloud_provider"`
	IntegrationType        string                `tfsdk:"integration_type"`
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
}

// Metadata returns the resource type name.
func (r *createAccount) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_account"
}

// Schema defines the schema for the resource.
func (r *createAccount) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"name": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"cloud_provider": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"integration_type": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"cluster_provider": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"cluster_url": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"project_id": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"zone": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"region": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"is_zonal_cluster": schema.BoolAttribute{
				Computed: false,
				Optional: true,
			},
			"service_account_json": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"ca_certificate": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"role": schema.StringAttribute{
				Optional: true,
				Computed: false,
			},
			"external_id": schema.StringAttribute{
				Optional: true,
				Computed: false,
			},
			"access_key": schema.StringAttribute{
				Optional: true,
				Computed: false,
			},
			"secret_key": schema.StringAttribute{
				Optional: true,
				Computed: false,
			},
			"agent_api_key": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"kube_install_cmd": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"helm_install_cmd": schema.StringAttribute{
				Computed: true,
				Required: false,
			},
			"create_secret_kubectl_cmd": schema.StringAttribute{
				Computed: true,
				Required: false,
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
		resp.Diagnostics.AddError("Unable to create account", err.Error())
		return
	}

	if plan.IntegrationType == "AGENT_BASED" {
		agentInstallationCommand, err := account.GetAgentInstallationCommand(plan.Name)
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
		resp.Diagnostics.AddError(
			"Error fetching sedai account",
			"Could not fetch account with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = basetypes.NewStringValue(fetchedAccount.ID)
	state.Name = fetchedAccount.Name
	state.CloudProvider = fetchedAccount.AccountDetails.CloudProvider
	state.IntegrationType = fetchedAccount.AccountDetails.IntegrationType

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

	if plan.IntegrationType == "AGENT_BASED" {
		agentInstallationCommand, err := account.GetAgentInstallationCommand(plan.Name)
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

	_, err := account.DeleteAccount(state.Name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete account", err.Error())
		return
	}
}

func createCredentials(plan accountModel) interface{} {
	// Check if role and external id are provided.
	if plan.Role.ValueString() != "" && plan.ExternalId.ValueString() != "" {
		return credentials.NewAwsRoleCredentials(plan.Role.ValueString(), plan.ExternalId.ValueString())
	}

	// Check if cloud provider is KUBERNETES and cluster provider is GCP
	if plan.CloudProvider == "KUBERNETES" {
		if plan.ClusterProvider.ValueString() == "AWS" {
			if plan.IntegrationType == "AGENT_BASED" {
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
		Name:            plan.Name,
		CloudProvider:   plan.CloudProvider,
		Credentials:     credential,
		IntegrationType: plan.IntegrationType,
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

	return createAccountRequest
}
