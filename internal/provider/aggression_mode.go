package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/constants"
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// The aggression mode (PROD, EXPEDITE, TURBO; "Speed" in the UI) is the
// server's `optimizationMode.mode` on Kubernetes and ECS settings. It is
// not `optimization_mode`, which is the DATA_PILOT / CO_PILOT / AUTO setting.
//
// The SDK's settings saves never write it. It changes only through the
// aggressionmode switch functions, which also reset CPU and memory
// guardrails, so an apply switches first and saves the other fields after
// (SED-24715).

const aggressionModeSwitchHint = "If Sedai rejected the request, check that it runs 1.148 or later and that the API key has the admin role."

// aggressionModeValidator accepts every mode; used on kube_app_settings and
// sedai_resource_settings.
func aggressionModeValidator() validator.String {
	return stringvalidator.OneOf(aggressionModeStrings(constants.ValidAggressionModes)...)
}

// ecsAggressionModeValidators refuse TURBO with the reason: the server stores
// TURBO on ECS without an error, so this is the plan-time guard.
func ecsAggressionModeValidators() []validator.String {
	return []validator.String{aggressionModeValidator(), ecsNoTurboValidator{}}
}

type ecsNoTurboValidator struct{}

func (ecsNoTurboValidator) Description(context.Context) string {
	return "TURBO is for Kubernetes workloads only"
}

