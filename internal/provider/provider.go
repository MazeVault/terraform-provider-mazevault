package provider

import (
	"context"
	"fmt"
	"os"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/MazeVault/maze-core/terraform-provider-mazevault/internal/datasources"
	"github.com/MazeVault/maze-core/terraform-provider-mazevault/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure MazeVaultProvider satisfies various provider interfaces.
var _ provider.Provider = &MazeVaultProvider{}

// MazeVaultProvider defines the provider implementation.
type MazeVaultProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MazeVaultProviderModel describes the provider data model.
type MazeVaultProviderModel struct {
	ServerURL    types.String `tfsdk:"server_url"`
	APIToken     types.String `tfsdk:"api_token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	SkipTLS      types.Bool   `tfsdk:"skip_tls_verify"`
	Timeout      types.String `tfsdk:"timeout"`
}

func (p *MazeVaultProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mazevault"
	resp.Version = p.version
}

func (p *MazeVaultProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "MazeVault server URL",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Service account client ID",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Service account client secret",
				Optional:            true,
				Sensitive:           true,
			},
			"skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS verification (for development only)",
				Optional:            true,
			},
			"timeout": schema.StringAttribute{
				MarkdownDescription: "API timeout duration (e.g. 30s)",
				Optional:            true,
			},
		},
	}
}

func (p *MazeVaultProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MazeVaultProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	serverURL := os.Getenv("MAZEVAULT_SERVER_URL")
	if !data.ServerURL.IsNull() {
		serverURL = data.ServerURL.ValueString()
	}

	if serverURL == "" {
		resp.Diagnostics.AddError(
			"Missing Server URL",
			"The server_url configuration or MAZEVAULT_SERVER_URL environment variable is required",
		)
	}

	apiToken := os.Getenv("MAZEVAULT_API_TOKEN")
	if !data.APIToken.IsNull() {
		apiToken = data.APIToken.ValueString()
	}

	clientID := os.Getenv("MAZEVAULT_CLIENT_ID")
	if !data.ClientID.IsNull() {
		clientID = data.ClientID.ValueString()
	}

	clientSecret := os.Getenv("MAZEVAULT_CLIENT_SECRET")
	if !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	if apiToken == "" && (clientID == "" || clientSecret == "") {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"Either api_token or client_id/client_secret must be provided via configuration or environment variables",
		)
	}

	// Create a new MazeVault SDK client
	c := mazevault.NewClient(serverURL)

	if apiToken != "" {
		c.SetToken(apiToken)
	} else if clientID != "" && clientSecret != "" {
		// Exchange client credentials for a bearer token
		if err := c.ClientCredentials(clientID, clientSecret); err != nil {
			resp.Diagnostics.AddError("Authentication Error",
				fmt.Sprintf("client_credentials grant failed: %s", err))
			return
		}
	}

	// Pass the unified SDK client to all resources and data sources.
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *MazeVaultProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Core
		resources.NewOrganizationResource,
		resources.NewProjectResource,
		resources.NewProjectSettingsResource,
		// Secrets
		resources.NewSecretResource,
		resources.NewSecretLinkResource,
		resources.NewSharedSecretResource,
		// Identity
		resources.NewServiceIdentityResource,
		resources.NewAPITokenResource,
		// PKI
		resources.NewCAResource,
		resources.NewCAAccountResource,
		resources.NewCertificateResource,
		resources.NewCertificateTemplateResource,
		resources.NewConfigTemplateResource,
		// RBAC
		resources.NewRoleResource,
		resources.NewGroupMappingResource,
		resources.NewUserResource,
		resources.NewUserRoleResource,
		// Integrations
		resources.NewIntegrationResource,
		resources.NewIntegrationGroupResource,
		resources.NewConsistencyGroupResource,
		resources.NewKeytabResource,
		// Identity Providers & Environments
		resources.NewIdentityProviderResource,
		resources.NewEnvironmentResource,
		// Policies
		resources.NewRenewalPolicyResource,
		resources.NewApprovalPolicyResource,
		// Rotation & Deployments
		resources.NewRotationConfigResource,
		resources.NewRotationWorkflowResource,
		resources.NewDeploymentResource,
	}
}

func (p *MazeVaultProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// PKI
		datasources.NewProjectCertificatesDataSource,
		datasources.NewProjectCAsDataSource,
		datasources.NewProjectCertificateTemplatesDataSource,
		datasources.NewProjectCSRsDataSource,
		// Secrets & Consistency
		datasources.NewConsistencyStatusDataSource,
		// Organizations & Projects
		datasources.NewOrganizationDataSource,
		datasources.NewProjectDataSource,
		// Audit & Rotation
		datasources.NewAuditLogsDataSource,
		datasources.NewRotationExecutionsDataSource,
		datasources.NewRenewalQueueDataSource,
		// RBAC
		datasources.NewUsersDataSource,
		datasources.NewRolesDataSource,
		// Integrations & Certificates
		datasources.NewIntegrationsDataSource,
		datasources.NewCertificateDataSource,
		// CA Accounts & Environments
		datasources.NewCAAccountsDataSource,
		datasources.NewEnvironmentsDataSource,
		// Secrets read-only
		datasources.NewSecretDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MazeVaultProvider{
			version: version,
		}
	}
}
