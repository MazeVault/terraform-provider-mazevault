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

var _ resource.Resource = &RotationConfigResource{}

func NewRotationConfigResource() resource.Resource {
	return &RotationConfigResource{}
}

type RotationConfigResource struct {
	client *mazevault.Client
}

type RotationConfigResourceModel struct {
	ID               types.String `tfsdk:"id"`
	SecretID         types.String `tfsdk:"secret_id"`
	Environment      types.String `tfsdk:"environment"`
	TTLHours         types.Int64  `tfsdk:"ttl_hours"`
	RotationStrategy types.String `tfsdk:"rotation_strategy"`
	// target_environment is round-tripped from the backend GET response.
	TargetEnvironment types.String `tfsdk:"target_environment"`
}

func (r *RotationConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rotation_config"
}

func (r *RotationConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secret_id": schema.StringAttribute{
				Required: true,
			},
			"environment": schema.StringAttribute{
				Required: true,
			},
			"ttl_hours": schema.Int64Attribute{
				Required: true,
			},
			"rotation_strategy": schema.StringAttribute{
				Optional: true,
			},
			"target_environment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Override the target environment for secret write-back. Defaults to the rotation config's own environment.",
			},
		},
	}
}

func (r *RotationConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *RotationConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateRotationConfigRequest{
		SecretID:             data.SecretID.ValueString(),
		Enabled:              true,
		RotationIntervalDays: int(data.TTLHours.ValueInt64() / 24),
	}

	cfg, err := r.client.CreateRotationConfig(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create rotation config: %s", err))
		return
	}

	data.ID = types.StringValue(cfg.ID)
	// Ensure Optional+Computed fields have a known value after create.
	if data.TargetEnvironment.IsNull() || data.TargetEnvironment.IsUnknown() {
		if cfg.TargetEnvironment != "" {
			data.TargetEnvironment = types.StringValue(cfg.TargetEnvironment)
		} else {
			data.TargetEnvironment = types.StringValue("")
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetRotationConfig(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read rotation config %s: %s", data.ID.ValueString(), err))
		return
	}
	if cfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.ID = types.StringValue(cfg.ID)
	data.SecretID = types.StringValue(cfg.SecretID)
	if cfg.TargetEnvironment != "" {
		data.TargetEnvironment = types.StringValue(cfg.TargetEnvironment)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.UpdateRotationConfigRequest{
		Enabled:              true,
		RotationIntervalDays: int(data.TTLHours.ValueInt64() / 24),
	}

	cfg, err := r.client.UpdateRotationConfig(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update rotation config %s: %s", data.ID.ValueString(), err))
		return
	}

	data.SecretID = types.StringValue(cfg.SecretID)
	if cfg.TargetEnvironment != "" {
		data.TargetEnvironment = types.StringValue(cfg.TargetEnvironment)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RotationConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RotationConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRotationConfig(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete rotation config %s: %s", data.ID.ValueString(), err))
	}
}
