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

var _ resource.Resource = &KeytabResource{}

func NewKeytabResource() resource.Resource { return &KeytabResource{} }

type KeytabResource struct{ client *mazevault.Client }

type KeytabModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	KeytabBase64 types.String `tfsdk:"keytab_base64"`
	Principal    types.String `tfsdk:"principal"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (r *KeytabResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keytab"
}

func (r *KeytabResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Imports a Kerberos keytab file into MazeVault for agent authentication.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the keytab.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the keytab.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the keytab's purpose.",
			},
			"keytab_base64": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Base64-encoded keytab file content.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"principal": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Kerberos principal extracted from the keytab.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO-8601 timestamp when the keytab was imported.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *KeytabResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeytabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeytabModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.ImportKeytab(&mazevault.ImportKeytabRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		KeytabB64:   data.KeytabBase64.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Import Keytab Error", fmt.Sprintf("Unable to import keytab: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.Principal = types.StringValue(created.Principal)
	data.CreatedAt = types.StringValue(created.CreatedAt.Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeytabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeytabModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	kt, err := r.client.GetKeytab(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Keytab Error", fmt.Sprintf("Unable to read keytab: %s", err))
		return
	}
	if kt == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.Name = types.StringValue(kt.Name)
	data.Description = types.StringValue(kt.Description)
	data.Principal = types.StringValue(kt.Principal)
	data.CreatedAt = types.StringValue(kt.CreatedAt.Format(time.RFC3339))
	// keytab_base64 is write-only; preserve from state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeytabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KeytabModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.client.UpdateKeytab(state.ID.ValueString(), &mazevault.UpdateKeytabRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Update Keytab Error", fmt.Sprintf("Unable to update keytab: %s", err))
		return
	}
	plan.ID = state.ID
	plan.Principal = state.Principal
	plan.CreatedAt = state.CreatedAt
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KeytabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeytabModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteKeytab(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Keytab Error", fmt.Sprintf("Unable to delete keytab: %s", err))
	}
}
