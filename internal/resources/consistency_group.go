package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConsistencyGroupResource{}
var _ resource.ResourceWithImportState = &ConsistencyGroupResource{}

type ConsistencyGroupResource struct {
	client *mazevault.Client
}

type ConsistencyGroupResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	SecretKeys   types.List   `tfsdk:"secret_keys"`
	Environments types.List   `tfsdk:"environments"`
}

func NewConsistencyGroupResource() resource.Resource {
	return &ConsistencyGroupResource{}
}

func (r *ConsistencyGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_consistency_group"
}

func (r *ConsistencyGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a consistency group for validating secrets across environments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the consistency group.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project this group belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the consistency group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the consistency group.",
			},
			"secret_keys": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of secret keys to monitor.",
			},
			"environments": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of environments to check consistency across.",
			},
		},
	}
}

func (r *ConsistencyGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ConsistencyGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConsistencyGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var secretKeys, environments []string
	resp.Diagnostics.Append(data.SecretKeys.ElementsAs(ctx, &secretKeys, false)...)
	resp.Diagnostics.Append(data.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateConsistencyGroupRequest{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		SecretKeys:   secretKeys,
		Environments: environments,
	}

	group, err := r.client.CreateConsistencyGroup(data.ProjectID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create consistency group: %s", err))
		return
	}

	data.ID = types.StringValue(group.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsistencyGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConsistencyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetConsistencyGroup(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read consistency group %s: %s", data.ID.ValueString(), err))
		return
	}
	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	data.ProjectID = types.StringValue(group.ProjectID)

	secretKeysList, diags := types.ListValueFrom(ctx, types.StringType, group.SecretKeys)
	resp.Diagnostics.Append(diags...)
	environmentsList, diags := types.ListValueFrom(ctx, types.StringType, group.Environments)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.SecretKeys = secretKeysList
		data.Environments = environmentsList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsistencyGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConsistencyGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var secretKeys, environments []string
	resp.Diagnostics.Append(data.SecretKeys.ElementsAs(ctx, &secretKeys, false)...)
	resp.Diagnostics.Append(data.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.CreateConsistencyGroupRequest{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		SecretKeys:   secretKeys,
		Environments: environments,
	}

	group, err := r.client.UpdateConsistencyGroup(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update consistency group %s: %s", data.ID.ValueString(), err))
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConsistencyGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConsistencyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteConsistencyGroup(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete consistency group %s: %s", data.ID.ValueString(), err))
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ConsistencyGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
