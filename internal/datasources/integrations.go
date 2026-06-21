package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IntegrationsDataSource{}

func NewIntegrationsDataSource() datasource.DataSource { return &IntegrationsDataSource{} }

type IntegrationsDataSource struct{ client *mazevault.Client }

type IntegrationsDataModel struct {
	ProjectID    types.String      `tfsdk:"project_id"`
	Integrations []IntegrationItem `tfsdk:"integrations"`
}

type IntegrationItem struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Provider types.String `tfsdk:"provider"`
}

func (d *IntegrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

func (d *IntegrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all integrations configured for a MazeVault project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true, MarkdownDescription: "Project ID."},
			"integrations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":       schema.StringAttribute{Computed: true},
						"name":     schema.StringAttribute{Computed: true},
						"type":     schema.StringAttribute{Computed: true},
						"provider": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *IntegrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IntegrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IntegrationsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	integrations, err := d.client.ListIntegrations(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Integrations Error", fmt.Sprintf("Unable to list integrations: %s", err))
		return
	}
	items := make([]IntegrationItem, 0, len(integrations))
	for _, i := range integrations {
		items = append(items, IntegrationItem{
			ID:       types.StringValue(i.ID),
			Name:     types.StringValue(i.Name),
			Type:     types.StringValue(i.Type),
			Provider: types.StringValue(i.Provider),
		})
	}
	data.Integrations = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