func (v ecsNoTurboValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (ecsNoTurboValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.ValueString() == string(constants.AggressionModeTurbo) {
		resp.Diagnostics.AddAttributeError(req.Path, "TURBO is not available for ECS",
			`aggression_mode = "TURBO" is for Kubernetes workloads only. Use "PROD" or "EXPEDITE" in ecs_app_settings.`)
	}
}

// aggressionModeDescription is the shared schema text for aggression_mode.
func aggressionModeDescription(what, valid, guardrails string) string {
	return fmt.Sprintf("How aggressively Sedai optimizes %s. Valid: %s. Not the same as `optimization_mode` "+
		"(Data Pilot / Co-Pilot / Autopilot). A change is applied as a mode switch, before the other settings "+
		"are saved: the switch resets %s to the mode's defaults, and a value set in this block is written again "+
		"after it. If the save fails after the switch, the apply fails and the mode's default guardrails stay until the "+
		"next apply, which writes the configured values again without switching. "+
		"Must agree with `is_prod` when both are set (`is_prod = true` only with `PROD`). Needs Sedai 1.148 or "+
		"later and an admin-role API key.", what, valid, guardrails)
}

func aggressionModeStrings(modes []constants.AggressionMode) []string {
	out := make([]string, 0, len(modes))
	for _, m := range modes {
		out = append(out, string(m))
	}
	return out
}

// aggressionModeIsProdConflict returns an error message when is_prod and
// aggression_mode disagree in one block, else "". A save applies isProd
// after the mode, so `is_prod = true` would silently turn TURBO into PROD.
// Setting only one of them stays valid, so existing configs are unchanged.
func aggressionModeIsProdConflict(block string, isProd basetypes.BoolValue, mode basetypes.StringValue) string {
	if isProd.IsNull() || isProd.IsUnknown() || mode.IsNull() || mode.IsUnknown() {
		return ""
	}
	isProdMode := mode.ValueString() == string(constants.AggressionModeProd)
	if isProd.ValueBool() == isProdMode {
		return ""
	}
	if isProd.ValueBool() {
		return fmt.Sprintf("%s: is_prod = true requires aggression_mode = \"PROD\" (got %q). "+
			"Sedai saves is_prod after the mode, so this combination would be stored as PROD. "+
			"Remove is_prod or set it to false.", block, mode.ValueString())
	}
	return fmt.Sprintf("%s: is_prod = false requires aggression_mode = \"EXPEDITE\" or \"TURBO\" (got \"PROD\"). "+
		"Sedai saves is_prod after the mode, so this combination would be stored as EXPEDITE. "+
		"Remove is_prod or set it to true.", block)
}

// validateAggressionModeConflicts runs aggressionModeIsProdConflict on the
// kube and ECS blocks of account and group settings.
func validateAggressionModeConflicts(kube *kubeAppSettingsModel, ecs *ecsAppSettingsModel) []string {
	var errs []string
	if kube != nil {
		if msg := aggressionModeIsProdConflict("kube_app_settings", kube.IsProd, kube.AggressionMode); msg != "" {
			errs = append(errs, msg)
		}
	}
	if ecs != nil {
		if msg := aggressionModeIsProdConflict("ecs_app_settings", ecs.IsProd, ecs.AggressionMode); msg != "" {
			errs = append(errs, msg)
		}
	}
	return errs
}

// aggressionModeSwitch is one mode change an apply must make before saving.
type aggressionModeSwitch struct {
	ResourceType constants.AggressionModeResourceType
	Mode         constants.AggressionMode
}

// wantedAggressionModes reports whether the plan manages aggression_mode in
// either block, so callers skip the extra read when it doesn't.
func wantedAggressionModes(kube *kubeAppSettingsModel, ecs *ecsAppSettingsModel) bool {
	return (kube != nil && stringPtr(kube.AggressionMode) != nil) ||
		(ecs != nil && stringPtr(ecs.AggressionMode) != nil)
}

// pendingAggressionModeSwitches lists the switches needed to move the
// backend's current modes to the planned ones. A mode already in place is
// skipped: a switch would reset guardrails and add an audit row for nothing.
// Kubernetes comes first so a failure on it stops before ECS is touched.
func pendingAggressionModeSwitches(kube *kubeAppSettingsModel, ecs *ecsAppSettingsModel, currentKube, currentECS *string) []aggressionModeSwitch {
	var out []aggressionModeSwitch
	add := func(rt constants.AggressionModeResourceType, wanted basetypes.StringValue, current *string) {
		w := stringPtr(wanted)
		if w == nil || (current != nil && *current == *w) {
			return
		}
		out = append(out, aggressionModeSwitch{ResourceType: rt, Mode: constants.AggressionMode(*w)})
	}
	if kube != nil {
		add(constants.AggressionModeResourceTypeKube, kube.AggressionMode, currentKube)
	}
	if ecs != nil {
		add(constants.AggressionModeResourceTypeECS, ecs.AggressionMode, currentECS)
	}
	return out
}

// applyAggressionModeSwitches runs each switch in order through set and
// stops at the first error. The SDK's error already names the target and
// the resource type, so only the version and permission hint is added.
func applyAggressionModeSwitches(switches []aggressionModeSwitch, set func(constants.AggressionModeResourceType, constants.AggressionMode) error) error {
	for _, s := range switches {
		if err := set(s.ResourceType, s.Mode); err != nil {
			return withSwitchHint(err)
		}
	}
	return nil
}

func withSwitchHint(err error) error {
	return fmt.Errorf("%w. %s", err, aggressionModeSwitchHint)
}

// settleTiming controls the wait after a switch.
type settleTiming struct {
	interval, timeout time.Duration
	reads             int // matching reads in a row before saving
}

// Each Sedai API pod keeps its own near cache of settings. After a switch,
// another pod can return the settings from before it for about 10 s, and
// the SDK's save posts back the settings it read, so a stale read writes the
// old mode back. Seven matching reads 2 s apart span that window. A var so
// tests can shorten it.
var aggressionModeSettle = settleTiming{interval: 2 * time.Second, timeout: 90 * time.Second, reads: 7}

// awaitAggressionModes waits until inPlace reports the planned modes on
// aggressionModeSettle.reads reads in a row.
func awaitAggressionModes(ctx context.Context, what string, inPlace func() (bool, error)) error {
	deadline := time.Now().Add(aggressionModeSettle.timeout)
	for streak := 0; ; {
		ok, err := inPlace()
		if err != nil {
			return fmt.Errorf("could not read back the aggression mode of %s after the switch: %w", what, err)
		}
		if ok {
			streak++
		} else {
			streak = 0
		}
		if streak >= aggressionModeSettle.reads {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Sedai still returns the previous aggression mode of %s %s after the switch, "+
				"so the other settings were not saved; run terraform apply again", what, aggressionModeSettle.timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(aggressionModeSettle.interval):
		}
	}
}

// sectionModes returns the current Kubernetes and ECS modes of a settings
// response; nil when a section is missing.
func sectionModes(kube *sdksettings.KubeAppSettings, ecs *sdksettings.ECSAppSettings) (curKube, curECS *string) {
	if kube != nil {
		curKube = kube.AggressionMode
	}
	if ecs != nil {
		curECS = ecs.AggressionMode
	}
	return curKube, curECS
}
