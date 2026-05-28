package provider

import (
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TestResourceSettingsRequestFromPlan verifies the tfsdk model maps
// cleanly into the SDK request struct.
func TestResourceSettingsRequestFromPlan(t *testing.T) {
	plan := resourceSettingsModel{
		ResourceID:       basetypes.NewStringValue("abc/Deployment/cluster/ns/svc"),
		AvailabilityMode: basetypes.NewStringValue("CO_PILOT"),
		OptimizationMode: basetypes.NewStringValue("AUTO"),
		SedaiSyncEnabled: basetypes.NewBoolValue(true),
	}
	got := resourceSettingsRequestFromPlan(plan)
	if got == nil {
		t.Fatal("got nil settings from non-nil plan")
	}
	if got.AvailabilityMode != "CO_PILOT" {
		t.Errorf("AvailabilityMode: got %q", got.AvailabilityMode)
	}
	if got.OptimizationMode != "AUTO" {
		t.Errorf("OptimizationMode: got %q", got.OptimizationMode)
	}
	if !got.SedaiSyncEnabled {
		t.Errorf("SedaiSyncEnabled: got %v", got.SedaiSyncEnabled)
	}
	if _, ok := interface{}(got).(*resource.ResourceSettings); !ok {
		t.Error("expected *resource.ResourceSettings type")
	}
}

// TestResourceSettingsRequestFromPlan_EmptyModes verifies that null/empty
// modes pass through as empty strings so the SDK's read-modify-write keeps
// the existing backend value untouched (rather than overwriting to "").
func TestResourceSettingsRequestFromPlan_EmptyModes(t *testing.T) {
	plan := resourceSettingsModel{
		ResourceID:       basetypes.NewStringValue("res-1"),
		AvailabilityMode: basetypes.NewStringNull(),
		OptimizationMode: basetypes.NewStringNull(),
		SedaiSyncEnabled: basetypes.NewBoolNull(),
	}
	got := resourceSettingsRequestFromPlan(plan)
	if got.AvailabilityMode != "" {
		t.Errorf("AvailabilityMode: want empty, got %q", got.AvailabilityMode)
	}
	if got.OptimizationMode != "" {
		t.Errorf("OptimizationMode: want empty, got %q", got.OptimizationMode)
	}
	if got.SedaiSyncEnabled {
		t.Errorf("SedaiSyncEnabled: want false from null, got true")
	}
}
