package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CertificateRotationConfigResource{}
var _ resource.ResourceWithImportState = &CertificateRotationConfigResource{}

// NewCertificateRotationConfigResource returns a new mazevault_certificate_rotation_config resource.
func NewCertificateRotationConfigResource() resource.Resource {
	return &CertificateRotationConfigResource{}
}

// CertificateRotationConfigResource manages certificate rotation settings via
// GET/PUT /api/v1/certificates/:id/rotation/config.
//
// Because the backend has no hard-delete endpoint for certificate rotation configs,
// destroying this resource disables rotation (sets enabled = false) and removes the
// resource from Terraform state.  The rotation config record itself persists in the
// backend and can be re-imported with terraform import.
type CertificateRotationConfigResource struct {
	client *mazevault.Client
}

// CertRotationConfigModel is the Terraform state model for mazevault_certificate_rotation_config.
type CertRotationConfigModel struct {
	ID                  types.String              `tfsdk:"id"`
	CertificateID       types.String              `tfsdk:"certificate_id"`
	Enabled             types.Bool                `tfsdk:"enabled"`
	Schedule            types.String              `tfsdk:"schedule"`
	RenewalLeadDays     types.Int64               `tfsdk:"renewal_lead_days"`
	MaxRetryAttempts    types.Int64               `tfsdk:"max_retry_attempts"`
	RetryDelaySeconds   types.Int64               `tfsdk:"retry_delay_seconds"`
	TimeoutMinutes      types.Int64               `tfsdk:"timeout_minutes"`
	NotificationEmails  types.List                `tfsdk:"notification_emails"`
	PostRotationActions []PostRotationActionModel `tfsdk:"post_rotation_actions"`
	// Computed.
	ConfigInitialized types.Bool   `tfsdk:"config_initialized"`
	LastRotatedAt     types.String `tfsdk:"last_rotated_at"`
	NextRotationAt    types.String `tfsdk:"next_rotation_at"`
}

func (r *CertificateRotationConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_rotation_config"
}

