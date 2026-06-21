package resources

import (
	"context"
	"fmt"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var _ resource.Resource = &APITokenResource{}
var _ resource.ResourceWithConfigure = &APITokenResource{}

// NewAPITokenResource is a helper function to simplify the provider implementation.
func NewAPITokenResource() resource.Resource {
	return &APITokenResource{}
}

// APITokenResource is the resource implementation.
type APITokenResource struct {
	client *mazevault.Client
}

// APITokenResourceModel maps the resource schema data.
type APITokenResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Scopes    types.List   `tfsdk:"scopes"`
	Duration  types.String `tfsdk:"duration"`
	Token     types.String `tfsdk:"token"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

// Metadata returns the resource type name.
func (r *APITokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_token"
}

// Schema defines the schema for the resource.
func (r *APITokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MazeVault API Token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Token UUID",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Token name",
				Required:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "List of scopes/permissions",
				Optional:    true,
				ElementType: types.StringType,
			},
			"duration": schema.StringAttribute{
				Description: "Token validity duration (e.g. 24h)",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "The generated API token (sensitive)",
				Computed:    true,
				Sensitive:   true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Expiration timestamp",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *APITokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *APITokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan APITokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var scopes []string
	if !plan.Scopes.IsNull() {
		diags = plan.Scopes.ElementsAs(ctx, &scopes, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	duration := "24h"
	if !plan.Duration.IsNull() {
		duration = plan.Duration.ValueString()
	}

	token, err := r.client.CreateAPIToken(plan.Name.ValueString(), scopes, duration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating API token",
			"Could not create API token, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(token.ID)
	plan.Token = types.StringValue(token.Token)
	plan.ExpiresAt = types.StringValue(token.ExpiresAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *APITokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state APITokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// API tokens cannot be retrieved individually; list and match by ID.
	tokens, err := r.client.ListAPITokens()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list API tokens: %s", err))
		return
	}

	found := false
	for _, t := range tokens {
		if t.ID == state.ID.ValueString() {
			state.Name = types.StringValue(t.Name)
			state.ExpiresAt = types.StringValue(t.ExpiresAt.Format(time.RFC3339))
			found = true
			break
		}
	}

	if !found {
		// Token was revoked outside of Terraform — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *APITokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API Tokens are immutable usually. Force replacement?
	// For now, return error or implement replacement logic
	resp.Diagnostics.AddError("Update not supported", "API Tokens cannot be updated. Destroy and recreate.")
}

// Delete deletes the resource and removes the Terraform state.
func (r *APITokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state APITokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RevokeAPIToken(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error revoking API token",
			"Could not revoke API token, unexpected error: "+err.Error(),
		)
		return
	}
}
