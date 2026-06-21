package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IntegrationGroupResource{}

type IntegrationGroupResource struct {
	client *mazevault.Client
}

type IntegrationGroupResourceModel struct {
	ID          types.String              `tfsdk:"id"`
	ProjectID   types.String              `tfsdk:"project_id"`
	Name        types.String              `tfsdk:"name"`
	Description types.String              `tfsdk:"description"`
	Mappings    []IntegrationMappingModel `tfsdk:"mappings"`
}

type IntegrationMappingModel struct {
	IntegrationID types.String `tfsdk:"integration_id"`
	Environment   types.String `tfsdk:"environment"`
	Priority      types.Int64  `tfsdk:"priority"`
}

func NewIntegrationGroupResource() resource.Resource {
	return &IntegrationGroupResource{}
}

func (r *IntegrationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_group"
}

func (r *IntegrationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group of integrations mapped to environments for multi-KeyVault scenarios.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the integration group.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project this group belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the integration group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the integration group.",
			},
		},
		Blocks: map[string]schema.Block{
			"mappings": schema.ListNestedBlock{
				Description: "List of integration to environment mappings.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"integration_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the integration (e.g., Azure KeyVault).",
						},
						"environment": schema.StringAttribute{
							Required:    true,
							Description: "The environment this integration serves (e.g., dev, staging, prod).",
						},
						"priority": schema.Int64Attribute{
							Optional:    true,
							Description: "Priority for failover scenarios.",
						},
					},
				},
			},
		},
	}
}

func (r *IntegrationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IntegrationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IntegrationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateIntegrationGroupRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}
	for _, m := range data.Mappings {
		createReq.Mappings = append(createReq.Mappings, mazevault.IntegrationMapping{
			IntegrationID: m.IntegrationID.ValueString(),
			Environment:   m.Environment.ValueString(),
			Priority:      int(m.Priority.ValueInt64()),
		})
	}

	group, err := r.client.CreateIntegrationGroup(data.ProjectID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create integration group: %s", err))
		return
	}

	data.ID = types.StringValue(group.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IntegrationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetIntegrationGroup(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read integration group %s: %s", data.ID.ValueString(), err))
		return
	}
	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	data.ProjectID = types.StringValue(group.ProjectID)

	data.Mappings = nil
	for _, m := range group.Mappings {
		data.Mappings = append(data.Mappings, IntegrationMappingModel{
			IntegrationID: types.StringValue(m.IntegrationID),
			Environment:   types.StringValue(m.Environment),
			Priority:      types.Int64Value(int64(m.Priority)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IntegrationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.CreateIntegrationGroupRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}
	for _, m := range data.Mappings {
		updateReq.Mappings = append(updateReq.Mappings, mazevault.IntegrationMapping{
			IntegrationID: m.IntegrationID.ValueString(),
			Environment:   m.Environment.ValueString(),
			Priority:      int(m.Priority.ValueInt64()),
		})
	}

	group, err := r.client.UpdateIntegrationGroup(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update integration group %s: %s", data.ID.ValueString(), err))
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Description = types.StringValue(group.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IntegrationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIntegrationGroup(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete integration group %s: %s", data.ID.ValueString(), err))
	}
}
