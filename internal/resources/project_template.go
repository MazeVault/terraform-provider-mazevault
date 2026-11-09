package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectTemplateResource{}
var _ resource.ResourceWithValidateConfig = &ProjectTemplateResource{}
var _ resource.ResourceWithImportState = &ProjectTemplateResource{}

// projectTemplateSubtypes mirrors the frontend Project Template catalogue
// (ProjectTemplateModal/constants.ts). `certificates` is intentionally excluded:
// certificate-category templates auto-provision a CertificateTemplate and are
// managed by `mazevault_certificate_template`.
var projectTemplateSubtypes = map[string][]string{
	"secrets": {"secrets_manager", "database", "helm_chart", "configuration", "agent", "custom"},
	"keys":    {"aws_kms", "azure_key_vault", "hashicorp_vault", "ssh", "gpg", "custom"},
}

func NewProjectTemplateResource() resource.Resource { return &ProjectTemplateResource{} }

type ProjectTemplateResource struct{ client *mazevault.Client }

type ProjectTemplateModel struct {
	ID               types.String `tfsdk:"id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Category         types.String `tfsdk:"category"`
	Subtype          types.String `tfsdk:"subtype"`
	Version          types.String `tfsdk:"version"`
	ProtocolConfig   types.String `tfsdk:"protocol_config"`
	EnvironmentRules types.String `tfsdk:"environment_rules"`
	VariableSchema   types.String `tfsdk:"variable_schema"`
}

func (r *ProjectTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_template"
}

func (r *ProjectTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an organization-level **Project Template** of category `secrets` or `keys` (the entries shown " +
			"in the frontend *Project Templates* view). Templates pre-configure how projects of that kind are created: " +
			"protocol settings, environment rules and the variable schema prompted at project creation. " +
			"For certificate-category templates use `mazevault_certificate_template`, which additionally provisions the " +
			"issuing certificate template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the project template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization that owns the template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Human-readable description.",
			},
			"category": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template category: `secrets` or `keys`. Changing it forces a new template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"subtype": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("custom"),
				MarkdownDescription: "Category-specific subtype. `secrets`: `secrets_manager`, `database`, `helm_chart`, `configuration`, " +
					"`agent`, `custom`. `keys`: `aws_kms`, `azure_key_vault`, `hashicorp_vault`, `ssh`, `gpg`, `custom`. Defaults to `custom`.",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Free-form template version label.",
			},
			"protocol_config": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
				MarkdownDescription: "JSON object with protocol/provider defaults applied to projects created from the template. Build with `jsonencode()`.",
			},
			"environment_rules": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(`{"allow_all":true}`),
				MarkdownDescription: "JSON object restricting the environments a project created from the template may use. Defaults to `{\"allow_all\": true}`.",
			},
			"variable_schema": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("[]"),
				MarkdownDescription: "JSON array describing variables prompted when a project is instantiated from the template.",
			},
		},
	}
}

func (r *ProjectTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ProjectTemplateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.Category.IsNull() && !data.Category.IsUnknown() {
		category := data.Category.ValueString()
		subtypes, ok := projectTemplateSubtypes[category]
		if !ok {
			detail := fmt.Sprintf("%q is not supported; use \"secrets\" or \"keys\"", category)
			if category == "certificates" {
				detail = "certificate-category templates are managed by mazevault_certificate_template"
			}
			resp.Diagnostics.AddAttributeError(path.Root("category"), "Unsupported category", detail)
			return
		}
		if !data.Subtype.IsNull() && !data.Subtype.IsUnknown() && data.Subtype.ValueString() != "" {
			subtype := data.Subtype.ValueString()
			known := false
			for _, s := range subtypes {
				if s == subtype {
					known = true
					break
				}
			}
			if !known {
				resp.Diagnostics.AddAttributeError(path.Root("subtype"), "Unsupported subtype",
					fmt.Sprintf("%q is not a %s subtype; expected one of: %s", subtype, category, strings.Join(subtypes, ", ")))
			}
		}
	}
	for _, attr := range []struct {
		name  string
		value types.String
	}{{"protocol_config", data.ProtocolConfig}, {"environment_rules", data.EnvironmentRules}} {
		if !attr.value.IsNull() && !attr.value.IsUnknown() {
			if _, err := parseJSONObject(attr.name, attr.value.ValueString()); err != nil {
				resp.Diagnostics.AddAttributeError(path.Root(attr.name), "Invalid JSON object", err.Error())
			}
		}
	}
	if !data.VariableSchema.IsNull() && !data.VariableSchema.IsUnknown() {
		if _, err := parseJSONArray("variable_schema", data.VariableSchema.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("variable_schema"), "Invalid JSON array", err.Error())
		}
	}
}

