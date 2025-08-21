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
	_ datasource.DataSource              = &OrganizationMembersDataSource{}
	_ datasource.DataSourceWithConfigure = &OrganizationMembersDataSource{}
)

func NewOrganizationMembersDataSource() datasource.DataSource {
	return &OrganizationMembersDataSource{}
}

type OrganizationMembersDataSource struct {
	client *managementclient.ClientWithResponses
}

type OrganizationMembersDataSourceModel struct {
	OrganizationID types.String                        `tfsdk:"organization_id"`
	Members        []OrganizationMemberDataSourceModel `tfsdk:"members"`
}

type OrganizationMemberDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	UserID     types.String `tfsdk:"user_id"`
	MemberType types.String `tfsdk:"member_type"`
}

func (d *OrganizationMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_members"
}

func (d *OrganizationMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches all members of a Hatchet organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the organization.",
				Required:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "List of organization members.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the organization member.",
							Computed:            true,
						},
						"user_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the user.",
							Computed:            true,
						},
						"member_type": schema.StringAttribute{
							MarkdownDescription: "The type of member (OWNER).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationMembersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID, err := uuid.Parse(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Organization ID", err.Error())
		return
	}

	membersResp, err := d.client.OrganizationListMembersWithResponse(ctx, orgID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization members, got error: %s", err))
		return
	}

	if membersResp.StatusCode() != 200 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read organization members, got status: %d", membersResp.StatusCode()))
		return
	}

	if membersResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Organization members not found")
		return
	}

	members := make([]OrganizationMemberDataSourceModel, len(membersResp.JSON200.Rows))
	for i, member := range membersResp.JSON200.Rows {
		members[i] = OrganizationMemberDataSourceModel{
			ID:         types.StringValue(member.Metadata.Id),
			UserID:     types.StringValue(member.UserId.String()),
			MemberType: types.StringValue(string(member.MemberType)),
		}
	}

	data.Members = members

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
