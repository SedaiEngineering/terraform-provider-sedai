package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure interfaces are satisfied.
var (
	_ resource.Resource                = &accountSettings{}
	_ resource.ResourceWithImportState = &accountSettings{}
)

// AccountSettings is the resource constructor for `sedai_account_settings`.
func AccountSettings() resource.Resource {
	return &accountSettings{}
}

// accountSettings is the resource implementation. Resource type name:
// `sedai_account_settings`. Manages the top-level settings (availability /
// optimization mode + Sedai Sync toggle) for a single Sedai account.
//
// Unlike group settings, accounts come with settings already initialized,
// so there's no separate init call. Drift detection covers only the three
// modes; per-resource-type tuning on the backend is preserved untouched.
type accountSettings struct{}

type accountSettingsResourceModel struct {
	AccountID        basetypes.StringValue `tfsdk:"account_id"`
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`

	// Per-resource-type override blocks — same 7 as sedai_group_settings.
	// Mirror schema + behavior; field-level partial spec inside each.
	KubeAppSettings      *kubeAppSettingsModel      `tfsdk:"kube_app_settings"`
	BucketSettings       *bucketSettingsModel       `tfsdk:"bucket_settings"`
	AppSettings          *appSettingsModel          `tfsdk:"app_settings"`
	ContainerAppSettings *containerAppSettingsModel `tfsdk:"container_app_settings"`
	ECSAppSettings       *ecsAppSettingsModel       `tfsdk:"ecs_app_settings"`
	ServerlessSettings   *serverlessSettingsModel   `tfsdk:"serverless_settings"`
	VolumeSettings       *volumeSettingsModel       `tfsdk:"volume_settings"`
}

func (r *accountSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_settings"
}

func (r *accountSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the top-level settings for a Sedai account. Applies as the baseline default for every resource in the account. Only attributes specified here are tracked for drift — unmanaged per-resource-type tuning is preserved as-is.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "The Sedai account ID to configure. Typically `sedai_account.<name>.id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_mode": schema.StringAttribute{
				Required:    true,
				Description: "Availability mode. Valid values: `DATA_PILOT`, `CO_PILOT`, `AUTO`.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"optimization_mode": schema.StringAttribute{
				Required:    true,
				Description: "Optimization mode. Valid values: `DATA_PILOT`, `CO_PILOT`, `AUTO`.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, Sedai auto-syncs the account's resources with the latest configuration. Defaults to false if omitted.",
			},
		},
		Blocks: map[string]schema.Block{
			// Per-resource-type override blocks — same 7 schemas as
			// sedai_group_settings. Block-builders are reused; no
			// duplication of block field definitions.
			"kube_app_settings":      kubeAppSettingsBlock(),
			"bucket_settings":        bucketSettingsBlock(),
			"app_settings":           appSettingsBlock(),
			"container_app_settings": containerAppSettingsBlock(),
			"ecs_app_settings":       ecsAppSettingsBlock(),
			"serverless_settings":    serverlessSettingsBlock(),
			"volume_settings":        volumeSettingsBlock(),
		},
	}
}

func (r *accountSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateTopLevelModeConflicts(plan.AvailabilityMode.ValueString(), plan.OptimizationMode.ValueString(), plan.AppSettings, plan.BucketSettings, plan.VolumeSettings, plan.ServerlessSettings); err != "" {
		resp.Diagnostics.AddError("Invalid mode combination", err)
		return
	}

	if err := account.UpdateAccountSettings(plan.AccountID.ValueString(), accountSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to set account settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *accountSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := account.GetAccountSettings(state.AccountID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read account settings",
			"Could not fetch settings for account "+state.AccountID.ValueString()+": "+err.Error(),
		)
		return
	}
	if settings == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
	state.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
	state.SedaiSyncEnabled = basetypes.NewBoolValue(settings.SedaiSyncEnabled)

	// Refresh per-resource-type blocks (field-level partial spec).
	kubeAppSettingsRefresh(state.KubeAppSettings, settings.KubeAppSettings)
	bucketSettingsRefresh(state.BucketSettings, settings.BucketSettings)
	appSettingsRefresh(state.AppSettings, settings.AppSettings)
	containerAppSettingsRefresh(state.ContainerAppSettings, settings.ContainerAppSettings)
	ecsAppSettingsRefresh(state.ECSAppSettings, settings.ECSAppSettings)
	serverlessSettingsRefresh(state.ServerlessSettings, settings.ServerlessSettings)
	volumeSettingsRefresh(state.VolumeSettings, settings.VolumeSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *accountSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateTopLevelModeConflicts(plan.AvailabilityMode.ValueString(), plan.OptimizationMode.ValueString(), plan.AppSettings, plan.BucketSettings, plan.VolumeSettings, plan.ServerlessSettings); err != "" {
		resp.Diagnostics.AddError("Invalid mode combination", err)
		return
	}

	if err := account.UpdateAccountSettings(plan.AccountID.ValueString(), accountSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to update account settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete resets the managed modes to DATA_PILOT (the most conservative,
// observe-only setting), the practical equivalent of "reset to platform
// defaults" given there's no explicit reset endpoint. The account itself
// stays intact.
func (r *accountSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reset := &account.AccountSettings{
		AvailabilityMode: "DATA_PILOT",
		OptimizationMode: "DATA_PILOT",
		SedaiSyncEnabled: false,
	}
	if err := account.UpdateAccountSettings(state.AccountID.ValueString(), reset); err != nil {
		resp.Diagnostics.AddWarning("Unable to reset account settings on destroy", err.Error())
	}
}

func (r *accountSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("account_id"), req, resp)
}

func accountSettingsRequestFromPlan(p accountSettingsResourceModel) *account.AccountSettings {
	return &account.AccountSettings{
		AvailabilityMode:     p.AvailabilityMode.ValueString(),
		OptimizationMode:     p.OptimizationMode.ValueString(),
		SedaiSyncEnabled:     p.SedaiSyncEnabled.ValueBool(),
		KubeAppSettings:      kubeAppSettingsToSDK(p.KubeAppSettings),
		BucketSettings:       bucketSettingsToSDK(p.BucketSettings),
		AppSettings:          appSettingsToSDK(p.AppSettings),
		ContainerAppSettings: containerAppSettingsToSDK(p.ContainerAppSettings),
		ECSAppSettings:       ecsAppSettingsToSDK(p.ECSAppSettings),
		ServerlessSettings:   serverlessSettingsToSDK(p.ServerlessSettings),
		VolumeSettings:       volumeSettingsToSDK(p.VolumeSettings),
	}
}
