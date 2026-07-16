package resources

import (
	"context"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &HSMDeviceResource{}

func NewHSMDeviceResource() resource.Resource { return &HSMDeviceResource{} }

type HSMDeviceResource struct{ client *mazevault.Client }

type HSMDeviceModel struct {
	ID           types.String `tfsdk:"id"`
	OrgID        types.String `tfsdk:"organization_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Provider     types.String `tfsdk:"provider_type"`
	Config       types.Map    `tfsdk:"config"`
	Environments types.List   `tfsdk:"environments"`
	AppliesTo    types.List   `tfsdk:"applies_to"`
	Priority     types.Int64  `tfsdk:"priority"`
	Status       types.String `tfsdk:"status"`
	HealthStatus types.String `tfsdk:"health_status"`
	IsKEKDevice  types.Bool   `tfsdk:"is_kek_device"`
}

func (r *HSMDeviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hsm_device"
}

func (r *HSMDeviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an organization-level HSM/KMS device for centralized key protection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":          schema.StringAttribute{Required: true},
			"description":   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
			"provider_type": schema.StringAttribute{Required: true},
			"config": schema.MapAttribute{
				Required: true, Sensitive: true, ElementType: types.StringType,
				Description: "Provider-specific connection config (encrypted at rest).",
			},
			"environments":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"applies_to":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"priority":      schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(100)},
			"status":        schema.StringAttribute{Computed: true},
			"health_status": schema.StringAttribute{Computed: true},
			"is_kek_device": schema.BoolAttribute{Computed: true},
		},
	}
}

func (r *HSMDeviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *mazevault.Client")
		return
	}
	r.client = client
}

func (r *HSMDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HSMDeviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.extractConfig(ctx, &data, &resp.Diagnostics)
	envs := r.extractStringList(ctx, data.Environments, &resp.Diagnostics)
	appliesTo := r.extractStringList(ctx, data.AppliesTo, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateHSMDevice(data.OrgID.ValueString(), &mazevault.CreateHSMDeviceRequest{
		Name: data.Name.ValueString(), Description: data.Description.ValueString(),
		Provider: data.Provider.ValueString(), Config: cfg,
		Environments: envs, AppliesTo: appliesTo, Priority: int(data.Priority.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HSM device", err.Error())
		return
	}
	r.mapToModel(ctx, result, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HSMDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HSMDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.GetHSMDevice(data.OrgID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	r.mapToModel(ctx, result, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HSMDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HSMDeviceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.extractConfig(ctx, &data, &resp.Diagnostics)
	envs := r.extractStringList(ctx, data.Environments, &resp.Diagnostics)
	appliesTo := r.extractStringList(ctx, data.AppliesTo, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateHSMDevice(data.OrgID.ValueString(), data.ID.ValueString(), &mazevault.UpdateHSMDeviceRequest{
		Name: data.Name.ValueString(), Description: data.Description.ValueString(),
		Config: cfg, Environments: envs, AppliesTo: appliesTo, Priority: int(data.Priority.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update HSM device", err.Error())
		return
	}
	r.mapToModel(ctx, result, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HSMDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HSMDeviceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteHSMDevice(data.OrgID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete HSM device", err.Error())
	}
}

func (r *HSMDeviceResource) extractConfig(ctx context.Context, data *HSMDeviceModel, diags *diag.Diagnostics) map[string]interface{} {
	cfg := map[string]interface{}{}
	if !data.Config.IsNull() {
		m := map[string]string{}
		diags.Append(data.Config.ElementsAs(ctx, &m, false)...)
		for k, v := range m {
			cfg[k] = v
		}
	}
	return cfg
}

func (r *HSMDeviceResource) extractStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)
	return result
}

func (r *HSMDeviceResource) mapToModel(ctx context.Context, device *mazevault.HSMDevice, data *HSMDeviceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(device.ID)
	data.Name = types.StringValue(device.Name)
	data.Description = types.StringValue(device.Description)
	data.Provider = types.StringValue(device.Provider)
	data.Status = types.StringValue(device.Status)
	data.HealthStatus = types.StringValue(device.HealthStatus)
	data.IsKEKDevice = types.BoolValue(device.IsKEKDevice)
	data.Priority = types.Int64Value(int64(device.Priority))

	envList, d := types.ListValueFrom(ctx, types.StringType, device.Environments)
	diags.Append(d...)
	data.Environments = envList

	atList, d := types.ListValueFrom(ctx, types.StringType, device.AppliesTo)
	diags.Append(d...)
	data.AppliesTo = atList
}
