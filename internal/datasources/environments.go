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

var _ datasource.DataSource = &EnvironmentsDataSource{}

func NewEnvironmentsDataSource() datasource.DataSource { return &EnvironmentsDataSource{} }

type EnvironmentsDataSource struct{ client *mazevault.Client }

type EnvironmentsDataModel struct {
	OrganizationID types.String      `tfsdk:"organization_id"`
	Environments   []EnvironmentItem `tfsdk:"environments"`
}

type EnvironmentItem struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	IsProduction types.Bool   `tfsdk:"is_production"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (d *EnvironmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *EnvironmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all environments in a MazeVault organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{Required: true, MarkdownDescription: "Organization ID."},
			"environments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"slug":          schema.StringAttribute{Computed: true},
						"is_production": schema.BoolAttribute{Computed: true},
						"created_at":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EnvironmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envs, err := d.client.ListEnvironments(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("List Environments Error", fmt.Sprintf("Unable to list environments: %s", err))
		return
	}
	items := make([]EnvironmentItem, 0, len(envs))
	for _, e := range envs {
		items = append(items, EnvironmentItem{
			ID:           types.StringValue(e.ID),
			Name:         types.StringValue(e.Name),
			Slug:         types.StringValue(e.Slug),
			IsProduction: types.BoolValue(e.IsProduction),
			CreatedAt:    types.StringValue(e.CreatedAt.Format(time.RFC3339)),
		})
	}
	data.Environments = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
