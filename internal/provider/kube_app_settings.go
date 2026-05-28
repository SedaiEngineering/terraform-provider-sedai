package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// kubeAppSettingsModel is the tfsdk view of a `kube_app_settings { … }`
// block. All fields are Optional in the schema, so all tfsdk types are
// the basetypes wrappers — null/known/unknown awareness is needed for
// field-level partial spec ("manage these fields, ignore the rest").
//
// One source of truth: this struct AND the schema function below AND the
// SDK's settings.KubeAppSettings must stay in lockstep on field names.
// When adding a new field, update all three or none.
type kubeAppSettingsModel struct {
	AvailabilityMode                   basetypes.StringValue  `tfsdk:"availability_mode"`
	OptimizationMode                   basetypes.StringValue  `tfsdk:"optimization_mode"`
	OptimizationFocus                  basetypes.StringValue  `tfsdk:"optimization_focus"`
	SedaiSyncEnabled                   basetypes.BoolValue    `tfsdk:"sedai_sync_enabled"`
	IsProd                             basetypes.BoolValue    `tfsdk:"is_prod"`
	IsOperationAllowed                 basetypes.BoolValue    `tfsdk:"is_operation_allowed"`
	HorizontalScalingEnabled           basetypes.BoolValue    `tfsdk:"horizontal_scaling_enabled"`
	HorizontalScalingMinReplicas       basetypes.Int64Value   `tfsdk:"horizontal_scaling_min_replicas"`
	HorizontalScalingMaxReplicas       basetypes.Int64Value   `tfsdk:"horizontal_scaling_max_replicas"`
	HorizontalScalingReplicaMultiplier basetypes.Int64Value   `tfsdk:"horizontal_scaling_replica_multiplier"`
	VerticalScalingEnabled             basetypes.BoolValue    `tfsdk:"vertical_scaling_enabled"`
	VerticalScalingMinCPUCores         basetypes.Float64Value `tfsdk:"vertical_scaling_min_cpu_cores"`
	VerticalScalingMinMemoryBytes      basetypes.Int64Value   `tfsdk:"vertical_scaling_min_memory_bytes"`
	PredictiveScalingEnabled           basetypes.BoolValue    `tfsdk:"predictive_scaling_enabled"`
	AutonomousActionWithoutTraffic     basetypes.BoolValue    `tfsdk:"autonomous_action_without_traffic"`
	MaxLatencyIncreasePct              basetypes.Int64Value   `tfsdk:"max_latency_increase_pct"`
	MaxCPUIncreasePct                  basetypes.Int64Value   `tfsdk:"max_cpu_increase_pct"`
	MaxMemoryIncreasePct               basetypes.Int64Value   `tfsdk:"max_memory_increase_pct"`
}

// kubeAppSettingsBlock returns the `kube_app_settings { … }` block schema
// used by `sedai_group_settings`, `sedai_account_settings`, and any future
// resources that bundle per-resource-type settings.
//
// Each settings resource imports this single function rather than declaring
// the block locally; that way schema changes (a new field, a new
// validator) land in one place and propagate uniformly.
func kubeAppSettingsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for Kubernetes workloads (Deployments, StatefulSets, etc.). Every field is optional; only the fields set here are tracked for drift. Unset fields leave the backend's existing kubeSettings values untouched, including UI-configured advanced tuning not exposed in this block.",
		Attributes: map[string]schema.Attribute{
			"availability_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Override availability mode for Kubernetes workloads. Valid: `DATA_PILOT`, `CO_PILOT`, `AUTO`.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"optimization_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Override optimization mode for Kubernetes workloads. Valid: `DATA_PILOT`, `CO_PILOT`, `AUTO`.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"optimization_focus": schema.StringAttribute{
				Optional:    true,
				Description: "What to optimize for. Valid: `COST`, `DURATION`, `COST_AND_DURATION`.",
				Validators:  []validator.String{optimizationFocusValidator()},
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Keep Sedai's view of Kubernetes workloads synced with cloud state.",
			},
			"is_prod": schema.BoolAttribute{
				Optional:    true,
				Description: "Mark Kubernetes workloads as production — production allows fewer aggressive optimizations.",
			},
			"is_operation_allowed": schema.BoolAttribute{
				Optional:    true,
				Description: "Master on/off for Sedai actions on Kubernetes workloads in this scope.",
			},
			"horizontal_scaling_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable horizontal scaling (replica count tuning) for Kubernetes workloads.",
			},
			"horizontal_scaling_min_replicas": schema.Int64Attribute{
				Optional:    true,
				Description: "Floor for replica count when horizontal scaling is enabled.",
			},
			"horizontal_scaling_max_replicas": schema.Int64Attribute{
				Optional:    true,
				Description: "Ceiling for replica count when horizontal scaling is enabled.",
			},
			"horizontal_scaling_replica_multiplier": schema.Int64Attribute{
				Optional:    true,
				Description: "Factor by which replicas scale during horizontal events. Kubernetes-specific.",
			},
			"vertical_scaling_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable vertical scaling (CPU / memory right-sizing) for Kubernetes workloads.",
			},
			"vertical_scaling_min_cpu_cores": schema.Float64Attribute{
				Optional:    true,
				Description: "Floor for CPU per container, in cores (fractional). Kubernetes-specific.",
			},
			"vertical_scaling_min_memory_bytes": schema.Int64Attribute{
				Optional:    true,
				Description: "Floor for memory per container, in bytes. Kubernetes-specific.",
			},
			"predictive_scaling_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable predictive scaling based on detected seasonality.",
			},
			"autonomous_action_without_traffic": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow Sedai to optimize even without recent traffic data.",
			},
			"max_latency_increase_pct": schema.Int64Attribute{
				Optional:    true,
				Description: "Guardrail: maximum acceptable latency increase % during optimization.",
			},
			"max_cpu_increase_pct": schema.Int64Attribute{
				Optional:    true,
				Description: "Guardrail: maximum acceptable CPU increase % during optimization.",
			},
			"max_memory_increase_pct": schema.Int64Attribute{
				Optional:    true,
				Description: "Guardrail: maximum acceptable memory increase % during optimization.",
			},
		},
	}
}

