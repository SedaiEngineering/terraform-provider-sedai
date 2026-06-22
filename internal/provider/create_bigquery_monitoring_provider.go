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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource            = &createBigQueryMonitoringProvider{}
	_ resource.ResourceWithImportState = &createBigQueryMonitoringProvider{}
)

func CreateBigQueryMonitoringProvider() resource.Resource {
	return &createBigQueryMonitoringProvider{}
}

type createBigQueryMonitoringProvider struct{}

type bigQueryMonitoringProviderModel struct {
	ID              basetypes.StringValue `tfsdk:"id"`
	AccountId       basetypes.StringValue `tfsdk:"account_id"`
	ProjectId       basetypes.StringValue `tfsdk:"project_id"`
	Name            basetypes.StringValue `tfsdk:"name"`
	IntegrationType basetypes.StringValue `tfsdk:"integration_type"`
}

func (r *createBigQueryMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_monitoring_provider"
}

func (r *createBigQueryMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"integration_type": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *createBigQueryMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bigQueryMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := monitoringProvider.AddBigQueryMonitoringProvider(buildBigQueryRequest(plan))
	if err != nil {
		if found := verifyMonitoringProviderCreated(plan.AccountId.ValueString(), "BQMONITORING"); found != nil {
			addVerifyWarning(resp, "BigQuery monitoring provider", plan.AccountId.ValueString(), found["id"].(string))
			plan.ID = basetypes.NewStringValue(found["id"].(string))
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}
		resp.Diagnostics.AddError("Unable to create BigQuery monitoring provider", err.Error())
		return
	}

	setBigQueryState(&plan, response)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createBigQueryMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bigQueryMonitoringProviderModel
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
		resp.Diagnostics.AddError("Unable to read BigQuery monitoring provider", err.Error())
		return
	}

	setBigQueryState(&state, fetched)
	state.AccountId = basetypes.NewStringValue(fetched["accountId"].(string))
	if details, ok := fetched["details"].(map[string]interface{}); ok {
		state.ProjectId = basetypes.NewStringValue(details["projectId"].(string))
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *createBigQueryMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bigQueryMonitoringProviderModel
	var state bigQueryMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	response, err := monitoringProvider.AddBigQueryMonitoringProvider(buildBigQueryRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Unable to update BigQuery monitoring provider", err.Error())
		return
	}

	setBigQueryState(&plan, response)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createBigQueryMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bigQueryMonitoringProviderModel
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

	if err := deleteMPGracefully(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete BigQuery monitoring provider", err.Error())
		return
	}
}

func (r *createBigQueryMonitoringProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildBigQueryRequest(plan bigQueryMonitoringProviderModel) monitoringProvider.CreateBigQueryMonitoringProviderRequest {
	req := monitoringProvider.CreateBigQueryMonitoringProviderRequest{
		AccountId: plan.AccountId.ValueString(),
		ProjectId: plan.ProjectId.ValueString(),
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
	return req
}

func setBigQueryState(m *bigQueryMonitoringProviderModel, r map[string]interface{}) {
	m.ID = basetypes.NewStringValue(r["id"].(string))
	m.Name = basetypes.NewStringValue(r["name"].(string))
	m.IntegrationType = basetypes.NewStringValue(r["integrationType"].(string))
}
