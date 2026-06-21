package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectCertificateTemplatesDataSource{}

func NewProjectCertificateTemplatesDataSource() datasource.DataSource {
	return &ProjectCertificateTemplatesDataSource{}
}

type ProjectCertificateTemplatesDataSource struct {
	client *mazevault.Client
}

type ProjectCertificateTemplatesModel struct {
	ProjectID types.String               `tfsdk:"project_id"`
	Templates []CertificateTemplateModel `tfsdk:"templates"`
}

type CertificateTemplateModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	ValidityPeriod types.String `tfsdk:"validity_period"`
	KeyUsage       types.List   `tfsdk:"key_usage"`
}

func (d *ProjectCertificateTemplatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_certificate_templates"
}

func (d *ProjectCertificateTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all certificate templates for a MazeVault project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the project."},
			"templates": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of certificate templates.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"type":            schema.StringAttribute{Computed: true},
						"validity_period": schema.StringAttribute{Computed: true},
						"key_usage":       schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *ProjectCertificateTemplatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectCertificateTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectCertificateTemplatesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpls, err := d.client.ListProjectCertificateTemplates(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List Certificate Templates Error",
			fmt.Sprintf("Unable to list project certificate templates: %s", err))
		return
	}

	var tmplModels []CertificateTemplateModel
	for _, tmpl := range tmpls {
		keyUsageList, diags := types.ListValueFrom(ctx, types.StringType, tmpl.KeyUsage)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		tmplModels = append(tmplModels, CertificateTemplateModel{
			ID:             types.StringValue(tmpl.ID),
			Name:           types.StringValue(tmpl.Name),
			Type:           types.StringValue(tmpl.Type),
			ValidityPeriod: types.StringValue(tmpl.ValidityPeriod),
			KeyUsage:       keyUsageList,
		})
	}

	data.Templates = tmplModels
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
