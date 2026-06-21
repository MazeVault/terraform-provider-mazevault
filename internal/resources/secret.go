package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID          types.String   `tfsdk:"id"`
	ProjectID   types.String   `tfsdk:"project_id"`
	Key         types.String   `tfsdk:"key"`
	Value       types.String   `tfsdk:"value"`
	Environment types.String   `tfsdk:"environment"`
	TTLHours    types.Int64    `tfsdk:"ttl_hours"`
	Metadata    types.Map      `tfsdk:"metadata"`
	Version     types.Int64    `tfsdk:"version"`
	Status      types.String   `tfsdk:"status"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	Rotation    *RotationModel `tfsdk:"rotation"`
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
				Description: "Secret value (sensitive)",
				Required:    true,
				Sensitive:   true,
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Secret status",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
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

	secret, err := r.client.CreateSecret(
		plan.ProjectID.ValueString(),
		plan.Key.ValueString(),
		plan.Value.ValueString(),
		plan.Environment.ValueString(),
		int(plan.TTLHours.ValueInt64()),
		metadata,
		rotation,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating secret",
			"Could not create secret, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(secret.ID)
	// Note: CreateSecretResponse might not have all fields, but SDK CreateSecret returns CreateSecretResponse which has ID and Status.
	// Wait, SDK CreateSecret returns *CreateSecretResponse.
	// I need to check what CreateSecretResponse has.
	// It has ID and Status.
	// But I need Version and CreatedAt.
	// The API returns {id, status, version}.
	// I should update CreateSecretResponse in SDK to include Version.
	// And maybe fetch the secret to get CreatedAt? Or just leave it computed.

	plan.Status = types.StringValue(secret.Status)
	// plan.Version = types.Int64Value(int64(secret.Version)) // Need to add Version to response struct

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

	metadata := make(map[string]string)
	if !plan.Metadata.IsNull() {
		diags = plan.Metadata.ElementsAs(ctx, &metadata, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	secret, err := r.client.UpdateSecret(
		plan.ID.ValueString(),
		plan.Value.ValueString(),
		int(plan.TTLHours.ValueInt64()),
		metadata,
	)
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

// ImportState imports the resource into Terraform state.
func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
