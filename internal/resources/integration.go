package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IntegrationResource{}

func NewIntegrationResource() resource.Resource { return &IntegrationResource{} }

type IntegrationResource struct{ client *mazevault.Client }

type IntegrationModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Provider  types.String `tfsdk:"provider"`
	// Azure DevOps config fields
	AzureOrg             types.String `tfsdk:"azure_org"`
	AzureProject         types.String `tfsdk:"azure_project"`
	AzurePAT             types.String `tfsdk:"azure_pat"`
	AzureMode            types.String `tfsdk:"azure_mode"`
	AzureVariableGroupID types.String `tfsdk:"azure_variable_group_id"`
	AzureRepo            types.String `tfsdk:"azure_repo"`
	AzureBranch          types.String `tfsdk:"azure_branch"`
	AzureFilePath        types.String `tfsdk:"azure_file_path"`
}

func (r *IntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *IntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                      schema.StringAttribute{Computed: true},
			"project_id":              schema.StringAttribute{Required: true},
			"name":                    schema.StringAttribute{Required: true},
			"type":                    schema.StringAttribute{Required: true},
			"provider":                schema.StringAttribute{Required: true},
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
	cfg := map[string]interface{}{}
	set := func(k string, v types.String) {
		if !v.IsNull() {
			cfg[k] = v.ValueString()
		}
	}
	set("azure_org", data.AzureOrg)
	set("azure_project", data.AzureProject)
	set("azure_pat", data.AzurePAT)
	set("azure_mode", data.AzureMode)
	set("azure_variable_group_id", data.AzureVariableGroupID)
	set("azure_repo", data.AzureRepo)
	set("azure_branch", data.AzureBranch)
	set("azure_file_path", data.AzureFilePath)
	created, err := r.client.CreateIntegration(data.ProjectID.ValueString(), data.Name.ValueString(), data.Type.ValueString(), data.Provider.ValueString(), cfg)
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
	list, err := r.client.ListIntegrations(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Integration Error", err.Error())
		return
	}
	for _, it := range list {
		if it.ID == data.ID.ValueString() {
			// reflect config back if needed
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
			return
		}
	}
	// Not found
	resp.Diagnostics.AddWarning("Integration Not Found", "The integration was not found on the server")
}

func (r *IntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state IntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg := map[string]interface{}{}
	set := func(k string, v types.String) {
		if !v.IsNull() {
			cfg[k] = v.ValueString()
		}
	}
	set("azure_org", plan.AzureOrg)
	set("azure_project", plan.AzureProject)
	set("azure_pat", plan.AzurePAT)
	set("azure_mode", plan.AzureMode)
	set("azure_variable_group_id", plan.AzureVariableGroupID)
	set("azure_repo", plan.AzureRepo)
	set("azure_branch", plan.AzureBranch)
	set("azure_file_path", plan.AzureFilePath)

	_, err := r.client.UpdateIntegration(state.ID.ValueString(), plan.Name.ValueString(), plan.Type.ValueString(), plan.Provider.ValueString(), cfg)
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
