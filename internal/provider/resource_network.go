package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stuttgart-things/terraform-provider-clusterbook/internal/client"
)

var (
	_ resource.Resource              = &networkResource{}
	_ resource.ResourceWithConfigure = &networkResource{}
)

type networkResource struct {
	client *client.Client
}

type networkModel struct {
	ID         types.String `tfsdk:"id"`
	NetworkKey types.String `tfsdk:"network_key"`
	IPFrom     types.Int64  `tfsdk:"ip_from"`
	IPTo       types.Int64  `tfsdk:"ip_to"`
	Total      types.Int64  `tfsdk:"total"`
	Available  types.Int64  `tfsdk:"available"`
	Assigned   types.Int64  `tfsdk:"assigned"`
	Pending    types.Int64  `tfsdk:"pending"`
}

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a network IP pool in clusterbook",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (same as network_key)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_key": schema.StringAttribute{
				Description: "Network subnet prefix (e.g. 10.31.106)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_from": schema.Int64Attribute{
				Description: "Start of IP range (last octet)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ip_to": schema.Int64Attribute{
				Description: "End of IP range (last octet)",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"total": schema.Int64Attribute{
				Description: "Total number of IPs in the pool",
				Computed:    true,
			},
			"available": schema.Int64Attribute{
				Description: "Number of available IPs",
				Computed:    true,
			},
			"assigned": schema.Int64Attribute{
				Description: "Number of assigned IPs",
				Computed:    true,
			},
			"pending": schema.Int64Attribute{
				Description: "Number of pending IPs",
				Computed:    true,
			},
		},
	}
}

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateNetwork(
		plan.NetworkKey.ValueString(),
		int(plan.IPFrom.ValueInt64()),
		int(plan.IPTo.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create network", err.Error())
		return
	}

	plan.ID = plan.NetworkKey

	// Read back pool stats
	r.refreshPoolStats(&plan, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pools, err := r.client.GetNetworks()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read networks", err.Error())
		return
	}

	found := false
	for _, pool := range pools {
		if pool.NetworkKey == state.NetworkKey.ValueString() {
			found = true
			state.Total = types.Int64Value(int64(pool.Total))
			state.Available = types.Int64Value(int64(pool.Available))
			state.Assigned = types.Int64Value(int64(pool.Assigned))
			state.Pending = types.Int64Value(int64(pool.Pending))
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Network pools are immutable; changes require replacement")
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetwork(state.NetworkKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete network", err.Error())
		return
	}
}

func (r *networkResource) refreshPoolStats(model *networkModel, resp *resource.CreateResponse) {
	pools, err := r.client.GetNetworks()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read networks after create", err.Error())
		return
	}

	for _, pool := range pools {
		if pool.NetworkKey == model.NetworkKey.ValueString() {
			model.Total = types.Int64Value(int64(pool.Total))
			model.Available = types.Int64Value(int64(pool.Available))
			model.Assigned = types.Int64Value(int64(pool.Assigned))
			model.Pending = types.Int64Value(int64(pool.Pending))
			return
		}
	}

	resp.Diagnostics.AddError("Network not found after create", "The network was created but could not be read back")
}
