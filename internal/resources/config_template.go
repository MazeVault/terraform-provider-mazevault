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

var _ resource.Resource = &ConfigTemplateResource{}

func NewConfigTemplateResource() resource.Resource { return &ConfigTemplateResource{} }

type ConfigTemplateResource struct{ client *mazevault.Client }

type ConfigTemplateModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Template       types.String `tfsdk:"content"`
	Format         types.String `tfsdk:"format"`
}

func (r *ConfigTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_template"
}

func (r *ConfigTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a configuration template that agents use to render secrets into application configuration files.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the config template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization that owns this template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the template's purpose.",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template body (Go template syntax). Use `{{ .Secrets.MY_SECRET }}` to inject secret values.",
			},
			"format": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Output format hint: `env`, `json`, `yaml`, or `toml`.",
			},
		},
	}
}

func (r *ConfigTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConfigTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateConfigTemplate(data.OrganizationID.ValueString(), &mazevault.CreateConfigTemplateRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Template:    data.Template.ValueString(),
		Format:      data.Format.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Config Template Error", fmt.Sprintf("Unable to create config template: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tmpl, err := r.client.GetConfigTemplate(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Config Template Error", fmt.Sprintf("Unable to read config template: %s", err))
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(tmpl.Name)
	data.Description = types.StringValue(tmpl.Description)
	data.Template = types.StringValue(tmpl.Template)
	data.Format = types.StringValue(tmpl.Format)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConfigTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateConfigTemplate(state.OrganizationID.ValueString(), state.ID.ValueString(), &mazevault.CreateConfigTemplateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Template:    plan.Template.ValueString(),
		Format:      plan.Format.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Config Template Error", fmt.Sprintf("Unable to update config template: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConfigTemplate(data.OrganizationID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Config Template Error", fmt.Sprintf("Unable to delete config template: %s", err))
	}
}
