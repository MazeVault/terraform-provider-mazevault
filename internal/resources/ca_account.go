package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CAAccountResource{}
var _ resource.ResourceWithValidateConfig = &CAAccountResource{}
var _ resource.ResourceWithImportState = &CAAccountResource{}

var supportedCAProviderTypes = map[string]struct{}{
	"digicert":    {},
	"sslmarket":   {},
	"rapidssl":    {},
	"sectigo":     {},
	"globalsign":  {},
	"entrust":     {},
	"letsencrypt": {},
	"zerossl":     {},
	"acme":        {},
	"internal":    {},
	"adcs":        {},
	"vault":       {},
	"smallstep":   {},
	"est":         {},
	"venafi":      {},
}

const supportedCAProviderTypeList = "digicert, sslmarket, rapidssl, sectigo, globalsign, entrust, letsencrypt, zerossl, acme, internal, adcs, vault, smallstep, est, venafi"

func NewCAAccountResource() resource.Resource { return &CAAccountResource{} }

type CAAccountResource struct{ client *mazevault.Client }

type CAAccountModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	APIKey         types.String `tfsdk:"api_key"`
	BaseURL        types.String `tfsdk:"base_url"`
	Email          types.String `tfsdk:"email"`
	DirectoryURL   types.String `tfsdk:"directory_url"`
	EABKid         types.String `tfsdk:"eab_kid"`
	EABHMACKey     types.String `tfsdk:"eab_hmac_key"`
	Settings       types.Map    `tfsdk:"settings"`
	Status         types.String `tfsdk:"status"`
}

// buildCredentials assembles the backend credentials map from the non-empty
// provider fields. ACME providers (letsencrypt/zerossl/acme) consume email,
// directory_url and the EAB pair; API-style providers use api_key/base_url.
func (data CAAccountModel) buildCredentials() map[string]string {
	cfg := map[string]string{}
	if v := data.APIKey.ValueString(); v != "" {
		cfg["api_key"] = v
	}
	if v := data.BaseURL.ValueString(); v != "" {
		cfg["base_url"] = v
	}
	if v := data.Email.ValueString(); v != "" {
		cfg["email"] = v
	}
	if v := data.DirectoryURL.ValueString(); v != "" {
		cfg["directory_url"] = v
	}
	if v := data.EABKid.ValueString(); v != "" {
		cfg["eab_kid"] = v
	}
	if v := data.EABHMACKey.ValueString(); v != "" {
		cfg["eab_hmac_key"] = v
	}
	return cfg
}

func (data CAAccountModel) buildSettings(ctx context.Context) (map[string]interface{}, diag.Diagnostics) {
	settings := map[string]interface{}{}
	if v := data.BaseURL.ValueString(); v != "" {
		settings["base_url"] = v
	}
	if !data.Settings.IsNull() && !data.Settings.IsUnknown() {
		var configured map[string]string
		diagnostics := data.Settings.ElementsAs(ctx, &configured, false)
		if diagnostics.HasError() {
			return settings, diagnostics
		}
		for key, value := range configured {
			settings[key] = value
		}
	}
	return settings, nil
}

func canonicalCAProviderType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateCanonicalCAProviderType(raw string) error {
	canonical := canonicalCAProviderType(raw)
	if canonical == "" {
		return fmt.Errorf("provider_type must be a non-empty canonical lowercase value without surrounding whitespace")
	}
	if raw == canonical {
		return nil
	}
	return fmt.Errorf("provider_type must be canonical lowercase without surrounding whitespace; use %q", canonical)
}

func validateCAProviderType(raw string) (string, error) {
	if err := validateCanonicalCAProviderType(raw); err != nil {
		return "", err
	}
	providerType := canonicalCAProviderType(raw)
	if _, ok := supportedCAProviderTypes[providerType]; !ok {
		return "", fmt.Errorf("provider_type %q is not supported; must be one of: %s", providerType, supportedCAProviderTypeList)
	}
	return providerType, nil
}

// validateCredentials ensures non-internal providers carry at least one
// credential. The backend rejects an empty credentials map with a generic
// "Credentials are required for this provider type" 400; validating here yields a
// clear, actionable message (e.g. ACME needs email + directory_url).
func (data CAAccountModel) validateCredentials() error {
	if data.ProviderType.IsNull() || data.ProviderType.IsUnknown() {
		return nil
	}
	providerType, err := validateCAProviderType(data.ProviderType.ValueString())
	if err != nil {
		return err
	}
	if providerType == "internal" {
		return nil
	}
	credentials := data.buildCredentials()
	switch providerType {
	case "letsencrypt", "zerossl", "acme":
		if strings.TrimSpace(credentials["email"]) == "" {
			return fmt.Errorf("provider_type %q requires email for ACME account registration", providerType)
		}
		if providerType == "acme" && strings.TrimSpace(credentials["directory_url"]) == "" {
			return fmt.Errorf("provider_type %q requires directory_url for custom ACME directory discovery", providerType)
		}
		return nil
	case "digicert", "sslmarket", "rapidssl", "sectigo", "globalsign", "entrust", "venafi":
		if strings.TrimSpace(credentials["api_key"]) == "" {
			return fmt.Errorf("provider_type %q requires api_key", providerType)
		}
		return nil
	}
	if len(credentials) == 0 {
		return fmt.Errorf("provider_type %q requires credentials: set api_key for commercial providers (e.g. digicert, venafi), or email + directory_url (plus eab_kid/eab_hmac_key when the ACME CA requires External Account Binding) for ACME providers (e.g. letsencrypt, zerossl, acme)", providerType)
	}
	return nil
}

