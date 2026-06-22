package provider

import (
	"context"
	"errors"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/credentials"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/monitoringProvider"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource            = &createCloudWatchMonitoringProvider{}
	_ resource.ResourceWithImportState = &createCloudWatchMonitoringProvider{}
)

func CreateCloudWatchMonitoringProvider() resource.Resource {
	return &createCloudWatchMonitoringProvider{}
}

type createCloudWatchMonitoringProvider struct{}

type cloudWatchMonitoringProviderModel struct {
	ID                     basetypes.StringValue `tfsdk:"id"`
	AccountId              basetypes.StringValue `tfsdk:"account_id"`
	Name                   basetypes.StringValue `tfsdk:"name"`
	IntegrationType        basetypes.StringValue `tfsdk:"integration_type"`
	UseAccountCredentials  basetypes.BoolValue   `tfsdk:"use_account_credentials"`
	AccessKey              basetypes.StringValue `tfsdk:"access_key"`
	SecretKey              basetypes.StringValue `tfsdk:"secret_key"`
	Role                   basetypes.StringValue `tfsdk:"role"`
	ExternalId             basetypes.StringValue `tfsdk:"external_id"`
	LbDimensions           basetypes.ListValue   `tfsdk:"lb_dimensions"`
	AppDimensions          basetypes.ListValue   `tfsdk:"app_dimensions"`
	InstanceDimensions     basetypes.ListValue   `tfsdk:"instance_dimensions"`
}

func (r *createCloudWatchMonitoringProvider) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudwatch_monitoring_provider"
}

func (r *createCloudWatchMonitoringProvider) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a CloudWatch monitoring provider for an AWS account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Monitoring provider ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Sedai account ID to associate this monitoring provider with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Monitoring provider name (populated by Sedai).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"integration_type": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Integration type (populated from the account).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_account_credentials": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Use the AWS credentials from the account. Defaults to true. Set to false to provide an explicit role or access key.",
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(
						path.MatchRoot("access_key"),
						path.MatchRoot("secret_key"),
					),
				},
			},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "AWS access key. Used when `use_account_credentials = false`.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("secret_key")),
					stringvalidator.ConflictsWith(path.MatchRoot("use_account_credentials")),
				},
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "AWS secret key. Used when `use_account_credentials = false`.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("access_key")),
					stringvalidator.ConflictsWith(path.MatchRoot("use_account_credentials")),
				},
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Description: "IAM role ARN override for CloudWatch access.",
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Description: "External ID for the IAM role override.",
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
		},
	}
}

func (r *createCloudWatchMonitoringProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudWatchMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpRequest := buildCloudWatchRequest(plan)
	response, err := monitoringProvider.AddCloudWatchMonitoringProvider(mpRequest)
	if err != nil {
		if found := verifyMonitoringProviderCreated(plan.AccountId.ValueString(), "CLOUDWATCH"); found != nil {
			addVerifyWarning(resp, "CloudWatch monitoring provider", plan.AccountId.ValueString(), found["id"].(string))
			plan.ID = basetypes.NewStringValue(found["id"].(string))
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			return
		}
		resp.Diagnostics.AddError("Unable to create CloudWatch monitoring provider", err.Error())
		return
	}

	result := response.(map[string]interface{})
	plan.ID = basetypes.NewStringValue(result["id"].(string))
	plan.Name = basetypes.NewStringValue(result["name"].(string))
	plan.IntegrationType = basetypes.NewStringValue(result["integrationType"].(string))
	plan.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "lbDimensions"))
	plan.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "appDimensions"))
	plan.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "instanceDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createCloudWatchMonitoringProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudWatchMonitoringProviderModel
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
		resp.Diagnostics.AddError("Unable to read CloudWatch monitoring provider", err.Error())
		return
	}
	if fetched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = basetypes.NewStringValue(fetched["id"].(string))
	state.Name = basetypes.NewStringValue(fetched["name"].(string))
	state.IntegrationType = basetypes.NewStringValue(fetched["integrationType"].(string))
	state.AccountId = basetypes.NewStringValue(fetched["accountId"].(string))
	state.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetched, "lbDimensions"))
	state.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetched, "appDimensions"))
	state.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(fetched, "instanceDimensions"))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *createCloudWatchMonitoringProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudWatchMonitoringProviderModel
	var state cloudWatchMonitoringProviderModel
	diags := req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	mpRequest := buildCloudWatchRequest(plan)
	response, err := monitoringProvider.AddCloudWatchMonitoringProvider(mpRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update CloudWatch monitoring provider", err.Error())
		return
	}

	result := response.(map[string]interface{})
	plan.ID = basetypes.NewStringValue(result["id"].(string))
	plan.Name = basetypes.NewStringValue(result["name"].(string))
	plan.IntegrationType = basetypes.NewStringValue(result["integrationType"].(string))
	plan.LbDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "lbDimensions"))
	plan.AppDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "appDimensions"))
	plan.InstanceDimensions, _ = types.ListValueFrom(ctx, types.StringType, getDimensionFromResponse(result, "instanceDimensions"))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *createCloudWatchMonitoringProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudWatchMonitoringProviderModel
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
		resp.Diagnostics.AddError("Unable to delete CloudWatch monitoring provider", err.Error())
		return
	}
}

func (r *createCloudWatchMonitoringProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildCloudWatchRequest(plan cloudWatchMonitoringProviderModel) monitoringProvider.CreateCloudWatchMonitoringProviderRequest {
	useAccountCreds := plan.UseAccountCredentials.IsNull() || plan.UseAccountCredentials.ValueBool()

	req := monitoringProvider.CreateCloudWatchMonitoringProviderRequest{
		ID:                    plan.ID.ValueString(),
		AccountId:             plan.AccountId.ValueString(),
		UseAccountCredentials: useAccountCreds,
	}

	if !useAccountCreds {
		if plan.Role.ValueString() != "" {
			req.Credentials = credentials.NewAwsRoleCredentials(plan.Role.ValueString(), plan.ExternalId.ValueString())
		} else if plan.AccessKey.ValueString() != "" && plan.SecretKey.ValueString() != "" {
			req.Credentials = credentials.NewAwsKeyCredentials(plan.AccessKey.ValueString(), plan.SecretKey.ValueString())
		}
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

	return req
}
