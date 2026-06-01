package provider

import (
	"testing"

	sdksettings "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/settings"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TestKubeAppSettingsToSDK_NilPassesThrough verifies the conversion
// helper short-circuits on nil — used by Create/Update when the user
// omits the kube_app_settings { } block from HCL.
func TestKubeAppSettingsToSDK_NilPassesThrough(t *testing.T) {
	if got := kubeAppSettingsToSDK(nil); got != nil {
		t.Errorf("nil input: want nil, got %v", got)
	}
}

// TestKubeAppSettingsToSDK_NullsBecomeNilPointers is the central guarantee
// of field-level partial spec: a TF model field that's basetypes.X.Null()
// converts to a nil pointer in the SDK struct, signaling the SDK to leave
// the backend's existing value untouched.
func TestKubeAppSettingsToSDK_NullsBecomeNilPointers(t *testing.T) {
	m := &kubeAppSettingsModel{
		AvailabilityMode:             basetypes.NewStringValue("AUTO"), // managed
		OptimizationMode:             basetypes.NewStringNull(),        // not managed
		HorizontalScalingMinReplicas: basetypes.NewInt64Value(3),       // managed
		HorizontalScalingMaxReplicas: basetypes.NewInt64Null(),         // not managed
		SedaiSyncEnabled:             basetypes.NewBoolValue(true),     // managed
		IsProd:                       basetypes.NewBoolNull(),          // not managed
	}
	got := kubeAppSettingsToSDK(m)
	if got.AvailabilityMode == nil || *got.AvailabilityMode != "AUTO" {
		t.Error("AvailabilityMode: want AUTO, got nil/different")
	}
	if got.OptimizationMode != nil {
		t.Error("OptimizationMode: want nil (null state), got non-nil")
	}
	if got.HorizontalScalingMinReplicas == nil || *got.HorizontalScalingMinReplicas != 3 {
		t.Error("HorizontalScalingMinReplicas: want 3, got nil/different")
	}
	if got.HorizontalScalingMaxReplicas != nil {
		t.Error("HorizontalScalingMaxReplicas: want nil (null state), got non-nil")
	}
	if got.SedaiSyncEnabled == nil || *got.SedaiSyncEnabled != true {
		t.Error("SedaiSyncEnabled: want true, got nil/different")
	}
	if got.IsProd != nil {
		t.Error("IsProd: want nil (null state), got non-nil")
	}
}

// TestKubeAppSettingsRefresh_OnlyUpdatesManagedFields verifies the Read
// flow only writes backend values onto state fields the user is already
// managing. Unmanaged (null) state fields stay null so plan output isn't
// polluted by every backend default.
func TestKubeAppSettingsRefresh_OnlyUpdatesManagedFields(t *testing.T) {
	autoFromBackend := "AUTO"
	prodFromBackend := true
	min5FromBackend := int64(5)

	// State: user manages availability_mode + is_prod. Everything else null.
	state := &kubeAppSettingsModel{
		AvailabilityMode:             basetypes.NewStringValue("CO_PILOT"),
		IsProd:                       basetypes.NewBoolValue(false),
		HorizontalScalingMinReplicas: basetypes.NewInt64Null(), // not managed
		OptimizationMode:             basetypes.NewStringNull(),
	}

	// Backend reports values for all fields.
	fetched := &sdksettings.KubeAppSettings{
		AvailabilityMode:             &autoFromBackend,
		IsProd:                       &prodFromBackend,
		HorizontalScalingMinReplicas: &min5FromBackend,
		OptimizationMode:             stringPtrLiteral("DATA_PILOT"),
	}

	kubeAppSettingsRefresh(state, fetched)

	// Managed fields got updated.
	if state.AvailabilityMode.ValueString() != "AUTO" {
		t.Errorf("AvailabilityMode should be refreshed to AUTO, got %q", state.AvailabilityMode.ValueString())
	}
	if !state.IsProd.ValueBool() {
		t.Errorf("IsProd should be refreshed to true")
	}

	// Unmanaged fields stay null — no spurious drift.
	if !state.HorizontalScalingMinReplicas.IsNull() {
		t.Errorf("HorizontalScalingMinReplicas: want null (unmanaged), got %v", state.HorizontalScalingMinReplicas)
	}
	if !state.OptimizationMode.IsNull() {
		t.Errorf("OptimizationMode: want null (unmanaged), got %q", state.OptimizationMode.ValueString())
	}
}

// TestKubeAppSettingsRefresh_NilGuards covers Read paths where state has
// no managed block (user omitted kube_app_settings in HCL) and / or the
// backend response has no kube section. Both cases: no-op, no panic.
func TestKubeAppSettingsRefresh_NilGuards(t *testing.T) {
	// Both nil — no-op.
	kubeAppSettingsRefresh(nil, nil)
	kubeAppSettingsRefresh(nil, &sdksettings.KubeAppSettings{})

	// Non-nil state with nil fetched — state untouched.
	s := &kubeAppSettingsModel{
		AvailabilityMode: basetypes.NewStringValue("CO_PILOT"),
	}
	kubeAppSettingsRefresh(s, nil)
	if s.AvailabilityMode.ValueString() != "CO_PILOT" {
		t.Errorf("state should be untouched when fetched is nil")
	}
}

// stringPtrLiteral is a tiny test helper for &"literal" — Go won't let
// you take the address of a string literal directly.
func stringPtrLiteral(s string) *string { return &s }
