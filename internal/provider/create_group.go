package provider

import (
	"context"
	"errors"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &group{}
	_ resource.ResourceWithImportState = &group{}
)

// Group is a helper function to simplify the provider implementation.
func Group() resource.Resource {
	return &group{}
}

// group is the resource implementation. Resource type name: `sedai_group`.
type group struct{}

type groupModel struct {
	ID              basetypes.StringValue `tfsdk:"id"`
	Name            basetypes.StringValue `tfsdk:"name"`
	Enabled         basetypes.BoolValue   `tfsdk:"enabled"`
	AutoRefresh     basetypes.BoolValue   `tfsdk:"auto_refresh"`
	ParentGroupId   basetypes.StringValue `tfsdk:"parent_group_id"`
	Tags            []tagBlockModel       `tfsdk:"tags"`
	Clusters        basetypes.ListValue   `tfsdk:"clusters"`
	CloudAccountIDs basetypes.ListValue   `tfsdk:"cloud_account_ids"`
	ResourceTypes   basetypes.ListValue   `tfsdk:"resource_types"`
	ResourceIDs     basetypes.ListValue   `tfsdk:"resource_ids"`
	Regions         basetypes.ListValue   `tfsdk:"regions"`
	Namespaces      basetypes.ListValue   `tfsdk:"namespaces"`

	// Computed: matched-resource counts populated from the backend on
	// every refresh. Mirrors the data-source side (data.sedai_group)
	// so customers can read the same numbers from either form.
	LambdaCount    basetypes.Int64Value `tfsdk:"lambda_count"`
	EC2Count       basetypes.Int64Value `tfsdk:"ec2_count"`
	ECSCount       basetypes.Int64Value `tfsdk:"ecs_count"`
	KubeCount      basetypes.Int64Value `tfsdk:"kube_count"`
	S3Count        basetypes.Int64Value `tfsdk:"s3_count"`
	EBSCount       basetypes.Int64Value `tfsdk:"ebs_count"`
	AzureVMCount   basetypes.Int64Value `tfsdk:"azure_vm_count"`
	AzureLBCount   basetypes.Int64Value `tfsdk:"azure_lb_count"`
	StreamingCount basetypes.Int64Value `tfsdk:"streaming_count"`
	// ResourceCounts is the full dynamic per-type tally (superset of the
	// named *_count fields above). Keyed by backend resource type.
	ResourceCounts basetypes.MapValue `tfsdk:"resource_counts"`
}

// tagBlockModel is one entry of a repeatable `tags { ... }` block. Matches
// the shape in the requirements doc: each block names a single tag key plus
// up to two bucket lists (exact-match values and regex-match values). Either
// or both lists may be omitted but the key is required.
type tagBlockModel struct {
	Key   string   `tfsdk:"key"`
	Exact []string `tfsdk:"exact"`
	Regex []string `tfsdk:"regex"`
}

