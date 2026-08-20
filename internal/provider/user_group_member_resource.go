// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ resource.Resource                = &UserGroupMemberResource{}
	_ resource.ResourceWithImportState = &UserGroupMemberResource{}
)

func NewUserGroupMemberResource() resource.Resource {
	return &UserGroupMemberResource{}
}

type UserGroupMemberResource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type UserGroupMemberResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	UserGroupID          types.String `tfsdk:"user_group_id"`
	Email                types.String `tfsdk:"email"`
	OrganizationMemberID types.String `tfsdk:"organization_member_id"`
}

func (r *UserGroupMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group_member"
}

func (r *UserGroupMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Adds an organization member to a user group, by email. The member must have already accepted their organization invite. Adding or removing a member triggers a tenant membership sync for that member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The composite ID of the group membership, formatted as `<user_group_id>/<email>`.",
				Computed:            true,
			},
			"user_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the organization member to add to the group. Must belong to a member who has already accepted their invite.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_member_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization member, resolved from `email`.",
				Computed:            true,
			},
		},
	}
}

func (r *UserGroupMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserGroupMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	groupID, err := uuid.Parse(data.UserGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID", err.Error())
		return
	}

	email := data.Email.ValueString()
	memberID, err := r.resolveMemberIDByEmail(ctx, orgID, email)
	if err != nil {
		resp.Diagnostics.AddError("Organization Member Not Found", err.Error())
		return
	}

	addResp, err := r.client.OrganizationUserGroupAddMemberWithResponse(ctx, orgID, groupID, managementclient.AddUserGroupMemberRequest{
		OrganizationMemberId: memberID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add user group member, got error: %s", err))
		return
	}

	if addResp.StatusCode() < 200 || addResp.StatusCode() >= 300 {
		if addResp.JSON400 != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add user group member: %s", addResp.JSON400.Description))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add user group member, got status: %d", addResp.StatusCode()))
		}
		return
	}

	data.OrganizationMemberID = types.StringValue(memberID.String())
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", groupID, email))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserGroupMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	groupID, err := uuid.Parse(data.UserGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID", err.Error())
		return
	}

	listResp, err := r.client.OrganizationUserGroupListMembersWithResponse(ctx, orgID, groupID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list user group members, got error: %s", err))
		return
	}

	if listResp.StatusCode() == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	if listResp.StatusCode() < 200 || listResp.StatusCode() >= 300 || listResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to list user group members, got status: %d", listResp.StatusCode()))
		return
	}

	found := false
	for _, member := range listResp.JSON200.Rows {
		if string(member.Email) == data.Email.ValueString() {
			data.OrganizationMemberID = types.StringValue(member.Metadata.Id)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserGroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"User group membership cannot be updated in place. To change the group or member, delete and recreate the resource.",
	)
}

func (r *UserGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserGroupMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	groupID, err := uuid.Parse(data.UserGroupID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User Group ID", err.Error())
		return
	}

	memberID, err := uuid.Parse(data.OrganizationMemberID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization Member ID", err.Error())
		return
	}

	removeResp, err := r.client.OrganizationUserGroupRemoveMemberWithResponse(ctx, orgID, groupID, memberID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove user group member, got error: %s", err))
		return
	}

	if removeResp.StatusCode() < 200 || removeResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to remove user group member, got status: %d", removeResp.StatusCode()))
		return
	}
}

func (r *UserGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be formatted as <user_group_id>/<email>",
		)
		return
	}

	var data UserGroupMemberResourceModel
	data.ID = types.StringValue(req.ID)
	data.UserGroupID = types.StringValue(parts[0])
	data.Email = types.StringValue(parts[1])

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// resolveMemberIDByEmail looks up the organization member ID for an accepted
// member by email. Group membership is keyed by organization member ID
// server-side, so this is required even though the resource is configured by
// email.
func (r *UserGroupMemberResource) resolveMemberIDByEmail(ctx context.Context, orgID uuid.UUID, email string) (uuid.UUID, error) {
	orgResp, err := r.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to read organization: %w", err)
	}

	if orgResp.StatusCode() < 200 || orgResp.StatusCode() >= 300 || orgResp.JSON200 == nil {
		return uuid.Nil, fmt.Errorf("unable to read organization, got status: %d", orgResp.StatusCode())
	}

	if orgResp.JSON200.Members != nil {
		for _, member := range *orgResp.JSON200.Members {
			if string(member.Email) == email {
				return uuid.Parse(member.Metadata.Id)
			}
		}
	}

	return uuid.Nil, fmt.Errorf("no accepted organization member found with email %q (members with a pending invite cannot be added to a group)", email)
}
