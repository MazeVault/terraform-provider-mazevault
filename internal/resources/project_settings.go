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

var _ resource.Resource = &ProjectSettingsResource{}

func NewProjectSettingsResource() resource.Resource { return &ProjectSettingsResource{} }

type ProjectSettingsResource struct{ client *mazevault.Client }

type ProjectSettingsModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	RetentionDays       types.Int64  `tfsdk:"retention_days"`
	SlackChannel        types.String `tfsdk:"slack_channel"`
	OwnerEmail          types.String `tfsdk:"owner_email"`
	SyncEnabled         types.Bool   `tfsdk:"sync_enabled"`
	SyncIntervalMinutes types.Int64  `tfsdk:"sync_interval_minutes"`
}

func (r *ProjectSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_settings"
}

func (r *ProjectSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages configurable settings for a MazeVault project. There is exactly one settings object per project — creating this resource will upsert the project settings.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project whose settings are being managed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"retention_days": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of days to retain secret versions and audit logs.",
			},
			"slack_channel": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Slack channel name for project-level notifications (e.g. `#infra-alerts`).",
			},
			"owner_email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Email address of the project owner used for escalation notifications.",
			},
			"sync_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether automatic secret synchronisation is enabled for this project.",
			},
			"sync_interval_minutes": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Interval in minutes between automatic synchronisation runs.",
			},
		},
	}
}

func (r *ProjectSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateProjectSettings(data.ProjectID.ValueString(), &mazevault.UpdateProjectSettingsRequest{
		RetentionDays: int(data.RetentionDays.ValueInt64()),
		SlackChannel:  data.SlackChannel.ValueString(),
		OwnerEmail:    data.OwnerEmail.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Create Project Settings Error", fmt.Sprintf("Unable to apply project settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	settings, err := r.client.GetProjectSettings(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Project Settings Error", fmt.Sprintf("Unable to read project settings: %s", err))
		return
	}
	data.RetentionDays = types.Int64Value(int64(settings.RetentionDays))
	data.SlackChannel = types.StringValue(settings.SlackChannel)
	data.OwnerEmail = types.StringValue(settings.OwnerEmail)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateProjectSettings(plan.ProjectID.ValueString(), &mazevault.UpdateProjectSettingsRequest{
		RetentionDays: int(plan.RetentionDays.ValueInt64()),
		SlackChannel:  plan.SlackChannel.ValueString(),
		OwnerEmail:    plan.OwnerEmail.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Project Settings Error", fmt.Sprintf("Unable to update project settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Settings cannot be deleted, only reset. Removing from Terraform state only.
}
