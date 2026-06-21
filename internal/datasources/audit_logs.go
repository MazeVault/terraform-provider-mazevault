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

var _ datasource.DataSource = &AuditLogsDataSource{}

func NewAuditLogsDataSource() datasource.DataSource { return &AuditLogsDataSource{} }

type AuditLogsDataSource struct{ client *mazevault.Client }

type AuditLogsDataModel struct {
	ProjectID types.String   `tfsdk:"project_id"`
	Limit     types.Int64    `tfsdk:"limit"`
	Offset    types.Int64    `tfsdk:"offset"`
	Logs      []AuditLogItem `tfsdk:"logs"`
}

type AuditLogItem struct {
	ID         types.String `tfsdk:"id"`
	UserID     types.String `tfsdk:"user_id"`
	Action     types.String `tfsdk:"action"`
	EntityType types.String `tfsdk:"entity_type"`
	EntityID   types.String `tfsdk:"entity_id"`
	Severity   types.String `tfsdk:"severity"`
	IPAddress  types.String `tfsdk:"ip_address"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (d *AuditLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_logs"
}

func (d *AuditLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads audit log entries. Optionally scoped to a specific project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Scope to a specific project."},
			"limit":      schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of entries to return (default 100)."},
			"offset":     schema.Int64Attribute{Optional: true, MarkdownDescription: "Pagination offset."},
			"logs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"user_id":     schema.StringAttribute{Computed: true},
						"action":      schema.StringAttribute{Computed: true},
						"entity_type": schema.StringAttribute{Computed: true},
						"entity_id":   schema.StringAttribute{Computed: true},
						"severity":    schema.StringAttribute{Computed: true},
						"ip_address":  schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *AuditLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuditLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AuditLogsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	limit := 100
	if !data.Limit.IsNull() {
		limit = int(data.Limit.ValueInt64())
	}
	offset := 0
	if !data.Offset.IsNull() {
		offset = int(data.Offset.ValueInt64())
	}
	var result *mazevault.ListAuditLogsResponse
	var err error
	if !data.ProjectID.IsNull() && data.ProjectID.ValueString() != "" {
		result, err = d.client.ListProjectAuditLogs(data.ProjectID.ValueString(), limit, offset)
	} else {
		result, err = d.client.ListAuditLogs(limit, offset)
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Audit Logs Error", fmt.Sprintf("Unable to read audit logs: %s", err))
		return
	}
	items := make([]AuditLogItem, 0, len(result.Logs))
	for _, l := range result.Logs {
		items = append(items, AuditLogItem{
			ID:         types.StringValue(l.ID),
			UserID:     types.StringValue(l.UserID),
			Action:     types.StringValue(l.Action),
			EntityType: types.StringValue(l.EntityType),
			EntityID:   types.StringValue(l.EntityID),
			Severity:   types.StringValue(l.Severity),
			IPAddress:  types.StringValue(l.IPAddress),
			CreatedAt:  types.StringValue(l.CreatedAt.Format(time.RFC3339)),
		})
	}
	data.Logs = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
