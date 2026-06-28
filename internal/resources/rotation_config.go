package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RotationConfigResource{}
var _ resource.ResourceWithImportState = &RotationConfigResource{}

func NewRotationConfigResource() resource.Resource {
	return &RotationConfigResource{}
}

type RotationConfigResource struct {
	client *mazevault.Client
}

// RotationConfigResourceModel is the Terraform state model for mazevault_rotation_config.
//
// BREAKING CHANGE (v2.0): The `environment` and `rotation_strategy` attributes have been
// removed because the backend API does not accept them.  Remove them from existing
// configurations and run `terraform state rm` + `terraform import` if needed.
type RotationConfigResourceModel struct {
	ID                   types.String              `tfsdk:"id"`
	SecretID             types.String              `tfsdk:"secret_id"`
	RotationIntervalDays types.Int64               `tfsdk:"rotation_interval_days"`
	Enabled              types.Bool                `tfsdk:"enabled"`
	Schedule             types.String              `tfsdk:"schedule"`
	NotificationEmails   types.List                `tfsdk:"notification_emails"`
	PostActions          []PostRotationActionModel `tfsdk:"post_actions"`
	TargetEnvironment    types.String              `tfsdk:"target_environment"`
	// Computed status fields — populated from the backend; not user-configurable.
	Status         types.String `tfsdk:"status"`
	LastRotatedAt  types.String `tfsdk:"last_rotated_at"`
	NextRotationAt types.String `tfsdk:"next_rotation_at"`
	LastError      types.String `tfsdk:"last_error"`
}

func (r *RotationConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_config"
}

func (r *RotationConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rotation configuration for a secret. " +
			"The rotation config controls scheduling, notification, and post-rotation actions " +
			"for automatic secret rotation.\n\n" +
			"> **Breaking change (v2.0):** The `environment` and `rotation_strategy` arguments " +
			"have been removed. Remove them from existing configurations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the rotation configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the secret to rotate. Changing this value forces resource replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rotation_interval_days": schema.Int64Attribute{
				Optional:    true,
				Description: "Rotation interval in days.  Mutually exclusive with `schedule`; set one or the other.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether automatic rotation is enabled for this secret.  Defaults to `true`.",
			},
			"schedule": schema.StringAttribute{
				Optional: true,
				Description: "Cron expression for scheduled rotation (e.g. `\"0 2 * * 0\"` for weekly Sundays at 02:00 UTC). " +
					"Mutually exclusive with `rotation_interval_days`.",
			},
			"notification_emails": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of e-mail addresses to notify after a rotation event.",
			},
			"target_environment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Override the target environment for secret write-back.  Defaults to the rotation config's own environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// Computed status attributes.
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current rotation status (`idle`, `running`, `failed`, etc.).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_rotated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp of the last successful rotation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"next_rotation_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp of the next scheduled rotation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_error": schema.StringAttribute{
				Computed:    true,
				Description: "Error message from the last failed rotation, if any.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"post_actions": schema.ListNestedBlock{
				Description: "Ordered list of actions to execute after a successful rotation.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Action type. Supported values: `spring_actuator_refresh`, `webhook`, `agent_command`, `shell_script`, `azure_keyvault`, `kubernetes_secret`, `iis_recycle`.",
						},
						"config": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "String key/value configuration map for the action.",
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Description: "Execution order (lower numbers run first).",
						},
						"on_failure": schema.StringAttribute{
							Optional:    true,
							Description: "Behaviour on failure: `continue`, `rollback`, or `notify`.",
						},
						"gateway_id": schema.StringAttribute{
							Optional:    true,
							Description: "Pin this action to a specific gateway ID instead of using environment-based routing.",
						},
						"target_environment": schema.StringAttribute{
							Optional:    true,
							Description: "Override the environment context used when resolving which gateway executes this action.",
						},
					},
				},
			},
		},
	}
}

