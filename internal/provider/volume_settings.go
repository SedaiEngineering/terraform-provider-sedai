package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// volumeSettingsModel is the tfsdk view of `volume_settings { … }`.
// EBS volumes per spec — no AUTO for either availability or
// optimization mode (plan-time validators enforce).
type volumeSettingsModel struct {
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
}

// volumeSettingsBlock returns the schema for `volume_settings { … }`.
func volumeSettingsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS EBS volumes. Both `availability_mode` and `optimization_mode` are restricted to `DATA_PILOT` and `CO_PILOT` per spec — no AUTO for volumes.",
		Attributes: map[string]schema.Attribute{
			"availability_mode": schema.StringAttribute{
				Optional:    true,
				Description: "How Sedai manages volume availability. Valid: `DATA_PILOT`, `CO_PILOT`. No AUTO.",
				Validators:  []validator.String{noAutoConfigModeValidator()},
			},
			"optimization_mode": schema.StringAttribute{
				Optional:    true,
				Description: "How Sedai manages volume optimization. Valid: `DATA_PILOT`, `CO_PILOT`. No AUTO.",
				Validators:  []validator.String{noAutoConfigModeValidator()},
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Keep Sedai's view of EBS volumes synced with cloud state.",
			},
		},
	}
}

func volumeSettingsToSDK(m *volumeSettingsModel) *sdksettings.VolumeSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.VolumeSettings{
		AvailabilityMode: stringPtr(m.AvailabilityMode),
		OptimizationMode: stringPtr(m.OptimizationMode),
		SedaiSyncEnabled: boolPtr(m.SedaiSyncEnabled),
	}
}

func volumeSettingsRefresh(state *volumeSettingsModel, fetched *sdksettings.VolumeSettings) *volumeSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	return state
}
