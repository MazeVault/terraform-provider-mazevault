package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RenewalPolicyResource{}
var _ resource.ResourceWithImportState = &RenewalPolicyResource{}

func NewRenewalPolicyResource() resource.Resource { return &RenewalPolicyResource{} }

type RenewalPolicyResource struct{ client *mazevault.Client }

type RenewalPolicyModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	LeadDays         types.Int64  `tfsdk:"lead_days"`
	KeyReuseEnabled  types.Bool   `tfsdk:"key_reuse_enabled"`
	AutoApprove      types.Bool   `tfsdk:"auto_approve"`
	NotifyEmails     types.String `tfsdk:"notify_emails"`
	ValidityDuration types.Int64  `tfsdk:"validity_duration"`
}

func (r *RenewalPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_renewal_policy"
}

func (r *RenewalPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a certificate renewal policy for an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the renewal policy.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization that owns this policy.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Policy name (max 100 characters).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Optional human-readable description of the policy.",
			},
			"lead_days": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of days before expiry to trigger renewal (1–365).",
			},
			"key_reuse_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When true, the existing private key is reused for the renewed certificate instead of generating a new key pair.",
			},
			"auto_approve": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether renewal requests created by this policy are automatically approved without manual intervention.",
			},
			"notify_emails": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Comma-separated list of email addresses to notify when a renewal is triggered (e.g. `ops@example.com,security@example.com`).",
			},
			"validity_duration": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Requested certificate validity in days for renewed certificates. 0 means use the same validity as the previous certificate.",
			},
		},
	}
}

func (r *RenewalPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RenewalPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RenewalPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRenewalPolicy(data.OrganizationID.ValueString(), &mazevault.CreateRenewalPolicyRequest{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		LeadDays:         int(data.LeadDays.ValueInt64()),
		KeyReuseEnabled:  data.KeyReuseEnabled.ValueBool(),
		AutoApprove:      data.AutoApprove.ValueBool(),
		NotifyEmails:     data.NotifyEmails.ValueString(),
		ValidityDuration: int(data.ValidityDuration.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Renewal Policy Error", fmt.Sprintf("Unable to create renewal policy: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Description = types.StringValue(created.Description)
	data.KeyReuseEnabled = types.BoolValue(created.KeyReuseEnabled)
	data.AutoApprove = types.BoolValue(created.AutoApprove)
	data.NotifyEmails = types.StringValue(created.NotifyEmails)
	data.ValidityDuration = types.Int64Value(int64(created.ValidityDuration))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RenewalPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RenewalPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := r.client.GetRenewalPolicy(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Renewal Policy Error", fmt.Sprintf("Unable to read renewal policy: %s", err))
		return
	}
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(policy.Name)
	data.Description = types.StringValue(policy.Description)
	data.LeadDays = types.Int64Value(int64(policy.LeadDays))
	data.KeyReuseEnabled = types.BoolValue(policy.KeyReuseEnabled)
	data.AutoApprove = types.BoolValue(policy.AutoApprove)
	data.NotifyEmails = types.StringValue(policy.NotifyEmails)
	data.ValidityDuration = types.Int64Value(int64(policy.ValidityDuration))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RenewalPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RenewalPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := r.client.UpdateRenewalPolicy(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateRenewalPolicyRequest{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		LeadDays:         int(plan.LeadDays.ValueInt64()),
		KeyReuseEnabled:  plan.KeyReuseEnabled.ValueBool(),
		AutoApprove:      plan.AutoApprove.ValueBool(),
		NotifyEmails:     plan.NotifyEmails.ValueString(),
		ValidityDuration: int(plan.ValidityDuration.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Renewal Policy Error", fmt.Sprintf("Unable to update renewal policy: %s", err))
		return
	}
	plan.ID = state.ID
	plan.Description = types.StringValue(updated.Description)
	plan.KeyReuseEnabled = types.BoolValue(updated.KeyReuseEnabled)
	plan.AutoApprove = types.BoolValue(updated.AutoApprove)
	plan.NotifyEmails = types.StringValue(updated.NotifyEmails)
	plan.ValidityDuration = types.Int64Value(int64(updated.ValidityDuration))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RenewalPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RenewalPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRenewalPolicy(data.OrganizationID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Renewal Policy Error", fmt.Sprintf("Unable to delete renewal policy: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<organization_id>/<resource_id>"
func (r *RenewalPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<organization_id>/<resource_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
