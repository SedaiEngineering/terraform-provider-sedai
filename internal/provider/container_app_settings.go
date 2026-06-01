package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// containerAppSettingsModel is the tfsdk view of a
// `container_app_settings { … }` block. Mirrors the spec's "inherits app
// + adds container-specific" shape: all 7 app fields PLUS 7 container-
// specific extensions.
//
// We can't use Go struct embedding here — the plugin-framework's reflect
// can be finicky with anonymous fields. The field list is duplicated
// from appSettingsModel and that's the price for tooling compatibility.
type containerAppSettingsModel struct {
	// Inherited from app_settings:
	AvailabilityMode             basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode             basetypes.StringValue `tfsdk:"optimization_mode"`
	IsProd                       basetypes.BoolValue   `tfsdk:"is_prod"`
	SedaiSyncEnabled             basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
	HorizontalScalingEnabled     basetypes.BoolValue   `tfsdk:"horizontal_scaling_enabled"`
	HorizontalScalingMinReplicas basetypes.Int64Value  `tfsdk:"horizontal_scaling_min_replicas"`
	HorizontalScalingMaxReplicas basetypes.Int64Value  `tfsdk:"horizontal_scaling_max_replicas"`

	// Container-specific extensions:
	VerticalScalingEnabled         basetypes.BoolValue   `tfsdk:"vertical_scaling_enabled"`
	PredictiveScalingEnabled       basetypes.BoolValue   `tfsdk:"predictive_scaling_enabled"`
	OptimizationFocus              basetypes.StringValue `tfsdk:"optimization_focus"`
	MaxLatencyIncreasePct          basetypes.Int64Value  `tfsdk:"max_latency_increase_pct"`
	MaxCPUIncreasePct              basetypes.Int64Value  `tfsdk:"max_cpu_increase_pct"`
	MaxMemoryIncreasePct           basetypes.Int64Value  `tfsdk:"max_memory_increase_pct"`
	AutonomousActionWithoutTraffic basetypes.BoolValue   `tfsdk:"autonomous_action_without_traffic"`
}

// containerAppSettingsBlock returns the schema for the
// `container_app_settings { … }` block. Container-app modes allow AUTO
// per spec (unlike base app_settings). Schema is built by extending the
// shared app base attributes — same source of truth.
func containerAppSettingsBlock() schema.SingleNestedBlock {
	// Start from the shared base; container_app uses the full mode set.
	attrs := appSettingsBaseAttributes(settingsConfigModeValidator())

	// Add container-specific extensions on top.
	attrs["vertical_scaling_enabled"] = schema.BoolAttribute{
		Optional:    true,
		Description: "Enable vertical scaling (CPU / memory right-sizing).",
	}
	attrs["predictive_scaling_enabled"] = schema.BoolAttribute{
		Optional:    true,
		Description: "Enable predictive scaling based on detected seasonality.",
	}
	attrs["optimization_focus"] = schema.StringAttribute{
		Optional:    true,
		Description: "What to optimize for. Valid: `COST`, `DURATION`, `COST_AND_DURATION`.",
		Validators:  []validator.String{optimizationFocusValidator()},
	}
	attrs["max_latency_increase_pct"] = schema.Int64Attribute{
		Optional:    true,
		Description: "Guardrail: maximum acceptable latency increase % during optimization.",
	}
	attrs["max_cpu_increase_pct"] = schema.Int64Attribute{
		Optional:    true,
		Description: "Guardrail: maximum acceptable CPU increase % during optimization.",
	}
	attrs["max_memory_increase_pct"] = schema.Int64Attribute{
		Optional:    true,
		Description: "Guardrail: maximum acceptable memory increase % during optimization.",
	}
	attrs["autonomous_action_without_traffic"] = schema.BoolAttribute{
		Optional:    true,
		Description: "Allow Sedai to optimize even without recent traffic data.",
	}

	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for generic containerized workloads (containers running outside Kubernetes or ECS). Allows the full mode set including AUTO. For workload-specific tuning, use kube_app_settings or ecs_app_settings.",
		Attributes:  attrs,
	}
}

// containerAppSettingsToSDK converts the tfsdk model to the SDK struct.
// The embedded AppSettings inside ContainerAppSettings gets the
// inherited fields; the rest are container-specific.
func containerAppSettingsToSDK(m *containerAppSettingsModel) *sdksettings.ContainerAppSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.ContainerAppSettings{
		AppSettings: sdksettings.AppSettings{
			AvailabilityMode:             stringPtr(m.AvailabilityMode),
			OptimizationMode:             stringPtr(m.OptimizationMode),
			IsProd:                       boolPtr(m.IsProd),
			SedaiSyncEnabled:             boolPtr(m.SedaiSyncEnabled),
			HorizontalScalingEnabled:     boolPtr(m.HorizontalScalingEnabled),
			HorizontalScalingMinReplicas: int64Ptr(m.HorizontalScalingMinReplicas),
			HorizontalScalingMaxReplicas: int64Ptr(m.HorizontalScalingMaxReplicas),
		},
		VerticalScalingEnabled:         boolPtr(m.VerticalScalingEnabled),
		PredictiveScalingEnabled:       boolPtr(m.PredictiveScalingEnabled),
		OptimizationFocus:              stringPtr(m.OptimizationFocus),
		MaxLatencyIncreasePct:          int64Ptr(m.MaxLatencyIncreasePct),
		MaxCPUIncreasePct:              int64Ptr(m.MaxCPUIncreasePct),
		MaxMemoryIncreasePct:           int64Ptr(m.MaxMemoryIncreasePct),
		AutonomousActionWithoutTraffic: boolPtr(m.AutonomousActionWithoutTraffic),
	}
}

// containerAppSettingsRefresh updates only the managed fields.
func containerAppSettingsRefresh(state *containerAppSettingsModel, fetched *sdksettings.ContainerAppSettings) *containerAppSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	// Inherited app fields.
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshBoolIfManaged(&state.IsProd, fetched.IsProd)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	refreshBoolIfManaged(&state.HorizontalScalingEnabled, fetched.HorizontalScalingEnabled)
	refreshInt64IfManaged(&state.HorizontalScalingMinReplicas, fetched.HorizontalScalingMinReplicas)
	refreshInt64IfManaged(&state.HorizontalScalingMaxReplicas, fetched.HorizontalScalingMaxReplicas)

	// Container-specific extensions.
	refreshBoolIfManaged(&state.VerticalScalingEnabled, fetched.VerticalScalingEnabled)
	refreshBoolIfManaged(&state.PredictiveScalingEnabled, fetched.PredictiveScalingEnabled)
	refreshIfManaged(&state.OptimizationFocus, fetched.OptimizationFocus)
	refreshInt64IfManaged(&state.MaxLatencyIncreasePct, fetched.MaxLatencyIncreasePct)
	refreshInt64IfManaged(&state.MaxCPUIncreasePct, fetched.MaxCPUIncreasePct)
	refreshInt64IfManaged(&state.MaxMemoryIncreasePct, fetched.MaxMemoryIncreasePct)
	refreshBoolIfManaged(&state.AutonomousActionWithoutTraffic, fetched.AutonomousActionWithoutTraffic)
	return state
}
