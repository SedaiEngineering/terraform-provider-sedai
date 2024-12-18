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
	_ resource.Resource = &createGkeMonitoringProvider{}
)

func CreateGKEMonitoringProvider() resource.Resource {
	return &createGkeMonitoringProvider{}
}

type createGkeMonitoringProvider struct{}

type gkeMonitoringProviderModel struct {
	ID                  basetypes.StringValue `tfsdk:"id"`
	AccountId           basetypes.StringValue `tfsdk:"account_id"`
	ProjectId           basetypes.StringValue `tfsdk:"project_id"`
	Name                basetypes.StringValue `tfsdk:"name"`
	IntegrationType     basetypes.StringValue `tfsdk:"integration_type"`
	LbDimensions        basetypes.ListValue   `tfsdk:"lb_dimensions"`
	AppDimensions       basetypes.ListValue   `tfsdk:"app_dimensions"`
	InstanceDimensions  basetypes.ListValue   `tfsdk:"instance_dimensions"`
	RegionDimensions    basetypes.ListValue   `tfsdk:"region_dimensions"`
	ContainerDimensions basetypes.ListValue   `tfsdk:"container_dimensions"`
	NamespaceDimensions basetypes.ListValue   `tfsdk:"namespace_dimensions"`
	ClusterDimensions   basetypes.ListValue   `tfsdk:"cluster_dimensions"`
	ServiceAccountJson  basetypes.StringValue `tfsdk:"service_account_json"`
}

// Metadata returns the resource type name.
func (r *createGkeMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_gke_monitoring_provider"
}

// Schema defines the schema for the resource.
func (r *createGkeMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"project_id": schema.StringAttribute{
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
			"service_account_json": schema.StringAttribute{
				Computed: false,
				Required: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *createGkeMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gkeMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	monitoringProviderRequest := createGkeMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddGKEMonitoring(monitoringProviderRequest)
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

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *createGkeMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gkeMonitoringProviderModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetchedMonitoringProvider, err := monitoringProvider.GetMonitoringProviderById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read monitoring provider", err.Error())
		return
	}

	state.ID = basetypes.NewStringValue(fetchedMonitoringProvider["id"].(string))
	state.Name = basetypes.NewStringValue(fetchedMonitoringProvider["name"].(string))
	state.IntegrationType = basetypes.NewStringValue(fetchedMonitoringProvider["integrationType"].(string))
	state.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "lbDimensions"))
	state.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "appDimensions"))
	state.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "instanceDimensions"))
	state.RegionDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "regionDimensions"))
	state.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "containerDimensions"))
	state.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "namespaceDimensions"))
	state.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetchedMonitoringProvider, "clusterDimensions"))
	state.AccountId = basetypes.NewStringValue(fetchedMonitoringProvider["accountId"].(string))
	state.ProjectId = basetypes.NewStringValue(fetchedMonitoringProvider["details"].(map[string]interface{})["projectId"].(string))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *createGkeMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gkeMonitoringProviderModel
	var state gkeMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	monitoringProviderRequest := createGkeMonitoringProviderRequest(plan)
	response, err := monitoringProvider.AddGKEMonitoring(monitoringProviderRequest)
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

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state.
func (r *createGkeMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gkeMonitoringProviderModel

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

func createGkeMonitoringProviderRequest(plan gkeMonitoringProviderModel) monitoringProvider.CreateGKEMonitoringProviderRequest {
	//credential
	createGkeMonitoringProviderRequest := monitoringProvider.CreateGKEMonitoringProviderRequest{
		AccountId:   plan.AccountId.ValueString(),
		ProjectId:   plan.ProjectId.ValueString(),
		Credentials: credentials.NewGKEMonitoringCredentials(plan.ServiceAccountJson.String()),
	}

	// for updates
	if plan.ID.String() != "" {
		createGkeMonitoringProviderRequest.ID = plan.ID.ValueString()
	}

	if plan.Name.String() != "" {
		createGkeMonitoringProviderRequest.Name = plan.Name.ValueString()
	}
	if plan.IntegrationType.String() != "" {
		createGkeMonitoringProviderRequest.IntegrationType = plan.IntegrationType.ValueString()
	}
	if plan.LbDimensions.IsNull() {
		createGkeMonitoringProviderRequest.LbDimensions = convertFromBaseTypes(plan.LbDimensions)
	}
	if plan.AppDimensions.IsNull() {
		createGkeMonitoringProviderRequest.AppDimensions = convertFromBaseTypes(plan.AppDimensions)
	}
	if plan.InstanceDimensions.IsNull() {
		createGkeMonitoringProviderRequest.InstanceDimensions = convertFromBaseTypes(plan.InstanceDimensions)
	}
	if plan.RegionDimensions.IsNull() {
		createGkeMonitoringProviderRequest.RegionDimensions = convertFromBaseTypes(plan.RegionDimensions)
	}
	if plan.ContainerDimensions.IsNull() {
		createGkeMonitoringProviderRequest.ContainerDimensions = convertFromBaseTypes(plan.ContainerDimensions)
	}
	if plan.NamespaceDimensions.IsNull() {
		createGkeMonitoringProviderRequest.NamespaceDimensions = convertFromBaseTypes(plan.NamespaceDimensions)
	}
	if plan.ClusterDimensions.IsNull() {
		createGkeMonitoringProviderRequest.ClusterDimensions = convertFromBaseTypes(plan.ClusterDimensions)
	}

	return createGkeMonitoringProviderRequest
}

func convertFromBaseTypes(baseTypeString basetypes.ListValue) []string {
	elements := baseTypeString.Elements()
	result := make([]string, len(elements))
	for i, sv := range elements {
		result[i] = sv.String()
	}
	return result
}

func getDimensionFromResponse(response map[string]interface{}, key string) []string {
	dimensionsInterface, ok := response[key].([]interface{})
	if !ok {
		// Handle the error, maybe the field doesn't exist or is of unexpected type
		return nil
	}

	// Create a slice to hold the string values
	dimensions := make([]string, len(dimensionsInterface))

	// Convert each interface{} to string
	for i, v := range dimensionsInterface {
		dimensions[i], ok = v.(string)
		if !ok {
			continue
		}
	}

	return dimensions
}
