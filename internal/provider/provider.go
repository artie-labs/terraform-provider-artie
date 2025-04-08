package provider

import (
	"context"
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const DEFAULT_API_ENDPOINT = "https://api.artie.com"

// Ensure ArtieProvider satisfies various provider interfaces.
var _ provider.Provider = &ArtieProvider{}
var _ provider.ProviderWithFunctions = &ArtieProvider{}

type ArtieProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ArtieProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

type ArtieProviderData struct {
	Endpoint string
	APIKey   string
	version  string
}

func (a ArtieProviderData) NewClient() (artieclient.Client, error) {
	return artieclient.New(a.Endpoint, a.APIKey, a.version)
}

func (p *ArtieProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "artie"
	resp.Version = p.version
}

func (p *ArtieProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The Artie Terraform provider can be used to manage your data pipelines in Artie. You must already have an Artie account to use this provider. If you don't have one, you can request access at https://www.artie.com/.

To set up a new data pipeline, you'll need to create a few resources:
- Source Connector: this holds connection information for the source database.
- Destination Connector: this holds connection information for the destination database or data warehouse.
- Source Reader: this represents a process that reads data from a source connector and inserts it info Kafka. A Source Reader can be used by multiple Pipelines, e.g. to read from a single PostgreSQL replication slot and copy the data to multiple destinations.
- Pipeline: this represents a data pipeline that syncs data from a single source (e.g., PostgreSQL) to a single destination (e.g., Snowflake).
`,
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Artie API endpoint. This defaults to https://api.artie.com and should not need to be changed except when developing the provider.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Artie API key to authenticate requests to the Artie API. Generate an API key in the Artie web app at https://app.artie.com/settings. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ArtieProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var configData ArtieProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := DEFAULT_API_ENDPOINT
	if !configData.Endpoint.IsNull() {
		endpoint = configData.Endpoint.ValueString()
	}

	providerData := ArtieProviderData{
		Endpoint: endpoint,
		APIKey:   configData.APIKey.ValueString(),
		version:  p.version,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *ArtieProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeploymentResource,
		NewDestinationResource,
		NewSSHTunnelResource,
		NewConnectorResource,
		NewSourceReaderResource,
		NewPipelineResource,
	}
}

func (p *ArtieProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ArtieProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArtieProvider{
			version: version,
		}
	}
}
