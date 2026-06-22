package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RotationTemplateResource{}

// NewRotationTemplateResource returns a new mazevault_rotation_template resource.
func NewRotationTemplateResource() resource.Resource { return &RotationTemplateResource{} }

type RotationTemplateResource struct{ client *mazevault.Client }

type RotationTemplateModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	OrganizationID       types.String `tfsdk:"organization_id"`
	IsDefault            types.Bool   `tfsdk:"is_default"`
	RotationIntervalDays types.Int64  `tfsdk:"rotation_interval_days"`
	LeadTimeDays         types.Int64  `tfsdk:"lead_time_days"`
	GracePeriodDays      types.Int64  `tfsdk:"grace_period_days"`
	MaxRetryAttempts     types.Int64  `tfsdk:"max_retry_attempts"`
	TimeoutMinutes       types.Int64  `tfsdk:"timeout_minutes"`
}

func (r *RotationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_template"
}

func (r *RotationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a reusable rotation policy template. Templates can be applied to rotation configs to inherit consistent policy settings across secrets and certificates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the rotation template.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the template.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional description of what this template is for.",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Description: "Scope the template to a specific organization. Omit to apply globally.",
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is the default template. Only one template per organization can be the default.",
			},
			"rotation_interval_days": schema.Int64Attribute{
				Optional:    true,
				Description: "How frequently (in days) the secret or certificate should be rotated.",
			},
			"lead_time_days": schema.Int64Attribute{
				Optional:    true,
				Description: "Days before expiry to start the rotation. Applies to certificate renewal.",
			},
			"grace_period_days": schema.Int64Attribute{
				Optional:    true,
				Description: "Days after rotation to keep the previous value active for rollback.",
			},
			"max_retry_attempts": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of retry attempts on failure.",
			},
			"timeout_minutes": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum execution time in minutes before the rotation is considered timed out.",
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

	tmpl, err := r.client.CreateRotationTemplate(&mazevault.CreateRotationTemplateRequest{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		OrganizationID:       data.OrganizationID.ValueString(),
		IsDefault:            data.IsDefault.ValueBool(),
		RotationIntervalDays: int(data.RotationIntervalDays.ValueInt64()),
		LeadTimeDays:         int(data.LeadTimeDays.ValueInt64()),
		GracePeriodDays:      int(data.GracePeriodDays.ValueInt64()),
		MaxRetryAttempts:     int(data.MaxRetryAttempts.ValueInt64()),
		TimeoutMinutes:       int(data.TimeoutMinutes.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation template: %s", err))
		return
	}

	data.ID = types.StringValue(tmpl.ID)
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
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

	data.Name = types.StringValue(tmpl.Name)
	data.Description = types.StringValue(tmpl.Description)
	data.OrganizationID = types.StringValue(tmpl.OrganizationID)
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
	data.RotationIntervalDays = types.Int64Value(int64(tmpl.RotationIntervalDays))
	data.LeadTimeDays = types.Int64Value(int64(tmpl.LeadTimeDays))
	data.GracePeriodDays = types.Int64Value(int64(tmpl.GracePeriodDays))
	data.MaxRetryAttempts = types.Int64Value(int64(tmpl.MaxRetryAttempts))
	data.TimeoutMinutes = types.Int64Value(int64(tmpl.TimeoutMinutes))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RotationTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.UpdateRotationTemplate(data.ID.ValueString(), &mazevault.CreateRotationTemplateRequest{
		Name:                 data.Name.ValueString(),
		Description:          data.Description.ValueString(),
		OrganizationID:       data.OrganizationID.ValueString(),
		IsDefault:            data.IsDefault.ValueBool(),
		RotationIntervalDays: int(data.RotationIntervalDays.ValueInt64()),
		LeadTimeDays:         int(data.LeadTimeDays.ValueInt64()),
		GracePeriodDays:      int(data.GracePeriodDays.ValueInt64()),
		MaxRetryAttempts:     int(data.MaxRetryAttempts.ValueInt64()),
		TimeoutMinutes:       int(data.TimeoutMinutes.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation template %s: %s", data.ID.ValueString(), err))
		return
	}

	data.IsDefault = types.BoolValue(tmpl.IsDefault)
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
