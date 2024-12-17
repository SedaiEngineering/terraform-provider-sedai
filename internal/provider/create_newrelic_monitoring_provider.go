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
	_ resource.Resource = &createNewrelicMonitoringProvider{}
)

func CreateNewrelicMonitoringProvider() resource.Resource {
	return &createNewrelicMonitoringProvider{}
}

type createNewrelicMonitoringProvider struct{}

type newrelicMonitoringProviderModel struct {
	ID                  basetypes.StringValue `tfsdk:"id"`
	AccountId           basetypes.StringValue `tfsdk:"account_id"`
	Name                basetypes.StringValue `tfsdk:"name"`
	IntegrationType     basetypes.StringValue `tfsdk:"integration_type"`
	LbDimensions        basetypes.ListValue   `tfsdk:"lb_dimensions"`
	AppDimensions       basetypes.ListValue   `tfsdk:"app_dimensions"`
	InstanceDimensions  basetypes.ListValue   `tfsdk:"instance_dimensions"`
	ContainerDimensions basetypes.ListValue   `tfsdk:"container_dimensions"`
	NamespaceDimensions basetypes.ListValue   `tfsdk:"namespace_dimensions"`
	ClusterDimensions   basetypes.ListValue   `tfsdk:"cluster_dimensions"`
	ApiKey              basetypes.StringValue `tfsdk:"api_key"`
	NewRelicAccountId   basetypes.StringValue `tfsdk:"newrelic_account_id"`
	ApiServer           basetypes.StringValue `tfsdk:"api_server"`
}

// Metadata returns the resource type name.
func (r *createNewrelicMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_newrelic_monitoring_provider"
}

// Schema defines the schema for the resource.
func (r *createNewrelicMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"api_key": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"newrelic_account_id": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
			"api_server": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *createNewrelicMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan newrelicMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitoringProviderRequest := createNewrelicMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddNewRelicMonitoring(monitoringProviderRequest)
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
	plan.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	plan.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	plan.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *createNewrelicMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state newrelicMonitoringProviderModel

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
	state.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	state.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	state.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))
	state.AccountId = basetypes.NewStringValue(response["accountId"].(string))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *createNewrelicMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan newrelicMonitoringProviderModel
	var state newrelicMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	monitoringProviderRequest := createNewrelicMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddNewRelicMonitoring(monitoringProviderRequest)
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
	plan.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "containerDimensions"))
	plan.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "namespaceDimensions"))
	plan.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(response, "clusterDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state.
func (r *createNewrelicMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state newrelicMonitoringProviderModel

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

func createNewrelicMonitoringProviderRequest(plan newrelicMonitoringProviderModel) monitoringProvider.CreateNewRelicMonitoringProviderRequest {
	//credential
	createNewrelicMonitoringProviderRequest := monitoringProvider.CreateNewRelicMonitoringProviderRequest{
		AccountId:         plan.AccountId.ValueString(),
		NewRelicAccountId: plan.NewRelicAccountId.ValueString(),
		ApiServer:         plan.ApiServer.ValueString(),
		Credentials:       credentials.NewNewrelicCredentials(plan.ApiKey.ValueString()),
	}

	// for updates
	if plan.ID.String() != "" {
		createNewrelicMonitoringProviderRequest.ID = plan.ID.ValueString()
	}

	if plan.Name.String() != "" {
		createNewrelicMonitoringProviderRequest.Name = plan.Name.ValueString()
	}
	if plan.IntegrationType.String() != "" {
		createNewrelicMonitoringProviderRequest.IntegrationType = plan.IntegrationType.ValueString()
	}
	if plan.LbDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.LbDimensions = convertFromBaseTypes(plan.LbDimensions)
	}
	if plan.AppDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.AppDimensions = convertFromBaseTypes(plan.AppDimensions)
	}
	if plan.InstanceDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.InstanceDimensions = convertFromBaseTypes(plan.InstanceDimensions)
	}
	if plan.ContainerDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.ContainerDimensions = convertFromBaseTypes(plan.ContainerDimensions)
	}
	if plan.NamespaceDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.NamespaceDimensions = convertFromBaseTypes(plan.NamespaceDimensions)
	}
	if plan.ClusterDimensions.IsNull() {
		createNewrelicMonitoringProviderRequest.ClusterDimensions = convertFromBaseTypes(plan.ClusterDimensions)
	}

	return createNewrelicMonitoringProviderRequest
}
