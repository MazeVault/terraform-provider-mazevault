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

var _ datasource.DataSource = &RotationExecutionsDataSource{}

func NewRotationExecutionsDataSource() datasource.DataSource { return &RotationExecutionsDataSource{} }

type RotationExecutionsDataSource struct{ client *mazevault.Client }

type RotationExecutionsDataModel struct {
	Executions []RotationExecutionItem `tfsdk:"executions"`
}

type RotationExecutionItem struct {
	ID        types.String `tfsdk:"id"`
	ConfigID  types.String `tfsdk:"config_id"`
	Status    types.String `tfsdk:"status"`
	StartedAt types.String `tfsdk:"started_at"`
	Error     types.String `tfsdk:"error"`
}

func (d *RotationExecutionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_executions"
}

func (d *RotationExecutionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all rotation execution records.",
		Attributes: map[string]schema.Attribute{
			"executions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"config_id":  schema.StringAttribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
						"started_at": schema.StringAttribute{Computed: true},
						"error":      schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RotationExecutionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RotationExecutionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RotationExecutionsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	execs, err := d.client.ListRotationExecutions()
	if err != nil {
		resp.Diagnostics.AddError("Read Rotation Executions Error", fmt.Sprintf("Unable to list rotation executions: %s", err))
		return
	}
	items := make([]RotationExecutionItem, 0, len(execs))
	for _, e := range execs {
		items = append(items, RotationExecutionItem{
			ID:        types.StringValue(e.ID),
			ConfigID:  types.StringValue(e.ConfigID),
			Status:    types.StringValue(e.Status),
			StartedAt: types.StringValue(e.StartedAt.Format(time.RFC3339)),
			Error:     types.StringValue(e.Error),
		})
	}
	data.Executions = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
