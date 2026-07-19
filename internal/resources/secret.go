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

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithConfigure = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

// NewSecretResource is a helper function to simplify the provider implementation.
func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

// SecretResource is the resource implementation.
type SecretResource struct {
	client *mazevault.Client
}

// SecretResourceModel maps the resource schema data.
type SecretResourceModel struct {
	ID          types.String         `tfsdk:"id"`
	ProjectID   types.String         `tfsdk:"project_id"`
	Key         types.String         `tfsdk:"key"`
	Value       types.String         `tfsdk:"value"`
	ValueWO     types.String         `tfsdk:"value_wo"`
	ValueWOVer  types.Int64          `tfsdk:"value_wo_version"`
	Environment types.String         `tfsdk:"environment"`
	TTLHours    types.Int64          `tfsdk:"ttl_hours"`
	Metadata    types.Map            `tfsdk:"metadata"`
	Version     types.Int64          `tfsdk:"version"`
	Status      types.String         `tfsdk:"status"`
	CreatedAt   types.String         `tfsdk:"created_at"`
	Rotation    *RotationModel       `tfsdk:"rotation"`
	Generate    *SecretGenerateModel `tfsdk:"generate"`
}

// SecretGenerateModel requests server-side value generation.
type SecretGenerateModel struct {
	Length         types.Int64  `tfsdk:"length"`
	IncludeUpper   types.Bool   `tfsdk:"include_upper"`
	IncludeLower   types.Bool   `tfsdk:"include_lower"`
	IncludeDigits  types.Bool   `tfsdk:"include_digits"`
	IncludeSymbols types.Bool   `tfsdk:"include_symbols"`
	ExcludedChars  types.String `tfsdk:"excluded_chars"`
}

// toSDK converts the generation model to the SDK spec.
func (g *SecretGenerateModel) toSDK() *mazevault.SecretGenerate {
	spec := &mazevault.SecretGenerate{
		Length:        int(g.Length.ValueInt64()),
		ExcludedChars: g.ExcludedChars.ValueString(),
	}
	if !g.IncludeUpper.IsNull() && !g.IncludeUpper.IsUnknown() {
		v := g.IncludeUpper.ValueBool()
		spec.IncludeUpper = &v
	}
	if !g.IncludeLower.IsNull() && !g.IncludeLower.IsUnknown() {
		v := g.IncludeLower.ValueBool()
		spec.IncludeLower = &v
	}
	if !g.IncludeDigits.IsNull() && !g.IncludeDigits.IsUnknown() {
		v := g.IncludeDigits.ValueBool()
		spec.IncludeDigits = &v
	}
	if !g.IncludeSymbols.IsNull() && !g.IncludeSymbols.IsUnknown() {
		v := g.IncludeSymbols.ValueBool()
		spec.IncludeSymbols = &v
	}
	return spec
}

type RotationModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	Schedule           types.String `tfsdk:"schedule"`
	IntervalDays       types.Int64  `tfsdk:"interval_days"`
	NotificationEmails types.List   `tfsdk:"notification_emails"`
}

