package provider

import (
	"context"
	"strings"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure interfaces are satisfied.
var (
	_ resource.Resource                = &groupPriority{}
	_ resource.ResourceWithImportState = &groupPriority{}
)

// GroupPriority is the resource constructor for `sedai_group_priority`.
func GroupPriority() resource.Resource {
	return &groupPriority{}
}

// groupPriority is the resource implementation. Resource type name:
// `sedai_group_priority`. Manages the relative priority ordering of a set
// of sibling Sedai groups in a single batch.
//
// Priority is 1-based (1 = highest), matching the Sedai UI and the
// requirements doc. The SDK translates to/from the backend's 0-based
// representation internally.
//
// Delete is a no-op per spec — there is no "unset priority" API. The
// resource just leaves Terraform state when destroyed; backend priorities
// remain as last set.
type groupPriority struct{}

type groupPriorityResourceModel struct {
	ID              basetypes.StringValue     `tfsdk:"id"`
	GroupPriorities []groupPriorityBlockModel `tfsdk:"group_priorities"`
}

// groupPriorityBlockModel is one entry in the repeatable
// `group_priorities { … }` block on `sedai_group_priority`.
type groupPriorityBlockModel struct {
	GroupID  string `tfsdk:"group_id"`
	Priority int64  `tfsdk:"priority"`
}

func (r *groupPriority) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_priority"
}

func (r *groupPriority) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages relative priority ordering for a set of Sedai groups. Priority decides which group's settings take precedence when a resource belongs to multiple groups (lower number = higher priority; 1 is highest). All referenced groups must be enabled, otherwise the backend will reject the assignment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID = comma-separated group IDs managed by this resource. Used for import and identification in state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"group_priorities": schema.ListNestedBlock{
				Description: "Repeatable block — each entry assigns a priority to one group. The order of blocks does not matter; the `priority` value is what decides the ordering.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Required:    true,
							Description: "The Sedai group ID. Typically `sedai_group.<name>.id`. The referenced group must be enabled.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"priority": schema.Int64Attribute{
							Required:    true,
							Description: "Priority value. 1-based, where 1 is the highest priority. Must be >= 1.",
							Validators:  []validator.Int64{priorityValidator()},
						},
					},
				},
			},
		},
	}
}

func (r *groupPriority) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupPriorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statuses := r.applyPlan(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write confirmed priorities from the update response into state.
	// This avoids an immediate Read that would hit the stale 12h cache —
	// the response reflects what was actually accepted by the backend.
	confirmedByGroup := make(map[string]int, len(statuses))
	for _, s := range statuses {
		if s.Success {
			confirmedByGroup[s.GroupID] = s.RequestedPriority
		}
	}
	for i, entry := range plan.GroupPriorities {
		if confirmed, ok := confirmedByGroup[entry.GroupID]; ok {
			plan.GroupPriorities[i].Priority = int64(confirmed)
		}
	}

	plan.ID = basetypes.NewStringValue(joinGroupIDs(plan.GroupPriorities))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupPriority) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupPriorityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh each group's priority from the backend so drift via the UI is
	// detected. If the backend response doesn't surface a priority field
	// (Priority == nil on details), preserve the state value — better to
	// miss a drift event than to assert a false diff on every plan.
	missingAny := false
	for i, entry := range state.GroupPriorities {
		details, err := groups.GetGroupById(entry.GroupID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Unable to refresh group priority",
				"Could not fetch group "+entry.GroupID+": "+err.Error(),
			)
			continue
		}
		if details == nil {
			// Group was deleted out-of-band. Drop the whole priority resource
			// from state — the user's plan will need to be re-derived.
			missingAny = true
			continue
		}
		if details.Priority != nil {
			state.GroupPriorities[i].Priority = int64(*details.Priority)
		}
	}

	if missingAny {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = basetypes.NewStringValue(joinGroupIDs(state.GroupPriorities))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *groupPriority) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupPriorityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	statuses := r.applyPlan(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write confirmed priorities from the update response into state.
	// This avoids an immediate Read that would hit the stale 12h cache —
	// the response reflects what was actually accepted by the backend.
	confirmedByGroup := make(map[string]int, len(statuses))
	for _, s := range statuses {
		if s.Success {
			confirmedByGroup[s.GroupID] = s.RequestedPriority
		}
	}
	for i, entry := range plan.GroupPriorities {
		if confirmed, ok := confirmedByGroup[entry.GroupID]; ok {
			plan.GroupPriorities[i].Priority = int64(confirmed)
		}
	}

	plan.ID = basetypes.NewStringValue(joinGroupIDs(plan.GroupPriorities))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete is a no-op per the requirements doc — priorities have no "unset"
// concept on the backend, so we just remove the resource from Terraform
// state. Backend priorities remain as last set.
func (r *groupPriority) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState parses the import ID — a comma-separated list of group IDs —
// and writes a state shell with one priority entry per group ID. Priorities
// default to 0 (an invalid value the user is expected to fix). The next
// Read fills in the actual values from the backend.
func (r *groupPriority) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ids := splitGroupIDs(req.ID)
	if len(ids) == 0 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected a comma-separated list of group IDs (e.g. `abc123,def456`).",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	priorities := make([]groupPriorityBlockModel, 0, len(ids))
	for _, id := range ids {
		priorities = append(priorities, groupPriorityBlockModel{GroupID: id, Priority: 0})
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_priorities"), priorities)...)
}

// applyPlan converts the plan into SDK GroupPriority entries, calls the
// batch update, and reports any per-entry failures back as diagnostics.
func (r *groupPriority) applyPlan(_ context.Context, plan groupPriorityResourceModel, diags *diag.Diagnostics) []groups.GroupPriorityUpdateStatus {
	if len(plan.GroupPriorities) == 0 {
		diags.AddError("Empty group_priorities", "At least one `group_priorities` block is required.")
		return nil
	}

	sdkEntries := make([]groups.GroupPriority, 0, len(plan.GroupPriorities))
	for _, p := range plan.GroupPriorities {
		sdkEntries = append(sdkEntries, groups.GroupPriority{
			GroupID:  p.GroupID,
			Priority: int(p.Priority),
		})
	}

	statuses, err := groups.UpdateGroupPriorities(sdkEntries)
	if err != nil {
		diags.AddError("Unable to update group priorities", err.Error())
		return nil
	}
	for _, s := range statuses {
		if !s.Success {
			diags.AddError(
				"Group priority update rejected",
				"Group "+s.GroupID+": "+s.Message,
			)
		}
	}
	return statuses
}

// joinGroupIDs builds the composite resource ID (comma-separated group IDs)
// in the order the entries appear in the plan. The order of group_priorities
// blocks is the source of truth; sorting would let a re-order in HCL look
// like a no-op when it isn't.
func joinGroupIDs(entries []groupPriorityBlockModel) string {
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, e.GroupID)
	}
	return strings.Join(parts, ",")
}

// splitGroupIDs parses the composite import ID. Empty / whitespace-only
// entries are dropped.
func splitGroupIDs(id string) []string {
	raw := strings.Split(id, ",")
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

