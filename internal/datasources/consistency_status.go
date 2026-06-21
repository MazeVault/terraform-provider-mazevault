package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ConsistencyStatusDataSource{}

type ConsistencyStatusDataSource struct {
	client *mazevault.Client
}

type ConsistencyStatusDataSourceModel struct {
	ProjectID    types.String `tfsdk:"project_id"`
	Status       types.String `tfsdk:"status"`
	MissingCount types.Int64  `tfsdk:"missing_count"`
	Issues       types.List   `tfsdk:"issues"`
}

func NewConsistencyStatusDataSource() datasource.DataSource {
	return &ConsistencyStatusDataSource{}
}

func (d *ConsistencyStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consistency_status"
}

func (d *ConsistencyStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to check consistency status of a project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project to check.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Overall consistency status (healthy, warning, error).",
			},
			"missing_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of missing secrets.",
			},
			"issues": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of consistency issues found.",
			},
		},
	}
}

func (d *ConsistencyStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *ConsistencyStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConsistencyStatusDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.GetConsistencyStatus(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read consistency status for project %s: %s", data.ProjectID.ValueString(), err))
		return
	}

	data.Status = types.StringValue(status.Status)
	data.MissingCount = types.Int64Value(int64(status.MissingCount))

	issueList, diags := types.ListValueFrom(ctx, types.StringType, status.Issues)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.Issues = issueList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