// Metadata returns the resource type name.
func (r *SecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for the resource.
func (r *SecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MazeVault secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Secret UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Description: "Parent project ID",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description: "Secret key/name",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Legacy secret value input. This is marked sensitive, but Terraform still stores it in state. Prefer `value_wo` with Terraform 1.11+ or `generate` for new secrets.",
				Optional:    true,
				Sensitive:   true,
			},
			"value_wo": schema.StringAttribute{
				Description: "Write-only secret value input for Terraform 1.11+ ephemeral values. This value is read from configuration and is not stored in Terraform plan or state.",
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"value_wo_version": schema.Int64Attribute{
				Description: "Stored trigger for resending `value_wo`. Increment this when the write-only value should update the MazeVault secret.",
				Optional:    true,
			},
			"environment": schema.StringAttribute{
				Description: "Environment",
				Required:    true,
			},
			"ttl_hours": schema.Int64Attribute{
				Description: "Time-to-live in hours",
				Optional:    true,
			},
			"metadata": schema.MapAttribute{
				Description: "Key-value metadata",
				Optional:    true,
				ElementType: types.StringType,
			},
			"version": schema.Int64Attribute{
				Description: "Current version",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Secret status",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rotation": schema.SingleNestedAttribute{
				Description: "Rotation configuration",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Enable automatic rotation",
						Required:    true,
					},
					"schedule": schema.StringAttribute{
						Description: "Rotation schedule (cron)",
						Optional:    true,
					},
					"interval_days": schema.Int64Attribute{
						Description: "Rotation interval in days",
						Optional:    true,
					},
					"notification_emails": schema.ListAttribute{
						Description: "Emails to notify on rotation failure",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"generate": schema.SingleNestedAttribute{
				Description: "Request server-side value generation. Mutually exclusive with `value` and `value_wo`. The plaintext is generated inside MazeVault and, in Orchestrator Mode, offloaded to the external provider.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"length": schema.Int64Attribute{
						Description: "Length of the generated value. Defaults to 40.",
						Optional:    true,
					},
					"include_upper": schema.BoolAttribute{
						Description: "Include uppercase letters. Defaults to true.",
						Optional:    true,
					},
					"include_lower": schema.BoolAttribute{
						Description: "Include lowercase letters. Defaults to true.",
						Optional:    true,
					},
					"include_digits": schema.BoolAttribute{
						Description: "Include digits. Defaults to true.",
						Optional:    true,
					},
					"include_symbols": schema.BoolAttribute{
						Description: "Include DB/URL-safe symbols. Defaults to true.",
						Optional:    true,
					},
					"excluded_chars": schema.StringAttribute{
						Description: "Characters to exclude from the generated value.",
						Optional:    true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config SecretResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata := make(map[string]string)
	if !plan.Metadata.IsNull() {
		diags = plan.Metadata.ElementsAs(ctx, &metadata, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var rotation *mazevault.RotationConfig
	if plan.Rotation != nil {
		rotation = &mazevault.RotationConfig{
			Enabled:              plan.Rotation.Enabled.ValueBool(),
			Schedule:             plan.Rotation.Schedule.ValueString(),
			RotationIntervalDays: int(plan.Rotation.IntervalDays.ValueInt64()),
		}
		if !plan.Rotation.NotificationEmails.IsNull() {
			var emails []string
			diags = plan.Rotation.NotificationEmails.ElementsAs(ctx, &emails, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			rotation.NotificationEmails = emails
		}
	}

	// Exactly one of value / value_wo / generate must be supplied.
	hasValue := !plan.Value.IsNull() && plan.Value.ValueString() != ""
	hasValueWO := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown() && config.ValueWO.ValueString() != ""
	hasGenerate := plan.Generate != nil
	if sourceCount(hasValue, hasValueWO, hasGenerate) != 1 {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"Exactly one of 'value', 'value_wo', or 'generate' must be set.",
		)
		return
	}
	if hasValueWO && (plan.ValueWOVer.IsNull() || plan.ValueWOVer.IsUnknown()) {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"'value_wo_version' must be set when 'value_wo' is configured.",
		)
		return
	}
	if !hasValueWO && !plan.ValueWOVer.IsNull() && !plan.ValueWOVer.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"'value_wo_version' can only be set when 'value_wo' is configured.",
		)
		return
	}

	var secret *mazevault.CreateSecretResponse
	var err error
	if hasGenerate {
		secret, err = r.client.CreateSecretGenerated(
			plan.ProjectID.ValueString(),
			plan.Key.ValueString(),
			plan.Environment.ValueString(),
			int(plan.TTLHours.ValueInt64()),
			metadata,
			rotation,
			plan.Generate.toSDK(),
		)
	} else {
		value := plan.Value.ValueString()
		if hasValueWO {
			value = config.ValueWO.ValueString()
		}
		secret, err = r.client.CreateSecret(
			plan.ProjectID.ValueString(),
			plan.Key.ValueString(),
			value,
			plan.Environment.ValueString(),
			int(plan.TTLHours.ValueInt64()),
			metadata,
			rotation,
		)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating secret",
			"Could not create secret, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(secret.ID)
	plan.Status = types.StringValue(secret.Status)
	plan.Version = types.Int64Value(int64(secret.Version))

	// The create endpoint does not return created_at; fetch the full secret.
	full, err := r.client.GetSecretByID(secret.ID)
	if err == nil && full != nil {
		plan.CreatedAt = types.StringValue(full.CreatedAt.String())
		if plan.Version.IsNull() || plan.Version.IsUnknown() {
			plan.Version = types.Int64Value(int64(full.Version))
		}
	} else {
		// Fallback: set null so Terraform doesn't produce a permanent diff.
		// UseStateForUnknown preserves the previous known value on subsequent reads.
		plan.CreatedAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecretByID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading secret",
			"Could not read secret ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if secret == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ProjectID = types.StringValue(secret.ProjectID)
	state.Key = types.StringValue(secret.Key)
	state.Environment = types.StringValue(secret.Environment)
	state.TTLHours = types.Int64Value(int64(secret.TTLHours))

	if len(secret.Metadata) > 0 {
		metadata, diags := types.MapValueFrom(ctx, types.StringType, secret.Metadata)
		resp.Diagnostics.Append(diags...)
		state.Metadata = metadata
	} else {
		state.Metadata = types.MapNull(types.StringType)
	}

	state.Version = types.Int64Value(int64(secret.Version))
	state.Status = types.StringValue(secret.Status)
	state.CreatedAt = types.StringValue(secret.CreatedAt.String())
	if secret.TTLHours > 0 {
		state.TTLHours = types.Int64Value(int64(secret.TTLHours))
	} else {
		state.TTLHours = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SecretResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var config SecretResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	metadata := make(map[string]string)
	if !plan.Metadata.IsNull() {
		diags = plan.Metadata.ElementsAs(ctx, &metadata, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	hasValue := !plan.Value.IsNull() && plan.Value.ValueString() != ""
	hasValueWO := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown() && config.ValueWO.ValueString() != ""
	if sourceCount(hasValue, hasValueWO, plan.Generate != nil) != 1 {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"Exactly one of 'value', 'value_wo', or 'generate' must be set.",
		)
		return
	}
	if hasValueWO && (plan.ValueWOVer.IsNull() || plan.ValueWOVer.IsUnknown()) {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"'value_wo_version' must be set when 'value_wo' is configured.",
		)
		return
	}
	if !hasValueWO && !plan.ValueWOVer.IsNull() && !plan.ValueWOVer.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Secret Configuration",
			"'value_wo_version' can only be set when 'value_wo' is configured.",
		)
		return
	}

	var value *string
	if hasValue {
		v := plan.Value.ValueString()
		value = &v
	} else if hasValueWO && shouldSendWriteOnlyValue(state.ValueWOVer, plan.ValueWOVer) {
		v := config.ValueWO.ValueString()
		value = &v
	}

	secret, err := r.client.UpdateSecretWithRequest(plan.ID.ValueString(), mazevault.UpdateSecretRequest{
		Value:    value,
		TTLHours: int(plan.TTLHours.ValueInt64()),
		Metadata: metadata,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating secret",
			"Could not update secret, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Version = types.Int64Value(int64(secret.Version))
	plan.Status = types.StringValue(secret.Status)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state.
func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecretResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecret(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting secret",
			"Could not delete secret, unexpected error: "+err.Error(),
		)
		return
	}
}

func sourceCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func shouldSendWriteOnlyValue(previous, next types.Int64) bool {
	if next.IsNull() || next.IsUnknown() {
		return false
	}
	if previous.IsNull() || previous.IsUnknown() {
		return true
	}
	return previous.ValueInt64() != next.ValueInt64()
}

// ImportState imports the resource into Terraform state.
func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
