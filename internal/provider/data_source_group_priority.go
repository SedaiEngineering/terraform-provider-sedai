package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceGroupPriority{}

func DataSourceGroupPriority() datasource.DataSource {
	return &dataSourceGroupPriority{}
}

type dataSourceGroupPriority struct{}

type dataSourceGroupPriorityModel struct {
	GroupID  basetypes.StringValue `tfsdk:"group_id"`
	Priority basetypes.Int64Value  `tfsdk:"priority"`
}

func (d *dataSourceGroupPriority) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_priority"
}

func (d *dataSourceGroupPriority) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read the current priority of a Sedai group. Priority is 1-based — lower numbers are evaluated first when a resource matches multiple groups.",
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the group to read priority for.",
			},
			"priority": schema.Int64Attribute{
				Computed:    true,
				Description: "The group's current priority (1-based). Lower = higher precedence.",
			},
		},
	}
}

func (d *dataSourceGroupPriority) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceGroupPriorityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	details, err := groups.GetGroupById(config.GroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read group", err.Error())
		return
	}
	if details == nil {
		resp.Diagnostics.AddError(
			"Group not found",
			"No group with ID "+config.GroupID.ValueString()+" exists.",
		)
		return
	}

	if details.Priority != nil {
		config.Priority = basetypes.NewInt64Value(int64(*details.Priority))
	} else {
		config.Priority = basetypes.NewInt64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
