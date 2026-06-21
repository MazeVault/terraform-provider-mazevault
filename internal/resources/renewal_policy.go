package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RenewalPolicyResource{}

func NewRenewalPolicyResource() resource.Resource { return &RenewalPolicyResource{} }

type RenewalPolicyResource struct{ client *mazevault.Client }

type RenewalPolicyModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	Name             types.String `tfsdk:"name"`
	LeadDays         types.Int64  `tfsdk:"lead_days"`
	AutoApprove      types.Bool   `tfsdk:"auto_approve"`
	NotifyDaysBefore types.List   `tfsdk:"notify_days_before"`
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
				MarkdownDescription: "Policy name.",
			},
			"lead_days": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of days before expiry to trigger renewal.",
			},
			"auto_approve": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Whether renewals are automatically approved without manual intervention.",
			},
			"notify_days_before": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				MarkdownDescription: "Days before expiry at which to send notifications (e.g. [30, 14, 7]).",
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
		Name:        data.Name.ValueString(),
		LeadDays:    int(data.LeadDays.ValueInt64()),
		AutoApprove: data.AutoApprove.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Renewal Policy Error", fmt.Sprintf("Unable to create renewal policy: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
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
	data.LeadDays = types.Int64Value(int64(policy.LeadDays))
	data.AutoApprove = types.BoolValue(policy.AutoApprove)
	notifyList, diags := types.ListValueFrom(ctx, types.Int64Type, policy.NotifyDaysBefore)
	resp.Diagnostics.Append(diags...)
	data.NotifyDaysBefore = notifyList
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RenewalPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RenewalPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateRenewalPolicy(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateRenewalPolicyRequest{
		Name:        plan.Name.ValueString(),
		LeadDays:    int(plan.LeadDays.ValueInt64()),
		AutoApprove: plan.AutoApprove.ValueBool(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Renewal Policy Error", fmt.Sprintf("Unable to update renewal policy: %s", err))
		return
	}
	plan.ID = state.ID
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
