package datasources

import (
	"context"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TokensDataSource{}

func NewTokensDataSource() datasource.DataSource { return &TokensDataSource{} }

type TokensDataSource struct{ client *mazevault.Client }

type TokensDataSourceModel struct {
	Tokens []TokenItem `tfsdk:"tokens"`
}

type TokenItem struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Provider        types.String `tfsdk:"provider"`
	Status          types.String `tfsdk:"status"`
	StorageMode     types.String `tfsdk:"storage_mode"`
	ExpiresAt       types.String `tfsdk:"expires_at"`
	LeadTimeDays    types.Int64  `tfsdk:"lead_time_days"`
	IsExpired       types.Bool   `tfsdk:"is_expired"`
	DaysUntilExpiry types.Int64  `tfsdk:"days_until_expiry"`
}

func (d *TokensDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tokens"
}

func (d *TokensDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists managed external API tokens for the organization, including their expiry status.",
		Attributes: map[string]schema.Attribute{
			"tokens": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of managed tokens (values are never returned).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true},
						"name":              schema.StringAttribute{Computed: true},
						"provider":          schema.StringAttribute{Computed: true},
						"status":            schema.StringAttribute{Computed: true},
						"storage_mode":      schema.StringAttribute{Computed: true},
						"expires_at":        schema.StringAttribute{Computed: true},
						"lead_time_days":    schema.Int64Attribute{Computed: true},
						"is_expired":        schema.BoolAttribute{Computed: true},
						"days_until_expiry": schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *TokensDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *mazevault.Client")
		return
	}
	d.client = client
}

func (d *TokensDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TokensDataSourceModel

	result, err := d.client.ListManagedTokens()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list tokens", err.Error())
		return
	}

	data.Tokens = make([]TokenItem, 0, len(result))
	for _, t := range result {
		data.Tokens = append(data.Tokens, TokenItem{
			ID:              types.StringValue(t.ID),
			Name:            types.StringValue(t.Name),
			Provider:        types.StringValue(t.Provider),
			Status:          types.StringValue(t.Status),
			StorageMode:     types.StringValue(t.StorageMode),
			ExpiresAt:       types.StringValue(t.ExpiresAt.Format(time.RFC3339)),
			LeadTimeDays:    types.Int64Value(int64(t.LeadTimeDays)),
			IsExpired:       types.BoolValue(t.IsExpired),
			DaysUntilExpiry: types.Int64Value(int64(t.DaysUntilExpiry)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
