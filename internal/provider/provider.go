package provider

import (
	"context"
	"os"

	impl "github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/impl"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &sedaiProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sedaiProvider{
			version: version,
		}
	}
}

// sedaiProvider is the provider implementation.
type sedaiProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *sedaiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sedai"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *sedaiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Sedai provider manages Sedai accounts and monitoring providers. Authentication is configured via environment variables.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Sedai API base URL. Can also be set via the `SEDAI_BASE_URL` environment variable.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Sedai API token. Can also be set via the `SEDAI_API_TOKEN` environment variable.",
			},
		},
	}
}

// Configure reads credentials from the provider HCL block, falls back to
// environment variables, and pushes the resolved values into the SDK so
// every resource and data source uses the correct endpoint and token.
func (p *sedaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg struct {
		BaseURL  types.String `tfsdk:"base_url"`
		APIToken types.String `tfsdk:"api_token"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL  := cfg.BaseURL.ValueString()
	apiToken := cfg.APIToken.ValueString()

	// HCL takes precedence; fall back to env vars when not set in the block.
	if baseURL == "" {
		baseURL = os.Getenv("SEDAI_BASE_URL")
	}
	if apiToken == "" {
		apiToken = os.Getenv("SEDAI_API_TOKEN")
		if apiToken == "" {
			apiToken = os.Getenv("SEDAI_API_KEY")
		}
	}

	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Sedai base URL",
			"Set base_url in the provider block or export SEDAI_BASE_URL in the environment.",
		)
	}
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Sedai API token",
			"Set api_token in the provider block or export SEDAI_API_TOKEN (or SEDAI_API_KEY) in the environment.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	impl.SetConfig(baseURL, apiToken)
}

// DataSources defines the data sources implemented in the provider.
func (p *sedaiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		DataSourceGroup,
		DataSourceGroups,
		DataSourceAccount,
		DataSourceGroupSettings,
		DataSourceAccountSettings,
		DataSourceResourceSettings,
		DataSourceGroupPriority,
	}
}

// Resources defines the resources implemented in the provider.
func (p *sedaiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		CreateAccount,
		AccountSettings,
		Group,
		GroupSettings,
		GroupPriority,
		ResourceSettings,
		CreateGKEMonitoringProvider,
		CreateFPMonitoringProvider,
		CreateDatadogMonitoringProvider,
		CreateNewrelicMonitoringProvider,
		CreateVMMonitoringProvider,
		CreateCloudWatchMonitoringProvider,
		CreateAzureMonitoringProvider,
	}
}
