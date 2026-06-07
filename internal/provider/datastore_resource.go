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

// Ensure DatastoreResource implements resource.Resource.
var _ resource.Resource = &DatastoreResource{}

type DatastoreResource struct {
	client *client.Client
}

type DatastoreResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	HostID    types.String `tfsdk:"host_id"`
	Type      types.String `tfsdk:"type"`
	Path      types.String `tfsdk:"path"`
	Source    types.String `tfsdk:"source"`
	Target    types.String `tfsdk:"target"`
	Autostart types.Bool   `tfsdk:"autostart"`
	Scope     types.String `tfsdk:"scope"`
	Driver    types.String `tfsdk:"driver"`
	MountPath types.String `tfsdk:"mount_path"`
	State     types.String `tfsdk:"state"`
}

func NewDatastoreResource() resource.Resource {
	return &DatastoreResource{}
}

func (r *DatastoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastore"
}

func (r *DatastoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datastore (storage pool) on a Vapor Host connected to Cockpit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the storage pool / datastore.",
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
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of storage pool (e.g. `dir`, `netfs`, `logical`, `fs`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Target mount path for directory/netfs pool.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Source path or host directory (e.g., NFS server export source).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Target storage path.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"autostart": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to start the storage pool automatically on host boot.",
			},
			"scope": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("local"),
				MarkdownDescription: "Scope of the datastore (`local` or `shared`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Storage pool driver type.",
			},
			"mount_path": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resulting path where the pool is mounted.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status state (running, stopped).",
			},
		},
	}
}

func (r *DatastoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatastoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatastoreResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := client.CreateDatastoreRequest{
		HostID: data.HostID.ValueString(),
		Scope:  data.Scope.ValueString(),
		Pool: client.StoragePoolCreateRequest{
			Name:      data.Name.ValueString(),
			Type:      data.Type.ValueString(),
			Path:      data.Path.ValueString(),
			Source:    data.Source.ValueString(),
			Target:    data.Target.ValueString(),
			Autostart: data.Autostart.ValueBool(),
		},
	}

	apiResp, err := r.client.CreateDatastore(&apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Datastore",
			"Could not create datastore, unexpected error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(apiResp.ID)
	data.Driver = types.StringValue(apiResp.Driver)
	data.Scope = types.StringValue(apiResp.Scope)
	data.MountPath = types.StringValue(apiResp.MountPath)
	data.State = types.StringValue(apiResp.State)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatastoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatastoreResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.GetDatastore(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error Reading Datastore",
			"Could not read datastore, assuming it was deleted: "+err.Error(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	data.Driver = types.StringValue(apiResp.Driver)
	data.Scope = types.StringValue(apiResp.Scope)
	data.MountPath = types.StringValue(apiResp.MountPath)
	data.State = types.StringValue(apiResp.State)
	data.Autostart = types.BoolValue(apiResp.Autostart)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatastoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DatastoreResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Autostart.Equal(state.Autostart) {
		autostartVal := plan.Autostart.ValueBool()
		updateReq := client.StoragePoolUpdateRequest{
			Autostart: &autostartVal,
		}

		apiResp, err := r.client.UpdateDatastore(state.ID.ValueString(), &updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Datastore",
				"Could not update datastore: "+err.Error(),
			)
			return
		}

		state.Autostart = types.BoolValue(apiResp.Autostart)
		state.State = types.StringValue(apiResp.State)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DatastoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatastoreResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDatastore(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datastore",
			"Could not delete datastore, unexpected error: "+err.Error(),
		)
		return
	}
}
