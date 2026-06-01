package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/groups"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceGroups{}

// DataSourceGroups is the constructor for `sedai_groups` (data, plural —
// list all).
func DataSourceGroups() datasource.DataSource {
	return &dataSourceGroups{}
}

// dataSourceGroups is the data source implementation. Lists every group
// in the Sedai tenant, returning each as {id, name}. Used for discovery
// and bulk for_each operations.
type dataSourceGroups struct{}

type dataSourceGroupsModel struct {
	Groups []dataSourceGroupsEntry `tfsdk:"groups"`
}

type dataSourceGroupsEntry struct {
	ID   basetypes.StringValue `tfsdk:"id"`
	Name basetypes.StringValue `tfsdk:"name"`
}

func (d *dataSourceGroups) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *dataSourceGroups) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every Sedai group in the tenant. Useful for bulk operations (`for_each`) or for discovering groups created outside Terraform before importing them.",
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All groups visible to the caller. Each entry has `id` and `name`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The Sedai group ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The group name.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceGroups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	all, err := groups.GetAllGroups()
	if err != nil {
		resp.Diagnostics.AddError("Unable to list groups", err.Error())
		return
	}

	out := dataSourceGroupsModel{
		Groups: make([]dataSourceGroupsEntry, 0, len(all)),
	}
	for _, g := range all {
		out.Groups = append(out.Groups, dataSourceGroupsEntry{
			ID:   basetypes.NewStringValue(g.ID),
			Name: basetypes.NewStringValue(g.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}

// _ keeps imports honest — types is referenced only via the schema. Compile
// check: if we ever stop using ListNestedAttribute we'd want to drop this.
var _ = types.StringType
