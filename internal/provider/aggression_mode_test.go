package provider

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/constants"
	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// runStringValidators returns whether value passes every validator, and the
// first error's summary.
func runStringValidators(vs []validator.String, value string) (bool, string) {
	resp := &validator.StringResponse{}
	for _, v := range vs {
		v.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("aggression_mode"),
			ConfigValue: basetypes.NewStringValue(value),
		}, resp)
	}
	if resp.Diagnostics.HasError() {
		return false, resp.Diagnostics.Errors()[0].Summary()
	}
	return true, ""
}

func TestSettingsAggressionModeValidators(t *testing.T) {
	cases := []struct {
		value     string
		kubeValid bool
		ecsValid  bool
	}{
		{"PROD", true, true},
		{"EXPEDITE", true, true},
		{"TURBO", true, false}, // TURBO is Kubernetes only
		{"turbo", false, false},
		{"DATA_PILOT", false, false}, // optimization_mode values are not aggression modes
	}
	for _, c := range cases {
		if got, _ := runStringValidators([]validator.String{aggressionModeValidator()}, c.value); got != c.kubeValid {
			t.Errorf("kube %q: valid=%v, want %v", c.value, got, c.kubeValid)
		}
		if got, _ := runStringValidators(ecsAggressionModeValidators(), c.value); got != c.ecsValid {
			t.Errorf("ecs %q: valid=%v, want %v", c.value, got, c.ecsValid)
		}
	}
	// TURBO on ECS says why, not just which values are allowed.
	if _, summary := runStringValidators(ecsAggressionModeValidators(), "TURBO"); summary != "TURBO is not available for ECS" {
		t.Errorf("ecs TURBO: want the Kubernetes-only error, got %q", summary)
	}
}

func TestSettingsAggressionModeIsProdConflict(t *testing.T) {
	null := basetypes.NewStringNull()
	cases := []struct {
		name     string
		isProd   basetypes.BoolValue
		mode     basetypes.StringValue
		conflict bool
	}{
		{"only is_prod true (existing configs)", basetypes.NewBoolValue(true), null, false},
		{"only is_prod false (existing configs)", basetypes.NewBoolValue(false), null, false},
		{"only mode", basetypes.NewBoolNull(), basetypes.NewStringValue("TURBO"), false},
		{"true + PROD", basetypes.NewBoolValue(true), basetypes.NewStringValue("PROD"), false},
		{"false + EXPEDITE", basetypes.NewBoolValue(false), basetypes.NewStringValue("EXPEDITE"), false},
		{"false + TURBO", basetypes.NewBoolValue(false), basetypes.NewStringValue("TURBO"), false},
		{"true + TURBO", basetypes.NewBoolValue(true), basetypes.NewStringValue("TURBO"), true},
		{"true + EXPEDITE", basetypes.NewBoolValue(true), basetypes.NewStringValue("EXPEDITE"), true},
		{"false + PROD", basetypes.NewBoolValue(false), basetypes.NewStringValue("PROD"), true},
		{"unknown is_prod", basetypes.NewBoolUnknown(), basetypes.NewStringValue("TURBO"), false},
		{"unknown mode", basetypes.NewBoolValue(true), basetypes.NewStringUnknown(), false},
	}
	for _, c := range cases {
		msg := aggressionModeIsProdConflict("kube_app_settings", c.isProd, c.mode)
		if (msg != "") != c.conflict {
			t.Errorf("%s: conflict=%v (%q), want %v", c.name, msg != "", msg, c.conflict)
		}
		if msg != "" && !strings.HasPrefix(msg, "kube_app_settings:") {
			t.Errorf("%s: message should name the block, got %q", c.name, msg)
		}
	}
}

func TestSettingsValidateAggressionModeConflicts_ChecksBothBlocks(t *testing.T) {
	kube := &kubeAppSettingsModel{IsProd: basetypes.NewBoolValue(true), AggressionMode: basetypes.NewStringValue("TURBO")}
	ecs := &ecsAppSettingsModel{IsProd: basetypes.NewBoolValue(false), AggressionMode: basetypes.NewStringValue("PROD")}
	errs := validateAggressionModeConflicts(kube, ecs)
	if len(errs) != 2 || !strings.HasPrefix(errs[0], "kube_app_settings:") || !strings.HasPrefix(errs[1], "ecs_app_settings:") {
		t.Fatalf("want one error per block, got %v", errs)
	}
	if errs := validateAggressionModeConflicts(nil, nil); len(errs) != 0 {
		t.Errorf("no blocks: want no errors, got %v", errs)
	}
}

func strPtr(s string) *string { return &s }

