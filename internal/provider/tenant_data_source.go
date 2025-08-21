package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchetcloud/internal/api"
)

var (
	_ datasource.DataSource              = &TenantDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantDataSource{}
)

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

type TenantDataSource struct {
	client *managementclient.ClientWithResponses
}

type TenantDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Status         types.String `tfsdk:"status"`
	ArchivedAt     types.String `tfsdk:"archived_at"`
}

func (d *TenantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Hatchet tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tenant.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization this tenant belongs to.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the tenant (active, archived).",
				Computed:            true,
			},
			"archived_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the tenant was archived.",
				Computed:            true,
			},
		},
	}
}

func (d *TenantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HatchetCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HatchetCloudClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	apiClient, err := managementclient.NewClientWithResponses(
		fmt.Sprintf("https://%s", client.Endpoint),
		managementclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.Token))
			return nil
		}),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create API Client",
			"An unexpected error occurred when creating the API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	d.client = apiClient
}

func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TenantDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID", err.Error())
		return
	}

	tenantID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	orgResp, err := d.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	if orgResp.StatusCode() != 200 || orgResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Organization not found")
		return
	}

	var foundTenant *managementclient.OrganizationTenant
	for _, tenant := range orgResp.JSON200.Tenants {
		if tenant.Id == tenantID {
			foundTenant = &tenant
			break
		}
	}

	if foundTenant == nil {
		resp.Diagnostics.AddError("API Error", "Tenant not found in organization")
		return
	}

	data.Status = types.StringValue(string(foundTenant.Status))
	if foundTenant.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(foundTenant.ArchivedAt.String())
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
