package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CertificateDataSource{}

func NewCertificateDataSource() datasource.DataSource { return &CertificateDataSource{} }

type CertificateDataSource struct{ client *mazevault.Client }

type CertificateDataModel struct {
	ID             types.String `tfsdk:"id"`
	CommonName     types.String `tfsdk:"common_name"`
	SerialNumber   types.String `tfsdk:"serial_number"`
	CertificatePEM types.String `tfsdk:"certificate_pem"`
	ChainPEM       types.String `tfsdk:"chain_pem"`
	ChainSource    types.String `tfsdk:"chain_source"`
	ChainComplete  types.Bool   `tfsdk:"chain_complete"`
	Status         types.String `tfsdk:"status"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	Type           types.String `tfsdk:"certificate_type"`
	Thumbprint     types.String `tfsdk:"thumbprint_sha256"`
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a certificate from MazeVault by ID. The certificate inventory also holds imported OpenPGP keys (`certificate_type = \"pgp\"`); for such a record `certificate_pem` is the ASCII-armored public key, `serial_number` the key ID and `common_name` the primary user ID.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Required: true, MarkdownDescription: "Certificate ID."},
			"common_name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate Common Name (primary user ID of an OpenPGP key)."},
			"serial_number":     schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate serial number (key ID of an OpenPGP key). Unique per issuer only — use `thumbprint_sha256` to identify a certificate."},
			"certificate_pem":   schema.StringAttribute{Computed: true, MarkdownDescription: "PEM bundle: the certificate followed by its effective issuer chain (the same chain MazeVault exports and deploys)."},
			"chain_pem":         schema.StringAttribute{Computed: true, MarkdownDescription: "PEM issuer chain only, nearest issuer first and root last. Empty when no issuer is known."},
			"chain_source":      schema.StringAttribute{Computed: true, MarkdownDescription: "Where the issuer chain comes from: `stored` (captured at import/issuance), `inventory` (resolved from verified CA certificates in the organization) or `none`."},
			"chain_complete":    schema.BoolAttribute{Computed: true, MarkdownDescription: "True when the issuer chain ends at a self-signed root certificate."},
			"status":            schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate status (`active`, `revoked`, `expired`)."},
			"expires_at":        schema.StringAttribute{Computed: true, MarkdownDescription: "Expiry date (ISO-8601). An OpenPGP key that never expires reports `9999-12-31T00:00:00Z`."},
			"certificate_type":  schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate subtype (for example `ssl_tls`, `client_auth`), or `pgp` for an imported OpenPGP key."},
			"thumbprint_sha256": schema.StringAttribute{Computed: true, MarkdownDescription: "SHA-256 fingerprint of the certificate in lowercase hex — the identity MazeVault de-duplicates on. For an OpenPGP key: SHA-256 of its primary public key."},
		},
	}
}

func (d *CertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificateDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cert, err := d.client.GetCertificate(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Certificate Error", fmt.Sprintf("Unable to read certificate: %s", err))
		return
	}
	data.CommonName = types.StringValue(cert.CommonName)
	data.SerialNumber = types.StringValue(cert.SerialNumber)
	data.CertificatePEM = types.StringValue(cert.CertificatePEM)
	data.ChainPEM = types.StringValue(cert.ChainPEM)
	data.ChainSource = types.StringValue(cert.ChainSource)
	data.ChainComplete = types.BoolValue(cert.ChainComplete)
	data.Status = types.StringValue(cert.Status)
	data.ExpiresAt = types.StringValue(cert.ExpiresAt)
	data.Type = types.StringValue(cert.CertificateType)
	data.Thumbprint = types.StringValue(cert.ThumbprintSHA256)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
