package provider

import (
	"context"
	"strings"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &createGroup{}
	_ resource.ResourceWithImportState = &createGroup{}
)

// CreateGroup is a helper function to simplify the provider implementation.
func CreateGroup() resource.Resource {
	return &createGroup{}
}

// createGroup is the resource implementation.
type createGroup struct{}

type groupModel struct {
	ID                     basetypes.StringValue `tfsdk:"id"`
	Name                   string                `tfsdk:"name"`
	Enabled                basetypes.BoolValue   `tfsdk:"enabled"`
	AutoRefresh            basetypes.BoolValue   `tfsdk:"auto_refresh"`
	ParentGroupId          basetypes.StringValue `tfsdk:"parent_group_id"`
	Tags                   basetypes.MapValue    `tfsdk:"tags"`
	Cluster                basetypes.ListValue   `tfsdk:"cluster"`
	Cloud                  basetypes.ListValue   `tfsdk:"cloud"`
	ResourceType           basetypes.ListValue   `tfsdk:"resource_type"`
	ManuallyAddedResources basetypes.ListValue   `tfsdk:"manually_added_resources"`
	Region                 basetypes.ListValue   `tfsdk:"region"`
	Namespace              basetypes.ListValue   `tfsdk:"namespace"`
	Settings               *groupSettingsModel   `tfsdk:"settings"`
}

// groupSettingsModel is the nested block exposing the top-level group
// settings managed by Terraform. Resource-type-specific tuning (kube
// scaling, ECS, Lambda, etc.) is intentionally not modelled here in this
// iteration — the SDK preserves those nested sections opaquely so they
// remain UI-configurable without TF clobbering them.
type groupSettingsModel struct {
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
}

// Metadata returns the resource type name.
func (r *createGroup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_create_group"
}

