package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	managementclient "github.com/hatchet-dev/terraform-provider-hatchetcloud/internal/api"
)

var (
	_ datasource.DataSource              = &UserDataSource{}
	_ datasource.DataSourceWithConfigure = &UserDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *managementclient.ClientWithResponses
}

type UserDataSourceModel struct {
	Email types.String `tfsdk:"email"`
	ID    types.String `tfsdk:"id"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about a Hatchet user by email.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the user.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := data.Email.ValueString()
	if email == "" {
		resp.Diagnostics.AddError("Invalid Email", "Email is required")
		return
	}

	params := &managementclient.UserGetParams{
		Email: email,
	}

	userResp, err := d.client.UserGetWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	if userResp.StatusCode() < 200 || userResp.StatusCode() >= 300 || userResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("User not found or access denied. Status: %d", userResp.StatusCode()))
		return
	}

	user := userResp.JSON200
	data.ID = types.StringValue(user.Id.String())
	data.Email = types.StringValue(string(user.Email))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
