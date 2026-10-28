package resources

import (
	"context"
	"fmt"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CertificateTemplateResource{}
var _ resource.ResourceWithImportState = &CertificateTemplateResource{}

func NewCertificateTemplateResource() resource.Resource {
	return &CertificateTemplateResource{}
}

type CertificateTemplateResource struct {
	client *mazevault.Client
}

// CertificateTemplateResourceModel models a certificate-category Project Template.
//
// It is created via POST /api/v1/organizations/:id/templates with
// category="certificates" so it appears in the frontend "Project Templates" view.
// The backend automatically creates (or reuses) the underlying CertificateTemplate
// and returns its ID in protocol_config.certificate_template_id.
type CertificateTemplateResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ProjectID             types.String `tfsdk:"project_id"`
	OrganizationID        types.String `tfsdk:"organization_id"`
	Name                  types.String `tfsdk:"name"`
	Type                  types.String `tfsdk:"type"`
	ValidityPeriod        types.String `tfsdk:"validity_period"`
	KeyUsage              types.List   `tfsdk:"key_usage"`
	ChallengeType         types.String `tfsdk:"challenge_type"`
	DCVMethod             types.String `tfsdk:"dcv_method"`
	CAAccountID           types.String `tfsdk:"ca_account_id"`
	CAProductID           types.String `tfsdk:"ca_product_id"`
	OrderYears            types.Int64  `tfsdk:"order_years"`
	TechContact           types.Map    `tfsdk:"tech_contact"`
	CertificateTemplateID types.String `tfsdk:"certificate_template_id"`
}

func (r *CertificateTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_template"
}

