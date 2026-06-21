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

var _ resource.Resource = &EnvironmentResource{}

func NewEnvironmentResource() resource.Resource { return &EnvironmentResource{} }

type EnvironmentResource struct{ client *mazevault.Client }

type EnvironmentModel struct {
	OrganizationID         types.String `tfsdk:"organization_id"`
	Name                   types.String `tfsdk:"name"`
	Slug                   types.String `tfsdk:"slug"`
	IsProduction           types.Bool   `tfsdk:"is_production"`
	IncidentAutoEscalation types.Bool   `tfsdk:"incident_auto_escalation"`
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an environment tier within a MazeVault organization (e.g. `production`, `staging`, `development`).",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the organization that owns this environment.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Machine-readable environment name (e.g. `production`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "URL-safe slug for the environment. Derived from name if omitted.",
			},
			"is_production": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether this environment is considered production (affects rotation policies and alerting).",
			},
			"incident_auto_escalation": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether incidents in this environment are automatically escalated.",
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EnvironmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.AddEnvironment(data.OrganizationID.ValueString(), &mazevault.CreateEnvironmentRequest{
		Name:                   data.Name.ValueString(),
		Slug:                   data.Slug.ValueString(),
		IsProduction:           data.IsProduction.ValueBool(),
		IncidentAutoEscalation: data.IncidentAutoEscalation.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Environment Error", fmt.Sprintf("Unable to create environment: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	envs, err := r.client.ListEnvironments(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Environments Error", fmt.Sprintf("Unable to list environments: %s", err))
		return
	}
	for _, e := range envs {
		if e.Name == data.Name.ValueString() {
			data.Slug = types.StringValue(e.Slug)
			data.IsProduction = types.BoolValue(e.IsProduction)
			data.IncidentAutoEscalation = types.BoolValue(e.IncidentAutoEscalation)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EnvironmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateEnvironment(state.OrganizationID.ValueString(), state.Name.ValueString(), &mazevault.UpdateEnvironmentRequest{
		Name: plan.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Environment Error", fmt.Sprintf("Unable to update environment: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveEnvironment(data.OrganizationID.ValueString(), data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Environment Error", fmt.Sprintf("Unable to delete environment: %s", err))
	}
}
