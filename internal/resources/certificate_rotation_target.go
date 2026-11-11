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

var _ resource.Resource = &CertificateRotationTargetResource{}

// NewCertificateRotationTargetResource manages where a certificate is deployed
// after issuance and renewal (files, keystores, trust stores, containers,
// Kubernetes secrets, secret managers) and how the consumer is reloaded.
func NewCertificateRotationTargetResource() resource.Resource {
	return &CertificateRotationTargetResource{}
}

type CertificateRotationTargetResource struct {
	client *mazevault.Client
}

type CertificateRotationTargetModel struct {
	ID                    types.String `tfsdk:"id"`
	CertificateID         types.String `tfsdk:"certificate_id"`
	TargetType            types.String `tfsdk:"target_type"`
	AgentID               types.String `tfsdk:"agent_id"`
	IntegrationID         types.String `tfsdk:"integration_id"`
	DatabaseIntegrationID types.String `tfsdk:"database_integration_id"`
	ContentMode           types.String `tfsdk:"content_mode"`
	TargetConfigJSON      types.String `tfsdk:"target_config_json"`
	DeploymentWindowJSON  types.String `tfsdk:"deployment_window_json"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	SyncOnRenewal         types.Bool   `tfsdk:"sync_on_renewal"`
	Priority              types.Int64  `tfsdk:"priority"`
	LastSyncStatus        types.String `tfsdk:"last_sync_status"`
}

func (r *CertificateRotationTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_rotation_target"
}

func (r *CertificateRotationTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a deployment target of a certificate: where the certificate (and key) is delivered " +
			"on issuance and every renewal, and how the consuming service is reloaded. Agent-backed targets " +
			"(agent_file, agent_keystore, agent_truststore, agent_container_file, database, weblogic) require agent_id. " +
			"The target_config_json shape follows the backend TargetConfiguration union " +
			"(file, keystore, trust_store, container_file, kubernetes, secret_manager, database, weblogic). " +
			"A reload strategy inside the configuration (e.g. file.reload = {type = \"nginx_reload\", verify = {port = 443}}) " +
			"makes the agent reload the consumer, verify the served certificate and roll back on mismatch; " +
			"deployment_window_json confines deferred reloads to a maintenance window.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the deployment target.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"certificate_id": schema.StringAttribute{
				Required:      true,
				Description:   "UUID of the managed certificate.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"target_type": schema.StringAttribute{
				Required: true,
				Description: "Target type: secret_manager, kubernetes_secret, database, agent_keystore, agent_file, " +
					"agent_truststore, agent_container_file, weblogic.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"agent_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the agent on the destination host (required for agent-backed target types).",
			},
			"integration_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the integration (secret manager / Kubernetes) for non-agent targets.",
			},
			"database_integration_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the database integration for database wallet targets.",
			},
			"content_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("full_chain_with_key"),
				Description: "What is delivered: public_only, public_chain, full_chain_with_key (default), key_only, custom. " +
					"A target of an OpenPGP key (certificate_type \"pgp\") delivers exactly one ASCII-armored block and must set " +
					"public_only (public key) or key_only (private key) explicitly; the default is refused for it.",
			},
			"target_config_json": schema.StringAttribute{
				Required: true,
				Description: "Type-specific configuration as a JSON-encoded object (backend TargetConfiguration union). " +
					"Use jsonencode(). Example for agent_file: {\"file\":{\"cert_path\":\"/etc/nginx/ssl/site.pem\",\"key_path\":\"/etc/nginx/ssl/site.key\"," +
					"\"reload\":{\"type\":\"nginx_reload\",\"verify\":{\"port\":443}}}}.",
			},
			"deployment_window_json": schema.StringAttribute{
				Optional: true,
				Description: "Optional maintenance window for deferred reloads as JSON: " +
					"{\"days\":[\"Sat\",\"Sun\"],\"from\":\"02:00\",\"to\":\"04:00\",\"time_zone\":\"Europe/Prague\"}. " +
					"Files are written at rotation time; the reload waits for the window.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether this target participates in deployments.",
			},
			"sync_on_renewal": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Deploy automatically after every renewal.",
			},
			"priority": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Execution order among the certificate's targets (lower first).",
			},
			"last_sync_status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the last deployment to this target.",
			},
		},
	}
}

func (r *CertificateRotationTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func optionalString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	s := v.ValueString()
	return &s
}

func parseTargetJSONObject(field string, v types.String, required bool) (map[string]interface{}, error) {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		if required {
			return nil, fmt.Errorf("%s is required", field)
		}
		return nil, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil || out == nil {
		return nil, fmt.Errorf("%s must be a valid JSON object: %v", field, err)
	}
	return out, nil
}

func (r *CertificateRotationTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateRotationTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg, err := parseTargetJSONObject("target_config_json", data.TargetConfigJSON, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid target_config_json", err.Error())
		return
	}
	window, err := parseTargetJSONObject("deployment_window_json", data.DeploymentWindowJSON, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment_window_json", err.Error())
		return
	}
	enabled := data.Enabled.ValueBool()
	sync := data.SyncOnRenewal.ValueBool()
	created, err := r.client.CreateCertificateRotationTarget(data.CertificateID.ValueString(), &mazevault.CreateCertificateRotationTargetRequest{
		TargetType:            data.TargetType.ValueString(),
		AgentID:               optionalString(data.AgentID),
		IntegrationID:         optionalString(data.IntegrationID),
		DatabaseIntegrationID: optionalString(data.DatabaseIntegrationID),
		ContentMode:           data.ContentMode.ValueString(),
		TargetConfig:          cfg,
		Enabled:               &enabled,
		SyncOnRenewal:         &sync,
		Priority:              int(data.Priority.ValueInt64()),
		DeploymentWindow:      window,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create CertificateRotationTarget Error", err.Error())
		return
	}
	data.ID = types.StringValue(created.ID)
	data.LastSyncStatus = types.StringValue(created.LastSyncStatus)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateRotationTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateRotationTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	target, err := r.client.GetCertificateRotationTarget(data.CertificateID.ValueString(), data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read CertificateRotationTarget Error", err.Error())
		return
	}
	data.TargetType = types.StringValue(target.TargetType)
	data.ContentMode = types.StringValue(target.ContentMode)
	data.Enabled = types.BoolValue(target.Enabled)
	data.SyncOnRenewal = types.BoolValue(target.SyncOnRenewal)
	data.Priority = types.Int64Value(int64(target.Priority))
	data.LastSyncStatus = types.StringValue(target.LastSyncStatus)
	if target.AgentID != nil {
		data.AgentID = types.StringValue(*target.AgentID)
	}
	if target.IntegrationID != nil {
		data.IntegrationID = types.StringValue(*target.IntegrationID)
	}
	if target.DatabaseIntegrationID != nil {
		data.DatabaseIntegrationID = types.StringValue(*target.DatabaseIntegrationID)
	}
	// Round-trip the JSON documents only when the backend returned content;
	// the backend omits empty union members, so a semantically equal document
	// may differ textually. Keep the practitioner's text when it parses to the
	// same object to avoid perpetual diffs.
	if target.TargetConfig != nil && !jsonEquivalent(data.TargetConfigJSON.ValueString(), target.TargetConfig) {
		if b, err := json.Marshal(target.TargetConfig); err == nil {
			data.TargetConfigJSON = types.StringValue(string(b))
		}
	}
	if target.DeploymentWindow != nil {
		if !jsonEquivalent(data.DeploymentWindowJSON.ValueString(), target.DeploymentWindow) {
			if b, err := json.Marshal(target.DeploymentWindow); err == nil {
				data.DeploymentWindowJSON = types.StringValue(string(b))
			}
		}
	} else if !data.DeploymentWindowJSON.IsNull() {
		// The API no longer returns a window: clear it in state so a window that
		// was removed out of band (or cleared) does not linger and cause a
		// perpetual diff.
		data.DeploymentWindowJSON = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// jsonEquivalent reports whether text decodes to the same object as obj after
// dropping empty members (the backend serialises omitempty).
func jsonEquivalent(text string, obj map[string]interface{}) bool {
	if text == "" {
		return false
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return false
	}
	a, _ := json.Marshal(pruneEmpty(parsed))
	b, _ := json.Marshal(pruneEmpty(obj))
	return string(a) == string(b)
}

func pruneEmpty(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, val := range t {
			p := pruneEmpty(val)
			switch pv := p.(type) {
			case nil:
				continue
			case string:
				if pv == "" {
					continue
				}
			case bool:
				if !pv {
					continue
				}
			case float64:
				if pv == 0 {
					continue
				}
			case map[string]interface{}:
				if len(pv) == 0 {
					continue
				}
			case []interface{}:
				if len(pv) == 0 {
					continue
				}
			}
			out[k] = p
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, val := range t {
			out = append(out, pruneEmpty(val))
		}
		return out
	default:
		return v
	}
}

func (r *CertificateRotationTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CertificateRotationTargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg, err := parseTargetJSONObject("target_config_json", data.TargetConfigJSON, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid target_config_json", err.Error())
		return
	}
	window, err := parseTargetJSONObject("deployment_window_json", data.DeploymentWindowJSON, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment_window_json", err.Error())
		return
	}
	if window == nil {
		window = map[string]interface{}{} // empty object clears the window on the backend
	}
	enabled := data.Enabled.ValueBool()
	sync := data.SyncOnRenewal.ValueBool()
	priority := int(data.Priority.ValueInt64())
	updated, err := r.client.UpdateCertificateRotationTarget(data.CertificateID.ValueString(), data.ID.ValueString(), &mazevault.UpdateCertificateRotationTargetRequest{
		ContentMode:           data.ContentMode.ValueString(),
		TargetConfig:          cfg,
		AgentID:               optionalString(data.AgentID),
		IntegrationID:         optionalString(data.IntegrationID),
		DatabaseIntegrationID: optionalString(data.DatabaseIntegrationID),
		Enabled:               &enabled,
		SyncOnRenewal:         &sync,
		Priority:              &priority,
		// Always send the window (as a pointer) so an empty object actually
		// reaches the backend and clears a previously configured window.
		DeploymentWindow: &window,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update CertificateRotationTarget Error", err.Error())
		return
	}
	data.LastSyncStatus = types.StringValue(updated.LastSyncStatus)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateRotationTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateRotationTargetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCertificateRotationTarget(data.CertificateID.ValueString(), data.ID.ValueString()); err != nil {
		if mazevault.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete CertificateRotationTarget Error", err.Error())
	}
}
