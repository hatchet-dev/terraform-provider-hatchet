// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ datasource.DataSource              = &OrganizationDataSource{}
	_ datasource.DataSourceWithConfigure = &OrganizationDataSource{}
)

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

type OrganizationDataSource struct {
	client *managementclient.ClientWithResponses
}

type OrganizationDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Slug types.String `tfsdk:"slug"`
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Hatchet organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the organization.",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The slug of the organization.",
				Computed:            true,
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	apiClient, err := createAPIClient(client.Endpoint, client.Token, client.ProviderVersion)
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID", err.Error())
		return
	}

	orgResp, err := d.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	if orgResp.StatusCode() < 200 || orgResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read organization, got status: %d", orgResp.StatusCode()))
		return
	}

	if orgResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Organization not found")
		return
	}

	data.Name = types.StringValue(orgResp.JSON200.Name)
	data.Slug = types.StringValue(orgResp.JSON200.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
