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

var _ resource.Resource = &CertificateTemplateResource{}

func NewCertificateTemplateResource() resource.Resource {
	return &CertificateTemplateResource{}
}

type CertificateTemplateResource struct {
	client *mazevault.Client
}

type CertificateTemplateResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	ValidityPeriod types.String `tfsdk:"validity_period"`
	KeyUsage       types.List   `tfsdk:"key_usage"`
}

func (r *CertificateTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_template"
}

func (r *CertificateTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MazeVault certificate template that defines the profile for issued certificates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the certificate template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project this template belongs to.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique name for this certificate template.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Template type (e.g. `tls_server`, `tls_client`, `code_signing`).",
			},
			"validity_period": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Duration string for certificate validity (e.g. `8760h` for 1 year).",
			},
			"key_usage": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of X.509 key usage extensions to include (e.g. `digitalSignature`, `keyEncipherment`).",
			},
		},
	}
}

func (r *CertificateTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var keyUsage []string
	if !data.KeyUsage.IsNull() {
		resp.Diagnostics.Append(data.KeyUsage.ElementsAs(ctx, &keyUsage, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &mazevault.CreateCertificateTemplateRequest{
		Name:           data.Name.ValueString(),
		Type:           data.Type.ValueString(),
		ValidityPeriod: data.ValidityPeriod.ValueString(),
		KeyUsage:       keyUsage,
	}

	tmpl, err := r.client.CreateProjectCertificateTemplate(data.ProjectID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Create Template Error", fmt.Sprintf("Unable to create certificate template: %s", err))
		return
	}

	data.ID = types.StringValue(tmpl.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.GetCertificateTemplate(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Template Error", fmt.Sprintf("Unable to read certificate template: %s", err))
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(tmpl.Name)
	data.Type = types.StringValue(tmpl.Type)
	data.ValidityPeriod = types.StringValue(tmpl.ValidityPeriod)
	if tmpl.ProjectID != nil {
		data.ProjectID = types.StringValue(*tmpl.ProjectID)
	}

	keyUsageList, diags := types.ListValueFrom(ctx, types.StringType, tmpl.KeyUsage)
	resp.Diagnostics.Append(diags...)
	data.KeyUsage = keyUsageList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var keyUsage []string
	if !plan.KeyUsage.IsNull() {
		resp.Diagnostics.Append(plan.KeyUsage.ElementsAs(ctx, &keyUsage, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, err := r.client.UpdateCertificateTemplate(state.ID.ValueString(), &mazevault.CreateCertificateTemplateRequest{
		Name:           plan.Name.ValueString(),
		Type:           plan.Type.ValueString(),
		ValidityPeriod: plan.ValidityPeriod.ValueString(),
		KeyUsage:       keyUsage,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Template Error", fmt.Sprintf("Unable to update certificate template: %s", err))
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCertificateTemplate(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Template Error", fmt.Sprintf("Unable to delete certificate template: %s", err))
		return
	}
}
