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
	Status         types.String `tfsdk:"status"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a certificate from MazeVault by ID.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, MarkdownDescription: "Certificate ID."},
			"common_name":     schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate Common Name."},
			"serial_number":   schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate serial number."},
			"certificate_pem": schema.StringAttribute{Computed: true, MarkdownDescription: "PEM-encoded certificate chain."},
			"status":          schema.StringAttribute{Computed: true, MarkdownDescription: "Certificate status (`active`, `revoked`, `expired`)."},
			"expires_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Expiry date (ISO-8601)."},
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
	data.Status = types.StringValue(cert.Status)
	data.ExpiresAt = types.StringValue(cert.ExpiresAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
