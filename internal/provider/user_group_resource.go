// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	stringvalidators "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ resource.Resource                = &UserGroupResource{}
	_ resource.ResourceWithImportState = &UserGroupResource{}
)

func NewUserGroupResource() resource.Resource {
	return &UserGroupResource{}
}

type UserGroupResource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type UserGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Role        types.String `tfsdk:"role"`
	Tags        types.List   `tfsdk:"tags"`
	MemberCount types.Int64  `tfsdk:"member_count"`
}

func (r *UserGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (r *UserGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an organization user group. Members of a group are granted the group's tenant role on any tenant whose tags are a subset of the group's tags.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user group.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user group.",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The tenant role granted to group members when synced to a matching tenant. One of `OWNER`, `ADMIN`, `MEMBER`, `VIEWER`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidators.OneOf("OWNER", "ADMIN", "MEMBER", "VIEWER"),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags that determine which tenants this group's members are synced to.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"member_count": schema.Int64Attribute{
				MarkdownDescription: "The number of organization members in this group.",
				Computed:            true,
			},
		},
	}
}

func (r *UserGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	createReq := managementclient.CreateUserGroupRequest{
		Name: data.Name.ValueString(),
		Role: managementclient.TenantMemberRoleType(data.Role.ValueString()),
	}

	createResp, err := r.client.OrganizationUserGroupsCreateWithResponse(ctx, orgID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user group, got error: %s", err))
		return
	}

	if createResp.StatusCode() < 200 || createResp.StatusCode() >= 300 || createResp.JSON201 == nil {
		if createResp.JSON400 != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create user group: %s", createResp.JSON400.Description))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create user group, got status: %d", createResp.StatusCode()))
		}
		return
	}

	groupID, err := uuid.Parse(createResp.JSON201.Metadata.Id)
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID from API", err.Error())
		return
	}

	data.ID = types.StringValue(createResp.JSON201.Metadata.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tags []string
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		setTagsResp, err := r.client.OrganizationUserGroupSetTagsWithResponse(ctx, orgID, groupID, managementclient.SetTagsRequest{Tags: tags})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set user group tags, got error: %s", err))
			return
		}
		if setTagsResp.StatusCode() < 200 || setTagsResp.StatusCode() >= 300 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to set user group tags, got status: %d", setTagsResp.StatusCode()))
			return
		}
	}

	if err := r.refreshGroupData(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading user group state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.refreshGroupData(ctx, &data); err != nil {

		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user group state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	groupID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID", err.Error())
		return
	}

	updateReq := managementclient.UpdateUserGroupRequest{}
	name := plan.Name.ValueString()
	updateReq.Name = &name
	role := managementclient.TenantMemberRoleType(plan.Role.ValueString())
	updateReq.Role = &role

	updateResp, err := r.client.OrganizationUserGroupUpdateWithResponse(ctx, orgID, groupID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user group, got error: %s", err))
		return
	}

	if updateResp.StatusCode() < 200 || updateResp.StatusCode() >= 300 {
		if updateResp.JSON400 != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update user group: %s", updateResp.JSON400.Description))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update user group, got status: %d", updateResp.StatusCode()))
		}
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

	setTagsResp, err := r.client.OrganizationUserGroupSetTagsWithResponse(ctx, orgID, groupID, managementclient.SetTagsRequest{Tags: tags})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set user group tags, got error: %s", err))
		return
	}
	if setTagsResp.StatusCode() < 200 || setTagsResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to set user group tags, got status: %d", setTagsResp.StatusCode()))
		return
	}

	plan.ID = state.ID
	if err := r.refreshGroupData(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error reading user group state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	groupID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID", err.Error())
		return
	}

	deleteResp, err := r.client.OrganizationUserGroupDeleteWithResponse(ctx, orgID, groupID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user group, got error: %s", err))
		return
	}

	if deleteResp.StatusCode() < 200 || deleteResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete user group, got status: %d", deleteResp.StatusCode()))
		return
	}
}

func (r *UserGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// refreshGroupData fetches the current state of the user group from the API.
func (r *UserGroupResource) refreshGroupData(ctx context.Context, data *UserGroupResourceModel) error {
	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID from token: %w", err)
	}

	groupID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		return fmt.Errorf("invalid user group ID: %w", err)
	}

	getResp, err := r.client.OrganizationUserGroupGetWithResponse(ctx, orgID, groupID)
	if err != nil {
		return fmt.Errorf("unable to read user group: %w", err)
	}

	if getResp.StatusCode() == 404 {
		return fmt.Errorf("user group not found")
	}

	if getResp.StatusCode() < 200 || getResp.StatusCode() >= 300 || getResp.JSON200 == nil {
		return fmt.Errorf("unable to read user group, got status: %d", getResp.StatusCode())
	}

	data.Name = types.StringValue(getResp.JSON200.Name)
	data.Role = types.StringValue(string(getResp.JSON200.Role))
	data.MemberCount = types.Int64Value(int64(getResp.JSON200.MemberCount))

	tags := getResp.JSON200.Tags
	if tags == nil {
		tags = []string{}
	}
	tagList, diags := types.ListValueFrom(ctx, types.StringType, tags)
	if diags.HasError() {
		return fmt.Errorf("unable to convert tags to list")
	}
	data.Tags = tagList

	return nil
}
