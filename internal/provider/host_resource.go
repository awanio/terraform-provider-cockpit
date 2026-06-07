package provider

import (
	"context"
	"fmt"

	"github.com/awanio/terraform-provider-cockpit/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure HostResource implements resource.Resource.
var _ resource.Resource = &HostResource{}

type HostResource struct {
	client *client.Client
}

type HostResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Address  types.String `tfsdk:"address"`
	Port     types.Int64  `tfsdk:"port"`
	APIToken types.String `tfsdk:"api_token"`
	ParentID types.String `tfsdk:"parent_id"`
	Hostname types.String `tfsdk:"hostname"`
	IP       types.String `tfsdk:"ip"`
	Status   types.String `tfsdk:"status"`
}

func NewHostResource() resource.Resource {
	return &HostResource{}
}

func (r *HostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *HostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Vapor host connected to Cockpit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Vapor host IP address or hostname.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(7770),
				MarkdownDescription: "Vapor API port (defaults to 7770).",
			},
			"api_token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Vapor API token.",
			},
			"parent_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Inventory entity ID (Datacenter or Cluster UUID) to place the host under.",
			},
			"hostname": schema.StringAttribute{
				Computed: true,
			},
			"ip": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *HostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.AddHostRequest{
		Address:  data.Address.ValueString(),
		Port:     int(data.Port.ValueInt64()),
		APIToken: data.APIToken.ValueString(),
		ParentID: data.ParentID.ValueString(),
	}

	apiResp, err := r.client.AddHost(&apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Host",
			"Could not create host, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(apiResp.ID)
	data.Hostname = types.StringValue(apiResp.Hostname)
	data.IP = types.StringValue(apiResp.IP)
	data.Port = types.Int64Value(int64(apiResp.Port))
	data.Status = types.StringValue(apiResp.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.GetHost(data.ID.ValueString())
	if err != nil {
		// If resource is not found (assuming API error message mentions "not found" or similar), we should remove from state
		resp.Diagnostics.AddWarning(
			"Error Reading Host",
			"Could not read host, assuming it was deleted: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Hostname = types.StringValue(apiResp.Hostname)
	data.IP = types.StringValue(apiResp.IP)
	data.Port = types.Int64Value(int64(apiResp.Port))
	data.Status = types.StringValue(apiResp.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Cockpit host updates are usually not fully editable or require re-adding.
	// For simplicity, we just mark fields that cannot change as RequiresReplace, or force recreate.
	// Let's implement basic in-place update by reading plan and setting it if needed, or if API doesn't support PATCH host,
	// we just warn that update is not supported and recreate is required.
	resp.Diagnostics.AddError(
		"Host Update Not Supported",
		"In-place updates for hosts are not supported by the Cockpit API. Please recreate the host resource.",
	)
}

func (r *HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteHost(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Host",
			"Could not delete host, unexpected error: "+err.Error(),
		)
		return
	}
}
