package provider

import (
	"context"
	"os"

	"github.com/awanio/terraform-provider-cockpit/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure CockpitProvider implements provider.Provider.
var _ provider.Provider = &CockpitProvider{}

// CockpitProvider defines the provider implementation.
type CockpitProvider struct {
	// version is set to the provider version on release.
	version string
}

// CockpitProviderModel describes the provider data model.
type CockpitProviderModel struct {
	Host         types.String `tfsdk:"host"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *CockpitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cockpit"
	resp.Version = p.version
}

func (p *CockpitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider to interact with Cockpit Virtualization Management Platform.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Cockpit API endpoint (e.g. `http://localhost:7771/api/v1`). May also be set via the `COCKPIT_HOST` environment variable.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Service account client ID. May also be set via the `COCKPIT_CLIENT_ID` environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Service account client secret. May also be set via the `COCKPIT_CLIENT_SECRET` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *CockpitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CockpitProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read environment variables or fallback values
	host := os.Getenv("COCKPIT_HOST")
	clientID := os.Getenv("COCKPIT_CLIENT_ID")
	clientSecret := os.Getenv("COCKPIT_CLIENT_SECRET")

	if !data.Host.IsUnknown() && !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	if !data.ClientID.IsUnknown() && !data.ClientID.IsNull() {
		clientID = data.ClientID.ValueString()
	}

	if !data.ClientSecret.IsUnknown() && !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	if host == "" {
		host = "http://localhost:7771/api/v1"
	}

	apiClient := client.NewClient(host, clientID, clientSecret)

	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *CockpitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewVirtualMachineResource,
		NewKubernetesClusterResource,
		NewDatastoreResource,
		NewSwitchResource,
	}
}

func (p *CockpitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CockpitProvider{
			version: version,
		}
	}
}
