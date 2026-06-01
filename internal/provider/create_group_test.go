package provider

import (
	"context"
	"testing"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// newStringList is a test helper that builds a non-null ListValue<String>.
func newStringList(t *testing.T, values []string) basetypes.ListValue {
	t.Helper()
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, basetypes.NewStringValue(v))
	}
	lv, diags := basetypes.NewListValue(types.StringType, elements)
	if diags.HasError() {
		t.Fatalf("newStringList: %v", diags)
	}
	return lv
}

// TestBuildGroupDefinition_TagsExplicitBuckets verifies that tag blocks pass
// through verbatim — exact values go to exact, regex values go to regex, the
// user gets full control with no auto-routing.
func TestBuildGroupDefinition_TagsExplicitBuckets(t *testing.T) {
	plan := groupModel{
		Name: basetypes.NewStringValue("g"),
		Tags: []tagBlockModel{
			{Key: "env", Exact: []string{"prod", "production"}, Regex: nil},
			{Key: "team", Exact: nil, Regex: []string{"platform-*"}},
			{Key: "app", Exact: []string{"db"}, Regex: []string{"api-*"}},
		},
		Clusters:        basetypes.NewListNull(types.StringType),
		CloudAccountIDs: basetypes.NewListNull(types.StringType),
		ResourceTypes:   basetypes.NewListNull(types.StringType),
		ResourceIDs:     basetypes.NewListNull(types.StringType),
		Regions:         basetypes.NewListNull(types.StringType),
		Namespaces:      basetypes.NewListNull(types.StringType),
	}

	def, diags := buildGroupDefinition(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("buildGroupDefinition: %v", diags)
	}

	tagByKey := map[string]groups.GroupTag{}
	for _, tag := range def.Tags {
		tagByKey[tag.Key] = tag
	}
	if len(tagByKey) != 3 {
		t.Fatalf("want 3 tag keys, got %d (%v)", len(tagByKey), def.Tags)
	}

	env := tagByKey["env"]
	if len(env.Exact) != 2 || env.Exact[0] != "prod" || env.Exact[1] != "production" {
		t.Errorf("env.exact: want [prod production], got %v", env.Exact)
	}
	if len(env.Regex) != 0 {
		t.Errorf("env.regex: want [] (nil normalized), got %v", env.Regex)
	}

	team := tagByKey["team"]
	if len(team.Exact) != 0 {
		t.Errorf("team.exact: want [], got %v", team.Exact)
	}
	if len(team.Regex) != 1 || team.Regex[0] != "platform-*" {
		t.Errorf("team.regex: want [platform-*], got %v", team.Regex)
	}

	app := tagByKey["app"]
	if len(app.Exact) != 1 || app.Exact[0] != "db" {
		t.Errorf("app.exact: want [db], got %v", app.Exact)
	}
	if len(app.Regex) != 1 || app.Regex[0] != "api-*" {
		t.Errorf("app.regex: want [api-*], got %v", app.Regex)
	}
}

// TestBuildGroupDefinition_ParentGroupId verifies that an empty/null
// parent_group_id results in nil on the SDK side, and a real value flows
// through as *string.
func TestBuildGroupDefinition_ParentGroupId(t *testing.T) {
	base := func() groupModel {
		return groupModel{
			Name:            basetypes.NewStringValue("g"),
			Tags:            nil,
			Clusters:        basetypes.NewListNull(types.StringType),
			CloudAccountIDs: basetypes.NewListNull(types.StringType),
			ResourceTypes:   basetypes.NewListNull(types.StringType),
			ResourceIDs:     basetypes.NewListNull(types.StringType),
			Regions:         basetypes.NewListNull(types.StringType),
			Namespaces:      basetypes.NewListNull(types.StringType),
		}
	}

	// Null → nil
	p := base()
	p.ParentGroupId = basetypes.NewStringNull()
	def, _ := buildGroupDefinition(context.Background(), p)
	if def.ParentGroupId != nil {
		t.Errorf("null parent: want nil, got %v", *def.ParentGroupId)
	}

	// Empty string → still nil (don't send empty string to backend)
	p = base()
	p.ParentGroupId = basetypes.NewStringValue("")
	def, _ = buildGroupDefinition(context.Background(), p)
	if def.ParentGroupId != nil {
		t.Errorf("empty-string parent: want nil, got %v", *def.ParentGroupId)
	}

	// Real value → *string
	p = base()
	p.ParentGroupId = basetypes.NewStringValue("parent-xyz")
	def, _ = buildGroupDefinition(context.Background(), p)
	if def.ParentGroupId == nil || *def.ParentGroupId != "parent-xyz" {
		t.Errorf("set parent: want parent-xyz, got %v", def.ParentGroupId)
	}
}

