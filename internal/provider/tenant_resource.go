// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ resource.Resource                = &TenantResource{}
	_ resource.ResourceWithImportState = &TenantResource{}
)

func NewTenantResource() resource.Resource {
	return &TenantResource{}
}

type TenantResource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type TenantResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Slug       types.String `tfsdk:"slug"`
	Status     types.String `tfsdk:"status"`
	ArchivedAt types.String `tfsdk:"archived_at"`
}

func (r *TenantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hatchet tenant within an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tenant.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tenant.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The slug of the tenant.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *TenantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HatchetCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
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

	r.client = apiClient
	r.organizationID = client.OrganizationID
}

func (r *TenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TenantResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	createReq := managementclient.CreateNewTenantForOrganizationRequest{
		Name: data.Name.ValueString(),
		Slug: data.Slug.ValueString(),
	}

	tenantResp, err := r.client.OrganizationCreateTenantWithResponse(ctx, orgID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tenant, got error: %s", err))
		return
	}

	if tenantResp.StatusCode() < 200 || tenantResp.StatusCode() >= 300 {
		if tenantResp.JSON400 != nil && tenantResp.JSON400.Description == "tenant slug already in use" {
			resp.Diagnostics.AddError("Tenant slug already in use", "The tenant slug is already in use. Please choose a different slug.")
			return
		}

		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create tenant, got status: %d", tenantResp.StatusCode()))
		return
	}

	if tenantResp.JSON201 == nil {
		resp.Diagnostics.AddError("API Error", "Tenant creation failed")
		return
	}

	data.ID = types.StringValue(tenantResp.JSON201.Id.String())
	data.Status = types.StringValue(string(tenantResp.JSON201.Status))
	if tenantResp.JSON201.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(tenantResp.JSON201.ArchivedAt.String())
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TenantResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	orgResp, err := r.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	if orgResp.StatusCode() < 200 || orgResp.StatusCode() >= 300 || orgResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Organization not found")
		return
	}

	tenantID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	var foundTenant *managementclient.OrganizationTenant
	if orgResp.JSON200.Tenants != nil {
		for _, tenant := range *orgResp.JSON200.Tenants {
			if tenant.Id == tenantID {
				foundTenant = &tenant
				break
			}
		}
	}

	if foundTenant == nil {
		resp.State.RemoveResource(ctx)
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

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Tenant update is not currently supported through the Terraform provider.",
	)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TenantResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	deleteResp, err := r.client.OrganizationTenantDeleteWithResponse(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete tenant, got error: %s", err))
		return
	}

	if deleteResp.StatusCode() < 200 || deleteResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete tenant, got status: %d", deleteResp.StatusCode()))
		return
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