func (r *CertificateRotationConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the rotation (renewal) configuration for a certificate.\n\n" +
			"This resource controls when and how a certificate is automatically renewed: " +
			"lead time before expiry, retry behaviour, post-renewal notifications and actions.\n\n" +
			"> **No hard-delete:** Destroying this resource disables automatic rotation and " +
			"removes it from Terraform state.  The backend record is not deleted.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the rotation config record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the certificate whose renewal is being configured. Changing this value forces resource replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether automatic renewal is enabled.  Defaults to `true`.",
			},
			"schedule": schema.StringAttribute{
				Optional:    true,
				Description: "Cron expression for scheduled renewal (e.g. `\"0 3 * * *\"` for daily at 03:00 UTC). When set, takes precedence over `renewal_lead_days`-based scheduling.",
			},
			"renewal_lead_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of days before certificate expiry to start the renewal process.  Defaults to 30.",
			},
			"max_retry_attempts": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
				Description: "Maximum number of retry attempts on failure.  Defaults to 3.",
			},
			"retry_delay_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Delay in seconds between retry attempts.  Defaults to 60.",
			},
			"timeout_minutes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Maximum execution time in minutes before the renewal is considered timed out.  Defaults to 10.",
			},
			"notification_emails": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "E-mail addresses to notify after a renewal event.",
			},
			// Computed status fields.
			"config_initialized": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the rotation config has been fully initialised by the backend.",
			},
			"last_rotated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp of the last successful renewal.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"next_rotation_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 timestamp of the next scheduled renewal.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"post_rotation_actions": schema.ListNestedBlock{
				Description: "Ordered list of actions to execute after a successful renewal.",
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

func (r *CertificateRotationConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateRotationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertRotationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := buildCertRotationRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateCertificateRotationConfig(data.CertificateID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create certificate rotation config for %s: %s", data.CertificateID.ValueString(), err))
		return
	}

	populateCertRotationState(ctx, &data, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateRotationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertRotationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetCertificateRotationConfig(data.CertificateID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read certificate rotation config for %s: %s", data.CertificateID.ValueString(), err))
		return
	}
	if cfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	populateCertRotationState(ctx, &data, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateRotationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CertRotationConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq2 := buildCertRotationRequest(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg2, err := r.client.UpdateCertificateRotationConfig(data.CertificateID.ValueString(), updateReq2)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update certificate rotation config for %s: %s", data.CertificateID.ValueString(), err))
		return
	}

	populateCertRotationState(ctx, &data, cfg2, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete disables automatic rotation (no hard-delete endpoint exists on the backend).
func (r *CertificateRotationConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertRotationConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	disabled := false
	disableReq := &mazevault.UpdateCertRotationConfigRequest{Enabled: &disabled}
	if _, err := r.client.UpdateCertificateRotationConfig(data.CertificateID.ValueString(), disableReq); err != nil {
		// Treat 404 as already gone.
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to disable certificate rotation config for %s: %s", data.CertificateID.ValueString(), err))
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// buildCertRotationRequest converts the Terraform model to the SDK request struct.
func buildCertRotationRequest(ctx context.Context, data *CertRotationConfigModel, d *diag.Diagnostics) *mazevault.UpdateCertRotationConfigRequest {
	enabled := data.Enabled.ValueBool()
	renewalLeadDays := int(data.RenewalLeadDays.ValueInt64())
	maxRetry := int(data.MaxRetryAttempts.ValueInt64())
	retryDelay := int(data.RetryDelaySeconds.ValueInt64())
	timeout := int(data.TimeoutMinutes.ValueInt64())

	req := &mazevault.UpdateCertRotationConfigRequest{
		Enabled:           &enabled,
		Schedule:          data.Schedule.ValueString(),
		RenewalLeadDays:   &renewalLeadDays,
		MaxRetryAttempts:  &maxRetry,
		RetryDelaySeconds: &retryDelay,
		TimeoutMinutes:    &timeout,
	}

	var emails []string
	d.Append(data.NotificationEmails.ElementsAs(ctx, &emails, false)...)
	req.NotificationEmails = emails

	for _, a := range data.PostRotationActions {
		cfg := make(map[string]string)
		d.Append(a.Config.ElementsAs(ctx, &cfg, false)...)
		req.PostRotationActions = append(req.PostRotationActions, mazevault.PostRotationActionWF{
			Type:              a.Type.ValueString(),
			Config:            cfg,
			Order:             int(a.Order.ValueInt64()),
			OnFailure:         a.OnFailure.ValueString(),
			GatewayID:         a.GatewayID.ValueString(),
			TargetEnvironment: a.TargetEnvironment.ValueString(),
		})
	}

	return req
}

// populateCertRotationState reads the API response back into the TF state model.
func populateCertRotationState(ctx context.Context, data *CertRotationConfigModel, cfg *mazevault.CertificateRotationConfigDetail, d *diag.Diagnostics) {
	data.ID = types.StringValue(cfg.ID)
	data.CertificateID = types.StringValue(cfg.CertificateID)
	data.ConfigInitialized = types.BoolValue(cfg.ConfigInitialized)
	data.Enabled = types.BoolValue(cfg.Enabled)
	data.Schedule = types.StringValue(cfg.Schedule)
	data.RenewalLeadDays = types.Int64Value(int64(cfg.RenewalLeadDays))
	data.MaxRetryAttempts = types.Int64Value(int64(cfg.MaxRetryAttempts))
	data.RetryDelaySeconds = types.Int64Value(int64(cfg.RetryDelaySeconds))
	data.TimeoutMinutes = types.Int64Value(int64(cfg.TimeoutMinutes))

	emails, diags := types.ListValueFrom(ctx, types.StringType, cfg.NotificationEmails)
	d.Append(diags...)
	if !d.HasError() {
		data.NotificationEmails = emails
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
	}
	if cfg.NextRotationAt != nil {
		data.NextRotationAt = types.StringValue(cfg.NextRotationAt.Format("2006-01-02T15:04:05Z07:00"))
	} else if cfg.Resource != nil && cfg.Resource.NextDueAt != nil {
		data.NextRotationAt = types.StringValue(cfg.Resource.NextDueAt.Format("2006-01-02T15:04:05Z07:00"))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import by the certificate UUID: terraform import mazevault_certificate_rotation_config.x <certificate-uuid>
func (r *CertificateRotationConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("certificate_id"), req, resp)
}