// Schema defines the schema for the resource.
func (r *createGroup) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Sedai resource group. A group bundles cloud resources matching the filters (tags, clusters, regions, namespaces, resource types) so settings and policies can be applied at the group level.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Sedai group ID (assigned by the backend after creation).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Group name. Must be unique within the Sedai tenant.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the group is active. Disabled groups don't trigger optimization actions. Defaults to `false`; set to `true` to activate the group after creation.",
			},
			"auto_refresh": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, Sedai periodically re-evaluates the group's filters and adds/removes resources as the cloud inventory changes.",
			},
			"parent_group_id": schema.StringAttribute{
				Optional:    true,
				Description: "Parent group ID. Set this to create a subgroup; omit for a top-level group.",
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.ListType{ElemType: types.StringType},
				Description: "Tag filters. Map of tag key to list of values. Values containing `*` are sent as regex matchers; all other values are sent as exact matchers. Example: `{ app = [\"db-backend\", \"api-*\"] }`.",
			},
			"cluster": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Fully-qualified cluster identifiers (e.g. AWS ECS cluster ARNs) to include in the group.",
			},
			"cloud": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Cloud account IDs to include in the group.",
			},
			"resource_type": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Resource type filters. Valid values: `AWS_EBS`, `AWS_EC2`, `AWS_ECS`, `AWS_LAMBDA`, `AWS_S3`, `AWS_TAGS`, `AZURE_LB`, `AZURE_VM`, `KUBERNETES_CRONJOB`, `KUBERNETES_DEAMONSET`, `KUBERNETES_DEPLOYMENT`, `KUBERNETES_STATEFULSET`.",
			},
			"manually_added_resources": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Resource IDs to include in the group regardless of filter matches.",
			},
			"region": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Cloud region filters (e.g. `us-east-1`).",
			},
			"namespace": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Kubernetes namespace filters.",
			},
			"settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Top-level group settings (availability mode, optimization mode, Sedai Sync). Omit this block to leave settings unmanaged by Terraform — UI-configured advanced settings (per-resource-type tuning) are preserved either way.",
				Attributes: map[string]schema.Attribute{
					"availability_mode": schema.StringAttribute{
						Required:    true,
						Description: "Availability mode. Valid values: `DATA_PILOT` (monitor only), `CO_PILOT` (recommend, manual approval), `AUTO` (Sedai acts autonomously).",
					},
					"optimization_mode": schema.StringAttribute{
						Required:    true,
						Description: "Optimization mode. Valid values: `DATA_PILOT` (monitor only), `CO_PILOT` (recommend, manual approval), `AUTO` (Sedai acts autonomously).",
					},
					"sedai_sync_enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "When true, Sedai auto-syncs the group's resources with the latest configuration. Defaults to false if omitted.",
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *createGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	def, diags := buildGroupDefinition(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := groups.CreateGroup(def)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create group", err.Error())
		return
	}

	plan.ID = basetypes.NewStringValue(created.ID)

	// Reconcile enable state with the plan. We always make this call (even
	// when the plan says false) because the backend's default after create is
	// not contractually guaranteed — explicit is safer than implicit.
	if err := groups.EnableOrDisableGroup(created.ID, plan.Enabled.ValueBool()); err != nil {
		// The group exists; save state first so destroy works, then surface
		// the partial failure so the user can retry the toggle.
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		resp.Diagnostics.AddError("Group created but failed to set enable state", err.Error())
		return
	}

	// If the user configured a settings block, initialize and apply it. We
	// init defensively even though it might already be a no-op — keeps the
	// subsequent UpdateGroupSettings call from failing with "not initialized".
	if plan.Settings != nil {
		if err := groups.InitializeGroupSettings(created.ID); err != nil {
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			resp.Diagnostics.AddError("Group created but failed to initialize settings", err.Error())
			return
		}
		if err := groups.UpdateGroupSettings(created.ID, settingsFromPlan(plan.Settings)); err != nil {
			resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			resp.Diagnostics.AddError("Group created but failed to apply settings", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *createGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetched, err := groups.GetGroupById(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching Sedai group",
			"Could not fetch group with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	if fetched == nil || fetched.Definition == nil {
		// Group was deleted out-of-band. Drop it from state so Terraform plans to recreate.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(applyDefinitionToModel(ctx, &state, fetched)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = basetypes.NewStringValue(fetched.GroupId)
	state.Name = fetched.Name

	// Refresh settings from backend ONLY if state already has a settings
	// block — otherwise the user opted out of managing settings via TF and
	// we'd cause spurious drift by populating it.
	if state.Settings != nil {
		settings, err := groups.GetGroupSettings(state.ID.ValueString())
		if err != nil {
			// Don't fail the whole Read for settings refresh issues — surface
			// a warning so the user knows drift detection is degraded.
			resp.Diagnostics.AddWarning("Unable to refresh group settings", err.Error())
		} else if settings != nil {
			state.Settings.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
			state.Settings.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
			state.Settings.SedaiSyncEnabled = basetypes.NewBoolValue(settings.SedaiSyncEnabled)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *createGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	def, diags := buildGroupDefinition(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// URL uses the OLD name (from state) for lookup; the new name lives in the
	// payload. Renames work transparently this way.
	if err := groups.UpdateGroup(state.ID.ValueString(), state.Name, def); err != nil {
		resp.Diagnostics.AddError("Unable to update group", err.Error())
		return
	}

	// Sync enable state only if it changed — saves an API call on definition-
	// only edits, which are the common case.
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		if err := groups.EnableOrDisableGroup(state.ID.ValueString(), plan.Enabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Unable to update group enable state", err.Error())
			return
		}
	}

	// Sync settings if the user is managing them. When the user removes the
	// `settings` block, we intentionally do NOT clear settings on the backend
	// — UI-configured values stay, TF just stops tracking them.
	if plan.Settings != nil {
		if state.Settings == nil {
			// First time managing settings for this group — init before update.
			if err := groups.InitializeGroupSettings(state.ID.ValueString()); err != nil {
				resp.Diagnostics.AddError("Unable to initialize group settings", err.Error())
				return
			}
		}
		if !settingsEqual(plan.Settings, state.Settings) {
			if err := groups.UpdateGroupSettings(state.ID.ValueString(), settingsFromPlan(plan.Settings)); err != nil {
				resp.Diagnostics.AddError("Unable to update group settings", err.Error())
				return
			}
		}
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *createGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := groups.DeleteGroup(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete group", err.Error())
		return
	}
}

// ImportState lets `terraform import sedai_create_group.<name> <group-id>`
// adopt an existing Sedai group into Terraform state. The supplied ID is
// written into the `id` attribute; Terraform then calls Read to populate the
// rest of the model from the backend.
func (r *createGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildGroupDefinition converts the Terraform plan model into a SDK
// GroupDefinition. Tag values containing "*" are routed to the regex bucket;
// all others go to exact.
func buildGroupDefinition(ctx context.Context, plan groupModel) (*groups.GroupDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	def := groups.NewGroupDefinition(plan.Name)

	if plan.AutoRefresh.ValueBool() {
		def.AutoRefresh = true
	}
	if !plan.ParentGroupId.IsNull() && !plan.ParentGroupId.IsUnknown() && plan.ParentGroupId.ValueString() != "" {
		v := plan.ParentGroupId.ValueString()
		def.ParentGroupId = &v
	}

	def.Cluster = listToStrings(plan.Cluster)
	def.Cloud = listToStrings(plan.Cloud)
	def.ResourceType = listToStrings(plan.ResourceType)
	def.ManuallyAddedResources = listToStrings(plan.ManuallyAddedResources)
	def.Region = listToStrings(plan.Region)
	def.Namespace = listToStrings(plan.Namespace)

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		raw := map[string][]string{}
		diags.Append(plan.Tags.ElementsAs(ctx, &raw, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for key, values := range raw {
			tag := groups.GroupTag{Key: key, Regex: []string{}, Exact: []string{}}
			for _, v := range values {
				if strings.Contains(v, "*") {
					tag.Regex = append(tag.Regex, v)
				} else {
					tag.Exact = append(tag.Exact, v)
				}
			}
			def.Tags = append(def.Tags, tag)
		}
	}

	return def, diags
}

// applyDefinitionToModel populates the Terraform model fields from a
// GroupDetails fetched from the backend. Tag regex+exact buckets are merged
// back into the map<string,list<string>> form used in HCL.
func applyDefinitionToModel(ctx context.Context, state *groupModel, fetched *groups.GroupDetails) diag.Diagnostics {
	var diags diag.Diagnostics
	def := fetched.Definition

	// Refresh enable state from the backend if available. If IsActive is nil
	// (backend didn't surface the field under a name we recognize), leave
	// state.Enabled untouched — better to miss a drift event than to assert a
	// false `true -> false` diff on every plan.
	if fetched.IsActive != nil {
		state.Enabled = basetypes.NewBoolValue(*fetched.IsActive)
	}

	state.AutoRefresh = basetypes.NewBoolValue(def.AutoRefresh)
	if def.ParentGroupId != nil && *def.ParentGroupId != "" {
		state.ParentGroupId = basetypes.NewStringValue(*def.ParentGroupId)
	} else {
		state.ParentGroupId = basetypes.NewStringNull()
	}

	state.Cluster = stringsToList(def.Cluster)
	state.Cloud = stringsToList(def.Cloud)
	state.ResourceType = stringsToList(def.ResourceType)
	state.ManuallyAddedResources = stringsToList(def.ManuallyAddedResources)
	state.Region = stringsToList(def.Region)
	state.Namespace = stringsToList(def.Namespace)

	if len(def.Tags) == 0 {
		state.Tags = basetypes.NewMapNull(types.ListType{ElemType: types.StringType})
		return diags
	}
	tagsRaw := map[string][]string{}
	for _, t := range def.Tags {
		merged := append([]string{}, t.Exact...)
		merged = append(merged, t.Regex...)
		tagsRaw[t.Key] = merged
	}
	tagsMap, d := basetypes.NewMapValueFrom(ctx, types.ListType{ElemType: types.StringType}, tagsRaw)
	diags.Append(d...)
	state.Tags = tagsMap
	return diags
}

// listToStrings converts a Terraform ListValue of strings into a plain
// []string. Returns an empty slice (not nil) so the resulting JSON is `[]`.
func listToStrings(lv basetypes.ListValue) []string {
	if lv.IsNull() || lv.IsUnknown() {
		return []string{}
	}
	elements := lv.Elements()
	out := make([]string, 0, len(elements))
	for _, sv := range elements {
		if s, ok := sv.(basetypes.StringValue); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

// settingsFromPlan converts the TF-side groupSettingsModel into the SDK's
// GroupSettings. Empty / null mode strings are passed through as empty
// strings — UpdateGroupSettings skips empty modes (read-modify-write keeps
// the existing backend value), so a partial update is safe.
func settingsFromPlan(m *groupSettingsModel) *groups.GroupSettings {
	if m == nil {
		return nil
	}
	return &groups.GroupSettings{
		AvailabilityMode: m.AvailabilityMode.ValueString(),
		OptimizationMode: m.OptimizationMode.ValueString(),
		SedaiSyncEnabled: m.SedaiSyncEnabled.ValueBool(),
	}
}

// settingsEqual returns true when two tfsdk settings models have identical
// values. Nil vs non-nil is a difference. Used by Update to skip an API
// call when plan and state already match.
func settingsEqual(a, b *groupSettingsModel) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.AvailabilityMode.ValueString() == b.AvailabilityMode.ValueString() &&
		a.OptimizationMode.ValueString() == b.OptimizationMode.ValueString() &&
		a.SedaiSyncEnabled.ValueBool() == b.SedaiSyncEnabled.ValueBool()
}

// stringsToList converts a []string back into a Terraform ListValue. Nil/empty
// input becomes a typed null list so Terraform doesn't diff `[]` vs `null`.
func stringsToList(values []string) basetypes.ListValue {
	if len(values) == 0 {
		return basetypes.NewListNull(types.StringType)
	}
	elements := make([]basetypes.StringValue, 0, len(values))
	for _, v := range values {
		elements = append(elements, basetypes.NewStringValue(v))
	}
	lv, _ := basetypes.NewListValueFrom(context.Background(), types.StringType, elements)
	return lv
}
