package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectCertificatesDataSource{}

func NewProjectCertificatesDataSource() datasource.DataSource {
	return &ProjectCertificatesDataSource{}
}

type ProjectCertificatesDataSource struct {
	client *mazevault.Client
}

type ProjectCertificatesModel struct {
	ProjectID    types.String       `tfsdk:"project_id"`
	Certificates []CertificateModel `tfsdk:"certificates"`
}

type CertificateModel struct {
	ID           types.String `tfsdk:"id"`
	CommonName   types.String `tfsdk:"common_name"`
	SerialNumber types.String `tfsdk:"serial_number"`
	Status       types.String `tfsdk:"status"`
	ExpiryDate   types.String `tfsdk:"expiry_date"`
}

func (d *ProjectCertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_certificates"
}

func (d *ProjectCertificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all certificates for a MazeVault project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the project."},
			"certificates": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of certificates in the project.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"common_name":   schema.StringAttribute{Computed: true},
						"serial_number": schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"expiry_date":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ProjectCertificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectCertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectCertificatesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certs, err := d.client.ListProjectCertificates(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List Certificates Error",
			fmt.Sprintf("Unable to list project certificates: %s", err))
		return
	}

	var certModels []CertificateModel
	for _, cert := range certs {
		certModels = append(certModels, CertificateModel{
			ID:           types.StringValue(cert.ID),
			CommonName:   types.StringValue(cert.CommonName),
			SerialNumber: types.StringValue(cert.SerialNumber),
			Status:       types.StringValue(cert.Status),
			ExpiryDate:   types.StringValue(cert.ExpiresAt),
		})
	}

	data.Certificates = certModels
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
