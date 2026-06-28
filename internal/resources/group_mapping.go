package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupMappingResource{}
var _ resource.ResourceWithImportState = &GroupMappingResource{}

func NewGroupMappingResource() resource.Resource {
	return &GroupMappingResource{}
}

type GroupMappingResource struct {
	client *mazevault.Client
}

type GroupMappingResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	RoleID    types.String `tfsdk:"role_id"`
}

func (r *GroupMappingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_mapping"
}

func (r *GroupMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Maps an identity provider group to a MazeVault RBAC role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the group mapping.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"group_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Identity provider group name (e.g. Azure AD group name or LDAP CN).",
			},
			"role_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the MazeVault RBAC role to assign to the group.",
			},
		},
	}
}

func (r *GroupMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := r.client.CreateGroupMapping(&mazevault.CreateGroupMappingRequest{
		GroupName: data.GroupName.ValueString(),
		RoleID:    data.RoleID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Group Mapping Error", fmt.Sprintf("Unable to create group mapping: %s", err))
		return
	}

	data.ID = types.StringValue(mapping.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mapping, err := r.client.GetGroupMapping(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Group Mapping Error", fmt.Sprintf("Unable to read group mapping: %s", err))
		return
	}
	if mapping == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.GroupName = types.StringValue(mapping.GroupName)
	data.RoleID = types.StringValue(mapping.RoleID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMappingResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Group mappings are immutable; changes require destroy and recreate.
}

func (r *GroupMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGroupMapping(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Group Mapping Error", fmt.Sprintf("Unable to delete group mapping: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *GroupMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
