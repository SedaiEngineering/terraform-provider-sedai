package provider

import (
	"context"
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
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
	_ resource.Resource                      = &groupSettings{}
	_ resource.ResourceWithImportState       = &groupSettings{}
	_ resource.ResourceWithConfigValidators  = &groupSettings{}
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
	ID               basetypes.StringValue `tfsdk:"id"`
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
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier (mirrors group_id). Required by Terraform's import mechanism.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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

// ConfigValidators moves validateTopLevelModeConflicts to plan time so the
// error surfaces before any resource is created, preventing partial-state
// failures when the invalid combination is detected mid-apply.
func (r *groupSettings) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{groupSettingsModeValidator{}}
}

type groupSettingsModeValidator struct{}

func (v groupSettingsModeValidator) Description(_ context.Context) string {
	return "Validates that top-level modes are compatible with per-resource-type blocks."
}

func (v groupSettingsModeValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v groupSettingsModeValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg groupSettingsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := validateTopLevelModeConflicts(
		cfg.AvailabilityMode.ValueString(), cfg.OptimizationMode.ValueString(),
		cfg.AppSettings, cfg.BucketSettings, cfg.VolumeSettings, cfg.ServerlessSettings,
	); err != "" {
		resp.Diagnostics.AddError("Invalid mode combination", err)
	}
}

func (r *groupSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mode conflict validation runs at plan time via ConfigValidators.

	// Settings must be initialized before the first update; the API is a
	// no-op if already initialized so we always call it from Create.
	if err := groups.InitializeGroupSettings(plan.GroupID.ValueString()); err != nil {
		// Init may have succeeded before the connection dropped — check whether
		// settings already exist before treating this as a hard failure.
		existing, _ := groups.GetGroupSettings(plan.GroupID.ValueString())
		if existing == nil {
			resp.Diagnostics.AddError("Unable to initialize group settings", err.Error())
			return
		}
		// Settings already exist on the backend — fall through to UpdateGroupSettings.
	}

	if err := groups.UpdateGroupSettings(plan.GroupID.ValueString(), groupSettingsRequestFromPlan(plan)); err != nil {
		// POST may have been processed before the connection dropped (EOF-during-POST).
		// Poll GetGroupSettings up to 3 times before treating this as a hard failure.
		adopted := false
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			existing, fetchErr := groups.GetGroupSettings(plan.GroupID.ValueString())
			if fetchErr == nil && existing != nil {
				resp.Diagnostics.AddWarning(
					"Group settings configured despite connection error",
					"Settings for group '"+plan.GroupID.ValueString()+"' were found on the "+
						"backend after a failed POST — the response was likely lost in transit. "+
						"Current state adopted; run terraform apply again to reconcile any drift.",
				)
				adopted = true
				break
			}
		}
		if !adopted {
			resp.Diagnostics.AddError("Unable to set group settings", err.Error())
			return
		}
	}

	plan.ID = plan.GroupID
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

	// Only populate mode fields when state has no prior value (null/unknown),
	// i.e. on the first read after import. When already known, the user
	// explicitly set these Required fields — don't overwrite with the backend's
	// normalized value, which can differ (e.g. DATA_PILOT → CO_PILOT) and would
	// cause perpetual drift on every subsequent plan.
	populateStringIfUnset(&state.AvailabilityMode, settings.AvailabilityMode)
	populateStringIfUnset(&state.OptimizationMode, settings.OptimizationMode)
	populateBoolIfUnset(&state.SedaiSyncEnabled, settings.SedaiSyncEnabled)

	// id mirrors group_id so terraform import can locate this resource.
	state.ID = state.GroupID

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

	if err := validateTopLevelModeConflicts(plan.AvailabilityMode.ValueString(), plan.OptimizationMode.ValueString(), plan.AppSettings, plan.BucketSettings, plan.VolumeSettings, plan.ServerlessSettings); err != "" {
		resp.Diagnostics.AddError("Invalid mode combination", err)
		return
	}

	if err := groups.UpdateGroupSettings(plan.GroupID.ValueString(), groupSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to update group settings", err.Error())
		return
	}

	plan.ID = plan.GroupID
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
// adopt existing settings. The import ID is the group_id; we write it into
// both `id` (Terraform's synthetic resource identifier) and `group_id` so
// the subsequent Read call has a non-empty group_id to query.
func (r *groupSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), req.ID)...)
}

// The *FromSDK functions below are full-populate mappers used by data sources.
// Unlike the *Refresh helpers (which only update managed/non-null fields for
// resource reads), these map every field from the SDK struct unconditionally —
// that is the correct semantic for a data source that returns everything.

