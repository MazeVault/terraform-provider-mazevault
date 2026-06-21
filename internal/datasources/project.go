package datasources

import (
	"context"
	"fmt"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource { return &ProjectDataSource{} }

type ProjectDataSource struct{ client *mazevault.Client }

type ProjectDataModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Environment    types.String `tfsdk:"environment"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing MazeVault project by ID.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, MarkdownDescription: "Project ID."},
			"name":            schema.StringAttribute{Computed: true, MarkdownDescription: "Project name."},
			"type":            schema.StringAttribute{Computed: true, MarkdownDescription: "Project type."},
			"environment":     schema.StringAttribute{Computed: true, MarkdownDescription: "Default environment."},
			"organization_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Owning organization ID."},
			"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp (RFC3339)."},
		},
	}
}

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := d.client.GetProject(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Project Error", fmt.Sprintf("Unable to read project: %s", err))
		return
	}
	data.Name = types.StringValue(project.Name)
	data.Type = types.StringValue(project.Type)
	data.Environment = types.StringValue(project.Environment)
	data.OrganizationID = types.StringValue(project.OrganizationID)
	data.CreatedAt = types.StringValue(project.CreatedAt.Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