// TestBuildGroupDefinition_ListFields verifies that ListValue fields are
// translated to plain []string, including the null-list case.
func TestBuildGroupDefinition_ListFields(t *testing.T) {
	plan := groupModel{
		Name:            basetypes.NewStringValue("g"),
		Tags:            nil,
		Clusters:        newStringList(t, []string{"c1", "c2"}),
		CloudAccountIDs: newStringList(t, []string{}),
		ResourceTypes:   newStringList(t, []string{"AWS_EC2"}),
		ResourceIDs:     basetypes.NewListNull(types.StringType),
		Regions:         newStringList(t, []string{"us-east-1"}),
		Namespaces:      basetypes.NewListNull(types.StringType),
	}

	def, diags := buildGroupDefinition(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if len(def.Cluster) != 2 || def.Cluster[0] != "c1" || def.Cluster[1] != "c2" {
		t.Errorf("Cluster: got %v", def.Cluster)
	}
	if len(def.Cloud) != 0 || def.Cloud == nil {
		t.Errorf("Cloud: want empty non-nil, got %v", def.Cloud)
	}
	if len(def.ResourceType) != 1 || def.ResourceType[0] != "AWS_EC2" {
		t.Errorf("ResourceType: got %v", def.ResourceType)
	}
	if def.ManuallyAddedResources == nil {
		t.Error("ResourceIDs:     want empty non-nil for JSON shape")
	}
}

// TestApplyDefinitionToModel_TagsRoundTrip verifies tag blocks round-trip
// from backend to TF state preserving the exact/regex split.
func TestApplyDefinitionToModel_TagsRoundTrip(t *testing.T) {
	state := &groupModel{}
	fetched := &groups.GroupDetails{
		GroupId: "id-1",
		Name:    "g",
		Definition: &groups.GroupDefinition{
			Name: "g",
			Tags: []groups.GroupTag{
				{Key: "app", Exact: []string{"db-backend"}, Regex: []string{"api-*"}},
				{Key: "env", Exact: []string{"prod"}, Regex: []string{}},
			},
		},
	}

	diags := applyDefinitionToModel(context.Background(), state, fetched)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}

	if len(state.Tags) != 2 {
		t.Fatalf("want 2 tag blocks, got %d (%v)", len(state.Tags), state.Tags)
	}
	byKey := map[string]tagBlockModel{}
	for _, tb := range state.Tags {
		byKey[tb.Key] = tb
	}

	app := byKey["app"]
	if len(app.Exact) != 1 || app.Exact[0] != "db-backend" {
		t.Errorf("app.Exact: got %v", app.Exact)
	}
	if len(app.Regex) != 1 || app.Regex[0] != "api-*" {
		t.Errorf("app.Regex: got %v", app.Regex)
	}

	env := byKey["env"]
	if len(env.Exact) != 1 || env.Exact[0] != "prod" {
		t.Errorf("env.Exact: got %v", env.Exact)
	}
	if len(env.Regex) != 0 {
		t.Errorf("env.Regex: got %v", env.Regex)
	}
}

// TestApplyDefinitionToModel_NoTags verifies that an empty Tags slice in the
// definition results in nil state.Tags (not an empty slice) — keeps the plan
// output free of phantom block churn.
func TestApplyDefinitionToModel_NoTags(t *testing.T) {
	state := &groupModel{}
	fetched := &groups.GroupDetails{
		Definition: &groups.GroupDefinition{Tags: []groups.GroupTag{}},
	}

	diags := applyDefinitionToModel(context.Background(), state, fetched)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if state.Tags != nil {
		t.Errorf("expected nil tags slice, got %v", state.Tags)
	}
}

// TestApplyDefinitionToModel_EnabledRefresh verifies that when the SDK
// surfaces IsActive (non-nil), state.Enabled is updated from the backend.
// This is the drift-detection path: someone toggled the group in the UI and
// Read should pick that up.
func TestApplyDefinitionToModel_EnabledRefresh(t *testing.T) {
	active := false
	cases := []struct {
		name      string
		startWith bool
		isActive  bool
	}{
		{"true -> false", true, false},
		{"false -> true", false, true},
		{"true -> true (no-op)", true, true},
	}
	for _, c := range cases {
		state := &groupModel{Enabled: basetypes.NewBoolValue(c.startWith)}
		isActive := c.isActive
		fetched := &groups.GroupDetails{
			Definition: &groups.GroupDefinition{},
			IsActive:   &isActive,
		}
		diags := applyDefinitionToModel(context.Background(), state, fetched)
		if diags.HasError() {
			t.Fatalf("%s: %v", c.name, diags)
		}
		if state.Enabled.ValueBool() != c.isActive {
			t.Errorf("%s: want %v, got %v", c.name, c.isActive, state.Enabled.ValueBool())
		}
	}
	_ = active
}

