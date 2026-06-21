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

var _ datasource.DataSource = &RenewalQueueDataSource{}

func NewRenewalQueueDataSource() datasource.DataSource { return &RenewalQueueDataSource{} }

type RenewalQueueDataSource struct{ client *mazevault.Client }

type RenewalQueueDataModel struct {
	Items []RenewalQueueItemData `tfsdk:"items"`
}

type RenewalQueueItemData struct {
	ID            types.String `tfsdk:"id"`
	CertificateID types.String `tfsdk:"certificate_id"`
	Status        types.String `tfsdk:"status"`
	RequestedAt   types.String `tfsdk:"requested_at"`
}

func (d *RenewalQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_renewal_queue"
}

func (d *RenewalQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all certificates currently pending renewal in the renewal queue.",
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"certificate_id": schema.StringAttribute{Computed: true},
						"status":         schema.StringAttribute{Computed: true},
						"requested_at":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RenewalQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RenewalQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RenewalQueueDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	queue, err := d.client.GetRenewalQueue()
	if err != nil {
		resp.Diagnostics.AddError("Read Renewal Queue Error", fmt.Sprintf("Unable to read renewal queue: %s", err))
		return
	}
	items := make([]RenewalQueueItemData, 0, len(queue))
	for _, q := range queue {
		items = append(items, RenewalQueueItemData{
			ID:            types.StringValue(q.ID),
			CertificateID: types.StringValue(q.CertificateID),
			Status:        types.StringValue(q.Status),
			RequestedAt:   types.StringValue(q.RequestedAt.Format(time.RFC3339)),
		})
	}
	data.Items = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
