package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EntraRotationConfigResource{}
var _ resource.ResourceWithImportState = &EntraRotationConfigResource{}

// NewEntraRotationConfigResource returns a new mazevault_entra_rotation_config resource.
func NewEntraRotationConfigResource() resource.Resource {
	return &EntraRotationConfigResource{}
}

// EntraRotationConfigResource manages rotation settings for an Entra credential via
// POST/GET/PUT /api/v1/entra/credentials/:id/rotation-config.
//
// Because the backend has no hard-delete endpoint, Destroy disables rotation
// (rotation_enabled = false) and removes the resource from Terraform state.
type EntraRotationConfigResource struct {
	client *mazevault.Client
}

// EntraRotationConfigModel is the Terraform state model for mazevault_entra_rotation_config.
type EntraRotationConfigModel struct {
	ID                       types.String              `tfsdk:"id"`
	CredentialID             types.String              `tfsdk:"credential_id"`
	RotationEnabled          types.Bool                `tfsdk:"rotation_enabled"`
	RotationDaysBeforeExpiry types.Int64               `tfsdk:"rotation_days_before_expiry"`
	GracePeriodDays          types.Int64               `tfsdk:"grace_period_days"`
	KVIntegrationIDs         types.List                `tfsdk:"kv_integration_ids"`
	SecretName               types.String              `tfsdk:"secret_name"`
	SpringEndpoints          types.List                `tfsdk:"spring_endpoints"`
	WebhookURLs              types.List                `tfsdk:"webhook_urls"`
	PostRotationActions      []PostRotationActionModel `tfsdk:"post_rotation_actions"`
	StagedRotationEnabled    types.Bool                `tfsdk:"staged_rotation_enabled"`
	SoakWindowHours          types.Int64               `tfsdk:"soak_window_hours"`
	// Computed.
	LastRotatedAt types.String `tfsdk:"last_rotated_at"`
}

func (r *EntraRotationConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entra_rotation_config"
}

func (r *EntraRotationConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the rotation configuration for a Microsoft Entra (Azure AD) application credential.\n\n" +
			"This resource controls when and how client secrets are automatically rotated: " +
			"lead time before expiry, Key Vault write-back, Spring Boot endpoint refresh, " +
			"staged rotation soak windows, and post-rotation webhook/agent actions.\n\n" +
			"> **No hard-delete:** Destroying this resource sets `rotation_enabled = false` " +
			"and removes it from Terraform state.  The backend record is not deleted.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Derived ID — equal to `credential_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"credential_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the Entra credential (app registration client secret or certificate) to configure. Changing this forces resource replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rotation_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether automatic rotation is enabled.  Defaults to `true`.",
			},
			"rotation_days_before_expiry": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of days before credential expiry to start the rotation.  Defaults to 30.",
			},
			"grace_period_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(7),
				Description: "Number of days the old credential remains active after the new one is created.  Defaults to 7.",
			},
			"kv_integration_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Read-only mirror of the Key Vault integration IDs derived from `post_rotation_actions` entries of type `azure_keyvault`. To configure Key Vault write-back, add an `azure_keyvault` action to `post_rotation_actions` instead of setting this field.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_name": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only mirror of the target Key Vault secret name derived from the `azure_keyvault` post-rotation action.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spring_endpoints": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Read-only mirror of the Spring Boot Actuator refresh URLs derived from `post_rotation_actions` entries of type `spring_actuator_refresh`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"webhook_urls": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Read-only mirror of the webhook URLs derived from `post_rotation_actions` entries of type `webhook`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"staged_rotation_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable secondary-first staged rotation.  When `true`, MazeVault rotates the secondary credential first and waits for the `soak_window_hours` soak period before promoting it to primary.",
			},
			"soak_window_hours": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(48),
				Description: "Minimum hours to wait between secondary verification and primary promotion during staged rotation.  Defaults to 48.",
			},
			"last_rotated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp of the last successful rotation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"post_rotation_actions": schema.ListNestedBlock{
				Description: "Ordered list of delivery actions to execute after a successful rotation. " +
					"This is the single source of truth for post-rotation delivery: if no blocks are " +
					"defined, nothing is delivered. Key Vault write-back and Spring refresh are expressed " +
					"as actions here (see `type`), not via separate top-level fields.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Action type. Supported values: `azure_keyvault` (config: `integration_id`, `secret_name`), `spring_actuator_refresh` (config: `actuator_url`), `webhook` (config: `url`), `kubernetes_secret` (config: `namespace`, `secret_name`, `key_name`), `agent_secret_sync`, `iis_recycle`.",
						},
						"config": schema.MapAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "String key/value configuration map for the action. The server may add normalized defaults (e.g. `file_format` for `agent_secret_sync`).",
						},
						"order": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "Execution order (lower numbers run first). Server-assigned when omitted.",
						},
						"on_failure": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Behaviour on failure: `continue`, `abort`, or `retry`. Defaults to `continue`.",
						},
						"gateway_id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Pin this action to a specific gateway ID instead of using environment-based routing.",
						},
						"target_environment": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Override the environment context used when resolving which gateway executes this action.",
						},
					},
				},
			},
		},
	}
}

