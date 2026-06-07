package provider

import (
	"context"
	"fmt"

	"github.com/awanio/terraform-provider-cockpit/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure SwitchResource implements resource.Resource.
var _ resource.Resource = &SwitchResource{}

type SwitchResource struct {
	client *client.Client
}

type SwitchResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	HostID     types.String `tfsdk:"host_id"`
	Mode       types.String `tfsdk:"mode"`
	Bridge     types.String `tfsdk:"bridge"`
	IPAddress  types.String `tfsdk:"ip_address"`
	Netmask    types.String `tfsdk:"netmask"`
	DHCPStart  types.String `tfsdk:"dhcp_start"`
	DHCPEnd    types.String `tfsdk:"dhcp_end"`
	Autostart  types.Bool   `tfsdk:"autostart"`
	Domain     types.String `tfsdk:"domain"`
	SwitchUUID types.String `tfsdk:"switch_uuid"`
	Driver     types.String `tfsdk:"driver"`
	State      types.String `tfsdk:"state"`
}

func NewSwitchResource() resource.Resource {
	return &SwitchResource{}
}

func (r *SwitchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch"
}

func (r *SwitchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Switch (virtual network) on a Vapor Host connected to Cockpit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the network switch.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target Vapor Host ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Virtual network mode (`nat`, `bridge`, `route`, `isolated`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bridge": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Bridge device name (e.g. virbr0).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "IP address of the bridge interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"netmask": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Netmask of the bridge interface network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dhcp_start": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "DHCP start IP address pool.",
			},
			"dhcp_end": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "DHCP end IP address pool.",
			},
			"autostart": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to start the switch automatically on host boot.",
			},
			"domain": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "DNS Domain name config for DHCP clients.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"switch_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Underlying network UUID on the host hypervisor.",
			},
			"driver": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Switch network driver type (e.g. NAT, Bridged).",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current run status (active, inactive).",
			},
		},
	}
}

func (r *SwitchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

func (r *SwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SwitchResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.CreateSwitchRequest{
		HostID: data.HostID.ValueString(),
		Network: client.NetworkCreateRequest{
			Name:      data.Name.ValueString(),
			Mode:      data.Mode.ValueString(),
			Bridge:    data.Bridge.ValueString(),
			IPAddress: data.IPAddress.ValueString(),
			Netmask:   data.Netmask.ValueString(),
			DHCPStart: data.DHCPStart.ValueString(),
			DHCPEnd:   data.DHCPEnd.ValueString(),
			Autostart: data.Autostart.ValueBool(),
			Domain:    data.Domain.ValueString(),
		},
	}

	apiResp, err := r.client.CreateSwitch(&apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Switch",
			"Could not create switch, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(apiResp.ID)
	data.SwitchUUID = types.StringValue(apiResp.SwitchUUID)
	data.Driver = types.StringValue(apiResp.Driver)
	data.State = types.StringValue(apiResp.State)
	data.Bridge = types.StringValue(apiResp.Bridge)
	data.Mode = types.StringValue(apiResp.Mode)
	data.IPAddress = types.StringValue(apiResp.IPAddress)
	data.Netmask = types.StringValue(apiResp.Netmask)
	data.DHCPStart = types.StringValue(apiResp.DHCPStart)
	data.DHCPEnd = types.StringValue(apiResp.DHCPEnd)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SwitchResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.GetSwitch(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Reading Switch",
			"Could not read switch, assuming it was deleted: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	data.SwitchUUID = types.StringValue(apiResp.SwitchUUID)
	data.Driver = types.StringValue(apiResp.Driver)
	data.State = types.StringValue(apiResp.State)
	data.Bridge = types.StringValue(apiResp.Bridge)
	data.Mode = types.StringValue(apiResp.Mode)
	data.IPAddress = types.StringValue(apiResp.IPAddress)
	data.Netmask = types.StringValue(apiResp.Netmask)
	data.DHCPStart = types.StringValue(apiResp.DHCPStart)
	data.DHCPEnd = types.StringValue(apiResp.DHCPEnd)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SwitchResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dynamic update of autostart or DHCP ranges
	if !plan.Autostart.Equal(state.Autostart) || !plan.DHCPStart.Equal(state.DHCPStart) || !plan.DHCPEnd.Equal(state.DHCPEnd) {
		autostartVal := plan.Autostart.ValueBool()
		updateReq := client.NetworkUpdateRequest{
			Autostart: &autostartVal,
			DHCPStart: plan.DHCPStart.ValueString(),
			DHCPEnd:   plan.DHCPEnd.ValueString(),
		}

		apiResp, err := r.client.UpdateSwitch(state.ID.ValueString(), &updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Switch",
				"Could not update switch: "+err.Error(),
			)
			return
		}

		state.Autostart = types.BoolValue(apiResp.Autostart)
		state.DHCPStart = types.StringValue(apiResp.DHCPStart)
		state.DHCPEnd = types.StringValue(apiResp.DHCPEnd)
		state.State = types.StringValue(apiResp.State)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SwitchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SwitchResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSwitch(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Switch",
			"Could not delete switch, unexpected error: "+err.Error(),
		)
		return
	}
}
