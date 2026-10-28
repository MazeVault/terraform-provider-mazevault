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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RotationTemplateResource{}
var _ resource.ResourceWithImportState = &RotationTemplateResource{}

// NewRotationTemplateResource returns a new mazevault_rotation_template resource.
func NewRotationTemplateResource() resource.Resource { return &RotationTemplateResource{} }

type RotationTemplateResource struct{ client *mazevault.Client }

type RotationTemplateModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	OrganizationID       types.String `tfsdk:"organization_id"`
	ProjectID            types.String `tfsdk:"project_id"`
	IsDefault            types.Bool   `tfsdk:"is_default"`
	RotationIntervalDays types.Int64  `tfsdk:"rotation_interval_days"`
	LeadTimeDays         types.Int64  `tfsdk:"lead_time_days"`
	GracePeriodDays      types.Int64  `tfsdk:"grace_period_days"`
	MaxRetryAttempts     types.Int64  `tfsdk:"max_retry_attempts"`
	RetryDelaySeconds    types.Int64  `tfsdk:"retry_delay_seconds"`
	TimeoutMinutes       types.Int64  `tfsdk:"timeout_minutes"`
	SecretCustodyMode    types.String `tfsdk:"secret_custody_mode"`
}

// buildRotationTemplateRequest maps the Terraform model to the SDK request payload.
func buildRotationTemplateRequest(data RotationTemplateModel) *mazevault.CreateRotationTemplateRequest {
	return &mazevault.CreateRotationTemplateRequest{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		OrganizationID:       data.OrganizationID.ValueString(),
		ProjectID:            data.ProjectID.ValueString(),
		IsDefault:            data.IsDefault.ValueBool(),
		RotationIntervalDays: int(data.RotationIntervalDays.ValueInt64()),
		LeadTimeDays:         int(data.LeadTimeDays.ValueInt64()),
		GracePeriodDays:      int(data.GracePeriodDays.ValueInt64()),
		MaxRetryAttempts:     int(data.MaxRetryAttempts.ValueInt64()),
		RetryDelaySeconds:    int(data.RetryDelaySeconds.ValueInt64()),
		TimeoutMinutes:       int(data.TimeoutMinutes.ValueInt64()),
		SecretCustodyMode:    data.SecretCustodyMode.ValueString(),
	}
}

// flattenRotationTemplate maps the server response onto the Terraform model. It is the
// single source of state mapping shared by Create, Read, and Update so their output can
// never drift. An empty project_id from the server is a null (not an empty string).
func flattenRotationTemplate(tmpl *mazevault.RotationConfigTemplate, data *RotationTemplateModel) {
	data.ID = types.StringValue(tmpl.ID)
	data.Name = types.StringValue(tmpl.Name)
	data.Description = types.StringValue(tmpl.Description)
	data.OrganizationID = types.StringValue(tmpl.OrganizationID)
	if tmpl.ProjectID == "" {
		data.ProjectID = types.StringNull()
	} else {
		data.ProjectID = types.StringValue(tmpl.ProjectID)
	}
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
	data.RotationIntervalDays = types.Int64Value(int64(tmpl.RotationIntervalDays))
	data.LeadTimeDays = types.Int64Value(int64(tmpl.LeadTimeDays))
	data.GracePeriodDays = types.Int64Value(int64(tmpl.GracePeriodDays))
	data.MaxRetryAttempts = types.Int64Value(int64(tmpl.MaxRetryAttempts))
	data.RetryDelaySeconds = types.Int64Value(int64(tmpl.RetryDelaySeconds))
	data.TimeoutMinutes = types.Int64Value(int64(tmpl.TimeoutMinutes))
	data.SecretCustodyMode = types.StringValue(tmpl.SecretCustodyMode)
}

func (r *RotationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_template"
}

func (r *RotationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a reusable rotation policy template — a named catalog entry of default rotation settings that operators can copy when creating rotation configs. Templates are a convenience catalog; they do not retroactively change rotation configs already created from them.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "The unique identifier of the rotation template.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the template.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional description of what this template is for.",
			},
			"organization_id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Organization that owns the template. Omit to use the authenticated organization; the server resolves and returns it.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optionally scope the template to a single project. Omit for an organization-wide template.",
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is the default template. Only one template per organization can be the default.",
			},
			"rotation_interval_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "How frequently (in days) the secret or certificate should be rotated.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"lead_time_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Days before expiry to start the rotation. Applies to certificate renewal.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"grace_period_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Days after rotation to keep the previous value active for rollback.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"max_retry_attempts": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of retry attempts on failure.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"retry_delay_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Delay in seconds between retry attempts on failure.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"timeout_minutes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum execution time in minutes before the rotation is considered timed out.",
				Validators:  []validator.Int64{NonNegativeInt64()},
			},
			"secret_custody_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Secret custody mode for rotations: local, transient, or delegated.",
			},
		},
	}
}

func (r *RotationTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RotationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RotationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.CreateRotationTemplate(buildRotationTemplateRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation template: %s", err))
		return
	}

	flattenRotationTemplate(tmpl, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RotationTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.GetRotationTemplate(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation template %s: %s", data.ID.ValueString(), err))
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenRotationTemplate(tmpl, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RotationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.UpdateRotationTemplate(data.ID.ValueString(), buildRotationTemplateRequest(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation template %s: %s", data.ID.ValueString(), err))
		return
	}

	flattenRotationTemplate(tmpl, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RotationTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRotationTemplate(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation template %s: %s", data.ID.ValueString(), err))
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *RotationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
