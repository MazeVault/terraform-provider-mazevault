package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RoleResource{}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	client *mazevault.Client
}

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MazeVault RBAC role with a set of permissions. Note: roles cannot be deleted via the API; removing this resource removes it from Terraform state only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the role.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Machine-readable name (e.g. `read-only`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable display name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description of the role's purpose.",
			},
			"permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of permission strings granted by this role.",
			},
		},
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	resp.Diagnostics.Append(data.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.CreateRole(&mazevault.CreateRoleRequest{
		Name:        data.Name.ValueString(),
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		Permissions: permissions,
	})
	if err != nil {
		// Only attempt state-adoption if the backend signals a conflict (409).
		// Any other error (5xx, auth failure, network) is a real failure and must
		// propagate so the operator is not silently surprised.
		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "already exists") {
			all, listErr := r.client.ListRoles()
			if listErr != nil {
				resp.Diagnostics.AddError("Create Role Error", fmt.Sprintf("Unable to create role: %s", err))
				return
			}
			for _, existing := range all {
				if existing.Name == data.Name.ValueString() {
					data.ID = types.StringValue(existing.ID)
					resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
					return
				}
			}
		}
		resp.Diagnostics.AddError("Create Role Error", fmt.Sprintf("Unable to create role: %s", err))
		return
	}

	data.ID = types.StringValue(role.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The MazeVault API exposes only GET /rbac/roles (list); there is no
	// GET /rbac/roles/:id endpoint. We list all roles and find by ID.
	roles, err := r.client.ListRoles()
	if err != nil {
		resp.Diagnostics.AddError("Read Role Error", fmt.Sprintf("Unable to list roles: %s", err))
		return
	}
	var found *mazevault.Role
	for i := range roles {
		if roles[i].ID == data.ID.ValueString() {
			found = &roles[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(found.Name)
	data.DisplayName = types.StringValue(found.DisplayName)
	data.Description = types.StringValue(found.Description)

	permsList, diags := types.ListValueFrom(ctx, types.StringType, found.Permissions)
	resp.Diagnostics.Append(diags...)
	data.Permissions = permsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RoleResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Role updates are not supported by the MazeVault API; the `name` field uses RequiresReplace.
}

func (r *RoleResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// The MazeVault API does not expose a role-delete endpoint.
	// Removing from Terraform state only; the role remains in MazeVault.
}