func (r *EntraRotationConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntraRotationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntraRotationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := buildEntraRotationRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.ConfigureEntraRotation(data.CredentialID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to configure Entra rotation for credential %s: %s", data.CredentialID.ValueString(), err))
		return
	}

	populateEntraRotationState(ctx, &data, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntraRotationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntraRotationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetEntraRotationConfig(data.CredentialID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read Entra rotation config for credential %s: %s", data.CredentialID.ValueString(), err))
		return
	}
	if cfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	populateEntraRotationState(ctx, &data, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntraRotationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntraRotationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildEntraRotationRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateEntraRotationConfig(data.CredentialID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update Entra rotation config for credential %s: %s", data.CredentialID.ValueString(), err))
		return
	}

	populateEntraRotationState(ctx, &data, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete disables rotation (no hard-delete endpoint on the backend).
func (r *EntraRotationConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntraRotationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	disableReq := &mazevault.ConfigureEntraRotationRequest{RotationEnabled: false}
	if _, err := r.client.UpdateEntraRotationConfig(data.CredentialID.ValueString(), disableReq); err != nil {
		if mazevault.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to disable Entra rotation for credential %s: %s", data.CredentialID.ValueString(), err))
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// buildEntraRotationRequest converts the Terraform model to the SDK request struct.
//
// post_rotation_actions is the single writable delivery surface. It is ALWAYS sent
// explicitly (as a possibly-empty array) so the backend never falls back to
// deriving actions from the legacy flat fields (which are read-only mirrors here).
func buildEntraRotationRequest(ctx context.Context, data *EntraRotationConfigModel, d *diag.Diagnostics) *mazevault.ConfigureEntraRotationRequest {
	soakWindowHours := int(data.SoakWindowHours.ValueInt64())
	stagedEnabled := data.StagedRotationEnabled.ValueBool()

	req := &mazevault.ConfigureEntraRotationRequest{
		RotationEnabled:          data.RotationEnabled.ValueBool(),
		RotationDaysBeforeExpiry: int(data.RotationDaysBeforeExpiry.ValueInt64()),
		GracePeriodDays:          int(data.GracePeriodDays.ValueInt64()),
		StagedRotationEnabled:    &stagedEnabled,
		SoakWindowHours:          &soakWindowHours,
	}

	actions := make([]mazevault.PostRotationActionWF, 0, len(data.PostRotationActions))
	for _, a := range data.PostRotationActions {
		cfg := make(map[string]string)
		if !a.Config.IsNull() && !a.Config.IsUnknown() {
			d.Append(a.Config.ElementsAs(ctx, &cfg, false)...)
		}
		actions = append(actions, mazevault.PostRotationActionWF{
			Type:              a.Type.ValueString(),
			Config:            cfg,
			Order:             int(a.Order.ValueInt64()),
			OnFailure:         a.OnFailure.ValueString(),
			GatewayID:         a.GatewayID.ValueString(),
			TargetEnvironment: a.TargetEnvironment.ValueString(),
		})
	}
	// Non-nil even when empty → serializes as "post_rotation_actions": [].
	req.PostRotationActions = actions

	return req
}

// populateEntraRotationState reads the API response back into the TF state model.
func populateEntraRotationState(ctx context.Context, data *EntraRotationConfigModel, cfg *mazevault.EntraRotationConfigDetail, d *diag.Diagnostics) {
	// Use credential_id as the TF resource ID since there is no separate config UUID.
	data.ID = types.StringValue(cfg.CredentialID)
	data.CredentialID = types.StringValue(cfg.CredentialID)
	data.RotationEnabled = types.BoolValue(cfg.RotationEnabled)
	data.RotationDaysBeforeExpiry = types.Int64Value(int64(cfg.RotationDaysBeforeExpiry))
	data.GracePeriodDays = types.Int64Value(int64(cfg.GracePeriodDays))
	data.SecretName = types.StringValue(cfg.SecretName)
	data.StagedRotationEnabled = types.BoolValue(cfg.StagedRotationEnabled)
	data.SoakWindowHours = types.Int64Value(int64(cfg.SoakWindowHours))

	kvIDs, diags := types.ListValueFrom(ctx, types.StringType, cfg.KVIntegrationIDs)
	d.Append(diags...)
	if !d.HasError() {
		data.KVIntegrationIDs = kvIDs
	}

	springs, diags := types.ListValueFrom(ctx, types.StringType, cfg.SpringEndpoints)
	d.Append(diags...)
	if !d.HasError() {
		data.SpringEndpoints = springs
	}

	webhooks, diags := types.ListValueFrom(ctx, types.StringType, cfg.WebhookURLs)
	d.Append(diags...)
	if !d.HasError() {
		data.WebhookURLs = webhooks
	}

	// Restore post_rotation_actions from server for drift detection.
	if len(cfg.PostRotationActions) > 0 {
		actions := make([]PostRotationActionModel, 0, len(cfg.PostRotationActions))
		for _, a := range cfg.PostRotationActions {
			cfgMap, cfgDiags := types.MapValueFrom(ctx, types.StringType, a.Config)
			d.Append(cfgDiags...)
			if d.HasError() {
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
		data.PostRotationActions = actions
	} else if data.PostRotationActions == nil {
		data.PostRotationActions = []PostRotationActionModel{}
	}

	if cfg.LastRotatedAt != nil {
		data.LastRotatedAt = types.StringValue(cfg.LastRotatedAt.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		data.LastRotatedAt = types.StringNull()
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import by the credential UUID: terraform import mazevault_entra_rotation_config.x <credential-uuid>
func (r *EntraRotationConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("credential_id"), req, resp)
}
