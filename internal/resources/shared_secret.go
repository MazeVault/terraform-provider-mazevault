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

var _ resource.Resource = &SharedSecretResource{}

func NewSharedSecretResource() resource.Resource { return &SharedSecretResource{} }

type SharedSecretResource struct{ client *mazevault.Client }

type SharedSecretModel struct {
	ID             types.String `tfsdk:"id"`
	SecretID       types.String `tfsdk:"secret_id"`
	ContentType    types.String `tfsdk:"content_type"`
	RecipientEmail types.String `tfsdk:"recipient_email"`
	TTLHours       types.Int64  `tfsdk:"ttl_hours"`
	MaxViews       types.Int64  `tfsdk:"max_views"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	ShareURL       types.String `tfsdk:"share_url"`
}

func (r *SharedSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_secret"
}

func (r *SharedSecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a one-time secret share link. The secret value is encrypted and stored temporarily; the share URL can be sent to a recipient to retrieve it once.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique share token.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secret_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of an existing MazeVault secret to share. Mutually exclusive with inline content.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"content_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "MIME type for the shared content (e.g. `text/plain`, `application/json`).",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"recipient_email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Email address of the intended recipient for audit trail purposes.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ttl_hours": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Time-to-live in hours. After this period the share link expires.",
			},
			"max_views": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of times the secret can be viewed. Defaults to 1.",
			},
			"expires_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ISO-8601 expiry timestamp. If not set, the backend applies a default TTL.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"share_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full URL that can be sent to the recipient to retrieve the secret.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *SharedSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SharedSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SharedSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateSharedSecret(&mazevault.CreateSharedSecretRequest{
		SecretID:       data.SecretID.ValueString(),
		ContentType:    data.ContentType.ValueString(),
		RecipientEmail: data.RecipientEmail.ValueString(),
		TTLHours:       int(data.TTLHours.ValueInt64()),
		MaxViews:       int(data.MaxViews.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Shared Secret Error", fmt.Sprintf("Unable to create shared secret: %s", err))
		return
	}
	data.ID = types.StringValue(created.ID)
	data.ShareURL = types.StringValue(created.ShareURL)
	data.ExpiresAt = types.StringValue(created.ExpiresAt.Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SharedSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SharedSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	meta, err := r.client.GetSharedSecretMetadata(data.ID.ValueString())
	if err != nil {
		// Secret may have been consumed or expired
		resp.State.RemoveResource(ctx)
		return
	}
	if meta == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	data.ExpiresAt = types.StringValue(meta.ExpiresAt.Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SharedSecretResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Shared secrets are immutable; changes require destroy + recreate.
}

func (r *SharedSecretResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Shared secrets auto-expire; removing from state only.
}
