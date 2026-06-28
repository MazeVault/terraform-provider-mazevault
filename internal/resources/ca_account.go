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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CAAccountResource{}
var _ resource.ResourceWithImportState = &CAAccountResource{}

func NewCAAccountResource() resource.Resource { return &CAAccountResource{} }

type CAAccountResource struct{ client *mazevault.Client }

type CAAccountModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ProviderType   types.String `tfsdk:"provider_type"`
	APIKey         types.String `tfsdk:"api_key"`
	BaseURL        types.String `tfsdk:"base_url"`
	Status         types.String `tfsdk:"status"`
}

func (r *CAAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca_account"
}

func (r *CAAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connects an external Certificate Authority (CA) account to a MazeVault organization. Supports DigiCert, Venafi, ACME, and SmallStep.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the CA account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization this CA account belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for this CA account.",
			},
			"provider_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "CA provider type: `digicert`, `venafi`, `acme`, `smallstep`, or `internal`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key or credential for the CA provider.",
			},
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base URL for the CA provider API (required for self-hosted CAs).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current connection status of the CA account.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CAAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CAAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg := map[string]interface{}{}
	if v := data.APIKey.ValueString(); v != "" {
		cfg["api_key"] = v
	}
	if v := data.BaseURL.ValueString(); v != "" {
		cfg["base_url"] = v
	}
	created, err := r.client.ConnectCAAccount(data.OrganizationID.ValueString(), &mazevault.CreateCAAccountRequest{
		Name:         data.Name.ValueString(),
		ProviderType: data.ProviderType.ValueString(),
		Config:       cfg,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create CA Account Error", fmt.Sprintf("Unable to create CA account: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Status = types.StringValue(created.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	acc, err := r.client.GetCAAccount(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read CA Account Error", fmt.Sprintf("Unable to read CA account: %s", err))
		return
	}
	if acc == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(acc.Name)
	data.ProviderType = types.StringValue(acc.ProviderType)
	data.Status = types.StringValue(acc.Status)
	// api_key and base_url are write-only; preserve from state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CAAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfgU := map[string]interface{}{}
	if v := plan.APIKey.ValueString(); v != "" {
		cfgU["api_key"] = v
	}
	if v := plan.BaseURL.ValueString(); v != "" {
		cfgU["base_url"] = v
	}
	if _, err := r.client.UpdateCAAccount(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateCAAccountRequest{
		Name:         plan.Name.ValueString(),
		ProviderType: plan.ProviderType.ValueString(),
		Config:       cfgU,
	}); err != nil {
		resp.Diagnostics.AddError("Update CA Account Error", fmt.Sprintf("Unable to update CA account: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CAAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CAAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCAAccount(data.OrganizationID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete CA Account Error", fmt.Sprintf("Unable to delete CA account: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<organization_id>/<resource_id>"
func (r *CAAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<organization_id>/<resource_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
