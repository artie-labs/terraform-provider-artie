package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &deploymentsDataSource{}
	_ datasource.DataSourceWithConfigure = &deploymentsDataSource{}
)

// NewDeploymentsDataSource is a helper function to simplify the provider implementation.
func NewDeploymentsDataSource() datasource.DataSource {
	return &deploymentsDataSource{}
}

type deploymentsModel struct {
	UUID string `tfsdk:"uuid"`
	Name string `tfsdk:"name"`
}

type deploymentsDataSourceModel struct {
	Deployments []deploymentsModel `tfsdk:"deployments"`
}

type deploymentsResponse struct {
	Deployments []deploymentsModel `json:"items"`
}

// deploymentsDataSource is the data source implementation.
type deploymentsDataSource struct {
	endpoint basetypes.StringValue
	apiKey   basetypes.StringValue
}

// Metadata returns the data source type name.
func (d *deploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

// Configure adds the provider configured client to the data source.
func (d *deploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderModel)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.endpoint = providerData.Endpoint
	d.apiKey = providerData.APIKey
}

// Schema defines the schema for the data source.
func (d *deploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"deployments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *deploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, fmt.Sprintf("Endpoint: %s", d.endpoint.ValueString()))
	var state deploymentsDataSourceModel

	apiReq, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments", d.endpoint.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployments", err.Error())
		return
	}

	apiReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey.ValueString()))
	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployments", err.Error())
		return
	}

	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unable to Read Deployments", fmt.Sprintf("Received status code %d", apiResp.StatusCode))
		return
	}

	bodyBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployments", err.Error())
		return
	}

	var deploymentsResp deploymentsResponse
	err = json.Unmarshal(bodyBytes, &deploymentsResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployments", err.Error())
		return
	}

	for _, deployment := range deploymentsResp.Deployments {
		deploymentState := deploymentsModel{
			UUID: deployment.UUID,
			Name: deployment.Name,
		}
		state.Deployments = append(state.Deployments, deploymentState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
