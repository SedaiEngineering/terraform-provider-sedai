package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/credentials"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/monitoringProvider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource = &createFpMonitoringProvider{}
)

func CreateFPMonitoringProvider() resource.Resource {
	return &createFpMonitoringProvider{}
}

type createFpMonitoringProvider struct{}

type fpMonitoringProviderModel struct {
	ID                     basetypes.StringValue `tfsdk:"id"`
	AccountId              basetypes.StringValue `tfsdk:"account_id"`
	Endpoint               basetypes.StringValue `tfsdk:"endpoint"`
	Name                   basetypes.StringValue `tfsdk:"name"`
	IntegrationType        basetypes.StringValue `tfsdk:"integration_type"`
	LbDimensions           basetypes.ListValue   `tfsdk:"lb_dimensions"`
	AppDimensions          basetypes.ListValue   `tfsdk:"app_dimensions"`
	InstanceDimensions     basetypes.ListValue   `tfsdk:"instance_dimensions"`
	RegionDimensions       basetypes.ListValue   `tfsdk:"region_dimensions"`
	ContainerDimensions    basetypes.ListValue   `tfsdk:"container_dimensions"`
	NamespaceDimensions    basetypes.ListValue   `tfsdk:"namespace_dimensions"`
	ClusterDimensions      basetypes.ListValue   `tfsdk:"cluster_dimensions"`
	OsDimensions           basetypes.ListValue   `tfsdk:"os_dimensions"`
	ArchitectureDimensions basetypes.ListValue   `tfsdk:"architecture_dimensions"`
	InstanceTypeDimensions basetypes.ListValue   `tfsdk:"instance_type_dimensions"`
	AzDimensions           basetypes.ListValue   `tfsdk:"az_dimensions"`
	EnvDimensions          basetypes.ListValue   `tfsdk:"env_dimensions"`
	// InstanceIdPattern      basetypes.StringValue `tfsdk:"instance_id_pattern"`
	BearerToken   basetypes.StringValue `tfsdk:"bearer_token"`
	TokenEndpoint basetypes.StringValue `tfsdk:"token_endpoint"`
	ClientID      basetypes.StringValue `tfsdk:"client_id"`
	ClientSecret  basetypes.StringValue `tfsdk:"client_secret"`
}

// Metadata returns the resource type name.
func (r *createFpMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_federated_prometheus_monitoring_provider"
}

// Schema defines the schema for the resource.
func (r *createFpMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"account_id": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"endpoint": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"integration_type": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"lb_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"app_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"instance_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"region_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"container_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"namespace_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"cluster_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"os_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"architecture_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"instance_type_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"az_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			"env_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
			},
			// "instance_id_pattern": schema.StringAttribute{
			// 	Computed: false,
			// 	Optional: true,
			// },
			"bearer_token": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"token_endpoint": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Computed: false,
				Optional: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *createFpMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fpMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitoringProviderRequest := createFpMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddFederatedPrometheusMonitoring(monitoringProviderRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create monitoring provider", err.Error())
		return
	}

	plan.ID = basetypes.NewStringValue(response["id"].(string))
	plan.Name = basetypes.NewStringValue(response["name"].(string))
	plan.IntegrationType = basetypes.NewStringValue(response["integrationType"].(string))
	plan.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "lbDimensions"))
	plan.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "appDimensions"))
	plan.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceDimensions"))
	plan.RegionDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "regionDimensions"))
	plan.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	plan.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	plan.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))
	plan.AzDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "azDimensions"))
	plan.EnvDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "envDimensions"))
	plan.OsDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "osDimensions"))
	plan.ArchitectureDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "architectureDimensions"))
	plan.InstanceTypeDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceTypeDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *createFpMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fpMonitoringProviderModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := monitoringProvider.GetMonitoringProviderById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read monitoring provider", err.Error())
		return
	}

	state.ID = basetypes.NewStringValue(response["id"].(string))
	state.Name = basetypes.NewStringValue(response["name"].(string))
	state.IntegrationType = basetypes.NewStringValue(response["integrationType"].(string))
	state.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "lbDimensions"))
	state.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "appDimensions"))
	state.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceDimensions"))
	state.RegionDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "regionDimensions"))
	state.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	state.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	state.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))
	state.AzDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "azDimensions"))
	state.EnvDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "envDimensions"))
	state.OsDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "osDimensions"))
	state.ArchitectureDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "architectureDimensions"))
	state.InstanceTypeDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceTypeDimensions"))
	state.AccountId = basetypes.NewStringValue(response["accountId"].(string))
	state.Endpoint = basetypes.NewStringValue(response["details"].(map[string]interface{})["endpoint"].(string))
	// state.InstanceIdPattern = basetypes.NewStringValue(response["instanceIdPattern"].(string))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *createFpMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fpMonitoringProviderModel
	var state fpMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	monitoringProviderRequest := createFpMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddFederatedPrometheusMonitoring(monitoringProviderRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create monitoring provider", err.Error())
		return
	}

	plan.ID = basetypes.NewStringValue(response["id"].(string))
	plan.Name = basetypes.NewStringValue(response["name"].(string))
	plan.IntegrationType = basetypes.NewStringValue(response["integrationType"].(string))
	plan.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "lbDimensions"))
	plan.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "appDimensions"))
	plan.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceDimensions"))
	plan.RegionDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "regionDimensions"))
	plan.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	plan.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	plan.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))
	plan.AzDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "azDimensions"))
	plan.EnvDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "envDimensions"))
	plan.OsDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "osDimensions"))
	plan.ArchitectureDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "architectureDimensions"))
	plan.InstanceTypeDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "instanceTypeDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state.
