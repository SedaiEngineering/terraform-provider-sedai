package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
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
	_ resource.Resource                = &groupSettings{}
	_ resource.ResourceWithImportState = &groupSettings{}
)

// GroupSettings is the resource constructor for `sedai_group_settings`.
func GroupSettings() resource.Resource {
	return &groupSettings{}
}

// groupSettings is the resource implementation. Resource type name:
// `sedai_group_settings`. Manages the top-level settings (availability /
// optimization mode + Sedai Sync toggle) for a single Sedai group.
//
// In this iteration the schema exposes the three top-level modes only;
// per-resource-type tuning (kube scaling, ECS, lambda, …) is preserved
// untouched on the backend by the SDK's read-modify-write logic. A later
// iteration will add the per-type nested blocks the requirements doc
// specifies (kube_app_settings, serverless_settings, etc.).
type groupSettings struct{}

type groupSettingsResourceModel struct {
	GroupID          basetypes.StringValue `tfsdk:"group_id"`
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`

	// Per-resource-type override blocks. Each is nil when the user omits
	// the corresponding `xxx_settings { … }` HCL block. Non-nil means
	// "manage at least one field in this block" — field-level partial
	// spec is enforced inside the model (null fields = unmanaged).
	KubeAppSettings      *kubeAppSettingsModel      `tfsdk:"kube_app_settings"`
	BucketSettings       *bucketSettingsModel       `tfsdk:"bucket_settings"`
	AppSettings          *appSettingsModel          `tfsdk:"app_settings"`
	ContainerAppSettings *containerAppSettingsModel `tfsdk:"container_app_settings"`
	ECSAppSettings       *ecsAppSettingsModel       `tfsdk:"ecs_app_settings"`
	ServerlessSettings   *serverlessSettingsModel   `tfsdk:"serverless_settings"`
	VolumeSettings       *volumeSettingsModel       `tfsdk:"volume_settings"`
}

func (r *groupSettings) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_settings"
}

func (r *groupSettings) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the top-level settings for a Sedai group. The provider auto-initializes the group's settings on first apply, so the customer never has to call the init API directly. Only attributes specified in this resource are tracked for drift — unmanaged per-resource-type tuning (kube scaling, etc.) is preserved as-is.",
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the group to configure. Typically `sedai_group.<name>.id`.",
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
				Description: "When true, Sedai auto-syncs the group's resources with the latest configuration. Defaults to false if omitted.",
			},
		},
		Blocks: map[string]schema.Block{
			// Per-resource-type override blocks. Add additional blocks here
			// as they land (serverless_settings, volume_settings, …) — each
			// is opt-in for the customer and field-level partial.
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

func (r *groupSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Settings must be initialized before the first update; the API is a
	// no-op if already initialized so we always call it from Create.
	if err := groups.InitializeGroupSettings(plan.GroupID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to initialize group settings", err.Error())
		return
	}

	if err := groups.UpdateGroupSettings(plan.GroupID.ValueString(), groupSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to set group settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := groups.GetGroupSettings(state.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read group settings",
			"Could not fetch settings for group "+state.GroupID.ValueString()+": "+err.Error(),
		)
		return
	}
	if settings == nil {
		// Settings not initialized — could mean the group itself is gone, OR
		// that init was never called. Drop from state so the next plan re-creates.
		resp.State.RemoveResource(ctx)
		return
	}

	state.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
	state.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
	state.SedaiSyncEnabled = basetypes.NewBoolValue(settings.SedaiSyncEnabled)

	// Refresh each per-resource-type block ONLY for fields the user is
	// already managing. Helper handles nil state block (user didn't include
	// the HCL block) and the field-level partial-spec contract.
	kubeAppSettingsRefresh(state.KubeAppSettings, settings.KubeAppSettings)
	bucketSettingsRefresh(state.BucketSettings, settings.BucketSettings)
	appSettingsRefresh(state.AppSettings, settings.AppSettings)
	containerAppSettingsRefresh(state.ContainerAppSettings, settings.ContainerAppSettings)
	ecsAppSettingsRefresh(state.ECSAppSettings, settings.ECSAppSettings)
	serverlessSettingsRefresh(state.ServerlessSettings, settings.ServerlessSettings)
	volumeSettingsRefresh(state.VolumeSettings, settings.VolumeSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *groupSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := groups.UpdateGroupSettings(plan.GroupID.ValueString(), groupSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to update group settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete resets the managed modes to DATA_PILOT (the most conservative,
// observe-only setting), which is the practical equivalent of "reset to
// inherited defaults" given the backend has no explicit reset endpoint.
// The group itself is left intact. Per-resource-type tuning untouched.
func (r *groupSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reset := &groups.GroupSettings{
		AvailabilityMode: "DATA_PILOT",
		OptimizationMode: "DATA_PILOT",
		SedaiSyncEnabled: false,
	}
	if err := groups.UpdateGroupSettings(state.GroupID.ValueString(), reset); err != nil {
		// Don't fail destroy on a settings reset error — the group still
		// owns its lifecycle. Log a warning instead.
		resp.Diagnostics.AddWarning("Unable to reset group settings on destroy", err.Error())
	}
}

// ImportState lets `terraform import sedai_group_settings.<name> <group-id>`
// adopt existing settings. The supplied ID is written into group_id; the
// Read flow then fetches the current values.
func (r *groupSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_id"), req, resp)
}

func groupSettingsRequestFromPlan(p groupSettingsResourceModel) *groups.GroupSettings {
	return &groups.GroupSettings{
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
