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

var _ datasource.DataSource = &CAAccountsDataSource{}

func NewCAAccountsDataSource() datasource.DataSource { return &CAAccountsDataSource{} }

type CAAccountsDataSource struct{ client *mazevault.Client }

type CAAccountsDataModel struct {
	OrganizationID types.String    `tfsdk:"organization_id"`
	CAAccounts     []CAAccountItem `tfsdk:"ca_accounts"`
}

type CAAccountItem struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderType types.String `tfsdk:"provider_type"`
	Status       types.String `tfsdk:"status"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (d *CAAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_accounts"
}

func (d *CAAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all CA accounts connected to a MazeVault organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{Required: true, MarkdownDescription: "Organization ID."},
			"ca_accounts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"provider_type": schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"created_at":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *CAAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CAAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CAAccountsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	accounts, err := d.client.ListCAAccounts(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read CA Accounts Error", fmt.Sprintf("Unable to list CA accounts: %s", err))
		return
	}
	items := make([]CAAccountItem, 0, len(accounts))
	for _, a := range accounts {
		items = append(items, CAAccountItem{
			ID:           types.StringValue(a.ID),
			Name:         types.StringValue(a.Name),
			ProviderType: types.StringValue(a.ProviderType),
			Status:       types.StringValue(a.Status),
			CreatedAt:    types.StringValue(a.CreatedAt.Format(time.RFC3339)),
		})
	}
	data.CAAccounts = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
