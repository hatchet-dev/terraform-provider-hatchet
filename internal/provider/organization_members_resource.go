package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openapi_types "github.com/oapi-codegen/runtime/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchetcloud/internal/api"
)

var (
	_ resource.Resource                = &OrganizationMembersResource{}
	_ resource.ResourceWithImportState = &OrganizationMembersResource{}
)

func NewOrganizationMembersResource() resource.Resource {
	return &OrganizationMembersResource{}
}

type OrganizationMembersResource struct {
	client         *managementclient.ClientWithResponses
	organizationID string
}

type OrganizationMembersResourceModel struct {
	UserIds []types.String `tfsdk:"user_ids"`
}

func (r *OrganizationMembersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_members"
}

func (r *OrganizationMembersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hatchet organization's members.",
		Attributes: map[string]schema.Attribute{
			"user_ids": schema.ListAttribute{
				MarkdownDescription: "The IDs of the users to add as members.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *OrganizationMembersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationMembersResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	userIds := data.UserIds
	if len(userIds) == 0 {
		resp.Diagnostics.AddError("Invalid User IDs", "User IDs are required")
		return
	}

	var userIdsSlice []openapi_types.UUID
	for _, userId := range userIds {
		userIdsSlice = append(userIdsSlice, openapi_types.UUID(uuid.MustParse(userId.ValueString())))
	}

	addReq := managementclient.AddOrganizationMembersRequest{
		UserIds: userIdsSlice,
	}

	memberResp, err := r.client.OrganizationUpdateMembersWithResponse(ctx, orgID, addReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add organization members, got error: %s", err))
		return
	}

	if memberResp.StatusCode() < 200 || memberResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add organization members, got status: %d", memberResp.StatusCode()))
		return
	}

	if memberResp.JSON201 == nil {
		resp.Diagnostics.AddError("API Error", "Member addition failed")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationMembersResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	membersResp, err := r.client.OrganizationListMembersWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization members, got error: %s", err))
		return
	}

	if membersResp.StatusCode() < 200 || membersResp.StatusCode() >= 300 || membersResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Unable to read organization members")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData OrganizationMembersResourceModel
	var stateData OrganizationMembersResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	// Convert current state and planned user IDs to sets for comparison
	stateUserIds := make(map[string]bool)
	for _, userId := range stateData.UserIds {
		stateUserIds[userId.ValueString()] = true
	}

	planUserIds := make(map[string]bool)
	for _, userId := range planData.UserIds {
		planUserIds[userId.ValueString()] = true
	}

	// Find users to add (in plan but not in state)
	var usersToAdd []openapi_types.UUID
	for _, userId := range planData.UserIds {
		userIdStr := userId.ValueString()
		if !stateUserIds[userIdStr] {
			usersToAdd = append(usersToAdd, openapi_types.UUID(uuid.MustParse(userIdStr)))
		}
	}

	// Find users to remove (in state but not in plan)
	var usersToRemove []openapi_types.UUID
	for _, userId := range stateData.UserIds {
		userIdStr := userId.ValueString()
		if !planUserIds[userIdStr] {
			usersToRemove = append(usersToRemove, openapi_types.UUID(uuid.MustParse(userIdStr)))
		}
	}

	// Add new members if any
	if len(usersToAdd) > 0 {
		addReq := managementclient.AddOrganizationMembersRequest{
			UserIds: usersToAdd,
		}

		addResp, err := r.client.OrganizationUpdateMembersWithResponse(ctx, orgID, addReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add organization members, got error: %s", err))
			return
		}

		if addResp.StatusCode() < 200 || addResp.StatusCode() >= 300 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add organization members, got status: %d", addResp.StatusCode()))
			return
		}
	}

	// Remove members if any
	if len(usersToRemove) > 0 {
		removeResp, err := r.client.OrganizationUpdateRemoveMembersWithResponse(ctx, orgID, managementclient.OrganizationUpdateRemoveMembersJSONRequestBody{
			UserIds: usersToRemove,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove organization members, got error: %s", err))
			return
		}

		if removeResp.StatusCode() < 200 || removeResp.StatusCode() >= 300 {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to remove organization members, got status: %d", removeResp.StatusCode()))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *OrganizationMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationMembersResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID from token", err.Error())
		return
	}

	userIds := data.UserIds
	if len(userIds) == 0 {
		resp.Diagnostics.AddError("Invalid User IDs", "User IDs are required")
		return
	}

	var userIdsSlice []openapi_types.UUID
	for _, userId := range userIds {
		userIdsSlice = append(userIdsSlice, openapi_types.UUID(uuid.MustParse(userId.ValueString())))
	}

	apiResp, err := r.client.OrganizationUpdateRemoveMembersWithResponse(ctx, orgID, managementclient.OrganizationUpdateRemoveMembersJSONRequestBody{
		UserIds: userIdsSlice,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove organization members, got error: %s", err))
		return
	}

	if apiResp.StatusCode() < 200 || apiResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to remove organization members, got status: %d", apiResp.StatusCode()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID should be the organization ID, but we'll use the one from the token
	if r.organizationID == "" {
		resp.Diagnostics.AddError("Import Error", "Organization ID not available from token")
		return
	}

	orgID, err := uuid.Parse(r.organizationID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Invalid Organization ID from token: %s", err.Error()))
		return
	}

	// Fetch current organization members to populate the state
	membersResp, err := r.client.OrganizationListMembersWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to fetch organization members: %s", err.Error()))
		return
	}

	if membersResp.StatusCode() < 200 || membersResp.StatusCode() >= 300 || membersResp.JSON200 == nil {
		resp.Diagnostics.AddError("Import Error", "Unable to fetch organization members from API")
		return
	}

	// Convert member user IDs to terraform state
	var userIds []types.String
	for _, member := range membersResp.JSON200.Rows {
		userIds = append(userIds, types.StringValue(member.UserId.String()))
	}

	// Set the state with current members
	data := OrganizationMembersResourceModel{
		UserIds: userIds,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
