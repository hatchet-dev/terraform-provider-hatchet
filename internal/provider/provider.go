package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &HatchetCloudProvider{}
)

type HatchetCloudProvider struct {
	version string
}

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
				MarkdownDescription: "Endpoint for the Hatchet Cloud instance",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Management token for the Hatchet Cloud instance. Can also be provided via HATCHET_TOKEN environment variable.",
				Optional:            true,
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

	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = "cloud.onhatchet.run"
	}

	token := data.Token.ValueString()
	if token == "" {
		token = os.Getenv("HATCHET_TOKEN")
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Token Configuration",
			"The provider cannot create the Hatchet Cloud API client as there is a missing or empty value for the token. "+
				"Set the token value in the configuration or use the HATCHET_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	client := &HatchetCloudClient{
		Endpoint: endpoint,
		Token:    token,
	}
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hatchet Cloud client", map[string]any{"success": true})
}

func (p *HatchetCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource,
		NewTenantResource,
		NewTenantAPITokenResource,
		NewOrganizationMemberResource,
	}
}

func (p *HatchetCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
		NewTenantDataSource,
		NewOrganizationMembersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HatchetCloudProvider{
			version: version,
		}
	}
}

type HatchetCloudClient struct {
	Endpoint string
	Token    string
}
