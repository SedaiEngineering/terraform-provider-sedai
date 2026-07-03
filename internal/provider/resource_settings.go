package provider

import (
	"context"
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/resource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure interfaces are satisfied.
var (
	_ tfresource.Resource                = &resourceSettings{}
	_ tfresource.ResourceWithImportState = &resourceSettings{}
)

// ResourceSettings is the resource constructor for `sedai_resource_settings`.
func ResourceSettings() tfresource.Resource {
	return &resourceSettings{}
}

// resourceSettings is the resource implementation. Resource type name:
// `sedai_resource_settings`. Manages the top-level settings (availability /
// optimization mode + Sedai Sync) for a single Sedai resource —
// the most-specific override level in the inheritance chain
// (platform → account → group → resource).
//
// Per-type guardrails (horizontal_scaling_min_replicas, max_latency_increase_pct,
// etc.) are not modelled in this iteration; they're preserved opaquely
// across writes via the SDK's read-modify-write logic.
type resourceSettings struct{}

type resourceSettingsModel struct {
	ResourceID       basetypes.StringValue `tfsdk:"resource_id"`
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
	ResourceType     basetypes.StringValue `tfsdk:"resource_type"`
}

func (r *resourceSettings) Metadata(_ context.Context, req tfresource.MetadataRequest, resp *tfresource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_settings"
}

func (r *resourceSettings) Schema(_ context.Context, _ tfresource.SchemaRequest, resp *tfresource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the top-level settings for a single Sedai resource — the most-specific override in the inheritance chain. Only attributes specified here are tracked for drift; per-resource-type tuning fields (scaling limits, etc.) are preserved untouched.",
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The Sedai resource ID. Often a slash-delimited path like `abc123/Deployment/prod-cluster/payments/payment-api`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"availability_mode": schema.StringAttribute{
				Required:    true,
				Description: "Availability mode. Valid values: `DATA_PILOT`, `CO_PILOT`, `AUTO`. Per-resource-type validity rules (e.g. buckets cannot be AUTO) apply at apply-time.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"optimization_mode": schema.StringAttribute{
				Required:    true,
				Description: "Optimization mode. Valid values: `DATA_PILOT`, `CO_PILOT`, `AUTO`. Per-resource-type validity rules apply at apply-time.",
				Validators:  []validator.String{settingsConfigModeValidator()},
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, Sedai auto-syncs this resource with the latest cloud-side configuration. Defaults to false if omitted.",
			},
			"resource_type": schema.StringAttribute{
				Computed:    true,
				Description: "Backend-detected resource type (e.g. `KubernetesAppResourceSettingDetail`, `ServerlessFunctionResourceSettingDetail`). Populated from the API response — read-only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceSettings) Create(ctx context.Context, req tfresource.CreateRequest, resp *tfresource.CreateResponse) {
	var plan resourceSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := resource.UpdateResourceSettings(plan.ResourceID.ValueString(), resourceSettingsRequestFromPlan(plan)); err != nil {
		// POST may have been processed before the connection dropped (EOF-during-POST).
		// Poll GetResourceSettings up to 3 times before treating this as a hard failure.
		adopted := false
		for i := 0; i < 3; i++ {
			time.Sleep(2 * time.Second)
			existing, fetchErr := resource.GetResourceSettings(plan.ResourceID.ValueString())
			if fetchErr == nil && existing != nil {
				resp.Diagnostics.AddWarning(
					"Resource settings configured despite connection error",
					"Settings for resource '"+plan.ResourceID.ValueString()+"' were found on the "+
						"backend after a failed POST — the response was likely lost in transit. "+
						"Current state adopted; run terraform apply again to reconcile any drift.",
				)
				adopted = true
				break
			}
		}
		if !adopted {
			resp.Diagnostics.AddError("Unable to set resource settings", err.Error())
			return
		}
	}

	// Read back to populate the computed resource_type from the backend.
	r.refreshFromBackend(&plan, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSettings) Read(ctx context.Context, req tfresource.ReadRequest, resp *tfresource.ReadResponse) {
	var state resourceSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := resource.GetResourceSettings(state.ResourceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read resource settings",
			"Could not fetch settings for resource "+state.ResourceID.ValueString()+": "+err.Error(),
		)
		return
	}
	if settings == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
	state.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
	populateBoolIfUnset(&state.SedaiSyncEnabled, settings.SedaiSyncEnabled)
	state.ResourceType = basetypes.NewStringValue(settings.ResourceType)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceSettings) Update(ctx context.Context, req tfresource.UpdateRequest, resp *tfresource.UpdateResponse) {
	var plan resourceSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := resource.UpdateResourceSettings(plan.ResourceID.ValueString(), resourceSettingsRequestFromPlan(plan)); err != nil {
		resp.Diagnostics.AddError("Unable to update resource settings", err.Error())
		return
	}

	r.refreshFromBackend(&plan, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete resets the managed modes to DATA_PILOT (the most conservative,
// observe-only setting) — the practical equivalent of "reset to inherited
// default" given the backend has no explicit reset endpoint. The resource
// itself is left intact; per-type guardrails preserved.
func (r *resourceSettings) Delete(ctx context.Context, req tfresource.DeleteRequest, resp *tfresource.DeleteResponse) {
	var state resourceSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reset := &resource.ResourceSettings{
		AvailabilityMode: "DATA_PILOT",
		OptimizationMode: "DATA_PILOT",
		SedaiSyncEnabled: false,
	}
	if err := resource.UpdateResourceSettings(state.ResourceID.ValueString(), reset); err != nil {
		resp.Diagnostics.AddWarning("Unable to reset resource settings on destroy", err.Error())
	}
}

func (r *resourceSettings) ImportState(ctx context.Context, req tfresource.ImportStateRequest, resp *tfresource.ImportStateResponse) {
	tfresource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

// refreshFromBackend re-reads the resource settings and populates the
// computed `resource_type` field on the plan/state. Called from Create
// and Update so users see the backend-detected type after applying.
// Failures here are warnings only — the primary write already succeeded.
func (r *resourceSettings) refreshFromBackend(state *resourceSettingsModel, diags interface {
	AddWarning(string, string)
}) {
	fetched, err := resource.GetResourceSettings(state.ResourceID.ValueString())
	if err != nil {
		diags.AddWarning("Unable to populate computed resource_type", err.Error())
		state.ResourceType = basetypes.NewStringValue("")
		return
	}
	if fetched != nil {
		state.ResourceType = basetypes.NewStringValue(fetched.ResourceType)
	} else {
		state.ResourceType = basetypes.NewStringValue("")
	}
}

func resourceSettingsRequestFromPlan(p resourceSettingsModel) *resource.ResourceSettings {
	return &resource.ResourceSettings{
		AvailabilityMode: p.AvailabilityMode.ValueString(),
		OptimizationMode: p.OptimizationMode.ValueString(),
		SedaiSyncEnabled: p.SedaiSyncEnabled.ValueBool(),
	}
}

