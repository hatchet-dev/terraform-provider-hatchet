// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "-")
	re = regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	randomSuffix := make([]byte, 5)
	for i := range randomSuffix {
		randomSuffix[i] = "abcdefghijklmnopqrstuvwxyz0123456789"[rand.Intn(36)]
	}

	return slug + "-" + string(randomSuffix)
}

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
	Region     types.String `tfsdk:"region"`
	Tags       types.List   `tfsdk:"tags"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "The slug of the tenant. If not provided, a slug will be generated from the name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Shard selector as `provider:cloud-region` or `provider:cloud-region:shard-name`, e.g. `aws:us-west-2`. The shard name is optional. When omitted, the server selects an eligible shard automatically.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags applied to this tenant. Management tokens can only create or access tenants whose tags are a subset of the token's own tags.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
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

	slug := data.Slug.ValueString()
	if slug == "" {
		slug = generateSlug(data.Name.ValueString())
		data.Slug = types.StringValue(slug)
	}

	createReq := managementclient.CreateNewTenantForOrganizationRequest{
		Name: data.Name.ValueString(),
		Slug: slug,
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() && data.Region.ValueString() != "" {
		region := data.Region.ValueString()
		createReq.Region = &region
	}
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Tags = &tags
	}

	tenantResp, err := r.client.OrganizationCreateTenantWithResponse(ctx, orgID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tenant, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "tenant create response", map[string]any{
		"status": tenantResp.StatusCode(),
		"body":   string(tenantResp.Body),
	})

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
	if tenantResp.JSON201.Region != nil {
		data.Region = types.StringValue(*tenantResp.JSON201.Region)
	} else {
		data.Region = types.StringValue("")
	}
	if tenantResp.JSON201.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(tenantResp.JSON201.ArchivedAt.String())
	} else {
		data.ArchivedAt = types.StringNull()
	}
	if tenantResp.JSON201.Tags != nil {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, *tenantResp.JSON201.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagList
	} else if data.Tags.IsNull() || data.Tags.IsUnknown() {
		// The create response doesn't echo tags back; only default to empty
		// when the plan didn't already give us a concrete value (e.g. tags
		// weren't set in config at all), so we don't clobber a known,
		// just-requested value with an empty list.
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, []string{})
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

	tflog.Debug(ctx, "organization get response", map[string]any{
		"status": orgResp.StatusCode(),
		"body":   string(orgResp.Body),
	})

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
		tflog.Debug(ctx, "tenant not found in organization tenant list", map[string]any{
			"tenant_id": tenantID.String(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	data.Status = types.StringValue(string(foundTenant.Status))
	if foundTenant.Name != nil {
		data.Name = types.StringValue(*foundTenant.Name)
	}
	if foundTenant.Slug != nil {
		data.Slug = types.StringValue(*foundTenant.Slug)
	}
	if foundTenant.Region != nil {
		data.Region = types.StringValue(*foundTenant.Region)
	}
	if foundTenant.Tags != nil {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, *foundTenant.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagList
	} else {
		data.Tags, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}
	if foundTenant.ArchivedAt != nil {
		data.ArchivedAt = types.StringValue(foundTenant.ArchivedAt.String())
	} else {
		data.ArchivedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	tenantID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if tags == nil {
		tags = []string{}
	}

	setTagsResp, err := r.client.OrganizationTenantSetTagsWithResponse(ctx, orgID, tenantID, managementclient.SetTagsRequest{Tags: tags})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update tenant tags, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "tenant set tags response", map[string]any{
		"status": setTagsResp.StatusCode(),
		"body":   string(setTagsResp.Body),
	})

	if setTagsResp.StatusCode() < 200 || setTagsResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update tenant tags, got status: %d", setTagsResp.StatusCode()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	tflog.Debug(ctx, "tenant delete response", map[string]any{
		"status": deleteResp.StatusCode(),
		"body":   string(deleteResp.Body),
	})

	if deleteResp.StatusCode() < 200 || deleteResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete tenant, got status: %d", deleteResp.StatusCode()))
		return
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
