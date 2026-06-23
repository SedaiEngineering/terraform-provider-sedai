package provider

import (
	"context"

	sdkresource "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceResourceSettings{}

func DataSourceResourceSettings() datasource.DataSource {
	return &dataSourceResourceSettings{}
}

type dataSourceResourceSettings struct{}

type dataSourceResourceSettingsModel struct {
	ResourceID       basetypes.StringValue `tfsdk:"resource_id"`
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`
	ResourceType     basetypes.StringValue `tfsdk:"resource_type"`
}

func (d *dataSourceResourceSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_settings"
}

func (d *dataSourceResourceSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read the current settings for a specific Sedai-managed resource. Returns the availability/optimization modes and the resource type as reported by the backend.",
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The Sedai resource ID to read settings for.",
			},
			"availability_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The resource's current availability mode (`DATA_PILOT`, `CO_PILOT`, or `AUTO`).",
			},
			"optimization_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The resource's current optimization mode (`DATA_PILOT`, `CO_PILOT`, or `AUTO`).",
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Sedai auto-syncs this resource with the latest configuration.",
			},
			"resource_type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type as reported by the backend (e.g. `AWS_LAMBDA`, `AWS_EC2`).",
			},
		},
	}
}

func (d *dataSourceResourceSettings) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceResourceSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := sdkresource.GetResourceSettings(config.ResourceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read resource settings", err.Error())
		return
	}
	if settings == nil {
		resp.Diagnostics.AddError(
			"Resource settings not found",
			"No settings exist for resource "+config.ResourceID.ValueString()+".",
		)
		return
	}

	config.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
	config.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
	config.SedaiSyncEnabled = basetypes.NewBoolValue(settings.SedaiSyncEnabled)
	config.ResourceType = basetypes.NewStringValue(settings.ResourceType)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
