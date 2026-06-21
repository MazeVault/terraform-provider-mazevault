package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CAResource{}

func NewCAResource() resource.Resource { return &CAResource{} }

type CAResource struct{ client *mazevault.Client }

type CAResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	Name       types.String `tfsdk:"name"`
	ValidYears types.Int64  `tfsdk:"valid_years"`
	KeySize    types.Int64  `tfsdk:"key_size"`
	Type       types.String `tfsdk:"type"`
	Status     types.String `tfsdk:"status"`
	ValidUntil types.String `tfsdk:"valid_until"`
}

func (r *CAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ca"
}

func (r *CAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Initializes an internal root Certificate Authority (CA) for a MazeVault project. This is a one-time operation — the CA cannot be modified or deleted after creation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the project CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project to initialize the CA in.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the root CA.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"valid_years": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of years the CA certificate will be valid.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"key_size": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "RSA key size for the CA (e.g. 2048, 4096).",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
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
			"valid_until": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 timestamp until which the CA certificate is valid.",
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

func (r *CAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CAResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := r.client.InitializeProjectCA(data.ProjectID.ValueString(), &mazevault.InitializeProjectCARequest{
		Name:       data.Name.ValueString(),
		ValidYears: int(data.ValidYears.ValueInt64()),
		KeySize:    int(data.KeySize.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create CA Error", fmt.Sprintf("Unable to initialize project CA: %s", err))
		return
	}

	data.ID = types.StringValue(ca.ID)
	data.Type = types.StringValue(ca.Type)
	data.Status = types.StringValue(ca.Status)
	data.ValidUntil = types.StringValue(ca.ValidUntil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CAResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ca, err := r.client.GetProjectCA(data.ProjectID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read CA Error", fmt.Sprintf("Unable to read project CA: %s", err))
		return
	}
	if ca == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(ca.Name)
	data.Type = types.StringValue(ca.Type)
	data.Status = types.StringValue(ca.Status)
	data.ValidUntil = types.StringValue(ca.ValidUntil)
	if ca.ProjectID != nil {
		data.ProjectID = types.StringValue(*ca.ProjectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CAResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// CA initialization parameters are immutable; all mutable fields use RequiresReplace.
}

func (r *CAResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// The MazeVault backend does not expose a delete endpoint for project CAs.
	// Removing from Terraform state only; the CA remains in MazeVault.
}
