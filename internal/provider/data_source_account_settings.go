package provider

import (
	"context"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/account"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &dataSourceAccountSettings{}

func DataSourceAccountSettings() datasource.DataSource {
	return &dataSourceAccountSettings{}
}

type dataSourceAccountSettings struct{}

type dataSourceAccountSettingsModel struct {
	AccountID        basetypes.StringValue `tfsdk:"account_id"`
	AvailabilityMode basetypes.StringValue `tfsdk:"availability_mode"`
	OptimizationMode basetypes.StringValue `tfsdk:"optimization_mode"`
	SedaiSyncEnabled basetypes.BoolValue   `tfsdk:"sedai_sync_enabled"`

	KubeAppSettings      *kubeAppSettingsModel      `tfsdk:"kube_app_settings"`
	BucketSettings       *bucketSettingsModel       `tfsdk:"bucket_settings"`
	AppSettings          *appSettingsModel          `tfsdk:"app_settings"`
	ContainerAppSettings *containerAppSettingsModel `tfsdk:"container_app_settings"`
	ECSAppSettings       *ecsAppSettingsModel       `tfsdk:"ecs_app_settings"`
	ServerlessSettings   *serverlessSettingsModel   `tfsdk:"serverless_settings"`
	VolumeSettings       *volumeSettingsModel       `tfsdk:"volume_settings"`
}

func (d *dataSourceAccountSettings) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_settings"
}

func (d *dataSourceAccountSettings) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read the current settings for a Sedai account. Returns the availability/optimization modes and all per-resource-type block values as they exist on the backend.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the account to read settings for.",
			},
			"availability_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The account's current availability mode (`DATA_PILOT`, `CO_PILOT`, or `AUTO`).",
			},
			"optimization_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The account's current optimization mode (`DATA_PILOT`, `CO_PILOT`, or `AUTO`).",
			},
			"sedai_sync_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Sedai auto-syncs the account's resources with the latest configuration.",
			},
		},
		Blocks: map[string]schema.Block{
			"kube_app_settings":      dataSourceKubeAppSettingsBlock(),
			"bucket_settings":        dataSourceBucketSettingsBlock(),
			"app_settings":           dataSourceAppSettingsBlock(),
			"container_app_settings": dataSourceContainerAppSettingsBlock(),
			"ecs_app_settings":       dataSourceECSAppSettingsBlock(),
			"serverless_settings":    dataSourceServerlessSettingsBlock(),
			"volume_settings":        dataSourceVolumeSettingsBlock(),
		},
	}
}

func (d *dataSourceAccountSettings) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceAccountSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := account.GetAccountSettings(config.AccountID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read account settings", err.Error())
		return
	}
	if settings == nil {
		resp.Diagnostics.AddError(
			"Account settings not found",
			"No settings exist for account "+config.AccountID.ValueString()+".",
		)
		return
	}

	config.AvailabilityMode = basetypes.NewStringValue(settings.AvailabilityMode)
	config.OptimizationMode = basetypes.NewStringValue(settings.OptimizationMode)
	config.SedaiSyncEnabled = basetypes.NewBoolValue(settings.SedaiSyncEnabled)
	config.KubeAppSettings = kubeAppSettingsFromSDK(settings.KubeAppSettings)
	config.BucketSettings = bucketSettingsFromSDK(settings.BucketSettings)
	config.AppSettings = appSettingsFromSDK(settings.AppSettings)
	config.ContainerAppSettings = containerAppSettingsFromSDK(settings.ContainerAppSettings)
	config.ECSAppSettings = ecsAppSettingsFromSDK(settings.ECSAppSettings)
	config.ServerlessSettings = serverlessSettingsFromSDK(settings.ServerlessSettings)
	config.VolumeSettings = volumeSettingsFromSDK(settings.VolumeSettings)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