// TestApplyDefinitionToModel_EnabledPreservedWhenNil verifies that when the
// backend response didn't surface an enabled-status field (IsActive == nil),
// state.Enabled is preserved as-is. This avoids spurious diffs from a
// stale-or-unknown backend shape.
func TestApplyDefinitionToModel_EnabledPreservedWhenNil(t *testing.T) {
	state := &groupModel{Enabled: basetypes.NewBoolValue(true)}
	fetched := &groups.GroupDetails{
		Definition: &groups.GroupDefinition{},
		IsActive:   nil,
	}
	diags := applyDefinitionToModel(context.Background(), state, fetched)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if state.Enabled.ValueBool() != true {
		t.Errorf("expected state.Enabled preserved as true, got %v", state.Enabled.ValueBool())
	}
}

// TestApplyDefinitionToModel_ParentGroupId verifies the round-trip of a nil
// ParentGroupId from the SDK (group is not a subgroup).
func TestApplyDefinitionToModel_ParentGroupId(t *testing.T) {
	parent := "parent-xyz"
	cases := []struct {
		name     string
		in       *string
		wantNull bool
		wantStr  string
	}{
		{"nil parent", nil, true, ""},
		{"set parent", &parent, false, "parent-xyz"},
	}
	for _, c := range cases {
		state := &groupModel{}
		fetched := &groups.GroupDetails{
			Definition: &groups.GroupDefinition{ParentGroupId: c.in},
		}
		diags := applyDefinitionToModel(context.Background(), state, fetched)
		if diags.HasError() {
			t.Fatalf("%s: %v", c.name, diags)
		}
		if c.wantNull {
			if !state.ParentGroupId.IsNull() {
				t.Errorf("%s: want null, got %q", c.name, state.ParentGroupId.ValueString())
			}
		} else {
			if state.ParentGroupId.ValueString() != c.wantStr {
				t.Errorf("%s: want %q, got %q", c.name, c.wantStr, state.ParentGroupId.ValueString())
			}
		}
	}
}

// TestListToStrings_NullAndUnknown verifies null/unknown lists become empty
// (not nil) slices, so the JSON shape is `[]` not `null`.
func TestListToStrings_NullAndUnknown(t *testing.T) {
	if got := listToStrings(basetypes.NewListNull(types.StringType)); len(got) != 0 || got == nil {
		t.Errorf("null list: want empty non-nil, got %v", got)
	}
	if got := listToStrings(basetypes.NewListUnknown(types.StringType)); len(got) != 0 || got == nil {
		t.Errorf("unknown list: want empty non-nil, got %v", got)
	}
}

// TestStringsToList_Empty verifies that []string{} becomes a typed-null list,
// so Terraform doesn't see drift between an unset field and an empty list.
func TestStringsToList_Empty(t *testing.T) {
	lv := stringsToList(nil)
	if !lv.IsNull() {
		t.Errorf("nil input: want null list, got %v", lv)
	}
	lv = stringsToList([]string{})
	if !lv.IsNull() {
		t.Errorf("empty input: want null list, got %v", lv)
	}
}

// TestGroupSettingsRequestFromPlan verifies the standalone
// sedai_group_settings resource correctly maps its tfsdk model into the
// SDK's GroupSettings struct.
func TestGroupSettingsRequestFromPlan(t *testing.T) {
	plan := groupSettingsResourceModel{
		GroupID:          basetypes.NewStringValue("grp-1"),
		AvailabilityMode: basetypes.NewStringValue("AUTO"),
		OptimizationMode: basetypes.NewStringValue("CO_PILOT"),
		SedaiSyncEnabled: basetypes.NewBoolValue(true),
	}
	got := groupSettingsRequestFromPlan(plan)
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
}

// TestStringsToListRoundTrip verifies that values survive the
// []string → ListValue → []string round-trip unchanged.
func TestStringsToListRoundTrip(t *testing.T) {
	want := []string{"alpha", "beta", "gamma"}
	lv := stringsToList(want)
	got := listToStrings(lv)
	if len(got) != len(want) {
		t.Fatalf("round-trip length: want %d, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("round-trip[%d]: want %s, got %s", i, want[i], got[i])
		}
	}
}