func TestSettingsPendingAggressionModeSwitches(t *testing.T) {
	kube := func(m string) *kubeAppSettingsModel {
		if m == "" {
			return &kubeAppSettingsModel{AggressionMode: basetypes.NewStringNull()}
		}
		return &kubeAppSettingsModel{AggressionMode: basetypes.NewStringValue(m)}
	}
	ecs := func(m string) *ecsAppSettingsModel {
		if m == "" {
			return &ecsAppSettingsModel{AggressionMode: basetypes.NewStringNull()}
		}
		return &ecsAppSettingsModel{AggressionMode: basetypes.NewStringValue(m)}
	}
	cases := []struct {
		name        string
		kube        *kubeAppSettingsModel
		ecs         *ecsAppSettingsModel
		curK, curE  *string
		want        []aggressionModeSwitch
		wantManaged bool
	}{
		{name: "no blocks", want: nil},
		{name: "blocks without aggression_mode", kube: kube(""), ecs: ecs(""), curK: strPtr("PROD"), want: nil},
		{name: "already in place", kube: kube("TURBO"), curK: strPtr("TURBO"), want: nil, wantManaged: true},
		{name: "kube changes", kube: kube("TURBO"), curK: strPtr("PROD"), wantManaged: true,
			want: []aggressionModeSwitch{{constants.AggressionModeResourceTypeKube, constants.AggressionModeTurbo}}},
		{name: "backend section missing still switches", kube: kube("EXPEDITE"), wantManaged: true,
			want: []aggressionModeSwitch{{constants.AggressionModeResourceTypeKube, constants.AggressionModeExpedite}}},
		{name: "kube before ecs", kube: kube("TURBO"), ecs: ecs("EXPEDITE"), curK: strPtr("PROD"), curE: strPtr("PROD"), wantManaged: true,
			want: []aggressionModeSwitch{
				{constants.AggressionModeResourceTypeKube, constants.AggressionModeTurbo},
				{constants.AggressionModeResourceTypeECS, constants.AggressionModeExpedite},
			}},
		{name: "only ecs changes", kube: kube("TURBO"), ecs: ecs("PROD"), curK: strPtr("TURBO"), curE: strPtr("EXPEDITE"), wantManaged: true,
			want: []aggressionModeSwitch{{constants.AggressionModeResourceTypeECS, constants.AggressionModeProd}}},
	}
	for _, c := range cases {
		if got := wantedAggressionModes(c.kube, c.ecs); got != c.wantManaged {
			t.Errorf("%s: wantedAggressionModes=%v, want %v", c.name, got, c.wantManaged)
		}
		got := pendingAggressionModeSwitches(c.kube, c.ecs, c.curK, c.curE)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestSettingsApplyAggressionModeSwitches_StopsAtFirstError(t *testing.T) {
	switches := []aggressionModeSwitch{
		{constants.AggressionModeResourceTypeKube, constants.AggressionModeTurbo},
		{constants.AggressionModeResourceTypeECS, constants.AggressionModeExpedite},
	}
	var calls []string
	sdkErr := errors.New("failed to switch the aggression mode of group g1 (KUBE): HTTP 500")
	err := applyAggressionModeSwitches(switches, func(rt constants.AggressionModeResourceType, m constants.AggressionMode) error {
		calls = append(calls, string(rt)+":"+string(m))
		return sdkErr
	})
	if len(calls) != 1 || calls[0] != "KUBE:TURBO" {
		t.Errorf("want only the first switch tried, got %v", calls)
	}
	// The SDK's message is kept once, wrapped, with the hint after it.
	if !errors.Is(err, sdkErr) || strings.Count(err.Error(), "group g1") != 1 || !strings.Contains(err.Error(), "1.148") {
		t.Errorf("error should wrap the SDK error once and add the version hint, got %v", err)
	}

	calls = nil
	if err := applyAggressionModeSwitches(switches, func(rt constants.AggressionModeResourceType, m constants.AggressionMode) error {
		calls = append(calls, string(rt)+":"+string(m))
		return nil
	}); err != nil || len(calls) != 2 {
		t.Errorf("want both switches and no error, got %v, %v", calls, err)
	}
}

func TestSettingsAggressionModeRefreshAndFromSDK(t *testing.T) {
	turbo := "TURBO"
	// Managed: refreshed from the backend.
	k := &kubeAppSettingsModel{AggressionMode: basetypes.NewStringValue("PROD")}
	kubeAppSettingsRefresh(k, &sdksettings.KubeAppSettings{AggressionMode: &turbo})
	if k.AggressionMode.ValueString() != "TURBO" {
		t.Errorf("kube managed: want TURBO, got %v", k.AggressionMode)
	}
	// Unmanaged: stays null.
	k = &kubeAppSettingsModel{AggressionMode: basetypes.NewStringNull()}
	kubeAppSettingsRefresh(k, &sdksettings.KubeAppSettings{AggressionMode: &turbo})
	if !k.AggressionMode.IsNull() {
		t.Errorf("kube unmanaged: want null, got %v", k.AggressionMode)
	}

	expedite := "EXPEDITE"
	e := &ecsAppSettingsModel{AggressionMode: basetypes.NewStringValue("PROD")}
	ecsAppSettingsRefresh(e, &sdksettings.ECSAppSettings{AggressionMode: &expedite})
	if e.AggressionMode.ValueString() != "EXPEDITE" {
		t.Errorf("ecs managed: want EXPEDITE, got %v", e.AggressionMode)
	}

	if got := kubeAppSettingsFromSDK(&sdksettings.KubeAppSettings{AggressionMode: &turbo}); got.AggressionMode.ValueString() != "TURBO" {
		t.Errorf("kube FromSDK: want TURBO, got %v", got.AggressionMode)
	}
	if got := ecsAppSettingsFromSDK(&sdksettings.ECSAppSettings{}); !got.AggressionMode.IsNull() {
		t.Errorf("ecs FromSDK without a mode: want null, got %v", got.AggressionMode)
	}
}

func TestSettingsRefreshResourceAggressionMode(t *testing.T) {
	s := basetypes.NewStringNull()
	refreshResourceAggressionMode(&s, "TURBO")
	if !s.IsNull() {
		t.Errorf("unmanaged: want null, got %v", s)
	}
	s = basetypes.NewStringValue("PROD")
	refreshResourceAggressionMode(&s, "TURBO")
	if s.ValueString() != "TURBO" {
		t.Errorf("managed: want TURBO, got %v", s)
	}
}
