package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &HatchetCloudProvider{}
)

// HatchetCloudProvider is the provider implementation.
type HatchetCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HatchetCloudProviderModel describes the provider data model.
type HatchetCloudProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *HatchetCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hatchetcloud"
	resp.Version = p.version
}

func (p *HatchetCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Hatchet Cloud API endpoint",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Hatchet Cloud API token",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *HatchetCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data HatchetCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default endpoint if not provided
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = "cloud.onhatchet.run"
	}

	// Example client configuration for data sources and resources
	client := &HatchetCloudClient{
		Endpoint: endpoint,
		Token:    data.Token.ValueString(),
	}
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hatchet Cloud client", map[string]any{"success": true})
}

func (p *HatchetCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Add your resources here
	}
}

func (p *HatchetCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add your data sources here
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HatchetCloudProvider{
			version: version,
		}
	}
}

// HatchetCloudClient is a simple client for the Hatchet Cloud API
type HatchetCloudClient struct {
	Endpoint string
	Token    string
}
