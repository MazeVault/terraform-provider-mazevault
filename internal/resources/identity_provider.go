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

var _ resource.Resource = &IdentityProviderResource{}

func NewIdentityProviderResource() resource.Resource { return &IdentityProviderResource{} }

type IdentityProviderResource struct{ client *mazevault.Client }

type IdentityProviderModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Config       types.String `tfsdk:"config"`
	SyncSchedule types.String `tfsdk:"sync_schedule"`
	Status       types.String `tfsdk:"status"`
}

func (r *IdentityProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_provider"
}

func (r *IdentityProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an identity provider (SAML, OIDC, LDAP, or SCIM) for MazeVault authentication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the identity provider.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for this identity provider.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Provider type: `saml`, `oidc`, `ldap`, or `scim`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"config": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "JSON-encoded provider configuration. For OIDC include `client_id`, `client_secret`, `tenant_id`; for SAML include `metadata_url`; for LDAP include `host`, `bind_dn`, `bind_password`.",
			},
			"sync_schedule": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cron expression for automatic user/group sync (e.g. `0 * * * *` for hourly).",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of the identity provider (`active`, `error`, `syncing`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *IdentityProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IdentityProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IdentityProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg := parseConfigJSON(data.Config.ValueString())
	created, err := r.client.CreateIdentityProvider(&mazevault.CreateIdentityProviderRequest{
		Name:         data.Name.ValueString(),
		Type:         data.Type.ValueString(),
		Config:       cfg,
		SyncSchedule: data.SyncSchedule.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Identity Provider Error", fmt.Sprintf("Unable to create identity provider: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Status = types.StringValue(created.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IdentityProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	idp, err := r.client.GetIdentityProvider(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Identity Provider Error", fmt.Sprintf("Unable to read identity provider: %s", err))
		return
	}
	if idp == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(idp.Name)
	data.Type = types.StringValue(idp.Type)
	data.SyncSchedule = types.StringValue(idp.SyncSchedule)
	data.Status = types.StringValue(idp.Status)
	// config is Sensitive/write-only; preserve from state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IdentityProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state IdentityProviderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfgU := parseConfigJSON(plan.Config.ValueString())
	if _, err := r.client.UpdateIdentityProvider(state.ID.ValueString(), &mazevault.CreateIdentityProviderRequest{
		Name:         plan.Name.ValueString(),
		Type:         plan.Type.ValueString(),
		Config:       cfgU,
		SyncSchedule: plan.SyncSchedule.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Identity Provider Error", fmt.Sprintf("Unable to update identity provider: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IdentityProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IdentityProviderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIdentityProvider(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Identity Provider Error", fmt.Sprintf("Unable to delete identity provider: %s", err))
	}
}