func (r *ProjectTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (data ProjectTemplateModel) buildInput() (*mazevault.ProjectTemplateInput, error) {
	protocol, err := parseJSONObject("protocol_config", data.ProtocolConfig.ValueString())
	if err != nil {
		return nil, err
	}
	rules, err := parseJSONObject("environment_rules", data.EnvironmentRules.ValueString())
	if err != nil {
		return nil, err
	}
	variables, err := parseJSONArray("variable_schema", data.VariableSchema.ValueString())
	if err != nil {
		return nil, err
	}
	return &mazevault.ProjectTemplateInput{
		Name:             data.Name.ValueString(),
		Description:      data.Description.ValueString(),
		Category:         data.Category.ValueString(),
		Subtype:          data.Subtype.ValueString(),
		Version:          data.Version.ValueString(),
		ProtocolConfig:   protocol,
		EnvironmentRules: rules,
		VariableSchema:   variables,
	}, nil
}

func (data *ProjectTemplateModel) applyFromAPI(tmpl *mazevault.ProjectTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Name = types.StringValue(tmpl.Name)
	data.Description = types.StringValue(tmpl.Description)
	data.Category = types.StringValue(tmpl.Category)
	data.Subtype = types.StringValue(tmpl.Subtype)
	data.Version = types.StringValue(tmpl.Version)

	protocol, err := reconcileJSONState(data.ProtocolConfig.ValueString(), orEmptyObject(tmpl.ProtocolConfig))
	if err != nil {
		diags.AddError("Invalid protocol_config from backend", err.Error())
		return diags
	}
	data.ProtocolConfig = types.StringValue(protocol)

	rules, err := reconcileJSONState(data.EnvironmentRules.ValueString(), orEmptyObject(tmpl.EnvironmentRules))
	if err != nil {
		diags.AddError("Invalid environment_rules from backend", err.Error())
		return diags
	}
	data.EnvironmentRules = types.StringValue(rules)

	variables := tmpl.VariableSchema
	if variables == nil {
		variables = []interface{}{}
	}
	schemaJSON, err := reconcileJSONState(data.VariableSchema.ValueString(), variables)
	if err != nil {
		diags.AddError("Invalid variable_schema from backend", err.Error())
		return diags
	}
	data.VariableSchema = types.StringValue(schemaJSON)
	return diags
}

func orEmptyObject(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return m
}

func (r *ProjectTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := data.buildInput()
	if err != nil {
		resp.Diagnostics.AddError("Invalid Project Template", err.Error())
		return
	}
	created, err := r.client.CreateOrgProjectTemplate(data.OrganizationID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Create Project Template Error", fmt.Sprintf("Unable to create project template: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(data.applyFromAPI(created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tmpl, err := r.client.GetOrgProjectTemplate(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Project Template Error", fmt.Sprintf("Unable to read project template: %s", err))
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(data.applyFromAPI(tmpl)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProjectTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input, err := plan.buildInput()
	if err != nil {
		resp.Diagnostics.AddError("Invalid Project Template", err.Error())
		return
	}
	updated, err := r.client.UpdateOrgProjectTemplate(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Project Template Error", fmt.Sprintf("Unable to update project template: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(plan.applyFromAPI(updated)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrgProjectTemplate(data.ID.ValueString()); err != nil && !mazevault.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Delete Project Template Error", fmt.Sprintf("Unable to delete project template: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<organization_id>/<template_id>"
func (r *ProjectTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<organization_id>/<template_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
