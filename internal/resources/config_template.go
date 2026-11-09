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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConfigTemplateResource{}
var _ resource.ResourceWithValidateConfig = &ConfigTemplateResource{}
var _ resource.ResourceWithImportState = &ConfigTemplateResource{}

// supportedConfigTemplateTypes mirrors the Configuration Management template
// types understood by the backend (services/config). naming_convention and
// secret_naming_policy are stored through the same endpoint but validated by
// their dedicated services, so they are accepted here as well.
var supportedConfigTemplateTypes = []string{
	"naming_convention",
	"secret_pattern",
	"secret_naming_policy",
	"discovery_policy",
	"exclusion_rule",
	"resource_visibility_policy",
	"parser",
}

func NewConfigTemplateResource() resource.Resource { return &ConfigTemplateResource{} }

type ConfigTemplateResource struct{ client *mazevault.Client }

type ConfigTemplateModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	TemplateType   types.String `tfsdk:"template_type"`
	Definition     types.String `tfsdk:"definition"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	Priority       types.Int64  `tfsdk:"priority"`
}

func isSupportedConfigTemplateType(value string) bool {
	for _, t := range supportedConfigTemplateTypes {
		if t == value {
			return true
		}
	}
	return false
}

func (r *ConfigTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_template"
}

func (r *ConfigTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an organization-level **Configuration Management template**: a typed policy document " +
			"such as a naming convention, secret pattern, secret naming policy, discovery policy, exclusion rule, " +
			"or resource visibility policy. The document is validated by the backend for the selected `template_type` " +
			"before it is persisted. Requires the `config_management` license feature and the `config:write` permission; " +
			"`resource_visibility_policy` additionally requires visibility-policy administrator rights.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the template.",
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
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the template's purpose.",
			},
			"template_type": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Policy type. One of `" + strings.Join(supportedConfigTemplateTypes, "`, `") +
					"`. Changing the type forces a new template.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"definition": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "JSON object with the type-specific policy document. Build it with `jsonencode()`. " +
					"The backend validates the document against the schema of `template_type`; validation errors are " +
					"reported at plan/apply time with the offending field.",
			},
			"is_active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the template is active. Inactive templates are ignored by enforcement.",
			},
			"is_default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Marks the template as the organization default for its type.",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
				MarkdownDescription: "Evaluation priority (lower runs first). Defaults to `100`.",
			},
		},
	}
}

func (r *ConfigTemplateResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.TemplateType.IsNull() && !data.TemplateType.IsUnknown() {
		if !isSupportedConfigTemplateType(data.TemplateType.ValueString()) {
			resp.Diagnostics.AddAttributeError(path.Root("template_type"), "Unsupported template_type",
				fmt.Sprintf("%q is not a Configuration Management template type; expected one of: %s",
					data.TemplateType.ValueString(), strings.Join(supportedConfigTemplateTypes, ", ")))
		}
	}
	if !data.Definition.IsNull() && !data.Definition.IsUnknown() {
		if _, err := parseJSONObject("definition", data.Definition.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("definition"), "Invalid definition", err.Error())
		}
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

// buildRequest converts the plan into the backend payload after running the
// backend validator so that definition errors surface with their field names
// instead of a generic HTTP 400.
func (r *ConfigTemplateResource) buildRequest(data ConfigTemplateModel) (*mazevault.CreateConfigTemplateRequest, error) {
	definition, err := parseJSONObject("definition", data.Definition.ValueString())
	if err != nil {
		return nil, err
	}
	templateType := data.TemplateType.ValueString()
	result, err := r.client.ValidateConfigTemplateDefinition(data.OrganizationID.ValueString(), templateType, definition)
	if err != nil {
		return nil, fmt.Errorf("unable to validate definition: %w", err)
	}
	if result != nil && !result.Valid {
		issues := make([]string, 0, len(result.Errors))
		for _, issue := range result.Errors {
			issues = append(issues, fmt.Sprintf("%s: %s", issue.Field, issue.Message))
		}
		return nil, fmt.Errorf("definition rejected by backend validator for template_type %q: %s", templateType, strings.Join(issues, "; "))
	}
	isActive := data.IsActive.ValueBool()
	isDefault := data.IsDefault.ValueBool()
	priority := int(data.Priority.ValueInt64())
	return &mazevault.CreateConfigTemplateRequest{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		TemplateType: templateType,
		Definition:   definition,
		IsActive:     &isActive,
		IsDefault:    &isDefault,
		Priority:     &priority,
	}, nil
}

func (r *ConfigTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload, err := r.buildRequest(data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Config Template", err.Error())
		return
	}
	created, err := r.client.CreateConfigTemplate(data.OrganizationID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Config Template Error", fmt.Sprintf("Unable to create config template: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(r.applyToModel(&data, created)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigTemplateResource) applyToModel(data *ConfigTemplateModel, tmpl *mazevault.ConfigTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Name = types.StringValue(tmpl.Name)
	data.Description = types.StringValue(tmpl.Description)
	data.TemplateType = types.StringValue(tmpl.TemplateType)
	data.IsActive = types.BoolValue(tmpl.IsActive)
	data.IsDefault = types.BoolValue(tmpl.IsDefault)
	data.Priority = types.Int64Value(int64(tmpl.Priority))
	definition, err := reconcileJSONState(data.Definition.ValueString(), tmpl.Definition)
	if err != nil {
		diags.AddError("Invalid definition from backend", err.Error())
		return diags
	}
	data.Definition = types.StringValue(definition)
	return diags
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
	resp.Diagnostics.Append(r.applyToModel(&data, tmpl)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConfigTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ConfigTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload, err := r.buildRequest(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Config Template", err.Error())
		return
	}
	updated, err := r.client.UpdateConfigTemplate(state.OrganizationID.ValueString(), state.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Config Template Error", fmt.Sprintf("Unable to update config template: %s", err))
		return
	}
	plan.ID = state.ID
	resp.Diagnostics.Append(r.applyToModel(&plan, updated)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConfigTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteConfigTemplate(data.OrganizationID.ValueString(), data.ID.ValueString()); err != nil && !mazevault.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Delete Config Template Error", fmt.Sprintf("Unable to delete config template: %s", err))
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: "<organization_id>/<resource_id>"
func (r *ConfigTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if parts == nil {
		resp.Diagnostics.AddError("Import Error",
			"Import ID must be in the format \"<organization_id>/<resource_id>\"")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
