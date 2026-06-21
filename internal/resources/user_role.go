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

var _ resource.Resource = &UserRoleResource{}

func NewUserRoleResource() resource.Resource { return &UserRoleResource{} }

type UserRoleResource struct{ client *mazevault.Client }

type UserRoleModel struct {
	ID          types.String `tfsdk:"id"`
	UserID      types.String `tfsdk:"user_id"`
	RoleID      types.String `tfsdk:"role_id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Environment types.String `tfsdk:"environment"`
}

func (r *UserRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_role"
}

func (r *UserRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns an RBAC role to a user, optionally scoped to a specific resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the role assignment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"user_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the user to assign the role to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"role_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the role to assign.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Project ID to scope the role assignment to. Omit for global assignment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"environment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Environment to scope the role assignment to (e.g. `production`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *UserRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserRoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	assignment, err := r.client.AssignRole(data.UserID.ValueString(), &mazevault.AssignRoleRequest{
		RoleID:      data.RoleID.ValueString(),
		ProjectID:   data.ProjectID.ValueString(),
		Environment: data.Environment.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Assign Role Error", fmt.Sprintf("Unable to assign role: %s", err))
		return
	}
	data.ID = types.StringValue(assignment.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserRoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	roles, err := r.client.GetUserRoles(data.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read User Roles Error", fmt.Sprintf("Unable to read user roles: %s", err))
		return
	}
	for _, ur := range roles {
		if ur.ID == data.ID.ValueString() {
			data.RoleID = types.StringValue(ur.RoleID)
			data.ProjectID = types.StringValue(ur.ProjectID)
			data.Environment = types.StringValue(ur.Environment)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *UserRoleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Role assignments are immutable; changes require destroy and recreate.
}

func (r *UserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserRoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveRoleAssignment(data.UserID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Remove Role Assignment Error", fmt.Sprintf("Unable to remove role assignment: %s", err))
	}
}
