package resources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EntraIntegrationScopeResource{}
var _ resource.ResourceWithValidateConfig = &EntraIntegrationScopeResource{}
var _ resource.ResourceWithImportState = &EntraIntegrationScopeResource{}

func NewEntraIntegrationScopeResource() resource.Resource { return &EntraIntegrationScopeResource{} }

type EntraIntegrationScopeResource struct{ client *mazevault.Client }

type EntraIntegrationScopeModel struct {
	ID             types.String `tfsdk:"id"`
	IntegrationID  types.String `tfsdk:"integration_id"`
	ProjectID      types.String `tfsdk:"project_id"`
	ScopeMode      types.String `tfsdk:"scope_mode"`
	SyncAfterApply types.Bool   `tfsdk:"sync_after_apply"`
	SelectedApps   types.List   `tfsdk:"selected_apps"`
}

type EntraSelectedAppModel struct {
	AppID       types.String `tfsdk:"app_id"`
	ObjectID    types.String `tfsdk:"object_id"`
	DisplayName types.String `tfsdk:"display_name"`
	TenantID    types.String `tfsdk:"tenant_id"`
	BindingKind types.String `tfsdk:"binding_kind"`
}

var entraSelectedAppAttrTypes = map[string]attr.Type{
	"app_id":       types.StringType,
	"object_id":    types.StringType,
	"display_name": types.StringType,
	"tenant_id":    types.StringType,
	"binding_kind": types.StringType,
}

func (r *EntraIntegrationScopeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entra_integration_scope"
}

func (r *EntraIntegrationScopeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Selects which Microsoft Entra ID application registrations a project-scoped Entra integration " +
			"(`mazevault_integration` with `provider_name = \"entra_id\"`) manages. In `selected_apps` mode only the listed " +
			"applications are synchronised into the project and become eligible for client-secret lifecycle management; " +
			"an application can be bound to one project at a time. After the scope is written the integration is " +
			"synchronised from Microsoft Graph (see `sync_after_apply`) so that `data.mazevault_entra_app_registration` " +
			"can resolve credential IDs for `mazevault_entra_rotation_config`. Destroying the resource only removes it from " +
			"Terraform state: the backend cannot hold an empty `selected_apps` scope and the provider never widens the scope " +
			"to `legacy_all` implicitly; destroy the integration to drop all bindings. Entra ID itself is never modified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Equals `integration_id` (one scope per integration).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"integration_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project Entra integration whose scope is managed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"project_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the project that owns the integration (resolved by the backend).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"scope_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(mazevault.EntraScopeModeSelectedApps),
				MarkdownDescription: "`selected_apps` (default) manages only `selected_apps`; `legacy_all` manages every application the " +
					"integration identity can see and ignores `selected_apps`.",
			},
			"sync_after_apply": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Run an application/credential synchronisation from Microsoft Graph after the scope is created or updated. Defaults to `true`.",
			},
			"selected_apps": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Applications selected into the project. Required (non-empty) when `scope_mode = \"selected_apps\"`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"app_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Entra application (client) ID.",
						},
						"object_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Entra object ID of the application registration (or service principal, see `binding_kind`).",
						},
						"display_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Display name shown in MazeVault until the first synchronisation refreshes it.",
						},
						"tenant_id": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							MarkdownDescription: "Tenant ID; defaults to the integration tenant.",
						},
						"binding_kind": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(mazevault.EntraBindingKindAppRegistration),
							MarkdownDescription: "`app_registration` (default) or `service_principal`.",
						},
					},
				},
			},
		},
	}
}

