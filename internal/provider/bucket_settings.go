package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// bucketSettingsModel is the tfsdk view of a `bucket_settings { … }`
// block. Per spec, buckets expose just `optimization_mode` (no AUTO) and
// `sedai_sync_enabled`; `availability_mode` is fixed at DATA_PILOT
// backend-side and intentionally not configurable.
type bucketSettingsModel struct {
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
}

// bucketSettingsBlock returns the schema for the `bucket_settings { … }`
// block. Reused by every settings resource that bundles per-type blocks.
func bucketSettingsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for S3 buckets. Buckets do not support AUTO for optimization — only DATA_PILOT and CO_PILOT. Availability is always DATA_PILOT backend-side and not configurable here.",
		Attributes: map[string]schema.Attribute{
			"optimization_mode": schema.StringAttribute{
				Optional:    true,
				Description: "How Sedai manages optimization for S3 buckets. Valid: `DATA_PILOT`, `CO_PILOT`. No AUTO.",
				Validators:  []validator.String{noAutoConfigModeValidator()},
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Keep Sedai's view of S3 buckets synced with cloud state.",
			},
		},
	}
}

// bucketSettingsToSDK converts the tfsdk model to the SDK request struct.
// Returns nil if the model is nil (user omitted the HCL block).
func bucketSettingsToSDK(m *bucketSettingsModel) *sdksettings.BucketSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.BucketSettings{
		OptimizationMode: stringPtr(m.OptimizationMode),
		SedaiSyncEnabled: boolPtr(m.SedaiSyncEnabled),
	}
}

// bucketSettingsRefresh updates only the fields the caller is already
// managing (non-null in state) with the corresponding values from the
// SDK response. Field-level partial spec preserved.
func bucketSettingsRefresh(state *bucketSettingsModel, fetched *sdksettings.BucketSettings) *bucketSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	return state
}
