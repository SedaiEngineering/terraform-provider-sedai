package provider

import (
	"context"
	"strings"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceAccount{}

// DataSourceAccount is the constructor for `sedai_account` (data).
func DataSourceAccount() datasource.DataSource {
	return &dataSourceAccount{}
}

// dataSourceAccount looks up a single existing Sedai account by name or ID.
type dataSourceAccount struct{}

type dataSourceAccountModel struct {
	// Inputs — exactly one must be set
	Name      basetypes.StringValue `tfsdk:"name"`
	AccountID basetypes.StringValue `tfsdk:"account_id"`

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
		Description: "Look up an existing Sedai account by name or by ID. " +
			"Exactly one of `name` or `account_id` must be set. " +
			"Use `account_id` when multiple accounts share the same name — it is always unambiguous. " +
			"Use this data source to reference accounts created outside Terraform (e.g. via the Sedai UI).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The exact name of the account to look up. Errors if multiple accounts share this name — use account_id instead.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("account_id"),
					}...),
				},
			},
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Sedai account ID to look up directly. Always unambiguous — use this when multiple accounts share the same name.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("name"),
					}...),
				},
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

	// Lookup by ID — direct, always unambiguous
	if !config.AccountID.IsNull() && !config.AccountID.IsUnknown() {
		a, err := account.SearchAccountsById(config.AccountID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Unable to fetch account", err.Error())
			return
		}
		if a == nil {
			resp.Diagnostics.AddError(
				"Account not found",
				"No Sedai account with ID '"+config.AccountID.ValueString()+"' exists.",
			)
			return
		}
		config.ID = basetypes.NewStringValue(a.ID)
		config.CloudProvider = basetypes.NewStringValue(a.AccountDetails.CloudProvider)
		config.IntegrationType = basetypes.NewStringValue(a.AccountDetails.IntegrationType)
		resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
		return
	}

	// Lookup by name — may fail if multiple accounts share the same name
	name := config.Name.ValueString()
	matches, err := account.SearchAccountsByName(name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to search accounts", err.Error())
		return
	}
	if len(matches) == 0 {
		resp.Diagnostics.AddError(
			"Account not found",
			"No Sedai account with name '"+name+"' exists.",
		)
		return
	}
	if len(matches) > 1 {
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		resp.Diagnostics.AddError(
			"Ambiguous account name",
			"Multiple accounts found with name '"+name+"'. IDs: ["+strings.Join(ids, ", ")+"] — "+
				"use account_id instead to look up unambiguously, or delete the duplicate "+
				"via the Sedai UI (DELETE /api/site/accounts/{id}).",
		)
		return
	}

	a := matches[0]
	config.ID = basetypes.NewStringValue(a.ID)
	config.CloudProvider = basetypes.NewStringValue(a.AccountDetails.CloudProvider)
	config.IntegrationType = basetypes.NewStringValue(a.AccountDetails.IntegrationType)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