func (r *EntraIntegrationScopeResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data EntraIntegrationScopeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mode := data.ScopeMode.ValueString()
	if !data.ScopeMode.IsNull() && !data.ScopeMode.IsUnknown() && mode != mazevault.EntraScopeModeSelectedApps && mode != mazevault.EntraScopeModeLegacyAll {
		resp.Diagnostics.AddAttributeError(path.Root("scope_mode"), "Invalid scope_mode",
			fmt.Sprintf("%q is not supported; use %q or %q", mode, mazevault.EntraScopeModeSelectedApps, mazevault.EntraScopeModeLegacyAll))
	}
	if data.SelectedApps.IsNull() || data.SelectedApps.IsUnknown() {
		return
	}
	apps, diags := r.selectedApps(ctx, data.SelectedApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	seen := map[string]struct{}{}
	for i, app := range apps {
		if kind := app.BindingKind.ValueString(); !app.BindingKind.IsNull() && !app.BindingKind.IsUnknown() && kind != "" &&
			kind != mazevault.EntraBindingKindAppRegistration && kind != mazevault.EntraBindingKindServicePrincipal {
			resp.Diagnostics.AddAttributeError(path.Root("selected_apps").AtListIndex(i).AtName("binding_kind"), "Invalid binding_kind",
				fmt.Sprintf("%q is not supported; use %q or %q", kind, mazevault.EntraBindingKindAppRegistration, mazevault.EntraBindingKindServicePrincipal))
		}
		if app.AppID.IsUnknown() {
			continue
		}
		key := strings.ToLower(app.AppID.ValueString())
		if _, dup := seen[key]; dup {
			resp.Diagnostics.AddAttributeError(path.Root("selected_apps").AtListIndex(i).AtName("app_id"), "Duplicate application",
				fmt.Sprintf("application %s is listed more than once", app.AppID.ValueString()))
		}
		seen[key] = struct{}{}
	}
}

func (r *EntraIntegrationScopeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EntraIntegrationScopeResource) selectedApps(ctx context.Context, list types.List) ([]EntraSelectedAppModel, diag.Diagnostics) {
	var apps []EntraSelectedAppModel
	if list.IsNull() || list.IsUnknown() {
		return apps, nil
	}
	diags := list.ElementsAs(ctx, &apps, false)
	return apps, diags
}

func (r *EntraIntegrationScopeResource) buildRequest(ctx context.Context, data EntraIntegrationScopeModel) (*mazevault.EntraIntegrationScopeRequest, diag.Diagnostics) {
	apps, diags := r.selectedApps(ctx, data.SelectedApps)
	if diags.HasError() {
		return nil, diags
	}
	mode := data.ScopeMode.ValueString()
	if mode == "" {
		mode = mazevault.EntraScopeModeSelectedApps
	}
	if mode == mazevault.EntraScopeModeSelectedApps && len(apps) == 0 {
		diags.AddAttributeError(path.Root("selected_apps"), "No applications selected",
			"scope_mode \"selected_apps\" requires at least one entry in selected_apps")
		return nil, diags
	}
	req := &mazevault.EntraIntegrationScopeRequest{
		ScopeMode:           mode,
		SelectedAppBindings: make([]mazevault.EntraSelectedAppBinding, 0, len(apps)),
	}
	for _, app := range apps {
		kind := app.BindingKind.ValueString()
		if kind == "" {
			kind = mazevault.EntraBindingKindAppRegistration
		}
		req.SelectedAppBindings = append(req.SelectedAppBindings, mazevault.EntraSelectedAppBinding{
			BindingKind: kind,
			AppID:       app.AppID.ValueString(),
			ObjectID:    app.ObjectID.ValueString(),
			TenantID:    app.TenantID.ValueString(),
			DisplayName: app.DisplayName.ValueString(),
		})
	}
	return req, diags
}

// applyState maps the backend scope state onto the model. Bindings are ordered
// by app_id so repeated reads are deterministic; display names configured by the
// user are preserved when the backend only differs in casing/whitespace.
func (r *EntraIntegrationScopeResource) applyState(ctx context.Context, data *EntraIntegrationScopeModel, state *mazevault.EntraIntegrationScopeState) diag.Diagnostics {
	// The integration id is the identity of this resource and comes from the configuration (or the import
	// id); the backend view only confirms it, so a response without it must not blank the state.
	integrationID := data.IntegrationID.ValueString()
	if integrationID == "" {
		integrationID = state.IntegrationID
	}
	data.ID = types.StringValue(integrationID)
	data.IntegrationID = types.StringValue(integrationID)
	data.ProjectID = types.StringValue(state.ProjectID)
	if state.ScopeMode != "" {
		data.ScopeMode = types.StringValue(state.ScopeMode)
	} else if data.ScopeMode.IsNull() || data.ScopeMode.IsUnknown() {
		data.ScopeMode = types.StringValue(mazevault.EntraScopeModeSelectedApps)
	}

	prior, diags := r.selectedApps(ctx, data.SelectedApps)
	if diags.HasError() {
		return diags
	}
	priorByApp := map[string]EntraSelectedAppModel{}
	priorOrder := map[string]int{}
	for i, app := range prior {
		key := strings.ToLower(app.AppID.ValueString())
		priorByApp[key] = app
		priorOrder[key] = i
	}

	bindings := append([]mazevault.EntraSelectedAppBindingView(nil), state.SelectedAppBindings...)
	sort.SliceStable(bindings, func(i, j int) bool {
		oi, iok := priorOrder[strings.ToLower(bindings[i].AppID)]
		oj, jok := priorOrder[strings.ToLower(bindings[j].AppID)]
		switch {
		case iok && jok:
			return oi < oj
		case iok:
			return true
		case jok:
			return false
		default:
			return bindings[i].AppID < bindings[j].AppID
		}
	})

	elements := make([]attr.Value, 0, len(bindings))
	for _, b := range bindings {
		displayName := b.DisplayName
		tenantID := b.TenantID
		if p, ok := priorByApp[strings.ToLower(b.AppID)]; ok {
			if strings.EqualFold(strings.TrimSpace(p.DisplayName.ValueString()), strings.TrimSpace(b.DisplayName)) {
				displayName = p.DisplayName.ValueString()
			}
			if p.TenantID.ValueString() == "" {
				tenantID = ""
			}
		}
		obj, objDiags := types.ObjectValue(entraSelectedAppAttrTypes, map[string]attr.Value{
			"app_id":       types.StringValue(b.AppID),
			"object_id":    types.StringValue(b.ObjectID),
			"display_name": types.StringValue(displayName),
			"tenant_id":    types.StringValue(tenantID),
			"binding_kind": types.StringValue(b.BindingKind),
		})
		diags.Append(objDiags...)
		elements = append(elements, obj)
	}
	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: entraSelectedAppAttrTypes}, elements)
	diags.Append(listDiags...)
	data.SelectedApps = list
	return diags
}

