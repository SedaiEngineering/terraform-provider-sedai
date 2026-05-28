package provider

import (
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TestAccountSettingsRequestFromPlan verifies the standalone
// sedai_account_settings resource correctly maps its tfsdk model into the
// SDK's AccountSettings struct.
func TestAccountSettingsRequestFromPlan(t *testing.T) {
	plan := accountSettingsResourceModel{
		AccountID:        basetypes.NewStringValue("acct-1"),
		AvailabilityMode: basetypes.NewStringValue("AUTO"),
		OptimizationMode: basetypes.NewStringValue("CO_PILOT"),
		SedaiSyncEnabled: basetypes.NewBoolValue(true),
	}
	got := accountSettingsRequestFromPlan(plan)
	if got == nil {
		t.Fatal("got nil settings from non-nil plan")
	}
	if got.AvailabilityMode != "AUTO" {
		t.Errorf("AvailabilityMode: got %q", got.AvailabilityMode)
	}
	if got.OptimizationMode != "CO_PILOT" {
		t.Errorf("OptimizationMode: got %q", got.OptimizationMode)
	}
	if !got.SedaiSyncEnabled {
		t.Errorf("SedaiSyncEnabled: got %v", got.SedaiSyncEnabled)
	}
	// Confirm the right type was returned (not GroupSettings).
	if _, ok := interface{}(got).(*account.AccountSettings); !ok {
		t.Error("expected *account.AccountSettings type")
	}
}
