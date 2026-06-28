package resources

import (
	"context"
	"encoding/json"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SecretRotationTargetResource{}

func NewSecretRotationTargetResource() resource.Resource {
	return &SecretRotationTargetResource{}
}

// SecretRotationTargetResource manages a deployment target in a secret's rotation pipeline.
//
// Each target describes WHERE and HOW a rotated secret value should be delivered:
// Kubernetes secrets, database password rotation, agent file sync, DevOps variables,
// or cloud vaults (AWS Secrets Manager, GCP Secret Manager, OCI Vault).
//
// Terraform example:
//
//	resource "mazevault_secret_rotation_target" "k8s" {
//	  secret_id   = mazevault_secret.db_pass.id
//	  target_type = "kubernetes_secret"
//	  priority    = 10
//	  enabled     = true
//	  config_json = jsonencode({
//	    namespace   = "production"
//	    secret_name = "app-db-creds"
//	    secret_key  = "password"
//	  })
//	}
type SecretRotationTargetResource struct {
	client *mazevault.Client
}

// SecretRotationTargetModel is the Terraform state model for a rotation target.
type SecretRotationTargetModel struct {
	ID               types.String `tfsdk:"id"`
	SecretID         types.String `tfsdk:"secret_id"`
	TargetType       types.String `tfsdk:"target_type"`
	TargetRole       types.String `tfsdk:"target_role"`
	Priority         types.Int64  `tfsdk:"priority"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	SyncOnSuccess    types.Bool   `tfsdk:"sync_on_success"`
	VerificationMode types.String `tfsdk:"verification_mode"`
	RecoveryPolicy   types.String `tfsdk:"recovery_policy"`
	BindingRef       types.String `tfsdk:"binding_ref"`
	// config_json stores the type-specific configuration as a JSON string.
	// Using a JSON string keeps the schema provider-agnostic and avoids
	// a top-level schema explosion for 7 different target type shapes.
	ConfigJSON         types.String `tfsdk:"config_json"`
	LastDeliveryStatus types.String `tfsdk:"last_delivery_status"`
}

func (r *SecretRotationTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_rotation_target"
}

func (r *SecretRotationTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a structured deployment target in a secret's rotation pipeline. " +
			"Targets execute in priority order after each rotation to deliver the new value " +
			"to external systems (Kubernetes, databases, agents, DevOps variables, cloud vaults).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the rotation target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_id": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the secret this deployment target belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Required:    true,
				Description: "Delivery mechanism. Valid values: kubernetes_secret, database_password, agent_sync, devops_variable, aws_secrets_manager, gcp_secret_manager, oci_vault.",
			},
			"target_role": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("distribution"),
				Description: "Lifecycle role. Valid values: distribution (default), verification, notification, post_action.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(10),
				Description: "Execution order within the same role. Lower numbers execute first.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether this target is active.",
			},
			"sync_on_success": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to sync this target after a successful rotation.",
			},
			"verification_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Description: "Post-delivery verification mode. Valid values: none, pull_and_verify, signature_verify.",
			},
			"recovery_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("manual"),
				Description: "Recovery policy on delivery failure. Valid values: manual, automatic.",
			},
			"binding_ref": schema.StringAttribute{
				Optional:    true,
				Description: "Optional external reference (integration ID, ARN, or resource path) used for auth.",
			},
			"config_json": schema.StringAttribute{
				Required: true,
				Description: "Type-specific configuration as a JSON-encoded string. Schema depends on target_type. " +
					"Use jsonencode() to build this value in Terraform.",
			},
			"last_delivery_status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the last delivery attempt (pending, success, failed).",
			},
		},
	}
}

func (r *SecretRotationTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SecretRotationTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretRotationTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(data.ConfigJSON.ValueString()), &cfg); err != nil || cfg == nil {
		resp.Diagnostics.AddError("Invalid config_json",
			fmt.Sprintf("config_json must be a valid JSON object: %v", err))
		return
	}

	enabled := data.Enabled.ValueBool()
	syncOnSuccess := data.SyncOnSuccess.ValueBool()
	created, err := r.client.CreateSecretRotationTarget(data.SecretID.ValueString(), &mazevault.CreateSecretRotationTargetRequest{
		TargetType:       data.TargetType.ValueString(),
		TargetRole:       data.TargetRole.ValueString(),
		Priority:         int(data.Priority.ValueInt64()),
		Enabled:          enabled,
		SyncOnSuccess:    syncOnSuccess,
		VerificationMode: data.VerificationMode.ValueString(),
		RecoveryPolicy:   data.RecoveryPolicy.ValueString(),
		BindingRef:       data.BindingRef.ValueString(),
		Config:           cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create SecretRotationTarget Error", err.Error())
		return
	}

	data.ID = types.StringValue(created.ID)
	data.LastDeliveryStatus = types.StringValue(created.LastDeliveryStatus)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretRotationTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretRotationTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target, err := r.client.GetSecretRotationTarget(data.SecretID.ValueString(), data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			// Resource no longer exists; remove from state so Terraform will recreate it.
			resp.State.RemoveResource(ctx)
			return
		}
		// Transient error (network, auth, 5xx) — surface as a real error, do NOT wipe state.
		resp.Diagnostics.AddError("Read SecretRotationTarget Error", err.Error())
		return
	}

	data.TargetType = types.StringValue(target.TargetType)
	data.TargetRole = types.StringValue(target.TargetRole)
	data.Priority = types.Int64Value(int64(target.Priority))
	data.Enabled = types.BoolValue(target.Enabled)
	data.SyncOnSuccess = types.BoolValue(target.SyncOnSuccess)
	data.VerificationMode = types.StringValue(target.VerificationMode)
	data.RecoveryPolicy = types.StringValue(target.RecoveryPolicy)
	data.BindingRef = types.StringValue(target.BindingRef)
	data.LastDeliveryStatus = types.StringValue(target.LastDeliveryStatus)

	// Round-trip the config as JSON so state drift detection works.
	if target.Config != nil {
		b, marshalErr := json.Marshal(target.Config)
		if marshalErr == nil {
			data.ConfigJSON = types.StringValue(string(b))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretRotationTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecretRotationTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(data.ConfigJSON.ValueString()), &cfg); err != nil || cfg == nil {
		resp.Diagnostics.AddError("Invalid config_json",
			fmt.Sprintf("config_json must be a valid JSON object: %v", err))
		return
	}

	targetType := data.TargetType.ValueString()
	targetRole := data.TargetRole.ValueString()
	priority := int(data.Priority.ValueInt64())
	enabled := data.Enabled.ValueBool()
	syncOnSuccess := data.SyncOnSuccess.ValueBool()
	verificationMode := data.VerificationMode.ValueString()
	recoveryPolicy := data.RecoveryPolicy.ValueString()
	bindingRef := data.BindingRef.ValueString()

	updated, err := r.client.UpdateSecretRotationTarget(
		data.SecretID.ValueString(),
		data.ID.ValueString(),
		&mazevault.UpdateSecretRotationTargetRequest{
			TargetType:       &targetType,
			TargetRole:       &targetRole,
			Priority:         &priority,
			Enabled:          &enabled,
			SyncOnSuccess:    &syncOnSuccess,
			VerificationMode: &verificationMode,
			RecoveryPolicy:   &recoveryPolicy,
			BindingRef:       &bindingRef,
			Config:           cfg,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Update SecretRotationTarget Error", err.Error())
		return
	}

	data.LastDeliveryStatus = types.StringValue(updated.LastDeliveryStatus)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretRotationTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretRotationTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecretRotationTarget(
		data.SecretID.ValueString(),
		data.ID.ValueString(),
	); err != nil {
		if mazevault.IsNotFoundError(err) {
			// Already deleted — idempotent destroy succeeds silently.
			return
		}
		resp.Diagnostics.AddError("Delete SecretRotationTarget Error", err.Error())
		return
	}
}
