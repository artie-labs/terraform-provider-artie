package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	UUID                 string `tfsdk:"uuid"`
	Name                 string `tfsdk:"name"`
	Status               string `tfsdk:"status"`
	LastUpdatedAt        string `tfsdk:"last_updated_at"`
	HasUndeployedChanges bool   `tfsdk:"has_undeployed_changes"`
}

type deploymentsDataSourceModel struct {
	Deployments []deploymentsModel `tfsdk:"deployments"`
}

type deploymentsResponse struct {
	Deployments []deploymentsModel `json:"items"`
}

type deploymentsDataSource struct {
	endpoint string
	apiKey   string
}

func (d *deploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *deploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.endpoint = providerData.Endpoint
	d.apiKey = providerData.APIKey
}

func (d *deploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Deployments Data Source",
		Attributes: map[string]schema.Attribute{
			"deployments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid":                   schema.StringAttribute{Computed: true},
						"name":                   schema.StringAttribute{Computed: true},
						"status":                 schema.StringAttribute{Computed: true},
						"last_updated_at":        schema.StringAttribute{Computed: true},
						"has_undeployed_changes": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *deploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deploymentsDataSourceModel

	apiReq, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments", d.endpoint), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployments", err.Error())
		return
	}

	apiReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))
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
			UUID:                 deployment.UUID,
			Name:                 deployment.Name,
			Status:               deployment.Status,
			LastUpdatedAt:        deployment.LastUpdatedAt,
			HasUndeployedChanges: deployment.HasUndeployedChanges,
		}
		state.Deployments = append(state.Deployments, deploymentState)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
