package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CAResource{}
var _ resource.ResourceWithImportState = &CAResource{}

func NewCAResource() resource.Resource { return &CAResource{} }

type CAResource struct{ client *mazevault.Client }

type CAResourceModel struct {
	ID             types.String `tfsdk:"id"`
	AccountID      types.String `tfsdk:"account_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ValidYears     types.Int64  `tfsdk:"valid_years"`
	KeySize        types.Int64  `tfsdk:"key_size"`
	KeyAlgorithm   types.String `tfsdk:"key_algorithm"`
	Type           types.String `tfsdk:"type"`
	Status         types.String `tfsdk:"status"`
	NotAfter       types.String `tfsdk:"not_after"`
	SerialNumber   types.String `tfsdk:"serial_number"`
	OCSPURL        types.String `tfsdk:"ocsp_url"`
	CRLURL         types.String `tfsdk:"crl_url"`
}

func (r *CAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca"
}

func (r *CAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the single internal root Certificate Authority (CA) for a MazeVault organization. " +
			"Each organization has at most one internal CA; certificate templates and issued certificates chain to it. " +
			"The CA parameters are immutable after creation — changing any of them replaces the CA. Deleting the " +
			"resource soft-deletes the internal CA (rejected while active certificates were issued by it).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the internal CA (the certificate authority).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"account_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the internal CA account (OrganizationCAAccount). Pass this as `organization_ca_account_id` on `mazevault_certificate` to issue certificates from this internal CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization to initialize the internal CA in.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Common name / display name of the internal root CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"valid_years": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of years the CA certificate will be valid.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"key_size": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Key size for the CA. For RSA use 2048 or 4096; for ECDSA use 256 (P-256), 384 (P-384) or 521 (P-521).",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"key_algorithm": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Key algorithm for the CA: `RSA` (default) or `ECDSA`. Pair with a matching `key_size` (RSA: 2048/4096; ECDSA: 256/384/521).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CA type as returned by the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current operational status of the CA.",
			},
			"not_after": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 timestamp until which the CA certificate is valid.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"serial_number": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Serial number of the CA certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ocsp_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OCSP responder URL published for certificates issued by this CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"crl_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CRL Distribution Point (CDP) URL published for certificates issued by this CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *CAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CAResource) applyComputed(data *CAResourceModel, ca *mazevault.InternalCA) {
	data.ID = types.StringValue(ca.ID)
	if ca.AccountID != "" {
		data.AccountID = types.StringValue(ca.AccountID)
	}
	data.Type = types.StringValue(ca.Type)
	data.Status = types.StringValue(ca.Status)
	data.NotAfter = types.StringValue(ca.NotAfter)
	data.SerialNumber = types.StringValue(ca.SerialNumber)
	data.OCSPURL = types.StringValue(ca.OCSPURL)
	data.CRLURL = types.StringValue(ca.CRLURL)
	if ca.KeyAlgorithm != "" {
		data.KeyAlgorithm = types.StringValue(ca.KeyAlgorithm)
	}
	if ca.KeySize > 0 {
		data.KeySize = types.Int64Value(int64(ca.KeySize))
	}
}

func (r *CAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := r.client.InitializeInternalCA(data.OrganizationID.ValueString(), &mazevault.InitializeInternalCARequest{
		CommonName:    data.Name.ValueString(),
		ValidityYears: int(data.ValidYears.ValueInt64()),
		KeySize:       int(data.KeySize.ValueInt64()),
		KeyAlgorithm:  data.KeyAlgorithm.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create CA Error", fmt.Sprintf("Unable to initialize internal CA: %s", err))
		return
	}

	r.applyComputed(&data, ca)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := r.client.GetInternalCA(data.OrganizationID.ValueString())
	if err != nil {
		if mazevault.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read CA Error", fmt.Sprintf("Unable to read internal CA: %s", err))
		return
	}
	if ca == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if ca.Name != "" {
		data.Name = types.StringValue(ca.Name)
	}
	r.applyComputed(&data, ca)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// The internal CA is immutable: every configurable attribute is marked
	// RequiresReplace, so Update is only reached for computed drift. Persist the
	// planned values unchanged.
	var data CAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteInternalCA(data.OrganizationID.ValueString()); err != nil {
		if mazevault.IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Delete CA Error", fmt.Sprintf("Unable to delete internal CA: %s", err))
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID is the organization ID (the internal CA is a singleton per org).
func (r *CAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), req.ID)...)
}
