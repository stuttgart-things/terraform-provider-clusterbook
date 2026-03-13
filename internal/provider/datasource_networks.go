package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stuttgart-things/terraform-provider-clusterbook/internal/client"
)

var (
	_ datasource.DataSource              = &networksDataSource{}
	_ datasource.DataSourceWithConfigure = &networksDataSource{}
)

type networksDataSource struct {
	client *client.Client
}

type networksDataSourceModel struct {
	Networks []networkInfoModel `tfsdk:"networks"`
}

type networkInfoModel struct {
	NetworkKey types.String  `tfsdk:"network_key"`
	Total      types.Float64 `tfsdk:"total"`
	Available  types.Float64 `tfsdk:"available"`
	Assigned   types.Float64 `tfsdk:"assigned"`
	Pending    types.Float64 `tfsdk:"pending"`
}

func NewNetworksDataSource() datasource.DataSource {
	return &networksDataSource{}
}

func (d *networksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all network pools in clusterbook",
		Attributes: map[string]schema.Attribute{
			"networks": schema.ListNestedAttribute{
				Description: "All network pools",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network_key": schema.StringAttribute{Computed: true},
						"total":       schema.Float64Attribute{Computed: true},
						"available":   schema.Float64Attribute{Computed: true},
						"assigned":    schema.Float64Attribute{Computed: true},
						"pending":     schema.Float64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *networksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	pools, err := d.client.GetNetworks()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list networks", err.Error())
		return
	}

	var state networksDataSourceModel
	state.Networks = make([]networkInfoModel, len(pools))
	for i, pool := range pools {
		state.Networks[i] = networkInfoModel{
			NetworkKey: types.StringValue(pool.NetworkKey),
			Total:      types.Float64Value(pool.Total),
			Available:  types.Float64Value(pool.Available),
			Assigned:   types.Float64Value(pool.Assigned),
			Pending:    types.Float64Value(pool.Pending),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
