package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RotationWorkflowResource{}

type RotationWorkflowResource struct {
	client *mazevault.Client
}

type RotationWorkflowResourceModel struct {
	ID                  types.String              `tfsdk:"id"`
	SecretID            types.String              `tfsdk:"secret_id"`
	Environment         types.String              `tfsdk:"environment"`
	TTLHours            types.Int64               `tfsdk:"ttl_hours"`
	RotationStrategy    types.String              `tfsdk:"rotation_strategy"`
	LinkedIntegrations  types.List                `tfsdk:"linked_integrations"`
	PostRotationActions []PostRotationActionModel `tfsdk:"post_rotation_actions"`
	// Fáze N: extended attributes for scope, target environment, grace period, and resource kind
	TargetEnvironment  types.String `tfsdk:"target_environment"`
	Scope              types.String `tfsdk:"scope"`
	GracePeriodMinutes types.Int64  `tfsdk:"grace_period_minutes"`
	ResourceKind       types.String `tfsdk:"resource_kind"`
}

type PostRotationActionModel struct {
	Type      types.String `tfsdk:"type"`
	Config    types.Map    `tfsdk:"config"`
	Order     types.Int64  `tfsdk:"order"`
	OnFailure types.String `tfsdk:"on_failure"`
}

func NewRotationWorkflowResource() resource.Resource {
	return &RotationWorkflowResource{}
}

func (r *RotationWorkflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_workflow"
}

func (r *RotationWorkflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rotation workflow for a secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the rotation workflow.",
			},
			"secret_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the secret to rotate.",
			},
			"environment": schema.StringAttribute{
				Required:    true,
				Description: "The environment this workflow applies to.",
			},
			"ttl_hours": schema.Int64Attribute{
				Required:    true,
				Description: "Time-to-live in hours for the secret.",
			},
			"rotation_strategy": schema.StringAttribute{
				Optional:    true,
				Description: "Strategy for rotation (e.g., atomic, blue_green).",
			},
			"target_environment": schema.StringAttribute{
				Optional:    true,
				Description: "Override the target environment for secret write-back. Defaults to the workflow's own environment.",
			},
			"scope": schema.StringAttribute{
				Optional:    true,
				Description: "Scope of this workflow: 'organization' or 'project'. Defaults to 'project'.",
			},
			"grace_period_minutes": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of minutes the old secret/credential remains active after rotation. 0 means hard cutover (default).",
			},
			"resource_kind": schema.StringAttribute{
				Optional:    true,
				Description: "Kind of rotation resource: 'secret', 'certificate', 'entra', 'ssh_key_pair', 'keytab', 'system_key'. Defaults to 'secret'.",
			},
			"linked_integrations": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of integration IDs to include in rotation.",
			},
		},
		Blocks: map[string]schema.Block{
			"post_rotation_actions": schema.ListNestedBlock{
				Description: "List of actions to perform after rotation.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Type of action (e.g., iis_recycle, api_notification).",
						},
						"config": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Configuration for the action.",
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Description: "Execution order.",
						},
						"on_failure": schema.StringAttribute{
							Optional:    true,
							Description: "Behavior on failure (continue, rollback, notify).",
						},
					},
				},
			},
		},
	}
}

func (r *RotationWorkflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RotationWorkflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RotationWorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var linkedIntegrations []string
	resp.Diagnostics.Append(data.LinkedIntegrations.ElementsAs(ctx, &linkedIntegrations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateRotationWorkflowRequest{
		SecretID:             data.SecretID.ValueString(),
		Enabled:              true,
		RotationIntervalDays: int(data.TTLHours.ValueInt64() / 24),
		RotationStrategy:     data.RotationStrategy.ValueString(),
		TargetEnvironment:    data.TargetEnvironment.ValueString(),
		GracePeriodMinutes:   int(data.GracePeriodMinutes.ValueInt64()),
		ResourceKind:         data.ResourceKind.ValueString(),
		LinkedIntegrations:   linkedIntegrations,
	}

	for _, action := range data.PostRotationActions {
		cfg := make(map[string]string)
		resp.Diagnostics.Append(action.Config.ElementsAs(ctx, &cfg, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.PostRotationActions = append(createReq.PostRotationActions, mazevault.PostRotationActionWF{
			Type:      action.Type.ValueString(),
			Config:    cfg,
			Order:     int(action.Order.ValueInt64()),
			OnFailure: action.OnFailure.ValueString(),
		})
	}

	wf, err := r.client.CreateRotationWorkflow(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation workflow: %s", err))
		return
	}

	data.ID = types.StringValue(wf.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationWorkflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RotationWorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wf, err := r.client.GetRotationWorkflow(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation workflow %s: %s", data.ID.ValueString(), err))
		return
	}
	if wf == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.SecretID = types.StringValue(wf.SecretID)
	data.Environment = types.StringValue(wf.Environment)
	data.RotationStrategy = types.StringValue(wf.RotationStrategy)
	data.TargetEnvironment = types.StringValue(wf.TargetEnvironment)
	data.ResourceKind = types.StringValue(wf.ResourceKind)
	data.GracePeriodMinutes = types.Int64Value(int64(wf.GracePeriodMinutes))

	linkedList, diags := types.ListValueFrom(ctx, types.StringType, wf.LinkedIntegrations)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.LinkedIntegrations = linkedList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationWorkflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RotationWorkflowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var linkedIntegrations []string
	resp.Diagnostics.Append(data.LinkedIntegrations.ElementsAs(ctx, &linkedIntegrations, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.CreateRotationWorkflowRequest{
		SecretID:             data.SecretID.ValueString(),
		Enabled:              true,
		RotationIntervalDays: int(data.TTLHours.ValueInt64() / 24),
		RotationStrategy:     data.RotationStrategy.ValueString(),
		TargetEnvironment:    data.TargetEnvironment.ValueString(),
		GracePeriodMinutes:   int(data.GracePeriodMinutes.ValueInt64()),
		ResourceKind:         data.ResourceKind.ValueString(),
		LinkedIntegrations:   linkedIntegrations,
	}

	for _, action := range data.PostRotationActions {
		cfg := make(map[string]string)
		resp.Diagnostics.Append(action.Config.ElementsAs(ctx, &cfg, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.PostRotationActions = append(updateReq.PostRotationActions, mazevault.PostRotationActionWF{
			Type:      action.Type.ValueString(),
			Config:    cfg,
			Order:     int(action.Order.ValueInt64()),
			OnFailure: action.OnFailure.ValueString(),
		})
	}

	wf, err := r.client.UpdateRotationWorkflow(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation workflow %s: %s", data.ID.ValueString(), err))
		return
	}

	data.SecretID = types.StringValue(wf.SecretID)
	data.TargetEnvironment = types.StringValue(wf.TargetEnvironment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationWorkflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RotationWorkflowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRotationWorkflow(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation workflow %s: %s", data.ID.ValueString(), err))
	}
}
