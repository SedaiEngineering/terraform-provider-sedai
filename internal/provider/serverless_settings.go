package provider

import (
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// serverlessSettingsModel is the tfsdk view of `serverless_settings { … }`.
// Lambda has its own shape — no inheritance from app / container_app
// because scaling and replica concepts don't apply.
//
// Validity restrictions (enforced at plan time):
//   - optimization_mode: DATA_PILOT or AUTO only (no CO_PILOT)
//   - concurrency_mode:  OFF or AUTO
//   - optimization_focus: COST | DURATION | COST_AND_DURATION
type serverlessSettingsModel struct {
	AvailabilityMode    basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode    basetypes.StringValue `tfsdk:"optimization_mode"`
	OptimizationFocus   basetypes.StringValue `tfsdk:"optimization_focus"`
	ConcurrencyMode     basetypes.StringValue `tfsdk:"concurrency_mode"`
	MaxCostChangePct    basetypes.Int64Value  `tfsdk:"max_cost_change_pct"`
	MaxLatencyChangePct basetypes.Int64Value  `tfsdk:"max_latency_change_pct"`
	SedaiSyncEnabled    basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
}

// serverlessSettingsBlock returns the schema for the
// `serverless_settings { … }` block. Each value validator is sourced
// from constants — no hardcoded enum strings here.
func serverlessSettingsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS Lambda functions. `optimization_mode` is restricted to DATA_PILOT and AUTO (no CO_PILOT). `concurrency_mode` is Lambda-specific (OFF or AUTO).",
		Attributes: map[string]schema.Attribute{
			"availability_mode": schema.StringAttribute{
				Optional:    true,
				Description: "How Sedai manages Lambda availability. Valid: `DATA_PILOT`, `CO_PILOT`, `AUTO`.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"optimization_mode": schema.StringAttribute{
				Optional:    true,
				Description: "How Sedai manages Lambda optimization. Valid: `DATA_PILOT`, `AUTO`. No CO_PILOT for Lambda.",
				Validators:  []validator.String{noCopilotConfigModeValidator()},
			},
			"optimization_focus": schema.StringAttribute{
				Optional:    true,
				Description: "What to optimize for. Valid: `COST`, `DURATION`, `COST_AND_DURATION`.",
				Validators:  []validator.String{optimizationFocusValidator()},
			},
			"concurrency_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Whether Sedai manages Lambda reserved concurrency. Valid: `OFF`, `AUTO`.",
				Validators:  []validator.String{concurrencyModeValidator()},
			},
			"max_cost_change_pct": schema.Int64Attribute{
				Optional:    true,
				Description: "Guardrail: maximum acceptable cost increase % when optimizing for performance.",
			},
			"max_latency_change_pct": schema.Int64Attribute{
				Optional:    true,
				Description: "Guardrail: maximum acceptable latency increase % when optimizing for cost.",
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Keep Sedai's view of Lambda functions synced with cloud state.",
			},
		},
	}
}

// serverlessSettingsToSDK converts the tfsdk model to the SDK struct.
func serverlessSettingsToSDK(m *serverlessSettingsModel) *sdksettings.ServerlessSettings {
	if m == nil {
		return nil
	}
	return &sdksettings.ServerlessSettings{
		AvailabilityMode:    stringPtr(m.AvailabilityMode),
		OptimizationMode:    stringPtr(m.OptimizationMode),
		OptimizationFocus:   stringPtr(m.OptimizationFocus),
		ConcurrencyMode:     stringPtr(m.ConcurrencyMode),
		MaxCostChangePct:    int64Ptr(m.MaxCostChangePct),
		MaxLatencyChangePct: int64Ptr(m.MaxLatencyChangePct),
		SedaiSyncEnabled:    boolPtr(m.SedaiSyncEnabled),
	}
}

// serverlessSettingsRefresh updates only the managed fields.
func serverlessSettingsRefresh(state *serverlessSettingsModel, fetched *sdksettings.ServerlessSettings) *serverlessSettingsModel {
	if state == nil || fetched == nil {
		return state
	}
	refreshIfManaged(&state.AvailabilityMode, fetched.AvailabilityMode)
	refreshIfManaged(&state.OptimizationMode, fetched.OptimizationMode)
	refreshIfManaged(&state.OptimizationFocus, fetched.OptimizationFocus)
	refreshIfManaged(&state.ConcurrencyMode, fetched.ConcurrencyMode)
	refreshInt64IfManaged(&state.MaxCostChangePct, fetched.MaxCostChangePct)
	refreshInt64IfManaged(&state.MaxLatencyChangePct, fetched.MaxLatencyChangePct)
	refreshBoolIfManaged(&state.SedaiSyncEnabled, fetched.SedaiSyncEnabled)
	return state
}
