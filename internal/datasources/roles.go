package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RolesDataSource{}

func NewRolesDataSource() datasource.DataSource { return &RolesDataSource{} }

type RolesDataSource struct{ client *mazevault.Client }

type RolesDataModel struct {
	Roles []RoleItem `tfsdk:"roles"`
}

type RoleItem struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	IsSystem    types.Bool   `tfsdk:"is_system_role"`
}

func (d *RolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *RolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all RBAC roles available in MazeVault.",
		Attributes: map[string]schema.Attribute{
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"name":           schema.StringAttribute{Computed: true},
						"display_name":   schema.StringAttribute{Computed: true},
						"description":    schema.StringAttribute{Computed: true},
						"is_system_role": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *RolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RolesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	roles, err := d.client.ListRoles()
	if err != nil {
		resp.Diagnostics.AddError("Read Roles Error", fmt.Sprintf("Unable to list roles: %s", err))
		return
	}
	items := make([]RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, RoleItem{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			DisplayName: types.StringValue(r.DisplayName),
			Description: types.StringValue(r.Description),
			IsSystem:    types.BoolValue(r.IsSystemRole),
		})
	}
	data.Roles = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
