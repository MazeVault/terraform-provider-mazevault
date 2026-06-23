package resources

import (
	"context"
	"fmt"
	"strings"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithConfigure = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

// NewOrganizationResource is a helper function to simplify the provider implementation.
func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

// OrganizationResource is the resource implementation.
type OrganizationResource struct {
	client *mazevault.Client
}

// OrganizationResourceModel maps the resource schema data.
type OrganizationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *OrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the resource.
func (r *OrganizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MazeVault organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Organization name",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Last update timestamp",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *OrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.CreateOrganization(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating organization",
			"Could not create organization, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(org.ID)
	plan.CreatedAt = types.StringValue(org.CreatedAt.String())
	plan.UpdatedAt = types.StringValue(org.UpdatedAt.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading organization",
			"Could not read organization ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if org == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(org.Name)
	state.CreatedAt = types.StringValue(org.CreatedAt.String())
	state.UpdatedAt = types.StringValue(org.UpdatedAt.String())

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.UpdateOrganization(plan.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating organization",
			"Could not update organization, unexpected error: "+err.Error(),
		)
		return
	}

	plan.UpdatedAt = types.StringValue(org.UpdatedAt.String())
	plan.CreatedAt = types.StringValue(org.CreatedAt.String())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state.
func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganization(state.ID.ValueString())
	if err != nil {
		errMsg := err.Error()
		// Treat expected "no DELETE endpoint" responses as soft removals: the
		// resource is gone from state but the backend object is preserved.
		// Any other error is a hard failure that must surface to the operator.
		if strings.Contains(errMsg, "404") ||
			strings.Contains(errMsg, "405") ||
			strings.Contains(errMsg, "not found") ||
			strings.Contains(errMsg, "not supported") {
			resp.Diagnostics.AddWarning(
				"Organization deletion not supported",
				"The MazeVault backend does not support deleting organizations via API. "+
					"The resource has been removed from Terraform state. Error: "+errMsg,
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting organization",
			"Could not delete organization ID "+state.ID.ValueString()+": "+errMsg,
		)
	}
}

// ImportState imports the resource into Terraform state.
func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
