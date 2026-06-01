package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure interfaces are satisfied.
var _ datasource.DataSource = &dataSourceGroup{}

// DataSourceGroup is the data source constructor for `sedai_group` (data).
func DataSourceGroup() datasource.DataSource {
	return &dataSourceGroup{}
}

// dataSourceGroup is the data source implementation. Data source type name:
// `sedai_group`. Looks up a single existing Sedai group by name.
//
// Use this when the group was created outside Terraform (e.g. via the
// Sedai UI) and you want to reference its ID from other Terraform-managed
// resources without adopting it as a managed `sedai_group` resource.
type dataSourceGroup struct{}

type dataSourceGroupModel struct {
	// Input
	Name basetypes.StringValue `tfsdk:"name"`

	// Computed outputs
	ID             basetypes.StringValue `tfsdk:"id"`
	Enabled        basetypes.BoolValue   `tfsdk:"enabled"`
	AutoRefresh    basetypes.BoolValue   `tfsdk:"auto_refresh"`
	ParentGroupID  basetypes.StringValue `tfsdk:"parent_group_id"`
	Clusters       basetypes.ListValue   `tfsdk:"clusters"`
	CloudAccountIDs basetypes.ListValue  `tfsdk:"cloud_account_ids"`
	ResourceTypes  basetypes.ListValue   `tfsdk:"resource_types"`
	ResourceIDs    basetypes.ListValue   `tfsdk:"resource_ids"`
	Regions        basetypes.ListValue   `tfsdk:"regions"`
	Namespaces     basetypes.ListValue   `tfsdk:"namespaces"`

	// Resource counts (one per resource type).
	LambdaCount    basetypes.Int64Value `tfsdk:"lambda_count"`
	EC2Count       basetypes.Int64Value `tfsdk:"ec2_count"`
	ECSCount       basetypes.Int64Value `tfsdk:"ecs_count"`
	KubeCount      basetypes.Int64Value `tfsdk:"kube_count"`
	S3Count        basetypes.Int64Value `tfsdk:"s3_count"`
	EBSCount       basetypes.Int64Value `tfsdk:"ebs_count"`
	AzureVMCount   basetypes.Int64Value `tfsdk:"azure_vm_count"`
	AzureLBCount   basetypes.Int64Value `tfsdk:"azure_lb_count"`
	StreamingCount basetypes.Int64Value `tfsdk:"streaming_count"`
	ResourceCounts basetypes.MapValue   `tfsdk:"resource_counts"`
}

func (d *dataSourceGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *dataSourceGroup) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Sedai group by name. Returns the group's ID, definition, and matched-resource counts. Use this to reference groups created outside Terraform (e.g. via the Sedai UI) without adopting them as Terraform-managed resources.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The exact name of the group to look up. Errors at plan time if no group with this name exists.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Sedai group ID.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the group is currently active.",
			},
			"auto_refresh": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Sedai periodically re-evaluates the group's filters.",
			},
			"parent_group_id": schema.StringAttribute{
				Computed:    true,
				Description: "Parent group ID, if this group is a subgroup. Empty / null otherwise.",
			},
			"clusters": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Cluster ID filters.",
			},
			"cloud_account_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Sedai account IDs the group is scoped to.",
			},
			"resource_types": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Resource type filters.",
			},
			"resource_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Manually added resource IDs.",
			},
			"regions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Cloud region filters.",
			},
			"namespaces": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Kubernetes namespace filters.",
			},
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
				Description: "All matched-resource counts keyed by backend resource type (e.g. `AWS_EC2`, `AZURE_VM`, `GCP_VM_INSTANCE`, `KUBERNETES`). Superset of the named `*_count` attributes.",
			},
		},
	}
}

func (d *dataSourceGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()

	// 1) Find the group by name to get its ID.
	matches, err := groups.SearchGroupsByName(name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to search groups", err.Error())
		return
	}
	if len(matches) == 0 {
		resp.Diagnostics.AddError(
			"Group not found",
			"No Sedai group with name "+name+" exists.",
		)
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError(
			"Ambiguous group name",
			"Multiple groups exist with name "+name+"; this data source requires a unique name.",
		)
		return
	}

	// 2) Fetch full details + counts.
	details, err := groups.GetGroupById(matches[0].ID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read group details", err.Error())
		return
	}
	if details == nil || details.Definition == nil {
		resp.Diagnostics.AddError(
			"Group disappeared mid-lookup",
			"Found the group by name but could not fetch its details. Try again.",
		)
		return
	}

	def := details.Definition
	config.ID = basetypes.NewStringValue(details.GroupId)
	if details.IsActive != nil {
		config.Enabled = basetypes.NewBoolValue(*details.IsActive)
	} else {
		config.Enabled = basetypes.NewBoolNull()
	}
	config.AutoRefresh = basetypes.NewBoolValue(def.AutoRefresh)
	if def.ParentGroupId != nil && *def.ParentGroupId != "" {
		config.ParentGroupID = basetypes.NewStringValue(*def.ParentGroupId)
	} else {
		config.ParentGroupID = basetypes.NewStringNull()
	}
	config.Clusters = stringsToList(def.Cluster)
	config.CloudAccountIDs = stringsToList(def.Cloud)
	config.ResourceTypes = stringsToList(denormalizeResourceTypes(def.ResourceType))
	config.ResourceIDs = stringsToList(def.ManuallyAddedResources)
	config.Regions = stringsToList(def.Region)
	config.Namespaces = stringsToList(def.Namespace)

	c := details.Counts
	config.LambdaCount = basetypes.NewInt64Value(c.LambdaCount)
	config.EC2Count = basetypes.NewInt64Value(c.EC2Count)
	config.ECSCount = basetypes.NewInt64Value(c.ECSCount)
	config.KubeCount = basetypes.NewInt64Value(c.KubeCount)
	config.S3Count = basetypes.NewInt64Value(c.S3Count)
	config.EBSCount = basetypes.NewInt64Value(c.EBSCount)
	config.AzureVMCount = basetypes.NewInt64Value(c.AzureVMCount)
	config.AzureLBCount = basetypes.NewInt64Value(c.AzureLBCount)
	config.StreamingCount = basetypes.NewInt64Value(c.StreamingCount)

	// Full dynamic map. Int64 elements → NewMapValue cannot error here.
	elems := make(map[string]attr.Value, len(details.ResourceCounts))
	for k, v := range details.ResourceCounts {
		elems[k] = basetypes.NewInt64Value(v)
	}
	m, _ := basetypes.NewMapValue(basetypes.Int64Type{}, elems)
	config.ResourceCounts = m

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
