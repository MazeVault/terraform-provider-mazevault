package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CAAccountResource{}
var _ resource.ResourceWithImportState = &CAAccountResource{}

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

// validateCredentials ensures non-internal providers carry at least one
// credential. The backend rejects an empty credentials map with a generic
// "Credentials are required for this provider type" 400; validating here yields a
// clear, actionable message (e.g. ACME needs email + directory_url).
func (data CAAccountModel) validateCredentials() error {
	if data.ProviderType.ValueString() == "internal" {
		return nil
	}
	if len(data.buildCredentials()) == 0 {
		return fmt.Errorf("provider_type %q requires credentials: set api_key for commercial providers (e.g. digicert, venafi), or email + directory_url (plus eab_kid/eab_hmac_key when the ACME CA requires External Account Binding) for ACME providers (e.g. letsencrypt, zerossl, acme)", data.ProviderType.ValueString())
	}
	return nil
}

func (r *CAAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_account"
}

func (r *CAAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connects an external Certificate Authority (CA) account to a MazeVault organization. Supports DigiCert, Venafi, ACME, and SmallStep.",
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
				MarkdownDescription: "CA provider type: `digicert`, `venafi`, `smallstep`, `letsencrypt`, `zerossl`, `acme`, or `internal`. Non-internal providers require credentials (see `api_key` for commercial providers; `email`/`directory_url`/EAB for ACME).",
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
	created, err := r.client.ConnectCAAccount(data.OrganizationID.ValueString(), &mazevault.CreateCAAccountRequest{
		DisplayName:  data.Name.ValueString(),
		Name:         data.Name.ValueString(),
		ProviderType: data.ProviderType.ValueString(),
		Credentials:  data.buildCredentials(),
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
	// write-only credentials; the backend never returns them. Preserve state.
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
	if _, err := r.client.UpdateCAAccount(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateCAAccountRequest{
		DisplayName:  plan.Name.ValueString(),
		Name:         plan.Name.ValueString(),
		ProviderType: plan.ProviderType.ValueString(),
		Credentials:  cfgU,
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
