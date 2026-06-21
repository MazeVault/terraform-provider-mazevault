package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}

func NewUserResource() resource.Resource { return &UserResource{} }

type UserResource struct{ client *mazevault.Client }

type UserModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	FullName types.String `tfsdk:"full_name"`
	Role     types.String `tfsdk:"role"`
	Password types.String `tfsdk:"password"`
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MazeVault user account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the user.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email address (login credential).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"full_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Full display name of the user.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Initial role assignment (`admin`, `member`, `viewer`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial password for the user account. Required by the MazeVault API for all users, including those who will later authenticate via SSO.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateUser(&mazevault.CreateUserRequest{
		Email:    data.Email.ValueString(),
		FullName: data.FullName.ValueString(),
		Password: data.Password.ValueString(),
		Role:     data.Role.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create User Error", fmt.Sprintf("Unable to create user: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	users, err := r.client.ListUsers()
	if err != nil {
		resp.Diagnostics.AddError("Read User Error", fmt.Sprintf("Unable to list users: %s", err))
		return
	}
	for _, u := range users {
		if u.ID == data.ID.ValueString() {
			data.Email = types.StringValue(u.Email)
			data.FullName = types.StringValue(u.FullName)
			data.Role = types.StringValue(u.Role)
			// password is write-only; preserve from state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *UserResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// User updates are not exposed via the MazeVault public API. name/isAdmin changes require delete+recreate.
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete User Error", fmt.Sprintf("Unable to delete user: %s", err))
	}
}
