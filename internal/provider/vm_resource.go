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

// Ensure VirtualMachineResource implements resource.Resource.
var _ resource.Resource = &VirtualMachineResource{}

type VirtualMachineResource struct {
	client *client.Client
}

type VMOSModel struct {
	Type    types.String `tfsdk:"type"`
	Variant types.String `tfsdk:"variant"`
	Arch    types.String `tfsdk:"arch"`
}

type VMDiskModel struct {
	Size   types.Int64  `tfsdk:"size"`
	Pool   types.String `tfsdk:"pool"`
	Format types.String `tfsdk:"format"`
	Bus    types.String `tfsdk:"bus"`
}

type VMStorageModel struct {
	Disks []VMDiskModel `tfsdk:"disks"`
	ISO   types.String  `tfsdk:"iso"`
}

type VMNetworkModel struct {
	Type   types.String `tfsdk:"type"`
	Source types.String `tfsdk:"source"`
	Model  types.String `tfsdk:"model"`
}

type VMResourceModel struct {
	ID           types.String     `tfsdk:"id"`
	HostID       types.String     `tfsdk:"host_id"`
	ClusterID    types.String     `tfsdk:"cluster_id"`
	DatacenterID types.String     `tfsdk:"datacenter_id"`
	Name         types.String     `tfsdk:"name"`
	Memory       types.Int64      `tfsdk:"memory"`
	VCPUs        types.Int64      `tfsdk:"vcpus"`
	OS           *VMOSModel       `tfsdk:"os"`
	Storage      *VMStorageModel  `tfsdk:"storage"`
	Networks     []VMNetworkModel `tfsdk:"networks"`
	Autostart    types.Bool       `tfsdk:"autostart"`
	Status       types.String     `tfsdk:"status"`
}

func NewVirtualMachineResource() resource.Resource {
	return &VirtualMachineResource{}
}

func (r *VirtualMachineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *VirtualMachineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Virtual Machine in Cockpit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"host_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target Vapor Host ID to run the VM on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target Cluster ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Target Datacenter ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Virtual Machine.",
			},
			"memory": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "RAM Memory size in MB.",
			},
			"vcpus": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of Virtual CPUs.",
			},
			"os": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional: true,
					},
					"variant": schema.StringAttribute{
						Optional: true,
					},
					"arch": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"storage": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"disks": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"size": schema.Int64Attribute{
									Required:            true,
									MarkdownDescription: "Disk size in GB.",
								},
								"pool": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Storage Pool name/ID.",
								},
								"format": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: "Disk format (e.g. qcow2, raw).",
								},
								"bus": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									MarkdownDescription: "Disk bus (e.g. virtio, ide).",
								},
							},
						},
					},
					"iso": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"networks": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"source": schema.StringAttribute{
							Required: true,
						},
						"model": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"autostart": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *VirtualMachineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VirtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.CreateVMRequest{
		HostID:       data.HostID.ValueString(),
		ClusterID:    data.ClusterID.ValueString(),
		DatacenterID: data.DatacenterID.ValueString(),
		Name:         data.Name.ValueString(),
		Memory:       int(data.Memory.ValueInt64()),
		VCPUs:        int(data.VCPUs.ValueInt64()),
		Autostart:    data.Autostart.ValueBool(),
	}

	if data.OS != nil {
		apiReq.OS = &client.VMOS{
			Type:    data.OS.Type.ValueString(),
			Variant: data.OS.Variant.ValueString(),
			Arch:    data.OS.Arch.ValueString(),
		}
	}

	if data.Storage != nil {
		apiReq.Storage = &client.VMStorage{
			ISO: data.Storage.ISO.ValueString(),
		}
		if len(data.Storage.Disks) > 0 {
			apiReq.Storage.Disks = make([]client.VMDisk, len(data.Storage.Disks))
			for i, d := range data.Storage.Disks {
				apiReq.Storage.Disks[i] = client.VMDisk{
					Size:   d.Size.ValueInt64(),
					Pool:   d.Pool.ValueString(),
					Format: d.Format.ValueString(),
					Bus:    d.Bus.ValueString(),
				}
			}
		}
	}

	if len(data.Networks) > 0 {
		apiReq.Networks = make([]client.VMNetwork, len(data.Networks))
		for i, n := range data.Networks {
			apiReq.Networks[i] = client.VMNetwork{
				Type:   n.Type.ValueString(),
				Source: n.Source.ValueString(),
				Model:  n.Model.ValueString(),
			}
		}
	}

	apiResp, err := r.client.CreateVM(&apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating VM",
			"Could not create VM, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(apiResp.ID)
	data.Status = types.StringValue(apiResp.Status)
	data.Memory = types.Int64Value(int64(apiResp.Memory))
	data.VCPUs = types.Int64Value(int64(apiResp.VCPUs))
	data.HostID = types.StringValue(apiResp.HostID)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.GetVM(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Reading VM",
			"Could not read VM, assuming it was deleted: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Status = types.StringValue(apiResp.Status)
	data.Memory = types.Int64Value(int64(apiResp.Memory))
	data.VCPUs = types.Int64Value(int64(apiResp.VCPUs))
	data.HostID = types.StringValue(apiResp.HostID)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VirtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// For simplicity, updates trigger a recreate, or we can just require replacement of non-updatable fields.
	// Since we defined stringplanmodifier.RequiresReplace() on host_id and target IDs,
	// let's just raise an update error for other fields to enforce clean TF workflow.
	resp.Diagnostics.AddError(
		"Virtual Machine Update Not Supported",
		"In-place updates for virtual machines are not supported yet by the provider. Please recreate the resource.",
	)
}

func (r *VirtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVM(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VM",
			"Could not delete VM, unexpected error: "+err.Error(),
		)
		return
	}
}