func (r *EntraIntegrationScopeResource) write(ctx context.Context, data *EntraIntegrationScopeModel) diag.Diagnostics {
	req, diags := r.buildRequest(ctx, *data)
	if diags.HasError() {
		return diags
	}
	state, err := r.client.UpdateEntraIntegrationScope(data.IntegrationID.ValueString(), req)
	if err != nil {
		diags.AddError("Update Entra Integration Scope Error", fmt.Sprintf("Unable to update Entra scope: %s", err))
		return diags
	}
	if data.SyncAfterApply.IsNull() || data.SyncAfterApply.IsUnknown() || data.SyncAfterApply.ValueBool() {
		if err := r.client.SyncEntraAppRegistrations(data.IntegrationID.ValueString()); err != nil {
			diags.AddError("Entra Synchronisation Error",
				fmt.Sprintf("Scope was saved but synchronising applications from Microsoft Graph failed: %s. "+
					"Re-run apply or set sync_after_apply = false to skip the synchronisation.", err))
			return diags
		}
		// Re-read so display names refreshed by the synchronisation land in state.
		if refreshed, err := r.client.GetEntraIntegrationScope(data.IntegrationID.ValueString()); err == nil && refreshed != nil {
			state = refreshed
		}
	}
	diags.Append(r.applyState(ctx, data, state)...)
	return diags
}

func (r *EntraIntegrationScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntraIntegrationScopeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntraIntegrationScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntraIntegrationScopeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, err := r.client.GetEntraIntegrationScope(data.IntegrationID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Entra Integration Scope Error", fmt.Sprintf("Unable to read Entra scope: %s", err))
		return
	}
	resp.Diagnostics.Append(r.applyState(ctx, &data, state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EntraIntegrationScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EntraIntegrationScopeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.write(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from state only. The backend cannot represent a
// `selected_apps` scope without applications, and silently switching to
// `legacy_all` would widen the managed set instead of narrowing it. The last
// applied selection therefore stays active until the integration itself is
// destroyed (which removes all bindings) or a new scope is applied.
func (r *EntraIntegrationScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntraIntegrationScopeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddWarning("Entra scope retained on backend",
		fmt.Sprintf("mazevault_entra_integration_scope for integration %s was removed from Terraform state; the last applied "+
			"application selection remains active in MazeVault. Destroy the mazevault_integration to remove all bindings.",
			data.IntegrationID.ValueString()))
}

// ImportState implements resource.ResourceWithImportState. The import ID is
// the integration ID.
func (r *EntraIntegrationScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("integration_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sync_after_apply"), true)...)
}