func kubeAppSettingsFromSDK(s *sdksettings.KubeAppSettings) *kubeAppSettingsModel {
	if s == nil {
		return nil
	}
	return &kubeAppSettingsModel{
		AvailabilityMode:                   nullableString(s.AvailabilityMode),
		OptimizationMode:                   nullableString(s.OptimizationMode),
		OptimizationFocus:                  nullableString(s.OptimizationFocus),
		SedaiSyncEnabled:                   nullableBool(s.SedaiSyncEnabled),
		IsProd:                             nullableBool(s.IsProd),
		IsOperationAllowed:                 nullableBool(s.IsOperationAllowed),
		HorizontalScalingEnabled:           nullableBool(s.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas:       nullableInt64(s.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas:       nullableInt64(s.HorizontalScalingMaxReplicas),
		HorizontalScalingReplicaMultiplier: nullableInt64(s.HorizontalScalingReplicaMultiplier),
		VerticalScalingEnabled:             nullableBool(s.VerticalScalingEnabled),
		VerticalScalingMinCPUCores:         nullableFloat64(s.VerticalScalingMinCPUCores),
		VerticalScalingMinMemoryBytes:      nullableInt64(s.VerticalScalingMinMemoryBytes),
		PredictiveScalingEnabled:           nullableBool(s.PredictiveScalingEnabled),
		MaxLatencyIncreasePct:              nullableInt64(s.MaxLatencyIncreasePct),
		MaxCPUIncreasePct:                  nullableInt64(s.MaxCPUIncreasePct),
		MaxMemoryIncreasePct:               nullableInt64(s.MaxMemoryIncreasePct),
	}
}

func bucketSettingsFromSDK(s *sdksettings.BucketSettings) *bucketSettingsModel {
	if s == nil {
		return nil
	}
	return &bucketSettingsModel{
		OptimizationMode: nullableString(s.OptimizationMode),
		SedaiSyncEnabled: nullableBool(s.SedaiSyncEnabled),
	}
}

func appSettingsFromSDK(s *sdksettings.AppSettings) *appSettingsModel {
	if s == nil {
		return nil
	}
	return &appSettingsModel{
		AvailabilityMode:             nullableString(s.AvailabilityMode),
		OptimizationMode:             nullableString(s.OptimizationMode),
		IsProd:                       nullableBool(s.IsProd),
		SedaiSyncEnabled:             nullableBool(s.SedaiSyncEnabled),
		HorizontalScalingEnabled:     nullableBool(s.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas: nullableInt64(s.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas: nullableInt64(s.HorizontalScalingMaxReplicas),
	}
}

func containerAppSettingsFromSDK(s *sdksettings.ContainerAppSettings) *containerAppSettingsModel {
	if s == nil {
		return nil
	}
	return &containerAppSettingsModel{
		AvailabilityMode:               nullableString(s.AvailabilityMode),
		OptimizationMode:               nullableString(s.OptimizationMode),
		IsProd:                         nullableBool(s.IsProd),
		SedaiSyncEnabled:               nullableBool(s.SedaiSyncEnabled),
		HorizontalScalingEnabled:       nullableBool(s.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas:   nullableInt64(s.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas:   nullableInt64(s.HorizontalScalingMaxReplicas),
		VerticalScalingEnabled:         nullableBool(s.VerticalScalingEnabled),
		PredictiveScalingEnabled:       nullableBool(s.PredictiveScalingEnabled),
		OptimizationFocus:              nullableString(s.OptimizationFocus),
		MaxLatencyIncreasePct:          nullableInt64(s.MaxLatencyIncreasePct),
		MaxCPUIncreasePct:              nullableInt64(s.MaxCPUIncreasePct),
		MaxMemoryIncreasePct:           nullableInt64(s.MaxMemoryIncreasePct),
		AutonomousActionWithoutTraffic: nullableBool(s.AutonomousActionWithoutTraffic),
	}
}

func ecsAppSettingsFromSDK(s *sdksettings.ECSAppSettings) *ecsAppSettingsModel {
	if s == nil {
		return nil
	}
	return &ecsAppSettingsModel{
		AvailabilityMode:                  nullableString(s.AvailabilityMode),
		OptimizationMode:                  nullableString(s.OptimizationMode),
		IsProd:                            nullableBool(s.IsProd),
		SedaiSyncEnabled:                  nullableBool(s.SedaiSyncEnabled),
		HorizontalScalingEnabled:          nullableBool(s.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas:      nullableInt64(s.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas:      nullableInt64(s.HorizontalScalingMaxReplicas),
		VerticalScalingEnabled:            nullableBool(s.VerticalScalingEnabled),
		PredictiveScalingEnabled:          nullableBool(s.PredictiveScalingEnabled),
		OptimizationFocus:                 nullableString(s.OptimizationFocus),
		MaxLatencyIncreasePct:             nullableInt64(s.MaxLatencyIncreasePct),
		MaxCPUIncreasePct:                 nullableInt64(s.MaxCPUIncreasePct),
		MaxMemoryIncreasePct:              nullableInt64(s.MaxMemoryIncreasePct),
		ServiceAutoscalingEnabled:         nullableBool(s.ServiceAutoscalingEnabled),
		HorizontalScalingReplicaIncrement: nullableInt64(s.HorizontalScalingReplicaIncrement),
		VerticalScalingMinCPU:             nullableInt64(s.VerticalScalingMinCPU),
		VerticalScalingMinMemory:          nullableInt64(s.VerticalScalingMinMemory),
	}
}

func serverlessSettingsFromSDK(s *sdksettings.ServerlessSettings) *serverlessSettingsModel {
	if s == nil {
		return nil
	}
	return &serverlessSettingsModel{
		AvailabilityMode:    nullableString(s.AvailabilityMode),
		OptimizationMode:    nullableString(s.OptimizationMode),
		OptimizationFocus:   nullableString(s.OptimizationFocus),
		ConcurrencyMode:     nullableString(s.ConcurrencyMode),
		MaxCostChangePct:    nullableInt64(s.MaxCostChangePct),
		MaxLatencyChangePct: nullableInt64(s.MaxLatencyChangePct),
		SedaiSyncEnabled:    nullableBool(s.SedaiSyncEnabled),
	}
}

func volumeSettingsFromSDK(s *sdksettings.VolumeSettings) *volumeSettingsModel {
	if s == nil {
		return nil
	}
	return &volumeSettingsModel{
		AvailabilityMode: nullableString(s.AvailabilityMode),
		OptimizationMode: nullableString(s.OptimizationMode),
		SedaiSyncEnabled: nullableBool(s.SedaiSyncEnabled),
	}
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

// validateTopLevelModeConflicts checks that the top-level availability/optimization
// modes are compatible with any per-type blocks the user is managing. The SDK fans
// the top-level mode into every section — if a section does not support a mode, writing
// it causes silent undefined behavior or perpetual drift. Returns "" when valid.
//
// Rules (per spec and provider validators):
//   - app_settings, bucket_settings, volume_settings: no AUTO for either mode
//   - bucket_settings: no availability_mode at all (backend ignores it)
//   - volume_settings: no AUTO for either mode
//   - serverless_settings: no CO_PILOT for optimization_mode
func validateTopLevelModeConflicts(availMode, optMode string,
	appSettings *appSettingsModel,
	bucketSettings *bucketSettingsModel,
	volumeSettings *volumeSettingsModel,
	serverlessSettings *serverlessSettingsModel,
) string {
	if appSettings != nil {
		if availMode == "AUTO" {
			return "availability_mode = AUTO cannot be combined with an app_settings block. " +
				"The app type does not support AUTO. Use DATA_PILOT or CO_PILOT at the top level, " +
				"or set app_settings.availability_mode separately."
		}
		if optMode == "AUTO" {
			return "optimization_mode = AUTO cannot be combined with an app_settings block. " +
				"The app type does not support AUTO."
		}
	}
	if bucketSettings != nil && optMode == "AUTO" {
		return "optimization_mode = AUTO cannot be combined with a bucket_settings block. " +
			"S3 buckets do not support AUTO — use DATA_PILOT or CO_PILOT."
	}
	if volumeSettings != nil {
		if availMode == "AUTO" {
			return "availability_mode = AUTO cannot be combined with a volume_settings block. " +
				"EBS volumes do not support AUTO — use DATA_PILOT or CO_PILOT."
		}
		if optMode == "AUTO" {
			return "optimization_mode = AUTO cannot be combined with a volume_settings block. " +
				"EBS volumes do not support AUTO — use DATA_PILOT or CO_PILOT."
		}
	}
	if serverlessSettings != nil && optMode == "CO_PILOT" {
		return "optimization_mode = CO_PILOT cannot be combined with a serverless_settings block. " +
			"Lambda functions do not support CO_PILOT — use DATA_PILOT or AUTO."
	}
	return ""
}
