package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-artie/internal/provider/models"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Deployment resource",
		Attributes: map[string]schema.Attribute{
			"uuid":             schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":             schema.StringAttribute{Required: true},
			"status":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"destination_uuid": schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"source": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{Required: true},
					"postgres_config": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"host":     schema.StringAttribute{Required: true},
							"port":     schema.Int32Attribute{Required: true},
							"user":     schema.StringAttribute{Required: true},
							"password": schema.StringAttribute{Required: true, Sensitive: true},
							"database": schema.StringAttribute{Required: true},
						},
					},
					"mysql_config": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"host":     schema.StringAttribute{Required: true},
							"port":     schema.Int32Attribute{Required: true},
							"user":     schema.StringAttribute{Required: true},
							"password": schema.StringAttribute{Required: true, Sensitive: true},
							"database": schema.StringAttribute{Required: true},
						},
					},
					"tables": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"uuid":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
								"name":                  schema.StringAttribute{Required: true},
								"schema":                schema.StringAttribute{Required: true},
								"enable_history_mode":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
								"individual_deployment": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
								"is_partitioned":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
							},
						},
					},
				},
			},
			"destination_config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"database":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					"schema":                    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					"dataset":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					"use_same_schema_as_source": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
					"schema_name_prefix":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
					"schema_override":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
				},
			},
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
	// Read Terraform plan data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Our API's create endpoint only accepts the source type, so we need to send two requests:
	// one to create the bare-bones deployment, then one to update it with the rest of the data
	payloadBytes, err := json.Marshal(map[string]any{"source": data.Source.Type.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	url := fmt.Sprintf("%s/deployments", r.endpoint)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "payload", string(payloadBytes))
	tflog.Info(ctx, "Creating deployment via API")
	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	ctx = tflog.SetField(ctx, "response", string(bodyBytes))
	tflog.Info(ctx, "Created deployment")

	var deployment models.DeploymentAPIModel
	err = json.Unmarshal(bodyBytes, &deployment)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	// Translate the Terraform plan into an API model and then fill in computed fields
	// from the API response of the newly created deployment
	model := models.DeploymentResourceToAPIModel(data)
	model.UUID = deployment.UUID
	model.Status = deployment.Status

	// Second API request: update the newly created deployment
	payload := map[string]any{
		"deploy":           model,
		"updateDeployOnly": true,
	}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	url = fmt.Sprintf("%s/deployments/%s", r.endpoint, deployment.UUID)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "payload", string(payloadBytes))
	tflog.Info(ctx, "Updating deployment via API")

	apiReq, err = http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	bodyBytes, err = r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	var deploymentResp models.DeploymentAPIResponse
	err = json.Unmarshal(bodyBytes, &deploymentResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DeploymentAPIToResourceModel(deploymentResp.Deployment, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments/%s", r.endpoint, data.UUID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	var deploymentResp models.DeploymentAPIResponse
	err = json.Unmarshal(bodyBytes, &deploymentResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DeploymentAPIToResourceModel(deploymentResp.Deployment, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"deploy":           models.DeploymentResourceToAPIModel(data),
		"updateDeployOnly": true,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	url := fmt.Sprintf("%s/deployments/%s", r.endpoint, data.UUID.ValueString())
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "payload", string(payloadBytes))
	tflog.Info(ctx, "Updating deployment via API")

	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	var deploymentResp models.DeploymentAPIResponse
	err = json.Unmarshal(bodyBytes, &deploymentResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DeploymentAPIToResourceModel(deploymentResp.Deployment, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/deployments/%s", r.endpoint, data.UUID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete Deployment", err.Error())
		return
	}

	_, err = r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete Deployment", err.Error())
		return
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}

func (r *DeploymentResource) handleAPIRequest(apiReq *http.Request) (bodyBytes []byte, err error) {
	apiReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	apiResp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		return []byte{}, err
	}

	if apiResp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("received status code %d", apiResp.StatusCode)
	}

	defer apiResp.Body.Close()
	return io.ReadAll(apiResp.Body)
}