func (r *CertificateTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a certificate-category **Project Template**. The template is created at the " +
			"organization level (resolved from `project_id`) with category `certificates`, so it appears in the " +
			"frontend **Project Templates** view. MazeVault automatically creates the underlying certificate " +
			"template (exposed as `certificate_template_id`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the project template.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of a certificate-management project. Used to resolve the organization the template belongs to. Changing this forces replacement.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"organization_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the organization that owns this template (resolved from the project).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unique name for this certificate template.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Certificate type / subtype (e.g. `ssl_tls`, `web_server`, `client_auth`, `code_signing`).",
			},
			"validity_period": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Duration string for certificate validity (e.g. `8760h` for 1 year). Converted to whole days for the template.",
			},
			"key_usage": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of X.509 key usage extensions to include (e.g. `digitalSignature`, `keyEncipherment`).",
			},
			"challenge_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ACME challenge policy used when the issuing CA account is an ACME provider. One of `auto`, `dns-01`, `http-01`, `tls-alpn-01`. Defaults to `auto`. Ignored for non-ACME issuers.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dcv_method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Per-template domain control validation method for external commercial CAs that require it (e.g. SSLMarket: `email`, `dns`, `file`). Empty means the provider/account default. Ignored by internal and ACME issuers.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ca_account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the organization CA account that issues certificates from this template. Required for automatic renewal.",
			},
			"ca_product_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of a product from the CA account catalog (commercial CAs such as SSLMarket). Determines validation level, wildcard/SAN limits and the maximum subscription length.",
			},
			"order_years": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Subscription length in years for commercial CAs with multi-year orders (SSLMarket, 1–6). This is the length of the paid order, not of the certificate; within the subscription the CA issues follow-up certificates itself and MazeVault adopts them without placing a new order. Empty means 1 year.",
			},
			"tech_contact": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Technical contact forwarded with commercial CA orders. Keys: `title`, `firstname`, `lastname`, `email`, `tel` (format `+420.123456789`). Individual requests may override it; empty means the CA account default.",
			},
			"certificate_template_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the underlying certificate template automatically created by MazeVault.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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

	// Certificate templates are organization-scoped project templates. Resolve the
	// organization from the project.
	project, err := r.client.GetProject(data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create Template Error", fmt.Sprintf("Unable to resolve project %s: %s", data.ProjectID.ValueString(), err))
		return
	}
	if project == nil {
		resp.Diagnostics.AddError("Create Template Error", fmt.Sprintf("Project %s not found", data.ProjectID.ValueString()))
		return
	}

	input := r.buildInput(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.CreateOrgProjectTemplate(project.OrganizationID, input)
	if err != nil {
		resp.Diagnostics.AddError("Create Template Error", fmt.Sprintf("Unable to create certificate project template: %s", err))
		return
	}

	data.ID = types.StringValue(tmpl.ID)
	data.OrganizationID = types.StringValue(project.OrganizationID)
	applyProjectTemplateComputed(&data, tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		// Fall back to resolving via the project (e.g. after import).
		if project, err := r.client.GetProject(data.ProjectID.ValueString()); err == nil && project != nil {
			orgID = project.OrganizationID
			data.OrganizationID = types.StringValue(orgID)
		}
	}

	tmpl, err := r.client.GetOrgProjectTemplate(orgID, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Template Error", fmt.Sprintf("Unable to read certificate project template: %s", err))
		return
	}
	if tmpl == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(tmpl.Name)
	// `type`/`validity_period` are stored in protocol_config as cert_type/validity_days
	// and are not round-tripped in the same representation, so preserve the
	// configured values instead of overwriting them.
	applyProjectTemplateComputed(&data, tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := r.buildInput(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.UpdateOrgProjectTemplate(state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Update Template Error", fmt.Sprintf("Unable to update certificate project template: %s", err))
		return
	}

	plan.ID = state.ID
	plan.OrganizationID = state.OrganizationID
	applyProjectTemplateComputed(&plan, tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOrgProjectTemplate(data.ID.ValueString()); err != nil && !mazevault.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Delete Template Error", fmt.Sprintf("Unable to delete certificate project template: %s", err))
		return
	}
}

// buildInput assembles the ProjectTemplateInput (category="certificates") from the
// resource model. The certificate profile fields live under protocol_config, which
// the backend converts into the underlying CertificateTemplate.
func (r *CertificateTemplateResource) buildInput(ctx context.Context, data *CertificateTemplateResourceModel, diags *diag.Diagnostics) *mazevault.ProjectTemplateInput {
	var keyUsage []string
	if !data.KeyUsage.IsNull() && !data.KeyUsage.IsUnknown() {
		diags.Append(data.KeyUsage.ElementsAs(ctx, &keyUsage, false)...)
	}

	protocol := map[string]interface{}{
		"cert_type":     data.Type.ValueString(),
		"key_algorithm": "RSA",
		"key_size":      2048,
	}
	if days, ok := parseValidityDays(data.ValidityPeriod.ValueString()); ok {
		protocol["validity_days"] = days
	}
	if len(keyUsage) > 0 {
		protocol["key_usage"] = keyUsage
	}
	if ct := data.ChallengeType.ValueString(); ct != "" && !data.ChallengeType.IsUnknown() {
		protocol["challenge_type"] = ct
	}
	if dm := data.DCVMethod.ValueString(); dm != "" && !data.DCVMethod.IsUnknown() {
		protocol["dcv_method"] = dm
	}
	if v := data.CAAccountID.ValueString(); v != "" && !data.CAAccountID.IsUnknown() {
		protocol["ca_account_id"] = v
	}
	if v := data.CAProductID.ValueString(); v != "" && !data.CAProductID.IsUnknown() {
		protocol["ca_product_id"] = v
	}
	if !data.OrderYears.IsNull() && !data.OrderYears.IsUnknown() && data.OrderYears.ValueInt64() > 0 {
		protocol["order_years"] = data.OrderYears.ValueInt64()
	}
	if !data.TechContact.IsNull() && !data.TechContact.IsUnknown() {
		contact := map[string]string{}
		diags.Append(data.TechContact.ElementsAs(ctx, &contact, false)...)
		if len(contact) > 0 {
			protocol["tech_contact"] = contact
		}
	}

	return &mazevault.ProjectTemplateInput{
		Name:           data.Name.ValueString(),
		Category:       "certificates",
		Subtype:        data.Type.ValueString(),
		ProtocolConfig: protocol,
	}
}

// applyProjectTemplateComputed populates computed fields from the API response.
func applyProjectTemplateComputed(data *CertificateTemplateResourceModel, tmpl *mazevault.ProjectTemplate) {
	if tmpl.ProtocolConfig != nil {
		if ctid, ok := tmpl.ProtocolConfig["certificate_template_id"].(string); ok && ctid != "" {
			data.CertificateTemplateID = types.StringValue(ctid)
		}
		if data.ChallengeType.IsNull() || data.ChallengeType.IsUnknown() {
			if ct, ok := tmpl.ProtocolConfig["challenge_type"].(string); ok && ct != "" {
				data.ChallengeType = types.StringValue(ct)
			}
		}
		if data.DCVMethod.IsNull() || data.DCVMethod.IsUnknown() {
			if dm, ok := tmpl.ProtocolConfig["dcv_method"].(string); ok && dm != "" {
				data.DCVMethod = types.StringValue(dm)
			}
		}
	}
	if data.ChallengeType.IsNull() || data.ChallengeType.IsUnknown() {
		data.ChallengeType = types.StringValue("auto")
	}
	if data.DCVMethod.IsUnknown() {
		data.DCVMethod = types.StringNull()
	}
	if data.CertificateTemplateID.IsUnknown() {
		data.CertificateTemplateID = types.StringNull()
	}
}

// parseValidityDays converts a Go duration string (e.g. "8760h") to whole days.
func parseValidityDays(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	if d, err := time.ParseDuration(s); err == nil {
		return int(d.Hours() / 24), true
	}
	return 0, false
}

// ImportState implements resource.ResourceWithImportState.
func (r *CertificateTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
