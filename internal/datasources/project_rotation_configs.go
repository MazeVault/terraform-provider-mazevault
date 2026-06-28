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

var _ datasource.DataSource = &ProjectRotationConfigsDataSource{}

// NewProjectRotationConfigsDataSource returns a new mazevault_project_rotation_configs data source.
func NewProjectRotationConfigsDataSource() datasource.DataSource {
	return &ProjectRotationConfigsDataSource{}
}

// ProjectRotationConfigsDataSource lists all rotation configs scoped to a project.
type ProjectRotationConfigsDataSource struct{ client *mazevault.Client }

// ProjectRotationConfigsModel is the Terraform state model.
type ProjectRotationConfigsModel struct {
	ProjectID types.String                `tfsdk:"project_id"`
	Configs   []ProjectRotationConfigItem `tfsdk:"configs"`
}

// ProjectRotationConfigItem is one rotation config entry within the project.
type ProjectRotationConfigItem struct {
	ID                   types.String `tfsdk:"id"`
	SecretID             types.String `tfsdk:"secret_id"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Schedule             types.String `tfsdk:"schedule"`
	RotationIntervalDays types.Int64  `tfsdk:"rotation_interval_days"`
	Status               types.String `tfsdk:"status"`
	LastRotatedAt        types.String `tfsdk:"last_rotated_at"`
	NextRotationAt       types.String `tfsdk:"next_rotation_at"`
}

func (d *ProjectRotationConfigsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_rotation_configs"
}

func (d *ProjectRotationConfigsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all rotation configurations associated with a project.\n\n" +
			"Use this data source to enumerate every secret rotation config within a project, " +
			"for example to audit coverage or to pass config IDs to other resources.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the project whose rotation configs should be listed.",
			},
			"configs": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of rotation configs belonging to the project.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                     schema.StringAttribute{Computed: true, Description: "Rotation config ID."},
						"secret_id":              schema.StringAttribute{Computed: true, Description: "Secret this config applies to."},
						"enabled":                schema.BoolAttribute{Computed: true, Description: "Whether automatic rotation is enabled."},
						"schedule":               schema.StringAttribute{Computed: true, Description: "Cron schedule expression."},
						"rotation_interval_days": schema.Int64Attribute{Computed: true, Description: "Rotation interval in days."},
						"status":                 schema.StringAttribute{Computed: true, Description: "Current rotation status."},
						"last_rotated_at":        schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp of the last successful rotation."},
						"next_rotation_at":       schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp of the next scheduled rotation."},
					},
				},
			},
		},
	}
}

func (d *ProjectRotationConfigsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectRotationConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProjectRotationConfigsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListProjectRotationConfigs(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Project Rotation Configs Error",
			fmt.Sprintf("Unable to list rotation configs for project %s: %s", data.ProjectID.ValueString(), err))
		return
	}

	items := make([]ProjectRotationConfigItem, 0, len(result))
	for _, c := range result {
		item := ProjectRotationConfigItem{
			ID:                   types.StringValue(c.ID),
			SecretID:             types.StringValue(c.SecretID),
			Enabled:              types.BoolValue(c.Enabled),
			Schedule:             types.StringValue(c.Schedule),
			RotationIntervalDays: types.Int64Value(int64(c.RotationIntervalDays)),
			Status:               types.StringValue(c.Status),
			LastRotatedAt:        types.StringValue(""),
			NextRotationAt:       types.StringValue(""),
		}
		if c.LastRotatedAt != nil {
			item.LastRotatedAt = types.StringValue(c.LastRotatedAt.Format(time.RFC3339))
		}
		if c.NextRotationAt != nil {
			item.NextRotationAt = types.StringValue(c.NextRotationAt.Format(time.RFC3339))
		}
		items = append(items, item)
	}

	data.Configs = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
