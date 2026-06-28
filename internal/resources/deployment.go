package resources

import (
	"context"
	"fmt"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource { return &DeploymentResource{} }

type DeploymentResource struct{ client *mazevault.Client }

type DeploymentModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	OSType         types.String `tfsdk:"os_type"`
	AgentMode      types.String `tfsdk:"agent_mode"`
	DeploymentType types.String `tfsdk:"deployment_type"`
	GatewayURL     types.String `tfsdk:"gateway_url"`
	AutoUpdate     types.Bool   `tfsdk:"auto_update"`
	Token          types.String `tfsdk:"token"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a MazeVault agent deployment package. This generates the bootstrap token and configuration needed to register an agent in a specific project/environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the deployment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization this deployment belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the deployment (e.g. hostname or workload identifier).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"os_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target OS type: `linux`, `windows`, or `darwin`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"agent_mode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Agent operating mode: `gateway` or `sidecar`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"deployment_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deployment type: `agent`, `k8s`, or `docker`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"gateway_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Gateway URL for tunnel-mode agents.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"auto_update": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the agent should automatically update itself.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Bootstrap token used by the agent to register with the server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 timestamp when the deployment was created.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateDeployment(&mazevault.CreateDeploymentRequest{
		Name:           data.Name.ValueString(),
		OSType:         data.OSType.ValueString(),
		AgentMode:      data.AgentMode.ValueString(),
		DeploymentType: data.DeploymentType.ValueString(),
		GatewayURL:     data.GatewayURL.ValueString(),
		AutoUpdate:     data.AutoUpdate.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Deployment Error", fmt.Sprintf("Unable to create deployment: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Token = types.StringValue(created.Token)
	data.CreatedAt = types.StringValue(created.CreatedAt.Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	dep, err := r.client.GetDeployment(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Deployment Error", fmt.Sprintf("Unable to read deployment: %s", err))
		return
	}
	if dep == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(dep.Name)
	data.OSType = types.StringValue(dep.OSType)
	data.AgentMode = types.StringValue(dep.AgentMode)
	data.DeploymentType = types.StringValue(dep.DeploymentType)
	data.GatewayURL = types.StringValue(dep.GatewayURL)
	data.AutoUpdate = types.BoolValue(dep.AutoUpdate)
	data.CreatedAt = types.StringValue(dep.CreatedAt.Format(time.RFC3339))
	// Token is returned only on creation; preserve from state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Deployments are immutable; all fields use RequiresReplace.
}

func (r *DeploymentResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Deleting a deployment from state only — there is no API endpoint to delete deployment records.
}

// ImportState implements resource.ResourceWithImportState.
func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
