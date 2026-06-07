package provider

import (
	"context"
	"fmt"

	"github.com/awanio/terraform-provider-cockpit/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure KubernetesClusterResource implements resource.Resource.
var _ resource.Resource = &KubernetesClusterResource{}

type KubernetesClusterResource struct {
	client *client.Client
}

type K8sControlPlaneModel struct {
	CPU           types.Int64    `tfsdk:"cpu"`
	MemoryMB      types.Int64    `tfsdk:"memory_mb"`
	StoragePoolID types.String   `tfsdk:"storage_pool_id"`
	StorageGB     types.Int64    `tfsdk:"storage_gb"`
	DesiredCount  types.Int64    `tfsdk:"desired_count"`
	HostIDs       []types.String `tfsdk:"host_ids"`
}

type K8sWorkersModel struct {
	CPU           types.Int64    `tfsdk:"cpu"`
	MemoryMB      types.Int64    `tfsdk:"memory_mb"`
	StoragePoolID types.String   `tfsdk:"storage_pool_id"`
	StorageGB     types.Int64    `tfsdk:"storage_gb"`
	DesiredCount  types.Int64    `tfsdk:"desired_count"`
	HostIDs       []types.String `tfsdk:"host_ids"`
}

type K8sCSITiersModel struct {
	LocalPathEnabled types.Bool   `tfsdk:"local_path_enabled"`
	NFSEnabled       types.Bool   `tfsdk:"nfs_enabled"`
	NFSServer        types.String `tfsdk:"nfs_server"`
	NFSPath          types.String `tfsdk:"nfs_path"`
	LonghornEnabled  types.Bool   `tfsdk:"longhorn_enabled"`
	LonghornReplicas types.Int64  `tfsdk:"longhorn_replicas"`
}

type KubernetesClusterResourceModel struct {
	ID           types.String          `tfsdk:"id"`
	Name         types.String          `tfsdk:"name"`
	Version      types.String          `tfsdk:"version"`
	Distribution types.String          `tfsdk:"distribution"`
	NetworkID    types.String          `tfsdk:"network_id"`
	CNI          types.String          `tfsdk:"cni"`
	SSHPublicKey types.String          `tfsdk:"ssh_public_key"`
	ImagePath    types.String          `tfsdk:"image_path"`
	Status       types.String          `tfsdk:"status"`
	KubeConfig   types.String          `tfsdk:"kubeconfig"`
	ControlPlane *K8sControlPlaneModel `tfsdk:"control_plane"`
	Workers      *K8sWorkersModel      `tfsdk:"workers"`
	CSITiers     *K8sCSITiersModel     `tfsdk:"csi_tiers"`
}

func NewKubernetesClusterResource() resource.Resource {
	return &KubernetesClusterResource{}
}

func (r *KubernetesClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (r *KubernetesClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kubernetes Cluster provisioned via Cockpit Auto-Provisioner.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Kubernetes cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "K3s version tag (e.g. `v1.36.1+k3s1`). Check supported versions first.",
			},
			"distribution": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("k3s"),
				MarkdownDescription: "Kubernetes distribution (only `k3s` supported currently).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network switch ID (UUID) for VM interface links.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cni": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("flannel"),
				MarkdownDescription: "CNI plugin (`flannel`, `calico`, or `cilium`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "SSH Public key configured on cluster nodes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional template/golden image file path.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the cluster.",
			},
			"kubeconfig": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Decrypted Kubeconfig YAML to access the cluster.",
			},
			"control_plane": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"cpu": schema.Int64Attribute{
						Required: true,
					},
					"memory_mb": schema.Int64Attribute{
						Required: true,
					},
					"storage_pool_id": schema.StringAttribute{
						Required: true,
					},
					"storage_gb": schema.Int64Attribute{
						Required: true,
					},
					"desired_count": schema.Int64Attribute{
						Required: true,
					},
					"host_ids": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"workers": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"cpu": schema.Int64Attribute{
						Required: true,
					},
					"memory_mb": schema.Int64Attribute{
						Required: true,
					},
					"storage_pool_id": schema.StringAttribute{
						Required: true,
					},
					"storage_gb": schema.Int64Attribute{
						Required: true,
					},
					"desired_count": schema.Int64Attribute{
						Required: true,
					},
					"host_ids": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"csi_tiers": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"local_path_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"nfs_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"nfs_server": schema.StringAttribute{
						Optional: true,
					},
					"nfs_path": schema.StringAttribute{
						Optional: true,
					},
					"longhorn_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"longhorn_replicas": schema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (r *KubernetesClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KubernetesClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.K8sClusterCreateRequest{
		Name:         data.Name.ValueString(),
		Version:      data.Version.ValueString(),
		Distribution: data.Distribution.ValueString(),
		NetworkID:    data.NetworkID.ValueString(),
		CNI:          data.CNI.ValueString(),
		SSHPublicKey: data.SSHPublicKey.ValueString(),
		ImagePath:    data.ImagePath.ValueString(),
	}

	if data.ControlPlane != nil {
		apiReq.ControlPlane = &client.K8sControlPlaneRequest{
			CPU:           int(data.ControlPlane.CPU.ValueInt64()),
			MemoryMB:      int(data.ControlPlane.MemoryMB.ValueInt64()),
			StoragePoolID: data.ControlPlane.StoragePoolID.ValueString(),
			StorageGB:     int(data.ControlPlane.StorageGB.ValueInt64()),
			DesiredCount:  int(data.ControlPlane.DesiredCount.ValueInt64()),
		}
		for _, hid := range data.ControlPlane.HostIDs {
			apiReq.ControlPlane.HostIDs = append(apiReq.ControlPlane.HostIDs, hid.ValueString())
		}
	}

	if data.Workers != nil {
		apiReq.Workers = &client.K8sWorkersRequest{
			CPU:           int(data.Workers.CPU.ValueInt64()),
			MemoryMB:      int(data.Workers.MemoryMB.ValueInt64()),
			StoragePoolID: data.Workers.StoragePoolID.ValueString(),
			StorageGB:     int(data.Workers.StorageGB.ValueInt64()),
			DesiredCount:  int(data.Workers.DesiredCount.ValueInt64()),
		}
		for _, hid := range data.Workers.HostIDs {
			apiReq.Workers.HostIDs = append(apiReq.Workers.HostIDs, hid.ValueString())
		}
	}

	if data.CSITiers != nil {
		apiReq.CSITiers = &client.K8sCSITiersRequest{
			LocalPathEnabled: data.CSITiers.LocalPathEnabled.ValueBool(),
			NFSEnabled:       data.CSITiers.NFSEnabled.ValueBool(),
			NFSServer:        data.CSITiers.NFSServer.ValueString(),
			NFSPath:          data.CSITiers.NFSPath.ValueString(),
			LonghornEnabled:  data.CSITiers.LonghornEnabled.ValueBool(),
			LonghornReplicas: int(data.CSITiers.LonghornReplicas.ValueInt64()),
		}
	}

	apiResp, err := r.client.CreateK8sCluster(&apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Kubernetes Cluster",
			"Could not create cluster, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(apiResp.ID)
	data.Status = types.StringValue(apiResp.Status)

	// Attempt to retrieve kubeconfig (might be empty/not ready immediately, but we try)
	kubeconfig, err := r.client.GetK8sClusterKubeconfig(apiResp.ID)
	if err == nil {
		data.KubeConfig = types.StringValue(kubeconfig)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KubernetesClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.GetK8sCluster(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Reading Kubernetes Cluster",
			"Could not read Kubernetes cluster, assuming it was deleted: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Status = types.StringValue(apiResp.Status)
	data.Version = types.StringValue(apiResp.Version)

	// Sync node pool counts
	for _, p := range apiResp.NodePools {
		if p.Role == "control-plane" && data.ControlPlane != nil {
			data.ControlPlane.DesiredCount = types.Int64Value(int64(p.DesiredCount))
		} else if p.Role == "worker" && data.Workers != nil {
			data.Workers.DesiredCount = types.Int64Value(int64(p.DesiredCount))
		}
	}

	// Retrieve latest kubeconfig
	kubeconfig, err := r.client.GetK8sClusterKubeconfig(data.ID.ValueString())
	if err == nil {
		data.KubeConfig = types.StringValue(kubeconfig)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KubernetesClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ID.ValueString()

	// 1. Check for Upgrade (version change)
	if !plan.Version.Equal(state.Version) {
		err := r.client.UpgradeK8sCluster(clusterID, plan.Version.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Upgrading Kubernetes Cluster",
				fmt.Sprintf("Could not upgrade cluster to version %s: %s", plan.Version.ValueString(), err.Error()),
			)
			return
		}
		state.Version = plan.Version
	}

	// 2. Fetch latest state from API to get NodePool IDs for scaling
	apiResp, err := r.client.GetK8sCluster(clusterID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Fetching Cluster Pools",
			"Could not retrieve cluster pools metadata: "+err.Error(),
		)
		return
	}

	var cpPoolID, workerPoolID string
	for _, p := range apiResp.NodePools {
		if p.Role == "control-plane" {
			cpPoolID = p.ID
		} else if p.Role == "worker" {
			workerPoolID = p.ID
		}
	}

	// 3. Check for Control Plane scaling
	if plan.ControlPlane != nil && state.ControlPlane != nil && cpPoolID != "" {
		if plan.ControlPlane.DesiredCount.ValueInt64() != state.ControlPlane.DesiredCount.ValueInt64() {
			err := r.client.ScaleNodePool(clusterID, cpPoolID, int(plan.ControlPlane.DesiredCount.ValueInt64()))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Scaling Control Plane",
					"Could not scale control plane pool: "+err.Error(),
				)
				return
			}
			state.ControlPlane.DesiredCount = plan.ControlPlane.DesiredCount
		}
	}

	// 4. Check for Worker Pool scaling
	if plan.Workers != nil && state.Workers != nil && workerPoolID != "" {
		if plan.Workers.DesiredCount.ValueInt64() != state.Workers.DesiredCount.ValueInt64() {
			err := r.client.ScaleNodePool(clusterID, workerPoolID, int(plan.Workers.DesiredCount.ValueInt64()))
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Scaling Workers",
					"Could not scale worker pool: "+err.Error(),
				)
				return
			}
			state.Workers.DesiredCount = plan.Workers.DesiredCount
		}
	}

	// Sync status
	state.Status = types.StringValue("scaling") // Or fetch latest status, but Scaling starts asynchronously

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KubernetesClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteK8sCluster(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Kubernetes Cluster",
			"Could not delete cluster, unexpected error: "+err.Error(),
		)
		return
	}
}
