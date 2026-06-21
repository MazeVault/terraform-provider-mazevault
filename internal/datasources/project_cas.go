package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectCAsDataSource{}

func NewProjectCAsDataSource() datasource.DataSource {
	return &ProjectCAsDataSource{}
}

type ProjectCAsDataSource struct {
	client *mazevault.Client
}

type ProjectCAsModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	CAs       []CAModel    `tfsdk:"cas"`
}

type CAModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Status     types.String `tfsdk:"status"`
	ValidUntil types.String `tfsdk:"valid_until"`
}

func (d *ProjectCAsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_cas"
}

func (d *ProjectCAsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all internal Certificate Authorities (CAs) for a MazeVault project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the project."},
			"cas": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of CAs in the project.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"type":        schema.StringAttribute{Computed: true},
						"status":      schema.StringAttribute{Computed: true},
						"valid_until": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ProjectCAsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectCAsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectCAsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cas, err := d.client.ListProjectCAs(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List Project CAs Error", fmt.Sprintf("Unable to list project CAs: %s", err))
		return
	}

	var caModels []CAModel
	for _, ca := range cas {
		caModels = append(caModels, CAModel{
			ID:         types.StringValue(ca.ID),
			Name:       types.StringValue(ca.Name),
			Type:       types.StringValue(ca.Type),
			Status:     types.StringValue(ca.Status),
			ValidUntil: types.StringValue(ca.ValidUntil),
		})
	}

	data.CAs = caModels
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
