package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceAccount{}

// DataSourceAccount is the constructor for `sedai_account` (data).
func DataSourceAccount() datasource.DataSource {
	return &dataSourceAccount{}
}

// dataSourceAccount looks up a single existing Sedai account by name.
// Returns identity + cloud-provider metadata. Use this to reference
// accounts created outside Terraform (e.g. via the Sedai UI's account-
// onboarding wizard) when building groups or settings on top.
type dataSourceAccount struct{}

type dataSourceAccountModel struct {
	// Input
	Name basetypes.StringValue `tfsdk:"name"`

	// Computed outputs
	ID              basetypes.StringValue `tfsdk:"id"`
	CloudProvider   basetypes.StringValue `tfsdk:"cloud_provider"`
	IntegrationType basetypes.StringValue `tfsdk:"integration_type"`
}

func (d *dataSourceAccount) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *dataSourceAccount) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Sedai account by name. Returns the account's ID plus cloud-provider metadata. Use this to reference accounts created outside Terraform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The exact name of the account to look up. Errors at plan time if no account with this name exists.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Sedai account ID.",
			},
			"cloud_provider": schema.StringAttribute{
				Computed:    true,
				Description: "The cloud provider of the account (`AWS`, `AZURE`, `GCP`, `KUBERNETES`).",
			},
			"integration_type": schema.StringAttribute{
				Computed:    true,
				Description: "The integration type (`AGENTLESS`, `AGENT_BASED`).",
			},
		},
	}
}

func (d *dataSourceAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceAccountModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	matches, err := account.SearchAccountsByName(name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to search accounts", err.Error())
		return
	}
	if len(matches) == 0 {
		resp.Diagnostics.AddError(
			"Account not found",
			"No Sedai account with name "+name+" exists.",
		)
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError(
			"Ambiguous account name",
			"Multiple accounts exist with name "+name+"; this data source requires a unique name.",
		)
		return
	}

	a := matches[0]
	config.ID = basetypes.NewStringValue(a.ID)
	config.CloudProvider = basetypes.NewStringValue(a.AccountDetails.CloudProvider)
	config.IntegrationType = basetypes.NewStringValue(a.AccountDetails.IntegrationType)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
