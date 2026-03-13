package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stuttgart-things/terraform-provider-clusterbook/internal/client"
)

var (
	_ resource.Resource              = &ipAssignmentResource{}
	_ resource.ResourceWithConfigure = &ipAssignmentResource{}
)

type ipAssignmentResource struct {
	client *client.Client
}

type ipAssignmentModel struct {
	ID         types.String `tfsdk:"id"`
	NetworkKey types.String `tfsdk:"network_key"`
	Cluster    types.String `tfsdk:"cluster"`
	Status     types.String `tfsdk:"status"`
	CreateDNS  types.Bool   `tfsdk:"create_dns"`
	IPAddress  types.String `tfsdk:"ip_address"`
}

func NewIPAssignmentResource() resource.Resource {
	return &ipAssignmentResource{}
}

func (r *ipAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_assignment"
}

func (r *ipAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IP assignment in clusterbook",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (network_key/ip_digit)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_key": schema.StringAttribute{
				Description: "Network subnet prefix (e.g. 10.31.105)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster": schema.StringAttribute{
				Description: "Cluster name to assign the IP to",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Assignment status (ASSIGNED or PENDING)",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ASSIGNED"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"create_dns": schema.BoolAttribute{
				Description: "Create a PowerDNS record for this assignment",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolRequiresReplace{},
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "The assigned IP address (computed)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ipAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip, err := r.client.FindAndAssignIP(
		plan.NetworkKey.ValueString(),
		plan.Cluster.ValueString(),
		plan.Status.ValueString(),
		plan.CreateDNS.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to assign IP", err.Error())
		return
	}

	// Extract last octet for ID
	parts := strings.Split(ip, ".")
	ipDigit := parts[len(parts)-1]

	plan.IPAddress = types.StringValue(ip)
	plan.ID = types.StringValue(plan.NetworkKey.ValueString() + "/" + ipDigit)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Look up IP by querying cluster info
	info, err := r.client.GetClusterInfo(state.Cluster.ValueString())
	if err != nil {
		// Cluster not found means the assignment was deleted externally
		if strings.Contains(err.Error(), "status 404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cluster info", err.Error())
		return
	}

	// Find our IP in the cluster's assignments
	found := false
	for _, ip := range info.IPs {
		if ip.IP == state.IPAddress.ValueString() {
			found = true
			// Update status (strip :DNS suffix for the status attribute)
			status := ip.Status
			hasDNS := strings.HasSuffix(status, ":DNS")
			if hasDNS {
				status = strings.TrimSuffix(status, ":DNS")
			}
			state.Status = types.StringValue(status)
			state.CreateDNS = types.BoolValue(hasDNS)
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes require replace, so Update should never be called
	resp.Diagnostics.AddError("Update not supported", "IP assignments are immutable; changes require replacement")
}

func (r *ipAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.ReleaseIP(state.NetworkKey.ValueString(), state.IPAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to release IP", err.Error())
		return
	}
}

// boolRequiresReplace implements planmodifier.Bool to force replacement on change.
type boolRequiresReplace struct{}

func (m boolRequiresReplace) Description(_ context.Context) string {
	return "Requires replacement if changed"
}

func (m boolRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "Requires replacement if changed"
}

func (m boolRequiresReplace) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}
