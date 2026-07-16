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

var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}

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
	// When a CSR is supplied use it as-is; otherwise ask the backend to generate
	// the key pair and CSR server-side so a certificate can be issued from
	// Terraform without an external key/CSR toolchain.
	if !data.CSRPEM.IsNull() && data.CSRPEM.ValueString() != "" {
		csr.CSRPEM = data.CSRPEM.ValueString()
		csr.CSRSourceMode = "uploaded_csr"
	} else {
		csr.CSRSourceMode = "mazevault_generated"
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

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RevokeCertificate(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Revoke Certificate Error", fmt.Sprintf("Unable to revoke certificate: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