func (r *createFpMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fpMonitoringProviderModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetchedAccount, err := account.SearchAccountsById(state.AccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read account", err.Error())
		return
	}

	if fetchedAccount == nil {
		return
	}

	deleteMonitoringProvider, err := monitoringProvider.DeleteMonitoringProvider(state.ID.ValueString())
	if err != nil || !deleteMonitoringProvider {
		resp.Diagnostics.AddError("Unable to delete monitoring provider", err.Error())
		return
	}

	return
}

func createFpCredentials(plan fpMonitoringProviderModel) interface{} {
	if plan.BearerToken.String() != "" {
		return credentials.NewFederatedPrometheusJWT(plan.BearerToken.ValueString())
	}

	if plan.TokenEndpoint.String() != "" && plan.ClientID.String() != "" && plan.ClientSecret.String() != "" {
		return credentials.NewFederatedPrometheusClientCredentials(plan.TokenEndpoint.ValueString(), plan.ClientID.ValueString(), plan.ClientSecret.ValueString())
	}

	return credentials.NewFederatedPrometheusNoAuth()
}

func createFpMonitoringProviderRequest(plan fpMonitoringProviderModel) monitoringProvider.CreateFederatedPrometheusMonitoringProviderRequest {
	//credential
	createFpMonitoringProviderRequest := monitoringProvider.CreateFederatedPrometheusMonitoringProviderRequest{
		AccountId:   plan.AccountId.ValueString(),
		Endpoint:    plan.Endpoint.ValueString(),
		Credentials: createFpCredentials(plan),
	}

	// for updates
	if plan.ID.String() != "" {
		createFpMonitoringProviderRequest.ID = plan.ID.ValueString()
	}

	if plan.Name.String() != "" {
		createFpMonitoringProviderRequest.Name = plan.Name.ValueString()
	}
	if plan.IntegrationType.String() != "" {
		createFpMonitoringProviderRequest.IntegrationType = plan.IntegrationType.ValueString()
	}
	if plan.LbDimensions.IsNull() {
		createFpMonitoringProviderRequest.LbDimensions = convertFromBaseTypes(plan.LbDimensions)
	}
	if plan.AppDimensions.IsNull() {
		createFpMonitoringProviderRequest.AppDimensions = convertFromBaseTypes(plan.AppDimensions)
	}
	if plan.InstanceDimensions.IsNull() {
		createFpMonitoringProviderRequest.InstanceDimensions = convertFromBaseTypes(plan.InstanceDimensions)
	}
	if plan.RegionDimensions.IsNull() {
		createFpMonitoringProviderRequest.RegionDimensions = convertFromBaseTypes(plan.RegionDimensions)
	}
	if plan.ContainerDimensions.IsNull() {
		createFpMonitoringProviderRequest.ContainerDimensions = convertFromBaseTypes(plan.ContainerDimensions)
	}
	if plan.NamespaceDimensions.IsNull() {
		createFpMonitoringProviderRequest.NamespaceDimensions = convertFromBaseTypes(plan.NamespaceDimensions)
	}
	if plan.ClusterDimensions.IsNull() {
		createFpMonitoringProviderRequest.ClusterDimensions = convertFromBaseTypes(plan.ClusterDimensions)
	}
	if plan.OsDimensions.IsNull() {
		createFpMonitoringProviderRequest.OsDimensions = convertFromBaseTypes(plan.OsDimensions)
	}
	if plan.ArchitectureDimensions.IsNull() {
		createFpMonitoringProviderRequest.ArchitectureDimensions = convertFromBaseTypes(plan.ArchitectureDimensions)
	}
	if plan.InstanceTypeDimensions.IsNull() {
		createFpMonitoringProviderRequest.InstanceTypeDimensions = convertFromBaseTypes(plan.InstanceTypeDimensions)
	}
	if plan.AzDimensions.IsNull() {
		createFpMonitoringProviderRequest.AzDimensions = convertFromBaseTypes(plan.AzDimensions)
	}
	if plan.EnvDimensions.IsNull() {
		createFpMonitoringProviderRequest.EnvDimensions = convertFromBaseTypes(plan.EnvDimensions)
	}
	// if plan.InstanceIdPattern.String() != "" {
	// 	createFpMonitoringProviderRequest.InstanceIdPattern = plan.InstanceIdPattern.ValueString()
	// }

	return createFpMonitoringProviderRequest
}
