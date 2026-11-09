package datasources

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EntraAppRegistrationDataSource{}

func NewEntraAppRegistrationDataSource() datasource.DataSource {
	return &EntraAppRegistrationDataSource{}
}

type EntraAppRegistrationDataSource struct{ client *mazevault.Client }

type EntraAppRegistrationDataModel struct {
	IntegrationID         types.String          `tfsdk:"integration_id"`
	AppID                 types.String          `tfsdk:"app_id"`
	ID                    types.String          `tfsdk:"id"`
	ObjectID              types.String          `tfsdk:"object_id"`
	TenantID              types.String          `tfsdk:"tenant_id"`
	ProjectID             types.String          `tfsdk:"project_id"`
	DisplayName           types.String          `tfsdk:"display_name"`
	ManagedByMazeVault    types.Bool            `tfsdk:"managed_by_mazevault"`
	SyncStatus            types.String          `tfsdk:"sync_status"`
	PrimaryCredentialID   types.String          `tfsdk:"primary_credential_id"`
	SecondaryCredentialID types.String          `tfsdk:"secondary_credential_id"`
	Credentials           []EntraCredentialItem `tfsdk:"credentials"`
}

type EntraCredentialItem struct {
	ID              types.String `tfsdk:"id"`
	KeyID           types.String `tfsdk:"key_id"`
	DisplayName     types.String `tfsdk:"display_name"`
	CredentialType  types.String `tfsdk:"credential_type"`
	IsPrimary       types.Bool   `tfsdk:"is_primary"`
	Status          types.String `tfsdk:"status"`
	LifecycleState  types.String `tfsdk:"lifecycle_state"`
	RotationEnabled types.Bool   `tfsdk:"rotation_enabled"`
	StartDate       types.String `tfsdk:"start_date"`
	EndDate         types.String `tfsdk:"end_date"`
}

func (d *EntraAppRegistrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entra_app_registration"
}

func (d *EntraAppRegistrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Microsoft Entra ID application registration synchronised into a project Entra integration " +
			"by its application (client) ID and exposes its MazeVault identifiers and credentials. Use it to feed " +
			"`credential_id` into `mazevault_entra_rotation_config` after `mazevault_entra_integration_scope` has selected " +
			"and synchronised the application. Secret values are never exposed.",
		Attributes: map[string]schema.Attribute{
			"integration_id": schema.StringAttribute{Required: true, MarkdownDescription: "ID of the project Entra integration."},
			"app_id":         schema.StringAttribute{Required: true, MarkdownDescription: "Entra application (client) ID to look up."},
			"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "MazeVault app registration ID."},
			"object_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Entra object ID."},
			"tenant_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Entra tenant ID."},
			"project_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Owning project ID."},
			"display_name":   schema.StringAttribute{Computed: true, MarkdownDescription: "Display name from Entra ID."},
			"managed_by_mazevault": schema.BoolAttribute{Computed: true,
				MarkdownDescription: "Whether MazeVault owns the credential lifecycle of this application."},
			"sync_status": schema.StringAttribute{Computed: true, MarkdownDescription: "Last synchronisation status."},
			"primary_credential_id": schema.StringAttribute{Computed: true,
				MarkdownDescription: "ID of the active primary client-secret credential (empty when none)."},
			"secondary_credential_id": schema.StringAttribute{Computed: true,
				MarkdownDescription: "ID of the most recently issued non-primary active client-secret credential (empty when none)."},
			"credentials": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "All credentials of the application, newest end date first.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"key_id":           schema.StringAttribute{Computed: true},
						"display_name":     schema.StringAttribute{Computed: true},
						"credential_type":  schema.StringAttribute{Computed: true},
						"is_primary":       schema.BoolAttribute{Computed: true},
						"status":           schema.StringAttribute{Computed: true},
						"lifecycle_state":  schema.StringAttribute{Computed: true},
						"rotation_enabled": schema.BoolAttribute{Computed: true},
						"start_date":       schema.StringAttribute{Computed: true},
						"end_date":         schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EntraAppRegistrationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EntraAppRegistrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EntraAppRegistrationDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apps, err := d.client.ListEntraAppRegistrations(data.IntegrationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Entra App Registrations Error", fmt.Sprintf("Unable to list Entra applications: %s", err))
		return
	}
	wanted := strings.ToLower(strings.TrimSpace(data.AppID.ValueString()))
	var app *mazevault.EntraAppRegistration
	for i := range apps {
		if strings.ToLower(apps[i].AppID) == wanted {
			app = &apps[i]
			break
		}
	}
	if app == nil {
		resp.Diagnostics.AddError("Entra Application Not Found",
			fmt.Sprintf("No synchronised application with app_id %s exists in integration %s. Select it with "+
				"mazevault_entra_integration_scope and make sure the integration has been synchronised.",
				data.AppID.ValueString(), data.IntegrationID.ValueString()))
		return
	}

	creds, err := d.client.ListEntraAppCredentials(app.ID)
	if err != nil {
		resp.Diagnostics.AddError("Read Entra Credentials Error", fmt.Sprintf("Unable to list credentials: %s", err))
		return
	}
	sort.SliceStable(creds, func(i, j int) bool { return creds[i].EndDate.After(creds[j].EndDate) })

	data.ID = types.StringValue(app.ID)
	data.ObjectID = types.StringValue(app.ObjectID)
	data.TenantID = types.StringValue(app.TenantID)
	data.ProjectID = types.StringValue(app.ProjectID)
	data.DisplayName = types.StringValue(app.DisplayName)
	data.ManagedByMazeVault = types.BoolValue(app.ManagedByMazeVault)
	data.SyncStatus = types.StringValue(app.SyncStatus)
	data.PrimaryCredentialID = types.StringValue("")
	data.SecondaryCredentialID = types.StringValue("")
	data.Credentials = make([]EntraCredentialItem, 0, len(creds))
	for _, c := range creds {
		isSecret := c.CredentialType == "client_secret"
		active := strings.EqualFold(c.Status, "active")
		if isSecret && active {
			if c.IsPrimary && data.PrimaryCredentialID.ValueString() == "" {
				data.PrimaryCredentialID = types.StringValue(c.ID)
			} else if !c.IsPrimary && data.SecondaryCredentialID.ValueString() == "" {
				data.SecondaryCredentialID = types.StringValue(c.ID)
			}
		}
		data.Credentials = append(data.Credentials, EntraCredentialItem{
			ID:              types.StringValue(c.ID),
			KeyID:           types.StringValue(c.KeyID),
			DisplayName:     types.StringValue(c.DisplayName),
			CredentialType:  types.StringValue(c.CredentialType),
			IsPrimary:       types.BoolValue(c.IsPrimary),
			Status:          types.StringValue(c.Status),
			LifecycleState:  types.StringValue(c.LifecycleState),
			RotationEnabled: types.BoolValue(c.RotationEnabled),
			StartDate:       types.StringValue(c.StartDate.Format(time.RFC3339)),
			EndDate:         types.StringValue(c.EndDate.Format(time.RFC3339)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
