package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

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

	serverURL := os.Getenv("MAZEVAULT_SERVER_URL")
	if !data.ServerURL.IsNull() {
		serverURL = data.ServerURL.ValueString()
	}

	if serverURL == "" {
		resp.Diagnostics.AddError(
			"Missing Server URL",
			"The server_url configuration or MAZEVAULT_SERVER_URL environment variable is required",
		)
		return
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
		return
	}

	// Resolve timeout — default matches SDK default (1 minute).
	timeoutDuration := time.Minute
	if !data.Timeout.IsNull() && data.Timeout.ValueString() != "" {
		d, err := time.ParseDuration(data.Timeout.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Timeout",
				fmt.Sprintf("Cannot parse timeout %q as a Go duration (e.g. \"30s\", \"2m\"): %s", data.Timeout.ValueString(), err),
			)
			return
		}
		timeoutDuration = d
	}

	// Build HTTP transport — honour skip_tls_verify for dev/test environments.
	// Clone DefaultTransport so all defaults (proxy-from-env, keepalive, HTTP/2
	// upgrade, dial/idle timeouts) are preserved; only InsecureSkipVerify is
	// toggled. #nosec G402 — intentionally user-controlled; only activated when
	// the operator explicitly sets skip_tls_verify = true.
	var transport http.RoundTripper = http.DefaultTransport
	if !data.SkipTLS.IsNull() && data.SkipTLS.ValueBool() {
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			cloned := dt.Clone()
			if cloned.TLSClientConfig == nil {
				cloned.TLSClientConfig = &tls.Config{} //nolint:gosec
			}
			cloned.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec
			transport = cloned
		} else {
			// Fallback: DefaultTransport has been replaced globally (e.g. in tests).
			transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			}
		}
	}

	// Create a new MazeVault SDK client with the resolved HTTP settings.
	c := mazevault.NewClient(serverURL)
	c.HTTPClient = &http.Client{
		Timeout:   timeoutDuration,
		Transport: transport,
	}

	if apiToken != "" {
		c.SetToken(apiToken)
	} else if clientID != "" && clientSecret != "" {
		// Exchange client credentials (OAuth2 client_credentials grant) for a bearer token.
		if err := c.ClientCredentials(clientID, clientSecret); err != nil {
			resp.Diagnostics.AddError("Authentication Error",
				fmt.Sprintf("client_credentials grant failed: %s", err))
			return
		}
	}

	// Pass the configured SDK client to all resources and data sources.
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
		resources.NewRotationTemplateResource,
		resources.NewDeploymentResource,
		// Sync Rules
		resources.NewSyncRuleResource,
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
