package resources

import (
	"context"
	"fmt"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TokenResource{}
var _ resource.ResourceWithImportState = &TokenResource{}
var _ resource.ResourceWithUpgradeState = &TokenResource{}

func NewTokenResource() resource.Resource { return &TokenResource{} }

type TokenResource struct{ client *mazevault.Client }

type TokenModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	TokenProvider        types.String `tfsdk:"token_provider"`
	Value                types.String `tfsdk:"value"`
	ValueWO              types.String `tfsdk:"value_wo"`
	ValueWOVersion       types.Int64  `tfsdk:"value_wo_version"`
	ProjectID            types.String `tfsdk:"project_id"`
	ExpiresAt            types.String `tfsdk:"expires_at"`
	LeadTimeDays         types.Int64  `tfsdk:"lead_time_days"`
	Status               types.String `tfsdk:"status"`
	StorageMode          types.String `tfsdk:"storage_mode"`
	ExternalProviderType types.String `tfsdk:"external_provider_type"`
	IsExpired            types.Bool   `tfsdk:"is_expired"`
	DaysUntilExpiry      types.Int64  `tfsdk:"days_until_expiry"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

type tokenModelV0 struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Provider             types.String `tfsdk:"provider"`
	Value                types.String `tfsdk:"value"`
	ProjectID            types.String `tfsdk:"project_id"`
	ExpiresAt            types.String `tfsdk:"expires_at"`
	LeadTimeDays         types.Int64  `tfsdk:"lead_time_days"`
	Status               types.String `tfsdk:"status"`
	StorageMode          types.String `tfsdk:"storage_mode"`
	ExternalProviderType types.String `tfsdk:"external_provider_type"`
	IsExpired            types.Bool   `tfsdk:"is_expired"`
	DaysUntilExpiry      types.Int64  `tfsdk:"days_until_expiry"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

func (r *TokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *TokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		MarkdownDescription: "Manages an external API token (e.g. Jira, Signi) with a mandatory expiry date. " +
			"MazeVault stores the value encrypted (or offloaded in Orchestrator Mode), tracks expiry, and raises " +
			"incidents before expiry so it can be renewed manually in the external system.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the token.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name (unique within the project scope).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Free-text description.",
			},
			"token_provider": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Free-text provider label (e.g. jira, signi, generic). Defaults to 'generic'.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"value": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Legacy external API token value input. Terraform stores this sensitive value in state. Prefer `value_wo` for new tokens. Changing or removing it forces recreation.",
				DeprecationMessage:  "Use value_wo with value_wo_version for new tokens. The legacy value argument remains supported for compatibility but is stored in Terraform state.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"value_wo": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				MarkdownDescription: "Terraform 1.11+ write-only external API token value. The value is available only during apply and is never stored in Terraform plan or state. Use only when creating a new token.",
			},
			"value_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Stored replacement trigger paired with `value_wo`. Incrementing it recreates the token and deletes its existing deployment targets.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional project assignment. Changing it forces recreation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"expires_at": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Mandatory RFC-3339 expiration date of the external token.",
			},
			"lead_time_days": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Days before expiry to raise an incident (defaults to 30).",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Token lifecycle status (active, expired, revoked).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"storage_mode": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Where the value is stored (local or external offload).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"external_provider_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "External provider type when offloaded.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_expired": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the token has passed its expiration date.",
			},
			"days_until_expiry": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Days remaining until expiry (negative if expired).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 creation timestamp.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 last-update timestamp.",
			},
		},
	}
}

func tokenSchemaV0() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                     schema.StringAttribute{Computed: true},
		"name":                   schema.StringAttribute{Required: true},
		"description":            schema.StringAttribute{Optional: true},
		"provider":               schema.StringAttribute{Optional: true, Computed: true},
		"value":                  schema.StringAttribute{Required: true, Sensitive: true},
		"project_id":             schema.StringAttribute{Optional: true},
		"expires_at":             schema.StringAttribute{Required: true},
		"lead_time_days":         schema.Int64Attribute{Optional: true, Computed: true},
		"status":                 schema.StringAttribute{Computed: true},
		"storage_mode":           schema.StringAttribute{Computed: true},
		"external_provider_type": schema.StringAttribute{Computed: true},
		"is_expired":             schema.BoolAttribute{Computed: true},
		"days_until_expiry":      schema.Int64Attribute{Computed: true},
		"created_at":             schema.StringAttribute{Computed: true},
		"updated_at":             schema.StringAttribute{Computed: true},
	}}
}

