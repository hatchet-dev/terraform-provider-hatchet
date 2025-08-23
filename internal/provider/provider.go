package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	managementclient "github.com/hatchet-dev/terraform-provider-hatchetcloud/internal/api"
)

var _ provider.Provider = &HatchetCloudProvider{}

type HatchetCloudProvider struct {
	version string
}

type HatchetCloudProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *HatchetCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hatchetcloud"
	resp.Version = p.version
}

func (p *HatchetCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "Management token for the Hatchet Cloud instance. Can also be provided via HATCHET_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *HatchetCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data HatchetCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("HATCHET_TOKEN")
	if token == "" {
		token = data.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Token Configuration",
			"The provider cannot create the Hatchet Cloud API client as there is a missing or empty value for the token. "+
				"Set the token value in the configuration or use the HATCHET_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	// Decode JWT to extract organization ID and endpoint
	claims, err := decodeJWT(token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Token",
			"The provided token is not a valid JWT or cannot be decoded. "+
				"Please ensure you are using a valid Hatchet Cloud management token.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if claims.Sub == "" {
		resp.Diagnostics.AddError(
			"Invalid Token Claims",
			"The JWT token does not contain a valid organization ID (sub claim).",
		)
		return
	}

	if claims.Issuer == "" {
		resp.Diagnostics.AddError(
			"Invalid Token Claims",
			"The JWT token does not contain a valid issuer.",
		)
		return
	}

	endpoint := claims.Issuer

	client := &HatchetCloudClient{
		Endpoint:        endpoint,
		Token:           token,
		OrganizationID:  claims.Sub,
		ProviderVersion: p.version,
	}
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hatchet Cloud client", map[string]any{
		"success":         true,
		"endpoint":        endpoint,
		"organization_id": claims.Sub,
	})
}

func (p *HatchetCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTenantResource,
		NewTenantAPITokenResource,
		NewOrganizationMembersResource,
	}
}

func (p *HatchetCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
		NewTenantDataSource,
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HatchetCloudProvider{
			version: version,
		}
	}
}

type HatchetCloudClient struct {
	Endpoint        string
	Token           string
	OrganizationID  string
	ProviderVersion string
}

// JWTClaims represents the structure of the JWT token
type JWTClaims struct {
	Sub string `json:"sub"` // Organization ID
	jwt.RegisteredClaims
}

// decodeJWT decodes the JWT token and extracts organization ID and endpoint
func decodeJWT(tokenString string) (*JWTClaims, error) {
	// Parse the token without verification (we just need the claims)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	return claims, nil
}

// createAPIClient creates a new API client with proper headers
func createAPIClient(endpoint, token, providerVersion string) (*managementclient.ClientWithResponses, error) {
	return managementclient.NewClientWithResponses(
		endpoint,
		managementclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			req.Header.Set("User-Agent", fmt.Sprintf("terraform-provider-hatchetcloud/%s", providerVersion))
			return nil
		}),
	)
}
