package resources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SecretLinkResource{}

func NewSecretLinkResource() resource.Resource { return &SecretLinkResource{} }

type SecretLinkResource struct{ client *mazevault.Client }

type SecretLinkModel struct {
	ID            types.String `tfsdk:"id"`
	SecretID      types.String `tfsdk:"secret_id"`
	IntegrationID types.String `tfsdk:"integration_id"`
	DatabaseUser  types.String `tfsdk:"database_username"`
	TargetPath    types.String `tfsdk:"target_path"`
	FileFormat    types.String `tfsdk:"file_format"`
	SecretKey     types.String `tfsdk:"secret_key"`
	VariableName  types.String `tfsdk:"variable_name"`
	Environment   types.String `tfsdk:"environment"`
}

func (r *SecretLinkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_link"
}

func (r *SecretLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Links a secret to an integration (database/agent/devops)",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Computed: true},
			"secret_id":         schema.StringAttribute{Required: true},
			"integration_id":    schema.StringAttribute{Required: true},
			"database_username": schema.StringAttribute{Optional: true},
			"target_path":       schema.StringAttribute{Optional: true},
			"file_format":       schema.StringAttribute{Optional: true},
			"secret_key":        schema.StringAttribute{Optional: true},
			"variable_name":     schema.StringAttribute{Optional: true},
			"environment":       schema.StringAttribute{Optional: true},
		},
	}
}

func (r *SecretLinkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *SecretLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretLinkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload := map[string]interface{}{
		"integration_id": data.IntegrationID.ValueString(),
	}
	set := func(k string, v types.String) {
		if !v.IsNull() && v.ValueString() != "" {
			payload[k] = v.ValueString()
		}
	}
	set("database_username", data.DatabaseUser)
	set("target_path", data.TargetPath)
	set("file_format", data.FileFormat)
	set("secret_key", data.SecretKey)
	set("variable_name", data.VariableName)
	set("environment", data.Environment)
	created, err := r.client.CreateSecretLink(data.SecretID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Secret Link Error", err.Error())
		return
	}
	data.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	links, err := r.client.ListSecretLinks(data.SecretID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Secret Link Error", err.Error())
		return
	}
	for _, l := range links {
		if l.ID == data.ID.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddWarning("Secret Link Not Found", "The secret link was not found on the server")
}

func (r *SecretLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Simplified: no update endpoint; re-create if needed
	var data SecretLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddWarning("No Update", "Updating secret links is not supported; recreate to change fields")
}

func (r *SecretLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretLinkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSecretLink(data.SecretID.ValueString(), data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Secret Link Error", err.Error())
		return
	}
}