func (r *RotationConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RotationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateRotationConfigRequest{
		SecretID:             data.SecretID.ValueString(),
		Enabled:              data.Enabled.ValueBool(),
		Schedule:             data.Schedule.ValueString(),
		RotationIntervalDays: int(data.RotationIntervalDays.ValueInt64()),
	}

	var emails []string
	resp.Diagnostics.Append(data.NotificationEmails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq.NotificationEmails = emails

	for _, a := range data.PostActions {
		cfg := make(map[string]string)
		resp.Diagnostics.Append(a.Config.ElementsAs(ctx, &cfg, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.PostActions = append(createReq.PostActions, mazevault.PostRotationActionWF{
			Type:              a.Type.ValueString(),
			Config:            cfg,
			Order:             int(a.Order.ValueInt64()),
			OnFailure:         a.OnFailure.ValueString(),
			GatewayID:         a.GatewayID.ValueString(),
			TargetEnvironment: a.TargetEnvironment.ValueString(),
		})
	}

	cfg, err := r.client.CreateRotationConfig(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation config: %s", err))
		return
	}

	data.ID = types.StringValue(cfg.ID)
	populateRotationConfigState(&data, cfg)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetRotationConfig(data.ID.ValueString())
	if err != nil {
		// Treat definitive 404s as a deleted resource; surface all other errors.
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read rotation config %s: %s", data.ID.ValueString(), err),
		)
		return
	}
	if cfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(cfg.ID)
	data.SecretID = types.StringValue(cfg.SecretID)
	data.Enabled = types.BoolValue(cfg.Enabled)
	data.Schedule = types.StringValue(cfg.Schedule)
	data.RotationIntervalDays = types.Int64Value(int64(cfg.RotationIntervalDays))

	emails, diags := types.ListValueFrom(ctx, types.StringType, cfg.NotificationEmails)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		data.NotificationEmails = emails
	}

	// Restore post_actions from the server response.
	if len(cfg.PostActions) > 0 {
		actions := make([]PostRotationActionModel, 0, len(cfg.PostActions))
		for _, a := range cfg.PostActions {
			cfgMap, cfgDiags := types.MapValueFrom(ctx, types.StringType, a.Config)
			resp.Diagnostics.Append(cfgDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			actions = append(actions, PostRotationActionModel{
				Type:              types.StringValue(a.Type),
				Config:            cfgMap,
				Order:             types.Int64Value(int64(a.Order)),
				OnFailure:         types.StringValue(a.OnFailure),
				GatewayID:         types.StringValue(a.GatewayID),
				TargetEnvironment: types.StringValue(a.TargetEnvironment),
			})
		}
		data.PostActions = actions
	} else if data.PostActions == nil {
		data.PostActions = []PostRotationActionModel{}
	}

	if cfg.TargetEnvironment != "" {
		data.TargetEnvironment = types.StringValue(cfg.TargetEnvironment)
	}

	populateRotationConfigState(&data, cfg)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.UpdateRotationConfigRequest{
		Enabled:              data.Enabled.ValueBool(),
		Schedule:             data.Schedule.ValueString(),
		RotationIntervalDays: int(data.RotationIntervalDays.ValueInt64()),
	}

	var emails []string
	resp.Diagnostics.Append(data.NotificationEmails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq.NotificationEmails = emails

	for _, a := range data.PostActions {
		cfg := make(map[string]string)
		resp.Diagnostics.Append(a.Config.ElementsAs(ctx, &cfg, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.PostActions = append(updateReq.PostActions, mazevault.PostRotationActionWF{
			Type:              a.Type.ValueString(),
			Config:            cfg,
			Order:             int(a.Order.ValueInt64()),
			OnFailure:         a.OnFailure.ValueString(),
			GatewayID:         a.GatewayID.ValueString(),
			TargetEnvironment: a.TargetEnvironment.ValueString(),
		})
	}

	cfg, err := r.client.UpdateRotationConfig(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update rotation config %s: %s", data.ID.ValueString(), err),
		)
		return
	}

	data.SecretID = types.StringValue(cfg.SecretID)
	if cfg.TargetEnvironment != "" {
		data.TargetEnvironment = types.StringValue(cfg.TargetEnvironment)
	}
	populateRotationConfigState(&data, cfg)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRotationConfig(data.ID.ValueString()); err != nil {
		// If already gone, treat as success.
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete rotation config %s: %s", data.ID.ValueString(), err),
		)
	}
}

// populateRotationConfigState copies computed/status fields from the API response
// into the Terraform state model.
func populateRotationConfigState(data *RotationConfigResourceModel, cfg *mazevault.RotationConfigDetail) {
	// Status and errors from the response.
	if cfg.Status != "" {
		data.Status = types.StringValue(cfg.Status)
	} else if cfg.Resource != nil && cfg.Resource.StatusSummary != "" {
		data.Status = types.StringValue(cfg.Resource.StatusSummary)
	}

	if cfg.LastError != "" {
		data.LastError = types.StringValue(cfg.LastError)
	}

	if cfg.LastRotatedAt != nil {
		data.LastRotatedAt = types.StringValue(cfg.LastRotatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
	if cfg.NextRotationAt != nil {
		data.NextRotationAt = types.StringValue(cfg.NextRotationAt.Format("2006-01-02T15:04:05Z07:00"))
	} else if cfg.Resource != nil && cfg.Resource.NextDueAt != nil {
		data.NextRotationAt = types.StringValue(cfg.Resource.NextDueAt.Format("2006-01-02T15:04:05Z07:00"))
	}
}

// emptyStringList returns an empty types.List of string elements.
// Used to initialise optional list attributes that the server returns as nil.
func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}

// ImportState implements resource.ResourceWithImportState.
func (r *RotationConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
