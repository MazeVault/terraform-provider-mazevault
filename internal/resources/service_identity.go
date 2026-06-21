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

var _ resource.Resource = &ServiceIdentityResource{}

func NewServiceIdentityResource() resource.Resource {
	return &ServiceIdentityResource{}
}

type ServiceIdentityResource struct {
	client *mazevault.Client
}

type ServiceIdentityResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DisplayName  types.String `tfsdk:"display_name"`
	Description  types.String `tfsdk:"description"`
	OwnerEmail   types.String `tfsdk:"owner_email"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (r *ServiceIdentityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_identity"
}

func (r *ServiceIdentityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MazeVault service identity (machine account) with auto-generated client credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the service identity.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable display name for the service identity.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional description of the service identity's purpose.",
			},
			"owner_email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Email of the owner responsible for this service identity.",
			},
			"client_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Auto-generated OAuth2 client ID for this service identity.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"client_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Auto-generated OAuth2 client secret. Only available at creation time.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *ServiceIdentityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ServiceIdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &mazevault.CreateServiceIdentityRequest{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		OwnerEmail:  data.OwnerEmail.ValueString(),
	}

	created, err := r.client.CreateServiceIdentity(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Create Service Identity Error", fmt.Sprintf("Unable to create service identity: %s", err))
		return
	}

	data.ID = types.StringValue(created.ID)
	data.ClientID = types.StringValue(created.ClientID)
	data.ClientSecret = types.StringValue(created.ClientSecret)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceIdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceIdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identity, err := r.client.GetServiceIdentity(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Service Identity Error", fmt.Sprintf("Unable to read service identity: %s", err))
		return
	}
	if identity == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.DisplayName = types.StringValue(identity.DisplayName)
	data.Description = types.StringValue(identity.Description)
	data.OwnerEmail = types.StringValue(identity.OwnerEmail)
	data.ClientID = types.StringValue(identity.ClientID)
	// ClientSecret is write-only; preserve existing state value

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceIdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ServiceIdentityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &mazevault.CreateServiceIdentityRequest{
		DisplayName: plan.DisplayName.ValueString(),
		Description: plan.Description.ValueString(),
		OwnerEmail:  plan.OwnerEmail.ValueString(),
	}

	updated, err := r.client.UpdateServiceIdentity(state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Update Service Identity Error", fmt.Sprintf("Unable to update service identity: %s", err))
		return
	}

	plan.ID = types.StringValue(updated.ID)
	plan.ClientID = types.StringValue(updated.ClientID)
	// Preserve the existing client secret in state (update does not rotate it)
	plan.ClientSecret = state.ClientSecret

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServiceIdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceIdentityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServiceIdentity(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Service Identity Error", fmt.Sprintf("Unable to delete service identity: %s", err))
	}
}
