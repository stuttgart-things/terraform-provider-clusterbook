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
	_ datasource.DataSource              = &clusterDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterDataSource{}
)

type clusterDataSource struct {
	client *client.Client
}

type clusterDataSourceModel struct {
	Name types.String    `tfsdk:"name"`
	IPs  []clusterIPModel `tfsdk:"ips"`
}

type clusterIPModel struct {
	Network types.String `tfsdk:"network"`
	IP      types.String `tfsdk:"ip"`
	Status  types.String `tfsdk:"status"`
}

func NewClusterDataSource() datasource.DataSource {
	return &clusterDataSource{}
}

func (d *clusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get IP assignment info for a cluster",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Cluster name",
				Required:    true,
			},
			"ips": schema.ListNestedAttribute{
				Description: "IP assignments for this cluster",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"network": schema.StringAttribute{Computed: true},
						"ip":      schema.StringAttribute{Computed: true},
						"status":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *clusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clusterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := d.client.GetClusterInfo(config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cluster info", err.Error())
		return
	}

	config.IPs = make([]clusterIPModel, len(info.IPs))
	for i, ip := range info.IPs {
		config.IPs[i] = clusterIPModel{
			Network: types.StringValue(ip.Network),
			IP:      types.StringValue(ip.IP),
			Status:  types.StringValue(ip.Status),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