func (r *TokenResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{Attributes: tokenSchemaV0().Attributes},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior tokenModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				upgraded := TokenModel{
					ID:                   prior.ID,
					Name:                 prior.Name,
					Description:          prior.Description,
					TokenProvider:        prior.Provider,
					Value:                prior.Value,
					ValueWO:              types.StringNull(),
					ValueWOVersion:       types.Int64Null(),
					ProjectID:            prior.ProjectID,
					ExpiresAt:            prior.ExpiresAt,
					LeadTimeDays:         prior.LeadTimeDays,
					Status:               prior.Status,
					StorageMode:          prior.StorageMode,
					ExternalProviderType: prior.ExternalProviderType,
					IsExpired:            prior.IsExpired,
					DaysUntilExpiry:      prior.DaysUntilExpiry,
					CreatedAt:            prior.CreatedAt,
					UpdatedAt:            prior.UpdatedAt,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}

func (r *TokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func reconcileOptionalString(apiValue string, current types.String) types.String {
	if apiValue != "" {
		return types.StringValue(apiValue)
	}
	if current.IsUnknown() || (!current.IsNull() && current.ValueString() != "") {
		return types.StringNull()
	}
	return current
}

func validateTokenValueSources(legacy, writeOnly types.String, version types.Int64, requireValue bool) string {
	if legacy.IsUnknown() || writeOnly.IsUnknown() {
		return "token value inputs must be known during apply"
	}

	hasLegacy := !legacy.IsNull() && legacy.ValueString() != ""
	hasWriteOnly := !writeOnly.IsNull() && writeOnly.ValueString() != ""
	if hasLegacy && hasWriteOnly {
		return "exactly one of 'value' or 'value_wo' may be set"
	}
	if requireValue && !hasLegacy && !hasWriteOnly {
		return "exactly one of 'value' or 'value_wo' must be set when creating a token"
	}
	if hasWriteOnly && (version.IsNull() || version.IsUnknown()) {
		return "'value_wo_version' must be set when 'value_wo' is configured"
	}
	if !hasWriteOnly && !version.IsNull() {
		return "'value_wo_version' can only be set when 'value_wo' is configured"
	}
	return ""
}

func (r *TokenResource) apply(data *TokenModel, t *mazevault.ManagedToken) {
	data.ID = types.StringValue(t.ID)
	data.Name = types.StringValue(t.Name)
	data.TokenProvider = types.StringValue(t.Provider)
	data.Status = types.StringValue(t.Status)
	data.StorageMode = types.StringValue(t.StorageMode)
	data.ExpiresAt = types.StringValue(t.ExpiresAt.Format(time.RFC3339))
	data.LeadTimeDays = types.Int64Value(int64(t.LeadTimeDays))
	data.IsExpired = types.BoolValue(t.IsExpired)
	data.DaysUntilExpiry = types.Int64Value(int64(t.DaysUntilExpiry))
	data.CreatedAt = types.StringValue(t.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(t.UpdatedAt.Format(time.RFC3339))
	data.Description = reconcileOptionalString(t.Description, data.Description)
	data.ProjectID = reconcileOptionalString(t.ProjectID, data.ProjectID)
	if t.ExternalProviderType != "" {
		data.ExternalProviderType = types.StringValue(t.ExternalProviderType)
	} else {
		data.ExternalProviderType = types.StringNull()
	}
}

func (r *TokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config TokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if message := validateTokenValueSources(data.Value, config.ValueWO, data.ValueWOVersion, true); message != "" {
		resp.Diagnostics.AddError("Invalid Token Configuration", message)
		return
	}

	value := data.Value.ValueString()
	if !config.ValueWO.IsNull() {
		value = config.ValueWO.ValueString()
	}

	createReq := &mazevault.CreateManagedTokenRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Provider:    data.TokenProvider.ValueString(),
		Value:       value,
		ProjectID:   data.ProjectID.ValueString(),
		ExpiresAt:   data.ExpiresAt.ValueString(),
	}
	if !data.LeadTimeDays.IsNull() && !data.LeadTimeDays.IsUnknown() {
		v := int(data.LeadTimeDays.ValueInt64())
		createReq.LeadTimeDays = &v
	}

	created, err := r.client.CreateManagedToken(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Create Token Error", fmt.Sprintf("Unable to create token: %s", err))
		return
	}
	r.apply(&data, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	t, err := r.client.GetManagedToken(data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Token Error", fmt.Sprintf("Unable to read token: %s", err))
		return
	}
	if t == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	r.apply(&data, t)
	// The legacy value remains in state for compatibility. The framework always
	// nulls value_wo before returning plan or state to Terraform.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state TokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config TokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if message := validateTokenValueSources(plan.Value, config.ValueWO, plan.ValueWOVersion, false); message != "" {
		resp.Diagnostics.AddError("Invalid Token Configuration", message)
		return
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	prov := plan.TokenProvider.ValueString()
	exp := plan.ExpiresAt.ValueString()
	updReq := &mazevault.UpdateManagedTokenRequest{
		Name:        &name,
		Description: &desc,
		Provider:    &prov,
		ExpiresAt:   &exp,
	}
	if !plan.LeadTimeDays.IsNull() && !plan.LeadTimeDays.IsUnknown() {
		v := int(plan.LeadTimeDays.ValueInt64())
		updReq.LeadTimeDays = &v
	}

	updated, err := r.client.UpdateManagedToken(state.ID.ValueString(), updReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Token Error", fmt.Sprintf("Unable to update token: %s", err))
		return
	}
	plan.Value = state.Value
	r.apply(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteManagedToken(data.ID.ValueString()); err != nil {
		if mazevault.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete Token Error", fmt.Sprintf("Unable to delete token: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *TokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
