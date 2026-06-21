package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectCSRsDataSource{}

func NewProjectCSRsDataSource() datasource.DataSource {
	return &ProjectCSRsDataSource{}
}

type ProjectCSRsDataSource struct {
	client *mazevault.Client
}

type ProjectCSRsModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	CSRs      []CSRModel   `tfsdk:"csrs"`
}

type CSRModel struct {
	ID           types.String `tfsdk:"id"`
	TemplateID   types.String `tfsdk:"template_id"`
	CSRPEM       types.String `tfsdk:"csr_pem"`
	RequestedCN  types.String `tfsdk:"requested_cn"`
	Status       types.String `tfsdk:"status"`
	IssuedCertID types.String `tfsdk:"issued_cert_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (d *ProjectCSRsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_csrs"
}

func (d *ProjectCSRsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all certificate signing requests (CSRs) for a MazeVault project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the project."},
			"csrs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"template_id":    schema.StringAttribute{Computed: true},
						"csr_pem":        schema.StringAttribute{Computed: true},
						"requested_cn":   schema.StringAttribute{Computed: true},
						"status":         schema.StringAttribute{Computed: true},
						"issued_cert_id": schema.StringAttribute{Computed: true},
						"created_at":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ProjectCSRsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectCSRsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectCSRsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	csrs, err := d.client.ListProjectCSRs(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List CSRs Error",
			fmt.Sprintf("Unable to list project CSRs: %s", err))
		return
	}

	var csrModels []CSRModel
	for _, csr := range csrs {
		templateID := types.StringNull()
		if csr.TemplateID != nil {
			templateID = types.StringValue(*csr.TemplateID)
		}
		issuedCertID := types.StringNull()
		if csr.IssuedCertID != nil {
			issuedCertID = types.StringValue(*csr.IssuedCertID)
		}
		csrModels = append(csrModels, CSRModel{
			ID:           types.StringValue(csr.ID),
			TemplateID:   templateID,
			CSRPEM:       types.StringValue(csr.CSRPEM),
			RequestedCN:  types.StringValue(csr.RequestedCN),
			Status:       types.StringValue(csr.Status),
			IssuedCertID: issuedCertID,
			CreatedAt:    types.StringValue(csr.CreatedAt),
		})
	}

	data.CSRs = csrModels
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
