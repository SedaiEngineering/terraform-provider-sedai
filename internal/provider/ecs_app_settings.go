package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ecsAppSettingsModel is the tfsdk view of `ecs_app_settings { … }`.
// Inheritance: ecs_app inherits container_app (which inherits app), so
// the field list here is app + container + ECS-specific.
//
// Type difference vs kube: ECS uses integer CPU "units" (1024 = 1 vCPU)
// and integer memory MB, not fractional cores and bytes. That's why
// vertical_scaling_min_cpu and _min_memory are Int64 here vs Float64 +
// Int64 on kube_app_settings.
type ecsAppSettingsModel struct {
	// Inherited from app + container.
	AvailabilityMode               basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode               basetypes.StringValue `tfsdk:"optimization_mode"`
	IsProd                         basetypes.BoolValue   `tfsdk:"is_prod"`
	SedaiSyncEnabled               basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
	HorizontalScalingEnabled       basetypes.BoolValue   `tfsdk:"horizontal_scaling_enabled"`
	HorizontalScalingMinReplicas   basetypes.Int64Value  `tfsdk:"horizontal_scaling_min_replicas"`
	HorizontalScalingMaxReplicas   basetypes.Int64Value  `tfsdk:"horizontal_scaling_max_replicas"`
	VerticalScalingEnabled         basetypes.BoolValue   `tfsdk:"vertical_scaling_enabled"`
	PredictiveScalingEnabled       basetypes.BoolValue   `tfsdk:"predictive_scaling_enabled"`
	OptimizationFocus              basetypes.StringValue `tfsdk:"optimization_focus"`
	MaxLatencyIncreasePct          basetypes.Int64Value  `tfsdk:"max_latency_increase_pct"`
	MaxCPUIncreasePct              basetypes.Int64Value  `tfsdk:"max_cpu_increase_pct"`
	MaxMemoryIncreasePct           basetypes.Int64Value  `tfsdk:"max_memory_increase_pct"`

	// ECS-specific.
	ServiceAutoscalingEnabled         basetypes.BoolValue  `tfsdk:"service_autoscaling_enabled"`
	HorizontalScalingReplicaIncrement basetypes.Int64Value `tfsdk:"horizontal_scaling_replica_increment"`
	VerticalScalingMinCPU             basetypes.Int64Value `tfsdk:"vertical_scaling_min_cpu"`
	VerticalScalingMinMemory          basetypes.Int64Value `tfsdk:"vertical_scaling_min_memory"`
}

// ecsAppSettingsBlock returns the schema for `ecs_app_settings { … }`.
// Reuses appSettingsBaseAttributes for the inherited app fields and
// adds the container + ECS-specific extensions.
func ecsAppSettingsBlock() schema.SingleNestedBlock {
	attrs := appSettingsBaseAttributes(settingsConfigModeValidator())

	// Container extensions.
	attrs["vertical_scaling_enabled"] = schema.BoolAttribute{Optional: true, Description: "Enable vertical scaling (CPU / memory right-sizing)."}
	attrs["predictive_scaling_enabled"] = schema.BoolAttribute{Optional: true, Description: "Enable predictive scaling."}
	attrs["optimization_focus"] = schema.StringAttribute{
		Optional:    true,
		Description: "What to optimize for. Valid: `COST`, `DURATION`, `COST_AND_DURATION`.",
		Validators:  []validator.String{optimizationFocusValidator()},
	}
	attrs["max_latency_increase_pct"] = schema.Int64Attribute{Optional: true, Description: "Guardrail: max latency increase %."}
	attrs["max_cpu_increase_pct"] = schema.Int64Attribute{Optional: true, Description: "Guardrail: max CPU increase %."}
	attrs["max_memory_increase_pct"] = schema.Int64Attribute{Optional: true, Description: "Guardrail: max memory increase %."}
	// ECS-specific extensions.
	attrs["service_autoscaling_enabled"] = schema.BoolAttribute{Optional: true, Description: "Enable ECS service autoscaling."}
	attrs["horizontal_scaling_replica_increment"] = schema.Int64Attribute{Optional: true, Description: "Fixed number of tasks added or removed per scaling event."}
	attrs["vertical_scaling_min_cpu"] = schema.Int64Attribute{Optional: true, Description: "Floor for CPU per task. ECS units (1024 = 1 vCPU)."}
	attrs["vertical_scaling_min_memory"] = schema.Int64Attribute{Optional: true, Description: "Floor for memory per task, in MB."}

	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS ECS service workloads.",
		Attributes:  attrs,
	}
}

// ecsAppSettingsToSDK converts the tfsdk model to the SDK struct.
func ecsAppSettingsToSDK(m *ecsAppSettingsModel) *sdksettings.ECSAppSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.ECSAppSettings{
		ContainerAppSettings: sdksettings.ContainerAppSettings{
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
		},
		ServiceAutoscalingEnabled:         boolPtr(m.ServiceAutoscalingEnabled),
		HorizontalScalingReplicaIncrement: int64Ptr(m.HorizontalScalingReplicaIncrement),
		VerticalScalingMinCPU:             int64Ptr(m.VerticalScalingMinCPU),
		VerticalScalingMinMemory:          int64Ptr(m.VerticalScalingMinMemory),
	}
}

// ecsAppSettingsRefresh updates only the managed fields.
func ecsAppSettingsRefresh(state *ecsAppSettingsModel, fetched *sdksettings.ECSAppSettings) *ecsAppSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	// Inherited app + container.
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshBoolIfManaged(&state.IsProd, fetched.IsProd)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	refreshBoolIfManaged(&state.HorizontalScalingEnabled, fetched.HorizontalScalingEnabled)
	refreshInt64IfManaged(&state.HorizontalScalingMinReplicas, fetched.HorizontalScalingMinReplicas)
	refreshInt64IfManaged(&state.HorizontalScalingMaxReplicas, fetched.HorizontalScalingMaxReplicas)
	refreshBoolIfManaged(&state.VerticalScalingEnabled, fetched.VerticalScalingEnabled)
	refreshBoolIfManaged(&state.PredictiveScalingEnabled, fetched.PredictiveScalingEnabled)
	refreshIfManaged(&state.OptimizationFocus, fetched.OptimizationFocus)
	refreshInt64IfManaged(&state.MaxLatencyIncreasePct, fetched.MaxLatencyIncreasePct)
	refreshInt64IfManaged(&state.MaxCPUIncreasePct, fetched.MaxCPUIncreasePct)
	refreshInt64IfManaged(&state.MaxMemoryIncreasePct, fetched.MaxMemoryIncreasePct)

	// ECS-specific.
	refreshBoolIfManaged(&state.ServiceAutoscalingEnabled, fetched.ServiceAutoscalingEnabled)
	refreshInt64IfManaged(&state.HorizontalScalingReplicaIncrement, fetched.HorizontalScalingReplicaIncrement)
	refreshInt64IfManaged(&state.VerticalScalingMinCPU, fetched.VerticalScalingMinCPU)
	refreshInt64IfManaged(&state.VerticalScalingMinMemory, fetched.VerticalScalingMinMemory)
	return state
}
