package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchetcloud/internal/api"
)

var (
	_ resource.Resource              = &OrganizationMemberResource{}
	_ resource.ResourceWithImportState = &OrganizationMemberResource{}
)

func NewOrganizationMemberResource() resource.Resource {
	return &OrganizationMemberResource{}
}

type OrganizationMemberResource struct {
	client *managementclient.ClientWithResponses
}

type OrganizationMemberResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	UserID         types.String `tfsdk:"user_id"`
	MemberType     types.String `tfsdk:"member_type"`
}

func (r *OrganizationMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_member"
}

func (r *OrganizationMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hatchet organization member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization member.",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user to add as a member.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_type": schema.StringAttribute{
				MarkdownDescription: "The type of member (OWNER).",
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

	r.client = apiClient
}

func (r *OrganizationMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID", err.Error())
		return
	}

	userID, err := uuid.Parse(data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid User ID", err.Error())
		return
	}

	addReq := managementclient.AddOrganizationMembersRequest{
		UserIds: []uuid.UUID{userID},
	}

	memberResp, err := r.client.OrganizationUpdateMembersWithResponse(ctx, orgID, addReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add organization member, got error: %s", err))
		return
	}

	if memberResp.StatusCode() != 201 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add organization member, got status: %d", memberResp.StatusCode()))
		return
	}

	if memberResp.JSON201 == nil {
		resp.Diagnostics.AddError("API Error", "Member addition failed")
		return
	}

	for _, member := range memberResp.JSON201.Rows {
		if member.UserId == userID {
			data.ID = types.StringValue(member.Metadata.Id)
			data.MemberType = types.StringValue(string(member.MemberType))
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID", err.Error())
		return
	}

	membersResp, err := r.client.OrganizationListMembersWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization members, got error: %s", err))
		return
	}

	if membersResp.StatusCode() != 200 || membersResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Unable to read organization members")
		return
	}

	memberID := data.ID.ValueString()
	memberFound := false
	for _, member := range membersResp.JSON200.Rows {
		if member.Metadata.Id == memberID {
			memberFound = true
			data.UserID = types.StringValue(member.UserId.String())
			data.MemberType = types.StringValue(string(member.MemberType))
			break
		}
	}

	if !memberFound {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Organization member update is not supported. Members are immutable once added.",
	)
}

func (r *OrganizationMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationMemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	memberID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Member ID", err.Error())
		return
	}

	deleteResp, err := r.client.OrganizationMemberDeleteWithResponse(ctx, memberID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete organization member, got error: %s", err))
		return
	}

	if deleteResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete organization member, got status: %d", deleteResp.StatusCode()))
		return
	}
}

func (r *OrganizationMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}