package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// appSettingsModel is the tfsdk view of an `app_settings { … }` block.
// Mirrors sdksettings.AppSettings field-for-field, all Optional / nullable
// for the field-level partial-spec contract.
type appSettingsModel struct {
	AvailabilityMode             basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode             basetypes.StringValue `tfsdk:"optimization_mode"`
	IsProd                       basetypes.BoolValue   `tfsdk:"is_prod"`
	SedaiSyncEnabled             basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
	HorizontalScalingEnabled     basetypes.BoolValue   `tfsdk:"horizontal_scaling_enabled"`
	HorizontalScalingMinReplicas basetypes.Int64Value  `tfsdk:"horizontal_scaling_min_replicas"`
	HorizontalScalingMaxReplicas basetypes.Int64Value  `tfsdk:"horizontal_scaling_max_replicas"`
}

// appSettingsBaseAttributes returns the schema attributes shared by every
// app-style block (app_settings AND container_app_settings — the latter
// adds more on top). The two callers may override individual attributes
// (e.g. mode validators differ) but share these definitions verbatim
// otherwise. Centralized to avoid 14 lines of duplication.
//
// modeValidator parameter lets each caller plug in the right validator:
// app_settings uses noAutoConfigModeValidator (per spec); container_app
// uses settingsConfigModeValidator (allows AUTO).
func appSettingsBaseAttributes(modeValidator validator.String) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"availability_mode": schema.StringAttribute{
			Optional:    true,
			Description: "How Sedai manages availability.",
			Validators:  []validator.String{modeValidator},
		},
		"optimization_mode": schema.StringAttribute{
			Optional:    true,
			Description: "How Sedai manages optimization.",
			Validators:  []validator.String{modeValidator},
		},
		"is_prod": schema.BoolAttribute{
			Optional:    true,
			Description: "Mark this scope's apps as production. Non-prod allows more aggressive optimization.",
		},
		"sedai_sync_enabled": schema.BoolAttribute{
			Optional:    true,
			Description: "Keep Sedai's view synced with cloud state.",
		},
		"horizontal_scaling_enabled": schema.BoolAttribute{
			Optional:    true,
			Description: "Enable horizontal scaling (replica count tuning).",
		},
		"horizontal_scaling_min_replicas": schema.Int64Attribute{
			Optional:    true,
			Description: "Floor for replica count when horizontal scaling is enabled.",
		},
		"horizontal_scaling_max_replicas": schema.Int64Attribute{
			Optional:    true,
			Description: "Ceiling for replica count when horizontal scaling is enabled.",
		},
	}
}

// appSettingsBlock returns the schema for `app_settings { … }`. App-level
// modes are restricted to DATA_PILOT / CO_PILOT per spec (no AUTO).
func appSettingsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for generic application workloads. Modes are limited to `DATA_PILOT` and `CO_PILOT` per spec — for AUTO, use a more specific block (kube_app_settings, ecs_app_settings, …).",
		Attributes:  appSettingsBaseAttributes(noAutoConfigModeValidator()),
	}
}

// appSettingsToSDK converts the tfsdk model into the SDK request struct.
func appSettingsToSDK(m *appSettingsModel) *sdksettings.AppSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.AppSettings{
		AvailabilityMode:             stringPtr(m.AvailabilityMode),
		OptimizationMode:             stringPtr(m.OptimizationMode),
		IsProd:                       boolPtr(m.IsProd),
		SedaiSyncEnabled:             boolPtr(m.SedaiSyncEnabled),
		HorizontalScalingEnabled:     boolPtr(m.HorizontalScalingEnabled),
		HorizontalScalingMinReplicas: int64Ptr(m.HorizontalScalingMinReplicas),
		HorizontalScalingMaxReplicas: int64Ptr(m.HorizontalScalingMaxReplicas),
	}
}

// appSettingsRefresh updates only the managed fields with the backend's
// values (field-level partial spec).
func appSettingsRefresh(state *appSettingsModel, fetched *sdksettings.AppSettings) *appSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshBoolIfManaged(&state.IsProd, fetched.IsProd)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	refreshBoolIfManaged(&state.HorizontalScalingEnabled, fetched.HorizontalScalingEnabled)
	refreshInt64IfManaged(&state.HorizontalScalingMinReplicas, fetched.HorizontalScalingMinReplicas)
	refreshInt64IfManaged(&state.HorizontalScalingMaxReplicas, fetched.HorizontalScalingMaxReplicas)
	return state
}
