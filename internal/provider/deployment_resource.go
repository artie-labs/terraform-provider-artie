package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithConfigure = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	endpoint string
	apiKey   string
}

type DeploymentResourceModel struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	LastUpdatedAt        types.String `tfsdk:"last_updated_at"`
	HasUndeployedChanges types.Bool   `tfsdk:"has_undeployed_changes"`
}

type DeploymentResponse struct {
	Deployment DeploymentResourceAPIModel `json:"deploy"`
}

type DeploymentResourceAPIModel struct {
	UUID                 string `json:"uuid"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
	LastUpdatedAt        string `json:"lastUpdatedAt"`
	HasUndeployedChanges bool   `json:"hasUndeployedChanges"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Deployment resource",
		Attributes: map[string]schema.Attribute{
			"uuid":                   schema.StringAttribute{Computed: true},
			"name":                   schema.StringAttribute{Required: true},
			"status":                 schema.StringAttribute{Computed: true, Optional: true},
			"last_updated_at":        schema.StringAttribute{Computed: true},
			"has_undeployed_changes": schema.BoolAttribute{Computed: true},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.endpoint = providerData.Endpoint
	r.apiKey = providerData.APIKey
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	// data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments/%s", r.endpoint, data.UUID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	apiReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	if apiResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Unable to Read Deployment", fmt.Sprintf("Received status code %d", apiResp.StatusCode))
		return
	}

	defer apiResp.Body.Close()
	bodyBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	var deploymentResp DeploymentResponse
	err = json.Unmarshal(bodyBytes, &deploymentResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	ctx = tflog.SetField(ctx, "name", deploymentResp.Deployment.Name)
	ctx = tflog.SetField(ctx, "status", deploymentResp.Deployment.Status)
	tflog.Info(ctx, "deployment_resp")

	data.Name = types.StringValue(deploymentResp.Deployment.Name)
	data.Status = types.StringValue(deploymentResp.Deployment.Status)
	data.LastUpdatedAt = types.StringValue(deploymentResp.Deployment.LastUpdatedAt)
	data.HasUndeployedChanges = types.BoolValue(deploymentResp.Deployment.HasUndeployedChanges)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
