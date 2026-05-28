package provider

import (
	"sort"
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/constants"
)

// TestValidSettingsConfigModeStrings verifies the validator's accepted set
// stays in sync with constants.ValidSettingsConfigModes — the single
// source of truth. If someone adds a new mode to the constants slice, this
// test fails until the validator-output check below is updated, forcing a
// conscious decision.
func TestValidSettingsConfigModeStrings(t *testing.T) {
	got := validSettingsConfigModeStrings()

	if len(got) != len(constants.ValidSettingsConfigModes) {
		t.Fatalf("length mismatch: validator returns %d, constants has %d", len(got), len(constants.ValidSettingsConfigModes))
	}

	// Build the expected set straight from the canonical constants slice.
	want := make([]string, 0, len(constants.ValidSettingsConfigModes))
	for _, m := range constants.ValidSettingsConfigModes {
		want = append(want, string(m))
	}
	sort.Strings(got)
	sort.Strings(want)
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d]: want %q, got %q", i, want[i], got[i])
		}
	}
}

// TestValidGroupResourceTypeStrings verifies that the validator's accepted
// set is constants.ValidGroupResourceTypes PLUS the user-facing
// KUBERNETES_DAEMONSET spelling alias. The alias is needed because the
// backend's wire format misspells it as DEAMONSET; HCL accepts either and
// buildGroupDefinition translates.
func TestValidGroupResourceTypeStrings(t *testing.T) {
	got := validGroupResourceTypeStrings()

	// Must include every value from the canonical constants slice.
	gotSet := map[string]struct{}{}
	for _, s := range got {
		gotSet[s] = struct{}{}
	}
	for _, t2 := range constants.ValidGroupResourceTypes {
		if _, ok := gotSet[string(t2)]; !ok {
			t.Errorf("validator missing canonical value %q", string(t2))
		}
	}

	// Must also include the correctly-spelled DAEMONSET alias.
	if _, ok := gotSet["KUBERNETES_DAEMONSET"]; !ok {
		t.Error("validator should accept KUBERNETES_DAEMONSET (correctly-spelled alias)")
	}

	// Must also include the backend's misspelled DEAMONSET (it's in the
	// constants list as the canonical wire spelling).
	if _, ok := gotSet["KUBERNETES_DEAMONSET"]; !ok {
		t.Error("validator should accept KUBERNETES_DEAMONSET (wire spelling)")
	}

	// Sanity: length = canonical + 1 alias.
	wantLen := len(constants.ValidGroupResourceTypes) + 1
	if len(got) != wantLen {
		t.Errorf("length: want %d (constants + DAEMONSET alias), got %d", wantLen, len(got))
	}
}
