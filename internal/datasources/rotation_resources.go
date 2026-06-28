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

var _ datasource.DataSource = &RotationResourcesDataSource{}

// NewRotationResourcesDataSource returns a new mazevault_rotation_resources data source.
func NewRotationResourcesDataSource() datasource.DataSource {
	return &RotationResourcesDataSource{}
}

// RotationResourcesDataSource lists all rotation resources registered in MazeVault,
// optionally filtered by kind and/or environment scope.
type RotationResourcesDataSource struct{ client *mazevault.Client }

// RotationResourcesModel is the Terraform state model for mazevault_rotation_resources.
type RotationResourcesModel struct {
	Kind             types.String           `tfsdk:"kind"`
	EnvironmentScope types.String           `tfsdk:"environment_scope"`
	Resources        []RotationResourceItem `tfsdk:"resources"`
}

// RotationResourceItem represents a single rotation resource entry in the list.
type RotationResourceItem struct {
	ID               types.String `tfsdk:"id"`
	ResourceKind     types.String `tfsdk:"resource_kind"`
	ResourceID       types.String `tfsdk:"resource_id"`
	ProjectID        types.String `tfsdk:"project_id"`
	EnvironmentScope types.String `tfsdk:"environment_scope"`
	DisplayName      types.String `tfsdk:"display_name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ManualOnly       types.Bool   `tfsdk:"manual_only"`
	StatusSummary    types.String `tfsdk:"status_summary"`
	NextDueAt        types.String `tfsdk:"next_due_at"`
}

func (d *RotationResourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_resources"
}

func (d *RotationResourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all rotation resources registered in MazeVault.\n\n" +
			"Use `kind` and/or `environment_scope` to filter the results.",
		Attributes: map[string]schema.Attribute{
			"kind": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by resource kind.  Supported values: `secret`, `certificate`, `entra_credential`, `ssh_key`.",
			},
			"environment_scope": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by environment scope (e.g. `staging`, `production`).",
			},
			"resources": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of rotation resources matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true, Description: "Rotation resource record ID."},
						"resource_kind":     schema.StringAttribute{Computed: true, Description: "Kind of the underlying resource."},
						"resource_id":       schema.StringAttribute{Computed: true, Description: "UUID of the underlying resource."},
						"project_id":        schema.StringAttribute{Computed: true, Description: "Project this resource belongs to."},
						"environment_scope": schema.StringAttribute{Computed: true, Description: "Environment scope of the resource."},
						"display_name":      schema.StringAttribute{Computed: true, Description: "Human-readable name."},
						"enabled":           schema.BoolAttribute{Computed: true, Description: "Whether automatic rotation is enabled."},
						"manual_only":       schema.BoolAttribute{Computed: true, Description: "Whether the resource only supports manual rotation."},
						"status_summary":    schema.StringAttribute{Computed: true, Description: "Short status string (e.g. `ok`, `overdue`, `error`)."},
						"next_due_at":       schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp of the next scheduled rotation."},
					},
				},
			},
		},
	}
}

func (d *RotationResourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RotationResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RotationResourcesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListRotationResources(data.Kind.ValueString(), data.EnvironmentScope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Rotation Resources Error",
			fmt.Sprintf("Unable to list rotation resources: %s", err))
		return
	}

	items := make([]RotationResourceItem, 0, len(result))
	for _, r := range result {
		item := RotationResourceItem{
			ID:               types.StringValue(r.ID),
			ResourceKind:     types.StringValue(r.ResourceKind),
			ResourceID:       types.StringValue(r.ResourceID),
			ProjectID:        types.StringValue(r.ProjectID),
			EnvironmentScope: types.StringValue(r.EnvironmentScope),
			DisplayName:      types.StringValue(r.DisplayName),
			Enabled:          types.BoolValue(r.Enabled),
			ManualOnly:       types.BoolValue(r.ManualOnly),
			StatusSummary:    types.StringValue(r.StatusSummary),
		}
		if r.NextDueAt != nil {
			item.NextDueAt = types.StringValue(r.NextDueAt.Format(time.RFC3339))
		} else {
			item.NextDueAt = types.StringValue("")
		}
		items = append(items, item)
	}

	data.Resources = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