func (r *CAAccountResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.ProviderType.IsNull() || data.ProviderType.IsUnknown() {
		return
	}
	if _, err := validateCAProviderType(data.ProviderType.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("provider_type"), "Invalid CA Account Provider Type", err.Error())
	}
}
func (r *CAAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_account"
}

func (r *CAAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connects a Certificate Authority (CA) account record or integration to a MazeVault organization. Supports external providers, self-hosted providers, and internal CA account references.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the CA account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization this CA account belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for this CA account.",
			},
			"provider_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "CA provider type. Accepted canonical lowercase values for Terraform-managed CA account creation are `digicert`, `sslmarket`, `rapidssl`, `sectigo`, `globalsign`, `entrust`, `letsencrypt`, `zerossl`, `acme`, `internal`, `adcs`, `vault`, `smallstep`, `est`, and `venafi`. Values must be lowercase with no surrounding whitespace. Use `mazevault_ca` for normal internal root CA lifecycle management; `internal` here represents an internal CA account record/reference. Non-internal providers require credentials (see `api_key` for commercial providers; `email`/`directory_url`/EAB for ACME). I.CA/PostSignum and PSD2 account creation remain disabled until their backend integrations report a connectable status.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key or credential for API-based CA providers (DigiCert, SSLMarket, ZeroSSL EAB, …).",
			},
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base URL for the CA provider API (required for self-hosted CAs).",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ACME account contact email. Required for ACME providers (`letsencrypt`, `zerossl`, `acme`).",
			},
			"directory_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ACME directory URL. Required for custom `acme` providers; defaulted for `letsencrypt`/`zerossl`.",
			},
			"eab_kid": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "External Account Binding (EAB) key identifier for ACME providers that require it (e.g. ZeroSSL).",
			},
			"eab_hmac_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "External Account Binding (EAB) HMAC key for ACME providers that require it. Write-only.",
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Non-secret provider settings persisted in the CA account metadata, such as `organization_id`, `container_id`, `server_platform_id`, `dcv_method`, `certificate_dcv_scope`, or SSLMarket contact defaults. Do not place API keys or passwords here.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current connection status of the CA account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CAAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *CAAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := data.validateCredentials(); err != nil {
		resp.Diagnostics.AddError("Invalid CA Account Credentials", err.Error())
		return
	}
	settings, diags := data.buildSettings(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerType := canonicalCAProviderType(data.ProviderType.ValueString())
	created, err := r.client.ConnectCAAccount(data.OrganizationID.ValueString(), &mazevault.CreateCAAccountRequest{
		DisplayName:  data.Name.ValueString(),
		Name:         data.Name.ValueString(),
		ProviderType: providerType,
		Credentials:  data.buildCredentials(),
		Settings:     settings,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create CA Account Error", fmt.Sprintf("Unable to create CA account: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Status = types.StringValue(created.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	acc, err := r.client.GetCAAccount(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read CA Account Error", fmt.Sprintf("Unable to read CA account: %s", err))
		return
	}
	if acc == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	name := acc.DisplayName
	if name == "" {
		name = acc.Name
	}
	if name != "" {
		data.Name = types.StringValue(name)
	}
	data.ProviderType = types.StringValue(acc.ProviderType)
	data.Status = types.StringValue(acc.Status)
	// api_key, base_url, email, directory_url, eab_kid, eab_hmac_key are
	// write-only credentials. settings are provider metadata; preserve state because
	// the SDK response does not currently expose backend metadata distinctly.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CAAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := plan.validateCredentials(); err != nil {
		resp.Diagnostics.AddError("Invalid CA Account Credentials", err.Error())
		return
	}
	cfgU := plan.buildCredentials()
	settings, diags := plan.buildSettings(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerType := canonicalCAProviderType(plan.ProviderType.ValueString())
	if _, err := r.client.UpdateCAAccount(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateCAAccountRequest{
		DisplayName:  plan.Name.ValueString(),
		Name:         plan.Name.ValueString(),
		ProviderType: providerType,
		Credentials:  cfgU,
		Settings:     settings,
	}); err != nil {
		resp.Diagnostics.AddError("Update CA Account Error", fmt.Sprintf("Unable to update CA account: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CAAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCAAccount(data.OrganizationID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete CA Account Error", fmt.Sprintf("Unable to delete CA account: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<organization_id>/<resource_id>"
func (r *CAAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<organization_id>/<resource_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
