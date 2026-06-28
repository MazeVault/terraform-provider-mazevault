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

var _ datasource.DataSource = &RotationResourceHistoryDataSource{}

// NewRotationResourceHistoryDataSource returns a new mazevault_rotation_resource_history data source.
func NewRotationResourceHistoryDataSource() datasource.DataSource {
	return &RotationResourceHistoryDataSource{}
}

// RotationResourceHistoryDataSource lists the rotation execution history for a single resource.
type RotationResourceHistoryDataSource struct{ client *mazevault.Client }

// RotationResourceHistoryModel is the Terraform state model.
type RotationResourceHistoryModel struct {
	Kind       types.String                  `tfsdk:"kind"`
	ResourceID types.String                  `tfsdk:"resource_id"`
	Executions []RotationResourceHistoryItem `tfsdk:"executions"`
}

// RotationResourceHistoryItem represents one execution run.
type RotationResourceHistoryItem struct {
	ID          types.String `tfsdk:"id"`
	ConfigID    types.String `tfsdk:"config_id"`
	Status      types.String `tfsdk:"status"`
	StartedAt   types.String `tfsdk:"started_at"`
	CompletedAt types.String `tfsdk:"completed_at"`
	Error       types.String `tfsdk:"error"`
}

func (d *RotationResourceHistoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_resource_history"
}

func (d *RotationResourceHistoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the rotation execution history for a specific resource.\n\n" +
			"Use this data source to audit past rotation attempts and diagnose failures.",
		Attributes: map[string]schema.Attribute{
			"kind": schema.StringAttribute{
				Required:    true,
				Description: "Resource kind.  Supported values: `secret`, `certificate`, `entra_credential`, `ssh_key`.",
			},
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the resource whose rotation history should be retrieved.",
			},
			"executions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Ordered list of rotation executions (most-recent first).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "Execution ID."},
						"config_id":    schema.StringAttribute{Computed: true, Description: "Rotation config that triggered this execution."},
						"status":       schema.StringAttribute{Computed: true, Description: "Status: `pending`, `running`, `success`, `failed`, `skipped`."},
						"started_at":   schema.StringAttribute{Computed: true, Description: "RFC 3339 start timestamp."},
						"completed_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 completion timestamp.  Empty if still running."},
						"error":        schema.StringAttribute{Computed: true, Description: "Error message if the execution failed.  Empty on success."},
					},
				},
			},
		},
	}
}

func (d *RotationResourceHistoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RotationResourceHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RotationResourceHistoryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetRotationResourceHistory(data.Kind.ValueString(), data.ResourceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Rotation Resource History Error",
			fmt.Sprintf("Unable to fetch rotation history for %s/%s: %s",
				data.Kind.ValueString(), data.ResourceID.ValueString(), err))
		return
	}

	items := make([]RotationResourceHistoryItem, 0, len(result))
	for _, e := range result {
		item := RotationResourceHistoryItem{
			ID:       types.StringValue(e.ID),
			ConfigID: types.StringValue(e.ConfigID),
			Status:   types.StringValue(e.Status),
			Error:    types.StringValue(e.Error),
		}
		if e.StartedAt != nil {
			item.StartedAt = types.StringValue(e.StartedAt.Format(time.RFC3339))
		} else {
			item.StartedAt = types.StringValue("")
		}
		if e.CompletedAt != nil {
			item.CompletedAt = types.StringValue(e.CompletedAt.Format(time.RFC3339))
		} else {
			item.CompletedAt = types.StringValue("")
		}
		items = append(items, item)
	}

	data.Executions = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
