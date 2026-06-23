package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

// Configure prepares a HashiCups API client for data sources and resources.
func (p *sedaiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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
