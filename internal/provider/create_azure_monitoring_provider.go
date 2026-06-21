package provider

import (
	"context"
	"errors"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/monitoringProvider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource            = &createAzureMonitoringProvider{}
	_ resource.ResourceWithImportState = &createAzureMonitoringProvider{}
)

func CreateAzureMonitoringProvider() resource.Resource {
	return &createAzureMonitoringProvider{}
}

type createAzureMonitoringProvider struct{}

type azureMonitoringProviderModel struct {
	ID                  basetypes.StringValue `tfsdk:"id"`
	AccountId           basetypes.StringValue `tfsdk:"account_id"`
	Name                basetypes.StringValue `tfsdk:"name"`
	IntegrationType     basetypes.StringValue `tfsdk:"integration_type"`
	LbDimensions        basetypes.ListValue   `tfsdk:"lb_dimensions"`
	AppDimensions       basetypes.ListValue   `tfsdk:"app_dimensions"`
	InstanceDimensions  basetypes.ListValue   `tfsdk:"instance_dimensions"`
	RegionDimensions    basetypes.ListValue   `tfsdk:"region_dimensions"`
	ContainerDimensions basetypes.ListValue   `tfsdk:"container_dimensions"`
	NamespaceDimensions basetypes.ListValue   `tfsdk:"namespace_dimensions"`
	ClusterDimensions   basetypes.ListValue   `tfsdk:"cluster_dimensions"`
}

func (r *createAzureMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_monitoring_provider"
}

func (r *createAzureMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an Azure Monitor monitoring provider for an Azure account. Uses the account's service principal credentials automatically.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Monitoring provider ID.",
			},
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Sedai account ID to associate this monitoring provider with.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Monitoring provider name (populated by Sedai).",
			},
			"integration_type": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Integration type (populated from the account).",
			},
			"lb_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Load balancer dimension filters.",
			},
			"app_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Application dimension filters.",
			},
			"instance_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Instance dimension filters.",
			},
			"region_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Region dimension filters.",
			},
			"container_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Container dimension filters.",
			},
			"namespace_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Namespace dimension filters.",
			},
			"cluster_dimensions": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Description: "Cluster dimension filters.",
			},
		},
	}
}

func (r *createAzureMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan azureMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpRequest := buildAzureMonitoringRequest(plan)
	response, err := monitoringProvider.AddAzureMonitoringProvider(mpRequest)
	if err != nil {
		if found := verifyMonitoringProviderCreated(plan.AccountId.ValueString(), "AZUREMONITOR"); found != nil {
			addVerifyWarning(resp, "Azure monitoring provider", plan.AccountId.ValueString(), found["id"].(string))
			plan.ID = basetypes.NewStringValue(found["id"].(string))
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}
		resp.Diagnostics.AddError("Unable to create Azure monitoring provider", err.Error())
		return
	}

	setAzureMonitoringState(ctx, &plan, response)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createAzureMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state azureMonitoringProviderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetched, err := monitoringProvider.GetMonitoringProviderById(state.ID.ValueString())
	if err != nil {
		var notFound *impl.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Azure monitoring provider", err.Error())
		return
	}
	if fetched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	setAzureMonitoringState(ctx, &state, fetched)
	state.AccountId = basetypes.NewStringValue(fetched["accountId"].(string))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *createAzureMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan azureMonitoringProviderModel
	var state azureMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	mpRequest := buildAzureMonitoringRequest(plan)
	response, err := monitoringProvider.AddAzureMonitoringProvider(mpRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Azure monitoring provider", err.Error())
		return
	}

	setAzureMonitoringState(ctx, &plan, response)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createAzureMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state azureMonitoringProviderModel
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

	deleted, err := monitoringProvider.DeleteMonitoringProvider(state.ID.ValueString())
	if err != nil || !deleted {
		resp.Diagnostics.AddError("Unable to delete Azure monitoring provider", err.Error())
		return
	}
}

func (r *createAzureMonitoringProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildAzureMonitoringRequest(plan azureMonitoringProviderModel) monitoringProvider.CreateAzureMonitoringProviderRequest {
	req := monitoringProvider.CreateAzureMonitoringProviderRequest{
		AccountId: plan.AccountId.ValueString(),
	}

	if plan.ID.ValueString() != "" {
		req.ID = plan.ID.ValueString()
	}
	if plan.Name.ValueString() != "" {
		req.Name = plan.Name.ValueString()
	}
	if plan.IntegrationType.ValueString() != "" {
		req.IntegrationType = plan.IntegrationType.ValueString()
	}
	if !plan.LbDimensions.IsNull() {
		req.LbDimensions = convertFromBaseTypes(plan.LbDimensions)
	}
	if !plan.AppDimensions.IsNull() {
		req.AppDimensions = convertFromBaseTypes(plan.AppDimensions)
	}
	if !plan.InstanceDimensions.IsNull() {
		req.InstanceDimensions = convertFromBaseTypes(plan.InstanceDimensions)
	}
	if !plan.RegionDimensions.IsNull() {
		req.RegionDimensions = convertFromBaseTypes(plan.RegionDimensions)
	}
	if !plan.ContainerDimensions.IsNull() {
		req.ContainerDimensions = convertFromBaseTypes(plan.ContainerDimensions)
	}
	if !plan.NamespaceDimensions.IsNull() {
		req.NamespaceDimensions = convertFromBaseTypes(plan.NamespaceDimensions)
	}
	if !plan.ClusterDimensions.IsNull() {
		req.ClusterDimensions = convertFromBaseTypes(plan.ClusterDimensions)
	}

	return req
}

func setAzureMonitoringState(ctx context.Context, m *azureMonitoringProviderModel, r map[string]interface{}) {
	m.ID = basetypes.NewStringValue(r["id"].(string))
	m.Name = basetypes.NewStringValue(r["name"].(string))
	m.IntegrationType = basetypes.NewStringValue(r["integrationType"].(string))
	m.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "lbDimensions"))
	m.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "appDimensions"))
	m.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "instanceDimensions"))
	m.RegionDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "regionDimensions"))
	m.ContainerDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "containerDimensions"))
	m.NamespaceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "namespaceDimensions"))
	m.ClusterDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(r, "clusterDimensions"))
}
