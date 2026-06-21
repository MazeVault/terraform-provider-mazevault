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

var _ resource.Resource = &ApprovalPolicyResource{}

func NewApprovalPolicyResource() resource.Resource { return &ApprovalPolicyResource{} }

type ApprovalPolicyResource struct{ client *mazevault.Client }

type ApprovalPolicyModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	Environments      types.List   `tfsdk:"environments"`
	RequiredApprovals types.Int64  `tfsdk:"required_approvals"`
	ApproversGroupID  types.String `tfsdk:"approvers_group_id"`
}

func (r *ApprovalPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_approval_policy"
}

func (r *ApprovalPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an approval policy for secret-rotation or certificate-issuance workflows within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the approval policy.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project that owns this policy.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Policy name.",
			},
			"environments": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of environment names this policy applies to. If empty, applies to all environments.",
			},
			"required_approvals": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Minimum number of approvals required before the request is executed.",
			},
			"approvers_group_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the group whose members can approve requests under this policy.",
			},
		},
	}
}

func (r *ApprovalPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApprovalPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ApprovalPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var environments []string
	resp.Diagnostics.Append(data.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateApprovalPolicy(data.ProjectID.ValueString(), &mazevault.CreateApprovalPolicyRequest{
		Name:              data.Name.ValueString(),
		Environments:      environments,
		RequiredApprovals: int(data.RequiredApprovals.ValueInt64()),
		ApproversGroupID:  data.ApproversGroupID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Approval Policy Error", fmt.Sprintf("Unable to create approval policy: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApprovalPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ApprovalPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := r.client.GetApprovalPolicy(data.ProjectID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Approval Policy Error", fmt.Sprintf("Unable to read approval policy: %s", err))
		return
	}
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(policy.Name)
	data.RequiredApprovals = types.Int64Value(int64(policy.RequiredApprovals))
	data.ApproversGroupID = types.StringValue(policy.ApproversGroupID)
	envList, diags := types.ListValueFrom(ctx, types.StringType, policy.Environments)
	resp.Diagnostics.Append(diags...)
	data.Environments = envList
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ApprovalPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ApprovalPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var environments []string
	resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateApprovalPolicy(state.ID.ValueString(), &mazevault.CreateApprovalPolicyRequest{
		Name:              plan.Name.ValueString(),
		Environments:      environments,
		RequiredApprovals: int(plan.RequiredApprovals.ValueInt64()),
		ApproversGroupID:  plan.ApproversGroupID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Approval Policy Error", fmt.Sprintf("Unable to update approval policy: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ApprovalPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteApprovalPolicy(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Approval Policy Error", fmt.Sprintf("Unable to delete approval policy: %s", err))
	}
}
