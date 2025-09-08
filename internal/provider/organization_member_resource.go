// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openapi_types "github.com/oapi-codegen/runtime/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchet/internal/api"
)

var (
	_ resource.Resource                = &OrganizationMemberResource{}
	_ resource.ResourceWithImportState = &OrganizationMemberResource{}
)

func NewOrganizationMemberResource() resource.Resource {
	return &OrganizationMemberResource{}
}

type OrganizationMemberResource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type OrganizationMemberResourceModel struct {
	Email        types.String `tfsdk:"email"`
	Role         types.String `tfsdk:"role"`
	Status       types.String `tfsdk:"status"`
	InviteID     types.String `tfsdk:"invite_id"`
	InviterEmail types.String `tfsdk:"inviter_email"`
	Expires      types.String `tfsdk:"expires"`
}

func (r *OrganizationMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *OrganizationMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages organization members by sending invitations and tracking their status. When a member is invited, they receive an email invitation and remain in 'PENDING' status until they accept the invitation.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the user to invite to the organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role of the member in the organization. Currently only 'OWNER' is supported.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the member: 'PENDING' (invited but not yet accepted), 'ACCEPTED' (active member), or 'REJECTED' (declined invitation).",
				Computed:            true,
			},
			"invite_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the invitation (available while invitation is pending).",
				Computed:            true,
			},
			"inviter_email": schema.StringAttribute{
				MarkdownDescription: "Email address of the user who sent the invitation.",
				Computed:            true,
			},
			"expires": schema.StringAttribute{
				MarkdownDescription: "When the invitation expires (only available while invitation is pending).",
				Computed:            true,
			},
		},
	}
}

func (r *OrganizationMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	// Validate role
	roleStr := data.Role.ValueString()
	if roleStr != "OWNER" {
		resp.Diagnostics.AddError("Invalid Role", "Currently only 'OWNER' role is supported")
		return
	}

	// Parse email
	email := openapi_types.Email(data.Email.ValueString())

	// Create invitation request
	inviteReq := managementclient.CreateOrganizationInviteRequest{
		InviteeEmail: email,
		Role:         managementclient.OrganizationMemberRoleType(roleStr),
	}

	// Send invitation
	inviteResp, err := r.client.OrganizationInviteCreateWithResponse(ctx, orgID, inviteReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create organization invitation, got error: %s", err))
		return
	}

	if inviteResp.StatusCode() < 200 || inviteResp.StatusCode() >= 300 {
		if inviteResp.JSON400 != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create organization invitation: %s", inviteResp.JSON400.Description))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create organization invitation, got status: %d", inviteResp.StatusCode()))
		}
		return
	}

	// Refresh data to get the created invitation
	if err := r.refreshMemberData(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Error reading member state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.refreshMemberData(ctx, &data); err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Member/invite no longer exists, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading member state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Organization member update is not currently supported. To change a member's role or email, please delete and recreate the resource.",
	)
}

func (r *OrganizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	// Parse email for removal
	email := openapi_types.Email(data.Email.ValueString())

	// Create removal request
	removeReq := managementclient.RemoveOrganizationMembersRequest{
		Emails: []openapi_types.Email{email},
	}

	// Remove member
	removeResp, err := r.client.OrganizationMemberDeleteWithResponse(ctx, orgID, removeReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove organization member, got error: %s", err))
		return
	}

	if removeResp.StatusCode() < 200 || removeResp.StatusCode() >= 300 {
		if removeResp.JSON400 != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to remove organization member: %s", removeResp.JSON400.Description))
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to remove organization member, got status: %d", removeResp.StatusCode()))
		}
		return
	}
}

func (r *OrganizationMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by email address
	email := strings.TrimSpace(req.ID)
	if email == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "Import ID must be a valid email address")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), email)...)
}

// refreshMemberData fetches the current state of the member/invite from the API
func (r *OrganizationMemberResource) refreshMemberData(ctx context.Context, data *OrganizationMemberResourceModel) error {
	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		return fmt.Errorf("invalid organization ID from token: %w", err)
	}

	email := data.Email.ValueString()

	// First, check if this email corresponds to an existing member
	orgResp, err := r.client.OrganizationGetWithResponse(ctx, orgID)
	if err != nil {
		return fmt.Errorf("unable to read organization: %w", err)
	}

	if orgResp.StatusCode() < 200 || orgResp.StatusCode() >= 300 || orgResp.JSON200 == nil {
		return fmt.Errorf("organization not found, got status: %d", orgResp.StatusCode())
	}

	// Check existing members
	if orgResp.JSON200.Members != nil {
		for _, member := range *orgResp.JSON200.Members {
			if string(member.Email) == email {
				// Found existing member
				data.Status = types.StringValue("ACCEPTED")
				data.Role = types.StringValue(string(member.Role))
				data.InviteID = types.StringNull()
				data.InviterEmail = types.StringNull()
				data.Expires = types.StringNull()
				return nil
			}
		}
	}

	// If not found in members, check pending invitations
	invitesResp, err := r.client.OrganizationInviteListWithResponse(ctx, orgID)
	if err != nil {
		return fmt.Errorf("unable to read organization invitations: %w", err)
	}

	if invitesResp.StatusCode() < 200 || invitesResp.StatusCode() >= 300 || invitesResp.JSON200 == nil {
		return fmt.Errorf("unable to read organization invitations, got status: %d", invitesResp.StatusCode())
	}

	// Check pending invitations
	for _, invite := range invitesResp.JSON200.Rows {
		if string(invite.InviteeEmail) == email {
			// Found pending invitation
			data.Status = types.StringValue(string(invite.Status))
			data.Role = types.StringValue(string(invite.Role))
			data.InviteID = types.StringValue(invite.Metadata.Id)
			data.InviterEmail = types.StringValue(string(invite.InviterEmail))
			data.Expires = types.StringValue(invite.Expires.String())
			return nil
		}
	}

	// If we get here, the member/invite doesn't exist anymore
	// Return a special error that indicates the resource should be removed
	return fmt.Errorf("member with email %s not found in organization members or pending invites", email)
}
