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
	_ resource.Resource                = &TenantAPITokenResource{}
	_ resource.ResourceWithImportState = &TenantAPITokenResource{}
)

func NewTenantAPITokenResource() resource.Resource {
	return &TenantAPITokenResource{}
}

type TenantAPITokenResource struct {
	client *managementclient.ClientWithResponses
}

type TenantAPITokenResourceModel struct {
	ID        types.String `tfsdk:"id"`
	TenantID  types.String `tfsdk:"tenant_id"`
	Name      types.String `tfsdk:"name"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	Token     types.String `tfsdk:"token"`
}

func (r *TenantAPITokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_api_token"
}

func (r *TenantAPITokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hatchet tenant API token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the API token.",
				Computed:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tenant this API token belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the API token.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration date of the API token (optional).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The API token value. This is only available immediately after creation.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *TenantAPITokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TenantAPITokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TenantAPITokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, err := uuid.Parse(data.TenantID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	createReq := managementclient.CreateTenantAPITokenRequest{
		Name: data.Name.ValueString(),
	}

	if !data.ExpiresAt.IsNull() {
		expiresIn := data.ExpiresAt.ValueString()
		createReq.ExpiresIn = &expiresIn
	}

	tokenResp, err := r.client.OrganizationTenantCreateApiTokenWithResponse(ctx, tenantID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API token, got error: %s", err))
		return
	}

	if tokenResp.StatusCode() < 200 || tokenResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create API token, got status: %d", tokenResp.StatusCode()))
		return
	}

	if tokenResp.JSON201 == nil {
		resp.Diagnostics.AddError("API Error", "API token creation failed")
		return
	}

	data.ID = types.StringValue(tokenResp.JSON201.Token)
	data.Token = types.StringValue(tokenResp.JSON201.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantAPITokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TenantAPITokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, err := uuid.Parse(data.TenantID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	tokensResp, err := r.client.OrganizationTenantListApiTokensWithResponse(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API tokens, got error: %s", err))
		return
	}

	if tokensResp.StatusCode() < 200 || tokensResp.StatusCode() >= 300 || tokensResp.JSON200 == nil {
		resp.Diagnostics.AddError("API Error", "Unable to read API tokens")
		return
	}

	tokenID := data.ID.ValueString()
	tokenFound := false
	for _, token := range *tokensResp.JSON200.Rows {
		if token.Metadata.Id == tokenID {
			tokenFound = true
			data.Name = types.StringValue(token.Name)
			if !token.ExpiresAt.IsZero() {
				data.ExpiresAt = types.StringValue(token.ExpiresAt.String())
			} else {
				data.ExpiresAt = types.StringNull()
			}
			break
		}
	}

	if !tokenFound {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TenantAPITokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"API token update is not supported. API tokens are immutable once created.",
	)
}

func (r *TenantAPITokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TenantAPITokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID, err := uuid.Parse(data.TenantID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Tenant ID", err.Error())
		return
	}

	tokenID, err := uuid.Parse(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Token ID", err.Error())
		return
	}

	deleteResp, err := r.client.OrganizationTenantDeleteApiTokenWithResponse(ctx, tenantID, tokenID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API token, got error: %s", err))
		return
	}

	if deleteResp.StatusCode() < 200 || deleteResp.StatusCode() >= 300 {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete API token, got status: %d", deleteResp.StatusCode()))
		return
	}
}

func (r *TenantAPITokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
