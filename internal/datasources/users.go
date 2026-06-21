package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource { return &UsersDataSource{} }

type UsersDataSource struct{ client *mazevault.Client }

type UsersDataModel struct {
	Users []UserItem `tfsdk:"users"`
}

type UserItem struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	FullName types.String `tfsdk:"full_name"`
	Role     types.String `tfsdk:"role"`
	Status   types.String `tfsdk:"status"`
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all users in the MazeVault organization.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.StringAttribute{Computed: true},
						"email":     schema.StringAttribute{Computed: true},
						"full_name": schema.StringAttribute{Computed: true},
						"role":      schema.StringAttribute{Computed: true},
						"status":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	users, err := d.client.ListUsers()
	if err != nil {
		resp.Diagnostics.AddError("Read Users Error", fmt.Sprintf("Unable to list users: %s", err))
		return
	}
	items := make([]UserItem, 0, len(users))
	for _, u := range users {
		items = append(items, UserItem{
			ID:       types.StringValue(u.ID),
			Email:    types.StringValue(u.Email),
			FullName: types.StringValue(u.FullName),
			Role:     types.StringValue(u.Role),
			Status:   types.StringValue(u.Status),
		})
	}
	data.Users = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
