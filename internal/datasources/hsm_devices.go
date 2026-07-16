package datasources

import (
	"context"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &HSMDevicesDataSource{}

func NewHSMDevicesDataSource() datasource.DataSource { return &HSMDevicesDataSource{} }

type HSMDevicesDataSource struct{ client *mazevault.Client }

type HSMDevicesDataSourceModel struct {
	OrgID   types.String    `tfsdk:"organization_id"`
	Devices []HSMDeviceItem `tfsdk:"devices"`
}

type HSMDeviceItem struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Provider     types.String `tfsdk:"provider_type"`
	Priority     types.Int64  `tfsdk:"priority"`
	Status       types.String `tfsdk:"status"`
	HealthStatus types.String `tfsdk:"health_status"`
	IsKEKDevice  types.Bool   `tfsdk:"is_kek_device"`
}

func (d *HSMDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hsm_devices"
}

func (d *HSMDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all HSM/KMS devices configured for an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The organization to list devices for.",
			},
			"devices": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of HSM devices.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"description":   schema.StringAttribute{Computed: true},
						"provider_type": schema.StringAttribute{Computed: true},
						"priority":      schema.Int64Attribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"health_status": schema.StringAttribute{Computed: true},
						"is_kek_device": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *HSMDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*mazevault.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *mazevault.Client")
		return
	}
	d.client = client
}

func (d *HSMDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HSMDevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ListHSMDevices(data.OrgID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list HSM devices", err.Error())
		return
	}

	data.Devices = make([]HSMDeviceItem, 0, len(result.Devices))
	for _, dev := range result.Devices {
		data.Devices = append(data.Devices, HSMDeviceItem{
			ID:           types.StringValue(dev.ID),
			Name:         types.StringValue(dev.Name),
			Description:  types.StringValue(dev.Description),
			Provider:     types.StringValue(dev.Provider),
			Priority:     types.Int64Value(int64(dev.Priority)),
			Status:       types.StringValue(dev.Status),
			HealthStatus: types.StringValue(dev.HealthStatus),
			IsKEKDevice:  types.BoolValue(dev.IsKEKDevice),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
