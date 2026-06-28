package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SyncRuleResource{}
var _ resource.ResourceWithImportState = &SyncRuleResource{}

// NewSyncRuleResource returns a new mazevault_sync_rule resource.
func NewSyncRuleResource() resource.Resource { return &SyncRuleResource{} }

type SyncRuleResource struct{ client *mazevault.Client }

type SyncRuleModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	ProjectID            types.String `tfsdk:"project_id"`
	IntegrationID        types.String `tfsdk:"integration_id"`
	TargetEnvironment    types.String `tfsdk:"target_environment"`
	SourcePath           types.String `tfsdk:"source_path"`
	KeyTransformTemplate types.String `tfsdk:"key_transform_template"`
	ConflictStrategy     types.String `tfsdk:"conflict_strategy"`
	SyncDirection        types.String `tfsdk:"sync_direction"`
	SyncMode             types.String `tfsdk:"sync_mode"`
}

func (r *SyncRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_rule"
}

func (r *SyncRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a synchronization rule between MazeVault and an external secret provider (KV, GitHub, GitLab, Azure KV, etc.).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the sync rule.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the sync rule.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project this sync rule belongs to.",
			},
			"integration_id": schema.StringAttribute{
				Required:    true,
				Description: "The integration (Azure KV, GitHub Actions, etc.) used for synchronization.",
			},
			"target_environment": schema.StringAttribute{
				Optional:    true,
				Description: "The MazeVault environment to synchronize secrets into (e.g. production, staging).",
			},
			"source_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path prefix in the external system (e.g. /secret/data/myapp or secrets/APP_).",
			},
			"key_transform_template": schema.StringAttribute{
				Optional:    true,
				Description: "Template for transforming external key names. Use {{key}} as placeholder (e.g. app-{{environment}}-{{key}}).",
			},
			"conflict_strategy": schema.StringAttribute{
				Optional:    true,
				Description: "How to resolve conflicts: manual_resolution, mazevault_wins, external_wins, most_recent_wins. Defaults to manual_resolution.",
			},
			"sync_direction": schema.StringAttribute{
				Optional:    true,
				Description: "Direction of synchronization: pull (external → MazeVault), push (MazeVault → external), bidirectional. Defaults to pull.",
			},
			"sync_mode": schema.StringAttribute{
				Optional:    true,
				Description: "Sync granularity: incremental (changes only) or full_sync (reconcile all). Defaults to incremental.",
			},
		},
	}
}

func (r *SyncRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SyncRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SyncRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.CreateSyncRule(&mazevault.CreateSyncRuleRequest{
		Name:                 data.Name.ValueString(),
		IntegrationID:        data.IntegrationID.ValueString(),
		ProjectID:            data.ProjectID.ValueString(),
		TargetEnvironment:    data.TargetEnvironment.ValueString(),
		SourcePath:           data.SourcePath.ValueString(),
		KeyTransformTemplate: data.KeyTransformTemplate.ValueString(),
		ConflictStrategy:     data.ConflictStrategy.ValueString(),
		SyncDirection:        data.SyncDirection.ValueString(),
		SyncMode:             data.SyncMode.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create sync rule: %s", err))
		return
	}

	data.ID = types.StringValue(rule.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SyncRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetSyncRule(data.ProjectID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sync rule %s: %s", data.ID.ValueString(), err))
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(rule.Name)
	data.IntegrationID = types.StringValue(rule.IntegrationID)
	data.TargetEnvironment = types.StringValue(rule.TargetEnvironment)
	data.SourcePath = types.StringValue(rule.SourcePath)
	data.KeyTransformTemplate = types.StringValue(rule.KeyTransformTemplate)
	data.ConflictStrategy = types.StringValue(rule.ConflictStrategy)
	data.SyncDirection = types.StringValue(rule.SyncDirection)
	data.SyncMode = types.StringValue(rule.SyncMode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SyncRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.UpdateSyncRule(data.ProjectID.ValueString(), data.ID.ValueString(), &mazevault.CreateSyncRuleRequest{
		Name:                 data.Name.ValueString(),
		IntegrationID:        data.IntegrationID.ValueString(),
		ProjectID:            data.ProjectID.ValueString(),
		TargetEnvironment:    data.TargetEnvironment.ValueString(),
		SourcePath:           data.SourcePath.ValueString(),
		KeyTransformTemplate: data.KeyTransformTemplate.ValueString(),
		ConflictStrategy:     data.ConflictStrategy.ValueString(),
		SyncDirection:        data.SyncDirection.ValueString(),
		SyncMode:             data.SyncMode.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update sync rule %s: %s", data.ID.ValueString(), err))
		return
	}

	data.Name = types.StringValue(rule.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SyncRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSyncRule(data.ProjectID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sync rule %s: %s", data.ID.ValueString(), err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<project_id>/<resource_id>"
func (r *SyncRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<project_id>/<resource_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