// Metadata returns the resource type name.
func (r *group) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema defines the schema for the resource.
func (r *group) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Default:     booldefault.StaticBool(true),
				Description: "Whether the group is active. Disabled groups don't trigger optimization actions. Defaults to `true` (group is active on creation).",
			},
			"auto_refresh": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "When true, Sedai periodically re-evaluates the group's filters and adds/removes resources as the cloud inventory changes. Defaults to `true`.",
			},
			"parent_group_id": schema.StringAttribute{
				Optional:    true,
				Description: "Parent group ID. Set this to create a subgroup; omit for a top-level group.",
			},
			"clusters": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Fully-qualified cluster identifiers (e.g. AWS ECS cluster ARNs) to include in the group.",
			},
			"cloud_account_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Sedai account IDs to scope the group to.",
			},
			"resource_types": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Resource type filters. Valid values: `AWS_LAMBDA`, `AWS_EC2`, `AWS_ECS`, `AWS_EBS`, `AWS_S3`, `AWS_TAGS`, `AWS_LB`, `AWS_ASG`, `AWS_RDS`, `AWS_DYNAMODB`, `AZURE_LB`, `AZURE_VM`, `AZURE_TAGS`, `AZURE_VMSS`, `AZURE_DISK`, `AZURE_STORAGE_BUCKET`, `GCP_VM_INSTANCE`, `GCP_DISK`, `GCP_SNAPSHOT`, `GCP_LB`, `GCP_BUCKET`, `GCP_TAGS`, `GCP_BACKEND_SERVICE`, `GCP_DATAFLOW_STREAMING_JOB`, `GCP_DATAFLOW_BATCH_JOB`, `GCP_BIGQUERY_TRANSFER_CONFIG`, `GCP_BIGQUERY_RESERVATION`, `GCP_BIGQUERY_ASSIGNMENT`, `KUBERNETES_DEPLOYMENT`, `KUBERNETES_STATEFULSET`, `KUBERNETES_DAEMONSET`, `KUBERNETES_CRONJOB`.",
				Validators:  []validator.List{groupResourceTypeListValidator()},
			},
			"resource_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Specific resource IDs to include in the group regardless of filter matches.",
			},
			"regions": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Cloud region filters (e.g. `us-east-1`).",
			},
			"namespaces": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Kubernetes namespace filters.",
			},

			// Computed counts — refreshed from the backend on every read.
			// Values reflect how many resources currently match the group's
			// filters per resource type.
			"lambda_count":    schema.Int64Attribute{Computed: true, Description: "Number of AWS Lambda functions matched."},
			"ec2_count":       schema.Int64Attribute{Computed: true, Description: "Number of AWS EC2 instances matched."},
			"ecs_count":       schema.Int64Attribute{Computed: true, Description: "Number of AWS ECS services matched."},
			"kube_count":      schema.Int64Attribute{Computed: true, Description: "Number of Kubernetes workloads matched."},
			"s3_count":        schema.Int64Attribute{Computed: true, Description: "Number of AWS S3 buckets matched."},
			"ebs_count":       schema.Int64Attribute{Computed: true, Description: "Number of AWS EBS volumes matched."},
			"azure_vm_count":  schema.Int64Attribute{Computed: true, Description: "Number of Azure VMs matched."},
			"azure_lb_count":  schema.Int64Attribute{Computed: true, Description: "Number of Azure load balancers matched."},
			"streaming_count": schema.Int64Attribute{Computed: true, Description: "Number of streaming resources matched."},
			"resource_counts": schema.MapAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "All matched-resource counts keyed by backend resource type (e.g. `AWS_EC2`, `AZURE_VM`, `GCP_VM_INSTANCE`, `KUBERNETES`). Superset of the named `*_count` attributes — includes types those don't cover. Only types with at least one matched resource appear.",
			},
		},
		Blocks: map[string]schema.Block{
			"tags": schema.ListNestedBlock{
				Description: "Repeatable tag filter. Each block names a single tag key and lists exact-match values, regex-match values, or both. A resource is included if any tag block matches.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "The tag key to match (e.g. `env`, `team`).",
						},
						"exact": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Exact-match values for this tag key.",
						},
						"regex": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Regex / wildcard-pattern values for this tag key (e.g. `platform-*`).",
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *group) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		refreshCountsFromBackend(&plan, created.ID)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		resp.Diagnostics.AddError("Group created but failed to set enable state", err.Error())
		return
	}

	// Populate Computed count attributes so Terraform doesn't error on
	// "unknown values after apply". A brand-new group typically has counts
	// of 0 until discovery runs, but the values must be known regardless.
	refreshCountsFromBackend(&plan, created.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *group) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fetched, err := groups.GetGroupById(state.ID.ValueString())
	if err != nil {
		var notFound *impl.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
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
	state.Name = basetypes.NewStringValue(fetched.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *group) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	if err := groups.UpdateGroup(state.ID.ValueString(), state.Name.ValueString(), def); err != nil {
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

	plan.ID = state.ID

	// Refresh Computed counts from the backend so plan ↔ state diffs don't
	// linger as "unknown" after an Update that didn't touch the filters.
	refreshCountsFromBackend(&plan, plan.ID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *group) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

// ImportState lets `terraform import sedai_group.<name> <group-id>`
// adopt an existing Sedai group into Terraform state. The supplied ID is
// written into the `id` attribute; Terraform then calls Read to populate the
// rest of the model from the backend.
func (r *group) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildGroupDefinition converts the Terraform plan model into a SDK
// GroupDefinition. Tag values containing "*" are routed to the regex bucket;
// all others go to exact.
func buildGroupDefinition(ctx context.Context, plan groupModel) (*groups.GroupDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	def := groups.NewGroupDefinition(plan.Name.ValueString())

	if plan.AutoRefresh.ValueBool() {
		def.AutoRefresh = true
	}
	if !plan.ParentGroupId.IsNull() && !plan.ParentGroupId.IsUnknown() && plan.ParentGroupId.ValueString() != "" {
		v := plan.ParentGroupId.ValueString()
		def.ParentGroupId = &v
	}

	def.Cluster = listToStrings(plan.Clusters)
	def.Cloud = listToStrings(plan.CloudAccountIDs)
	def.ResourceType = normalizeResourceTypes(listToStrings(plan.ResourceTypes))
	def.ManuallyAddedResources = listToStrings(plan.ResourceIDs)
	def.Region = listToStrings(plan.Regions)
	def.Namespace = listToStrings(plan.Namespaces)

	for _, t := range plan.Tags {
		tag := groups.GroupTag{
			Key:   t.Key,
			Exact: t.Exact,
			Regex: t.Regex,
		}
		if tag.Exact == nil {
			tag.Exact = []string{}
		}
		if tag.Regex == nil {
			tag.Regex = []string{}
		}
		def.Tags = append(def.Tags, tag)
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

	state.Clusters = stringsToList(def.Cluster)
	state.CloudAccountIDs = stringsToList(def.Cloud)
	state.ResourceTypes = stringsToList(denormalizeResourceTypes(def.ResourceType))
	state.ResourceIDs = stringsToList(def.ManuallyAddedResources)
	state.Regions = stringsToList(def.Region)
	state.Namespaces = stringsToList(def.Namespace)

	populateCounts(state, fetched)

	if len(def.Tags) == 0 {
		state.Tags = nil
		return diags
	}
	out := make([]tagBlockModel, 0, len(def.Tags))
	for _, t := range def.Tags {
		out = append(out, tagBlockModel{
			Key:   t.Key,
			Exact: t.Exact,
			Regex: t.Regex,
		})
	}
	state.Tags = out
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


// populateCounts writes the nine named count fields AND the full dynamic
// resource_counts map onto a groupModel from a groups.GroupDetails. Called
// from Read (via applyDefinitionToModel) and from Create / Update directly so
// Computed attributes never stay unknown past an apply.
func populateCounts(state *groupModel, d *groups.GroupDetails) {
	c := d.Counts
	state.LambdaCount = basetypes.NewInt64Value(c.LambdaCount)
	state.EC2Count = basetypes.NewInt64Value(c.EC2Count)
	state.ECSCount = basetypes.NewInt64Value(c.ECSCount)
	state.KubeCount = basetypes.NewInt64Value(c.KubeCount)
	state.S3Count = basetypes.NewInt64Value(c.S3Count)
	state.EBSCount = basetypes.NewInt64Value(c.EBSCount)
	state.AzureVMCount = basetypes.NewInt64Value(c.AzureVMCount)
	state.AzureLBCount = basetypes.NewInt64Value(c.AzureLBCount)
	state.StreamingCount = basetypes.NewInt64Value(c.StreamingCount)

	// Full dynamic map. Elements are all Int64 so NewMapValue cannot error
	// here; the diags are discarded safely.
	elems := make(map[string]attr.Value, len(d.ResourceCounts))
	for k, v := range d.ResourceCounts {
		elems[k] = basetypes.NewInt64Value(v)
	}
	m, _ := basetypes.NewMapValue(basetypes.Int64Type{}, elems)
	state.ResourceCounts = m
}

// refreshCountsFromBackend fetches the group's current counts and writes
// them onto state. Used by Create / Update right before saving state so
// Computed attributes become known. On failure, zeroes out the counts —
// the group exists, we just couldn't read the snapshot; a subsequent Read
// will correct.
func refreshCountsFromBackend(state *groupModel, groupID string) {
	details, err := groups.GetGroupById(groupID)
	if err != nil || details == nil {
		populateCounts(state, &groups.GroupDetails{})
		return
	}
	populateCounts(state, details)
}

// normalizeResourceTypes maps the user-facing (correct) spelling of
// KUBERNETES_DAEMONSET to the backend's required (misspelled)
// KUBERNETES_DEAMONSET on the way out. The requirements doc lists the
// correctly-spelled form; the backend has a long-standing typo we can't
// "fix". This translation keeps HCL clean.
func normalizeResourceTypes(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		if v == "KUBERNETES_DAEMONSET" {
			out[i] = "KUBERNETES_DEAMONSET"
		} else {
			out[i] = v
		}
	}
	return out
}

// denormalizeResourceTypes reverses normalizeResourceTypes for state refresh:
// backend's KUBERNETES_DEAMONSET → user-facing KUBERNETES_DAEMONSET.
func denormalizeResourceTypes(in []string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		if v == "KUBERNETES_DEAMONSET" {
			out[i] = "KUBERNETES_DAEMONSET"
		} else {
			out[i] = v
		}
	}
	return out
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
