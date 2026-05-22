package provider

import (
	"context"
	"sort"
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

// newTagsMap is a test helper that builds a non-null MapValue<List<String>>.
func newTagsMap(t *testing.T, raw map[string][]string) basetypes.MapValue {
	t.Helper()
	mv, diags := basetypes.NewMapValueFrom(context.Background(), types.ListType{ElemType: types.StringType}, raw)
	if diags.HasError() {
		t.Fatalf("newTagsMap: %v", diags)
	}
	return mv
}

// TestBuildGroupDefinition_TagAutoSplit verifies the core auto-split rule:
// tag values containing '*' go to regex[], all others to exact[].
func TestBuildGroupDefinition_TagAutoSplit(t *testing.T) {
	plan := groupModel{
		Name: "g",
		Tags: newTagsMap(t, map[string][]string{
			"app": {"db-backend", "api-*", "frontend"},
			"env": {"prod"},
		}),
		Cluster:                basetypes.NewListNull(types.StringType),
		Cloud:                  basetypes.NewListNull(types.StringType),
		ResourceType:           basetypes.NewListNull(types.StringType),
		ManuallyAddedResources: basetypes.NewListNull(types.StringType),
		Region:                 basetypes.NewListNull(types.StringType),
		Namespace:              basetypes.NewListNull(types.StringType),
	}

	def, diags := buildGroupDefinition(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("buildGroupDefinition: %v", diags)
	}

	// Find each tag by key (order of map iteration is not deterministic).
	tagByKey := map[string]groups.GroupTag{}
	for _, tag := range def.Tags {
		tagByKey[tag.Key] = tag
	}
	if len(tagByKey) != 2 {
		t.Fatalf("want 2 tag keys, got %d (%v)", len(tagByKey), def.Tags)
	}

	appTag, ok := tagByKey["app"]
	if !ok {
		t.Fatal("missing 'app' tag")
	}
	sort.Strings(appTag.Exact)
	if len(appTag.Exact) != 2 || appTag.Exact[0] != "db-backend" || appTag.Exact[1] != "frontend" {
		t.Errorf("app.exact: want [db-backend frontend], got %v", appTag.Exact)
	}
	if len(appTag.Regex) != 1 || appTag.Regex[0] != "api-*" {
		t.Errorf("app.regex: want [api-*], got %v", appTag.Regex)
	}

	envTag, ok := tagByKey["env"]
	if !ok {
		t.Fatal("missing 'env' tag")
	}
	if len(envTag.Exact) != 1 || envTag.Exact[0] != "prod" {
		t.Errorf("env.exact: want [prod], got %v", envTag.Exact)
	}
	if len(envTag.Regex) != 0 {
		t.Errorf("env.regex: want empty, got %v", envTag.Regex)
	}
}

// TestBuildGroupDefinition_ParentGroupId verifies that an empty/null
// parent_group_id results in nil on the SDK side, and a real value flows
// through as *string.
func TestBuildGroupDefinition_ParentGroupId(t *testing.T) {
	base := func() groupModel {
		return groupModel{
			Name:                   "g",
			Tags:                   basetypes.NewMapNull(types.ListType{ElemType: types.StringType}),
			Cluster:                basetypes.NewListNull(types.StringType),
			Cloud:                  basetypes.NewListNull(types.StringType),
			ResourceType:           basetypes.NewListNull(types.StringType),
			ManuallyAddedResources: basetypes.NewListNull(types.StringType),
			Region:                 basetypes.NewListNull(types.StringType),
			Namespace:              basetypes.NewListNull(types.StringType),
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
		Name:                   "g",
		Tags:                   basetypes.NewMapNull(types.ListType{ElemType: types.StringType}),
		Cluster:                newStringList(t, []string{"c1", "c2"}),
		Cloud:                  newStringList(t, []string{}),
		ResourceType:           newStringList(t, []string{"AWS_EC2"}),
		ManuallyAddedResources: basetypes.NewListNull(types.StringType),
		Region:                 newStringList(t, []string{"us-east-1"}),
		Namespace:              basetypes.NewListNull(types.StringType),
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
		t.Error("ManuallyAddedResources: want empty non-nil for JSON shape")
	}
}

// TestApplyDefinitionToModel_TagMerge verifies regex+exact buckets are merged
// back into a single list per key on the round-trip from backend to TF state.
func TestApplyDefinitionToModel_TagMerge(t *testing.T) {
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

	var got map[string][]string
	d := state.Tags.ElementsAs(context.Background(), &got, false)
	if d.HasError() {
		t.Fatalf("ElementsAs: %v", d)
	}

	if len(got["app"]) != 2 {
		t.Errorf("app: want 2 entries, got %v", got["app"])
	}
	if len(got["env"]) != 1 || got["env"][0] != "prod" {
		t.Errorf("env: want [prod], got %v", got["env"])
	}
}

// TestApplyDefinitionToModel_NoTags verifies that an empty Tags slice in the
// definition produces a typed null map (not an empty map) — this prevents
// Terraform from diffing `null` vs `{}` on refresh.
func TestApplyDefinitionToModel_NoTags(t *testing.T) {
	state := &groupModel{}
	fetched := &groups.GroupDetails{
		Definition: &groups.GroupDefinition{Tags: []groups.GroupTag{}},
	}

	diags := applyDefinitionToModel(context.Background(), state, fetched)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if !state.Tags.IsNull() {
		t.Errorf("expected null tags map, got %v", state.Tags)
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

// TestSettingsFromPlan_NilGivesNil ensures that passing a nil settings
// model doesn't panic and yields nil — the path Create/Update takes when
// the user omits the settings block from HCL.
func TestSettingsFromPlan_NilGivesNil(t *testing.T) {
	if got := settingsFromPlan(nil); got != nil {
		t.Errorf("nil input: want nil, got %v", got)
	}
}

// TestSettingsFromPlan_PopulatesFields verifies the mapping from the
// tfsdk model to the SDK request struct.
func TestSettingsFromPlan_PopulatesFields(t *testing.T) {
	plan := &groupSettingsModel{
		AvailabilityMode: basetypes.NewStringValue("AUTO"),
		OptimizationMode: basetypes.NewStringValue("CO_PILOT"),
		SedaiSyncEnabled: basetypes.NewBoolValue(true),
	}
	got := settingsFromPlan(plan)
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

// TestSettingsEqual covers all 4 quadrants of nil-handling plus a real
// diff and an exact match. This drives the Update flow's
// "skip the API call if nothing changed" optimization.
func TestSettingsEqual(t *testing.T) {
	mk := func(av, opt string, sync bool) *groupSettingsModel {
		return &groupSettingsModel{
			AvailabilityMode: basetypes.NewStringValue(av),
			OptimizationMode: basetypes.NewStringValue(opt),
			SedaiSyncEnabled: basetypes.NewBoolValue(sync),
		}
	}
	cases := []struct {
		name string
		a, b *groupSettingsModel
		want bool
	}{
		{"both nil", nil, nil, true},
		{"a nil, b set", nil, mk("AUTO", "AUTO", false), false},
		{"a set, b nil", mk("AUTO", "AUTO", false), nil, false},
		{"exact match", mk("AUTO", "CO_PILOT", true), mk("AUTO", "CO_PILOT", true), true},
		{"different availability", mk("AUTO", "AUTO", false), mk("CO_PILOT", "AUTO", false), false},
		{"different optimization", mk("AUTO", "AUTO", false), mk("AUTO", "CO_PILOT", false), false},
		{"different sync", mk("AUTO", "AUTO", false), mk("AUTO", "AUTO", true), false},
	}
	for _, c := range cases {
		if got := settingsEqual(c.a, c.b); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
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