// kubeAppSettingsToSDK builds an SDK-side *settings.KubeAppSettings from
// the tfsdk model. Each field uses the centralized basetypes-to-pointer
// helpers; null/unknown fields produce nil pointers so the SDK skips them.
//
// Returns nil if the model is nil (caller doesn't manage the block).
func kubeAppSettingsToSDK(m *kubeAppSettingsModel) *sdksettings.KubeAppSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.KubeAppSettings{
		AvailabilityMode:                   stringPtr(m.AvailabilityMode),
		OptimizationMode:                   stringPtr(m.OptimizationMode),
		OptimizationFocus:                  stringPtr(m.OptimizationFocus),
		SedaiSyncEnabled:                   boolPtr(m.SedaiSyncEnabled),
		IsProd:                             boolPtr(m.IsProd),
		IsOperationAllowed:                 boolPtr(m.IsOperationAllowed),
		HorizontalScalingEnabled:           boolPtr(m.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas:       int64Ptr(m.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas:       int64Ptr(m.HorizontalScalingMaxReplicas),
		HorizontalScalingReplicaMultiplier: int64Ptr(m.HorizontalScalingReplicaMultiplier),
		VerticalScalingEnabled:             boolPtr(m.VerticalScalingEnabled),
		VerticalScalingMinCPUCores:         float64Ptr(m.VerticalScalingMinCPUCores),
		VerticalScalingMinMemoryBytes:      int64Ptr(m.VerticalScalingMinMemoryBytes),
		PredictiveScalingEnabled:           boolPtr(m.PredictiveScalingEnabled),
		AutonomousActionWithoutTraffic:     boolPtr(m.AutonomousActionWithoutTraffic),
		MaxLatencyIncreasePct:              int64Ptr(m.MaxLatencyIncreasePct),
		MaxCPUIncreasePct:                  int64Ptr(m.MaxCPUIncreasePct),
		MaxMemoryIncreasePct:               int64Ptr(m.MaxMemoryIncreasePct),
	}
}

// kubeAppSettingsRefresh updates only the fields the caller is already
// managing (non-null in the existing state model) with the corresponding
// values from the SDK response. Non-managed fields stay null so plan
// output isn't polluted by backend-only values.
//
// Returns the same state pointer for caller convenience; nil-in / nil-out.
func kubeAppSettingsRefresh(state *kubeAppSettingsModel, fetched *sdksettings.KubeAppSettings) *kubeAppSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshIfManaged(&state.OptimizationFocus, fetched.OptimizationFocus)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	refreshBoolIfManaged(&state.IsProd, fetched.IsProd)
	refreshBoolIfManaged(&state.IsOperationAllowed, fetched.IsOperationAllowed)
	refreshBoolIfManaged(&state.HorizontalScalingEnabled, fetched.HorizontalScalingEnabled)
	refreshInt64IfManaged(&state.HorizontalScalingMinReplicas, fetched.HorizontalScalingMinReplicas)
	refreshInt64IfManaged(&state.HorizontalScalingMaxReplicas, fetched.HorizontalScalingMaxReplicas)
	refreshInt64IfManaged(&state.HorizontalScalingReplicaMultiplier, fetched.HorizontalScalingReplicaMultiplier)
	refreshBoolIfManaged(&state.VerticalScalingEnabled, fetched.VerticalScalingEnabled)
	refreshFloat64IfManaged(&state.VerticalScalingMinCPUCores, fetched.VerticalScalingMinCPUCores)
	refreshInt64IfManaged(&state.VerticalScalingMinMemoryBytes, fetched.VerticalScalingMinMemoryBytes)
	refreshBoolIfManaged(&state.PredictiveScalingEnabled, fetched.PredictiveScalingEnabled)
	refreshBoolIfManaged(&state.AutonomousActionWithoutTraffic, fetched.AutonomousActionWithoutTraffic)
	refreshInt64IfManaged(&state.MaxLatencyIncreasePct, fetched.MaxLatencyIncreasePct)
	refreshInt64IfManaged(&state.MaxCPUIncreasePct, fetched.MaxCPUIncreasePct)
	refreshInt64IfManaged(&state.MaxMemoryIncreasePct, fetched.MaxMemoryIncreasePct)
	return state
}

