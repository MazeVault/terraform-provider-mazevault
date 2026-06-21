package datasources

import (
	"context"
	"fmt"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecretDataSource{}

func NewSecretDataSource() datasource.DataSource { return &SecretDataSource{} }

type SecretDataSource struct{ client *mazevault.Client }

type SecretDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Key         types.String `tfsdk:"key"`
	Value       types.String `tfsdk:"value"`
	Environment types.String `tfsdk:"environment"`
	Version     types.Int64  `tfsdk:"version"`
}

func (d *SecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *SecretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a secret value from a MazeVault project. The value is marked Sensitive.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Secret ID. Provide either `id` or both `project_id` + `key`.",
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Project containing the secret.",
			},
			"key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Secret key name.",
			},
			"value": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Secret value.",
			},
			"environment": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Environment the secret belongs to.",
			},
			"version": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Current version number.",
			},
		},
	}
}

func (d *SecretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mazevault.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *SecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecretDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var secret *mazevault.Secret
	var err error
	// Detect unknown id early: ValueString() returns "" for unknown values, which
	// would silently fall through to the project_id/key branch and produce a
	// confusing "Missing project_id" error instead of a clear diagnostic.
	if !data.ID.IsNull() && data.ID.IsUnknown() {
		if data.ProjectID.IsNull() || data.ProjectID.ValueString() == "" ||
			data.Key.IsNull() || data.Key.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Unknown id",
				"The id attribute is not yet known (it references an unapplied value). "+
					"Either provide a known id, or set both project_id and key to look up a secret.",
			)
			return
		}
		// project_id + key are available — fall through to that lookup below.
	}
	if !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != "" {
		secret, err = d.client.GetSecretByID(data.ID.ValueString())
	} else {
		if data.ProjectID.IsNull() || data.ProjectID.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing project_id",
				"Either id or both project_id and key must be set to look up a secret.",
			)
			return
		}
		if data.Key.IsNull() || data.Key.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing key",
				"Either id or both project_id and key must be set to look up a secret.",
			)
			return
		}
		secret, err = d.client.GetSecret(data.ProjectID.ValueString(), data.Key.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Secret Error", fmt.Sprintf("Unable to read secret: %s", err))
		return
	}
	data.ID = types.StringValue(secret.ID)
	data.ProjectID = types.StringValue(secret.ProjectID)
	data.Key = types.StringValue(secret.Key)
	data.Value = types.StringValue(secret.Value)
	data.Environment = types.StringValue(secret.Environment)
	data.Version = types.Int64Value(int64(secret.Version))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
