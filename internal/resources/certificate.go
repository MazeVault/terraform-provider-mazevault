package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}
var _ resource.ResourceWithValidateConfig = &CertificateResource{}

// ValidateConfig enforces the private key custody contract at plan time.
func (r *CertificateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mode := data.KeyStorageMode.ValueString()
	if !data.KeyStorageMode.IsNull() && !data.KeyStorageMode.IsUnknown() && mode != "local" && mode != "external" {
		resp.Diagnostics.AddAttributeError(path.Root("key_storage_mode"), "Invalid key_storage_mode",
			fmt.Sprintf("%q is not supported; use \"local\" or \"external\"", mode))
	}
	hasCSR := !data.CSRPEM.IsNull() && !data.CSRPEM.IsUnknown() && data.CSRPEM.ValueString() != ""
	if hasCSR {
		for _, attr := range []struct {
			name  string
			value types.String
		}{{"key_storage_mode", data.KeyStorageMode}, {"key_integration_id", data.KeyIntegrationID}, {"key_secret_name", data.KeySecretName}} {
			if !attr.value.IsNull() && !attr.value.IsUnknown() && attr.value.ValueString() != "" {
				resp.Diagnostics.AddAttributeError(path.Root(attr.name), "Conflicts with csr_pem",
					"Private key custody attributes apply only to server-side generated key pairs; remove them when csr_pem is supplied.")
			}
		}
	}
	if !data.KeyStorageMode.IsUnknown() && mode != "external" {
		for _, attr := range []struct {
			name  string
			value types.String
		}{{"key_integration_id", data.KeyIntegrationID}, {"key_secret_name", data.KeySecretName}} {
			if !attr.value.IsNull() && !attr.value.IsUnknown() && attr.value.ValueString() != "" {
				resp.Diagnostics.AddAttributeError(path.Root(attr.name), "Requires key_storage_mode = \"external\"",
					fmt.Sprintf("%s is only meaningful when the private key is offloaded to an external secret manager.", attr.name))
			}
		}
	}
}

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	client *mazevault.Client
}

type CertificateResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	CommonName              types.String `tfsdk:"common_name"`
	ProjectID               types.String `tfsdk:"project_id"`
	TemplateID              types.String `tfsdk:"template_id"`
	CSRPEM                  types.String `tfsdk:"csr_pem"`
	TTL                     types.String `tfsdk:"ttl"`
	KeySize                 types.Int64  `tfsdk:"key_size"`
	SerialNumber            types.String `tfsdk:"serial_number"`
	CertificatePEM          types.String `tfsdk:"certificate_pem"`
	PrivateKeyPEM           types.String `tfsdk:"private_key_pem"`
	Status                  types.String `tfsdk:"status"`
	ExpiresAt               types.String `tfsdk:"expires_at"`
	OrganizationCAAccountID types.String `tfsdk:"organization_ca_account_id"`
	DCVMethod               types.String `tfsdk:"dcv_method"`
	CAProductID             types.String `tfsdk:"ca_product_id"`
	OrderYears              types.Int64  `tfsdk:"order_years"`
	RequestedValidityDays   types.Int64  `tfsdk:"requested_validity_days"`
	TechContact             types.Map    `tfsdk:"tech_contact"`
	Voucher                 types.String `tfsdk:"voucher"`
	KeyStorageMode          types.String `tfsdk:"key_storage_mode"`
	KeyIntegrationID        types.String `tfsdk:"key_integration_id"`
	KeySecretName           types.String `tfsdk:"key_secret_name"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Issues a certificate via a CSR submission. Deletion revokes the certificate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the issued certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"common_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Common Name (CN) for the certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the project to assign the issued certificate to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"template_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the certificate template to apply.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"csr_pem": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "PEM-encoded Certificate Signing Request. Required when using an externally generated key pair.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_storage_mode": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Private key custody for server-side generated key pairs (ignored when `csr_pem` is set). " +
					"`local` (default) encrypts the key with the MazeVault master key; `external` offloads it to a " +
					"secret-manager integration and keeps only a reference. **Orchestrator Mode requires `external`.**",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_integration_id": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "ID of the `mazevault_integration` (e.g. Azure Key Vault) that receives the private key when " +
					"`key_storage_mode = \"external\"`. When omitted the project's default secret-manager integration is used.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_secret_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the secret created in the external secret manager for the private key (`external` mode only). Defaults to a name derived from the common name.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ttl": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Requested certificate TTL (e.g. `8760h`). May be constrained by template settings.",
			},
			"key_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Key size in bits for server-side key generation (if `csr_pem` is omitted).",
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "X.509 serial number of the issued certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"certificate_pem": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "PEM-encoded certificate chain.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"private_key_pem": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "PEM-encoded private key. Only present when MazeVault generates the key pair server-side.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Certificate status (`active`, `revoked`, `expired`).",
			},
			"expires_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 expiry timestamp.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_ca_account_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the organization-level CA account to sign with, overriding the template's CA resolution. Use `mazevault_ca.account_id` to issue from the organization internal CA, or an external `mazevault_ca_account` id. When omitted, the template's CA (or the active internal CA fallback) is used.",
			},
			"dcv_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Domain validation method for commercial CAs (`email`, `dns`, `file`). Overrides the template; empty inherits template, then CA account. Changing it forces a new certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ca_product_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "CA account product to order (commercial CAs). Overrides the template product. Changing it forces a new certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"order_years": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Subscription length in years for multi-year commercial CA orders (SSLMarket, 1–6). Length of the paid order, not of the certificate. Changing it forces a new certificate.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"requested_validity_days": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Requested certificate validity in days for CAs that honour it. Changing it forces a new certificate.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"tech_contact": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Technical contact forwarded with the commercial CA order. Keys: `title`, `firstname`, `lastname`, `email`, `tel`. Overrides the template contact.",
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"voucher": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Voucher code applied to the commercial CA order. Inherited from the CA account when empty.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	csr := &mazevault.SubmitCSRRequest{
		RequestedCN: data.CommonName.ValueString(),
	}
	if !data.TemplateID.IsNull() {
		csr.TemplateID = data.TemplateID.ValueString()
	}
	if !data.ProjectID.IsNull() && !data.ProjectID.IsUnknown() && data.ProjectID.ValueString() != "" {
		csr.ProjectID = data.ProjectID.ValueString()
	}
	if !data.OrganizationCAAccountID.IsNull() && !data.OrganizationCAAccountID.IsUnknown() && data.OrganizationCAAccountID.ValueString() != "" {
		csr.OrganizationCAAccountID = data.OrganizationCAAccountID.ValueString()
	}
	if !data.KeySize.IsNull() && !data.KeySize.IsUnknown() {
		csr.RequestedKeySize = int(data.KeySize.ValueInt64())
	}
	// Commercial-CA order overrides (request -> template -> CA account precedence).
	if v := data.DCVMethod.ValueString(); v != "" && !data.DCVMethod.IsUnknown() {
		csr.DCVMethod = v
	}
	if v := data.CAProductID.ValueString(); v != "" && !data.CAProductID.IsUnknown() {
		csr.CAProductID = v
	}
	if !data.OrderYears.IsNull() && !data.OrderYears.IsUnknown() && data.OrderYears.ValueInt64() > 0 {
		years := int(data.OrderYears.ValueInt64())
		csr.OrderYears = &years
	}
	if !data.RequestedValidityDays.IsNull() && !data.RequestedValidityDays.IsUnknown() && data.RequestedValidityDays.ValueInt64() > 0 {
		days := int(data.RequestedValidityDays.ValueInt64())
		csr.RequestedValidityDays = &days
	}
	if !data.TechContact.IsNull() && !data.TechContact.IsUnknown() {
		contact := map[string]string{}
		resp.Diagnostics.Append(data.TechContact.ElementsAs(ctx, &contact, false)...)
		if len(contact) > 0 {
			csr.TechContact = contact
		}
	}
	if v := data.Voucher.ValueString(); v != "" && !data.Voucher.IsUnknown() {
		csr.Voucher = v
	}
	// When a CSR is supplied use it as-is; otherwise ask the backend to generate
	// the key pair and CSR server-side so a certificate can be issued from
	// Terraform without an external key/CSR toolchain.
	if !data.CSRPEM.IsNull() && data.CSRPEM.ValueString() != "" {
		csr.CSRPEM = data.CSRPEM.ValueString()
		csr.CSRSourceMode = "uploaded_csr"
	} else {
		csr.CSRSourceMode = "mazevault_generated"
		if v := data.KeyStorageMode.ValueString(); v != "" && !data.KeyStorageMode.IsUnknown() {
			csr.KeyStorageMode = v
		}
		if v := data.KeyIntegrationID.ValueString(); v != "" && !data.KeyIntegrationID.IsUnknown() {
			csr.KeyIntegrationID = v
		}
		if v := data.KeySecretName.ValueString(); v != "" && !data.KeySecretName.IsUnknown() {
			csr.KeySecretName = v
		}
	}

	cert, err := r.client.SubmitCSR(csr)
	if err != nil {
		resp.Diagnostics.AddError("Create Certificate Error", fmt.Sprintf("Unable to issue certificate: %s", err))
		return
	}

	data.ID = types.StringValue(cert.ID)
	data.SerialNumber = types.StringValue(cert.SerialNumber)
	data.CertificatePEM = types.StringValue(cert.CertificatePEM)
	if cert.PrivateKeyPEM != "" {
		data.PrivateKeyPEM = types.StringValue(cert.PrivateKeyPEM)
	} else {
		// No server-side key returned (e.g. CSR-based or externally-held key) —
		// the value must still be known after apply.
		data.PrivateKeyPEM = types.StringNull()
	}
	data.Status = types.StringValue(cert.Status)
	data.ExpiresAt = types.StringValue(cert.ExpiresAt)
	// Reflect the project the certificate was assigned to.
	if data.ProjectID.IsNull() || data.ProjectID.IsUnknown() {
		if cert.ProjectID != nil && *cert.ProjectID != "" {
			data.ProjectID = types.StringValue(*cert.ProjectID)
		} else {
			data.ProjectID = types.StringNull()
		}
	}
	// Reflect the CA account actually used. Preserve an explicitly configured
	// value; otherwise adopt whatever the backend resolved (may be empty for the
	// internal-CA fallback path).
	if data.OrganizationCAAccountID.IsNull() || data.OrganizationCAAccountID.IsUnknown() {
		if cert.OrganizationCAAccountID != nil && *cert.OrganizationCAAccountID != "" {
			data.OrganizationCAAccountID = types.StringValue(*cert.OrganizationCAAccountID)
		} else {
			data.OrganizationCAAccountID = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.GetCertificate(data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Certificate Error", fmt.Sprintf("Unable to read certificate: %s", err))
		return
	}
	if cert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.CommonName = types.StringValue(cert.CommonName)
	data.SerialNumber = types.StringValue(cert.SerialNumber)
	data.CertificatePEM = types.StringValue(cert.CertificatePEM)
	data.Status = types.StringValue(cert.Status)
	data.ExpiresAt = types.StringValue(cert.ExpiresAt)
	if cert.ProjectID != nil && *cert.ProjectID != "" {
		data.ProjectID = types.StringValue(*cert.ProjectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Certificates are immutable; all fields use RequiresReplace.
}

// Delete revokes the certificate where the issuing CA supports it. Certificates
// issued through an external CA account carry no internal CA link, so the
// backend cannot revoke them; those are archived instead so the record leaves
// the active inventory and destroy completes. Already-removed records are
// treated as gone.
func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	err := r.client.RevokeCertificate(id)
	switch {
	case err == nil, mazevault.IsNotFoundError(err):
		return
	case strings.Contains(err.Error(), "no associated CA"), strings.Contains(err.Error(), "already revoked"):
		if archiveErr := r.client.ArchiveCertificate(id); archiveErr != nil && !mazevault.IsNotFoundError(archiveErr) {
			resp.Diagnostics.AddError("Archive Certificate Error",
				fmt.Sprintf("Certificate cannot be revoked (%s) and archiving failed: %s", err, archiveErr))
			return
		}
		resp.Diagnostics.AddWarning("Certificate archived instead of revoked",
			fmt.Sprintf("Certificate %s was issued through an external CA account without revocation support in MazeVault; the record was archived. Revoke it at the issuing CA if required.", id))
	default:
		resp.Diagnostics.AddError("Revoke Certificate Error", fmt.Sprintf("Unable to revoke certificate: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
