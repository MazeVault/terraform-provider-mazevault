package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IntegrationResource{}
var _ resource.ResourceWithImportState = &IntegrationResource{}

func NewIntegrationResource() resource.Resource { return &IntegrationResource{} }

type IntegrationResource struct{ client *mazevault.Client }

type IntegrationModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Provider    types.String `tfsdk:"provider_name"`
	Environment types.String `tfsdk:"environment"`
	// Generic configuration (works for any provider: Azure Key Vault, Entra, etc.).
	Config                types.Map    `tfsdk:"config"`
	SensitiveConfig       types.Map    `tfsdk:"sensitive_config"`
	InteractiveAuthMethod types.String `tfsdk:"interactive_auth_method"`
	// Azure DevOps config fields (kept for backward compatibility).
	AzureOrg             types.String `tfsdk:"azure_org"`
	AzureProject         types.String `tfsdk:"azure_project"`
	AzurePAT             types.String `tfsdk:"azure_pat"`
	AzureMode            types.String `tfsdk:"azure_mode"`
	AzureVariableGroupID types.String `tfsdk:"azure_variable_group_id"`
	AzureRepo            types.String `tfsdk:"azure_repo"`
	AzureBranch          types.String `tfsdk:"azure_branch"`
	AzureFilePath        types.String `tfsdk:"azure_file_path"`
}

// buildIntegrationConfig merges the typed azure_* fields with the generic config
// and sensitive_config maps into a single config payload. The backend accepts an
// arbitrary config map and branches by type/provider, so this supports any
// integration type (Azure Key Vault, Entra, Azure DevOps, …).
func buildIntegrationConfig(ctx context.Context, d IntegrationModel, diags *diag.Diagnostics) map[string]interface{} {
	cfg := map[string]interface{}{}
	set := func(k string, v types.String) {
		if !v.IsNull() {
			cfg[k] = v.ValueString()
		}
	}
	set("azure_org", d.AzureOrg)
	set("azure_project", d.AzureProject)
	set("azure_pat", d.AzurePAT)
	set("azure_mode", d.AzureMode)
	set("azure_variable_group_id", d.AzureVariableGroupID)
	set("azure_repo", d.AzureRepo)
	set("azure_branch", d.AzureBranch)
	set("azure_file_path", d.AzureFilePath)

	if !d.Config.IsNull() && !d.Config.IsUnknown() {
		m := map[string]string{}
		diags.Append(d.Config.ElementsAs(ctx, &m, false)...)
		for k, v := range m {
			cfg[k] = v
		}
	}
	if !d.SensitiveConfig.IsNull() && !d.SensitiveConfig.IsUnknown() {
		m := map[string]string{}
		diags.Append(d.SensitiveConfig.ElementsAs(ctx, &m, false)...)
		for k, v := range m {
			cfg[k] = v
		}
	}
	return cfg
}

func (r *IntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project_id":    schema.StringAttribute{Required: true},
			"name":          schema.StringAttribute{Required: true},
			"type":          schema.StringAttribute{Required: true},
			"provider_name": schema.StringAttribute{Required: true},
			"environment": schema.StringAttribute{
				Required: true,
				// UpdateIntegration does not send environment in the PUT body;
				// changes to this field must recreate the resource.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"config": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Generic non-secret configuration for the integration (e.g. `vault_url`, `tenant_id`, `client_id` for Azure Key Vault / Entra). Merged with the typed `azure_*` fields.",
			},
			"sensitive_config": schema.MapAttribute{
				Optional:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Generic secret configuration values (e.g. `client_secret`). Write-only; not refreshed from the server.",
			},
			"interactive_auth_method": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Interactive auth method (e.g. `on_behalf_of`). When `on_behalf_of`, tenant/client credentials are injected server-side from the SSO configuration.",
			},
			"azure_org":               schema.StringAttribute{Optional: true},
			"azure_project":           schema.StringAttribute{Optional: true},
			"azure_pat":               schema.StringAttribute{Optional: true, Sensitive: true},
			"azure_mode":              schema.StringAttribute{Optional: true},
			"azure_variable_group_id": schema.StringAttribute{Optional: true},
			"azure_repo":              schema.StringAttribute{Optional: true},
			"azure_branch":            schema.StringAttribute{Optional: true},
			"azure_file_path":         schema.StringAttribute{Optional: true},
		},
	}
}

func (r *IntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *IntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg := buildIntegrationConfig(ctx, data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateIntegration(data.ProjectID.ValueString(), data.Name.ValueString(), data.Type.ValueString(), data.Provider.ValueString(), data.Environment.ValueString(), cfg, data.InteractiveAuthMethod.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Integration Error", err.Error())
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	it, err := r.client.GetIntegration(data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Integration Error", err.Error())
		return
	}
	if it == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	// Restore core fields from API response.
	data.Name = types.StringValue(it.Name)
	data.Type = types.StringValue(it.Type)
	data.Provider = types.StringValue(it.Provider)
	// Reflect Azure DevOps config fields (write-only fields like AzurePAT are preserved from state).
	if v, ok := it.Config["azure_org"].(string); ok {
		data.AzureOrg = types.StringValue(v)
	}
	if v, ok := it.Config["azure_project"].(string); ok {
		data.AzureProject = types.StringValue(v)
	}
	if v, ok := it.Config["azure_mode"].(string); ok {
		data.AzureMode = types.StringValue(v)
	}
	if v, ok := it.Config["azure_variable_group_id"].(string); ok {
		data.AzureVariableGroupID = types.StringValue(v)
	}
	if v, ok := it.Config["azure_repo"].(string); ok {
		data.AzureRepo = types.StringValue(v)
	}
	if v, ok := it.Config["azure_branch"].(string); ok {
		data.AzureBranch = types.StringValue(v)
	}
	if v, ok := it.Config["azure_file_path"].(string); ok {
		data.AzureFilePath = types.StringValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state IntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg := buildIntegrationConfig(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateIntegration(state.ID.ValueString(), plan.Name.ValueString(), plan.Type.ValueString(), plan.Provider.ValueString(), cfg, plan.InteractiveAuthMethod.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update Integration Error", err.Error())
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteIntegration(data.ProjectID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Integration Error", err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *IntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
