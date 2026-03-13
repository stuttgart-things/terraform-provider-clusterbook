package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stuttgart-things/terraform-provider-clusterbook/internal/client"
)

var _ provider.Provider = &clusterbookProvider{}

type clusterbookProvider struct{}

type clusterbookProviderModel struct {
	URL types.String `tfsdk:"url"`
}

func New() provider.Provider {
	return &clusterbookProvider{}
}

func (p *clusterbookProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clusterbook"
}

func (p *clusterbookProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for clusterbook IPAM",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The URL of the clusterbook API (e.g. http://localhost:8080)",
				Required:    true,
			},
		},
	}
}

func (p *clusterbookProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config clusterbookProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddError("Unknown URL", "The clusterbook URL must be known at configure time")
		return
	}

	c := client.New(config.URL.ValueString())
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *clusterbookProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIPAssignmentResource,
	}
}

func (p *clusterbookProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
