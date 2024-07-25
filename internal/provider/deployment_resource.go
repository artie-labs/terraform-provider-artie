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
			"uuid":                   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"company_uuid":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":                   schema.StringAttribute{Required: true},
			"status":                 schema.StringAttribute{Computed: true, Optional: true},
			"last_updated_at":        schema.StringAttribute{Computed: true},
			"destination_uuid":       schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"has_undeployed_changes": schema.BoolAttribute{Computed: true},
			"source": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Required: true},
					"config": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"host":          schema.StringAttribute{Required: true},
							"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
							"port":          schema.Int64Attribute{Required: true},
							"user":          schema.StringAttribute{Required: true},
							"database":      schema.StringAttribute{Required: true},
							"dynamodb": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"region":                schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
									"table_name":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
									"streams_arn":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
									"aws_access_key_id":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
									"aws_secret_access_key": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
								},
							},
							// TODO Password
						},
					},
					"tables": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"uuid":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
								"name":                  schema.StringAttribute{Required: true},
								"schema":                schema.StringAttribute{Required: true},
								"enable_history_mode":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
								"individual_deployment": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
								"is_partitioned":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
								"advanced_settings": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"alias":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
										"skip_delete":            schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
										"flush_interval_seconds": schema.Int64Attribute{Optional: true, Computed: true},
										"buffer_rows":            schema.Int64Attribute{Optional: true, Computed: true},
										"flush_size_kb":          schema.Int64Attribute{Optional: true, Computed: true},
										"autoscale_max_replicas": schema.Int64Attribute{Optional: true, Computed: true},
										"autoscale_target_value": schema.Int64Attribute{Optional: true, Computed: true},
										"k8s_request_cpu":        schema.Int64Attribute{Optional: true, Computed: true},
										"k8s_request_memory_mb":  schema.Int64Attribute{Optional: true, Computed: true},
										// TODO BigQueryPartitionSettings, MergePredicates, ExcludeColumns
									},
									Computed: true,
								},
							},
						},
					},
				},
			},
			"advanced_settings": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"drop_deleted_columns":               schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"include_artie_updated_at_column":    schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
					"include_database_updated_at_column": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"enable_heartbeats":                  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"enable_soft_delete":                 schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"flush_interval_seconds":             schema.Int64Attribute{Optional: true, Computed: true},
					"buffer_rows":                        schema.Int64Attribute{Optional: true, Computed: true},
					"flush_size_kb":                      schema.Int64Attribute{Optional: true, Computed: true},
					"publication_name_override":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"replication_slot_override":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"publication_auto_create_mode":       schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					// TODO PartitionRegex
				},
			},
			"destination_config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"database":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"schema":                    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"dataset":                   schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"use_same_schema_as_source": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
					"schema_name_prefix":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"schema_override":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"bucket_name":               schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"optional_prefix":           schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
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
	var data models.DeploymentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO implement Create

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.DeploymentResourceModel

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

	bodyBytes, err := r.handleAPIRequest(ctx, apiReq)
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
	var data models.DeploymentResourceModel

	// Read Terraform plan data into the model
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

	bodyBytes, err := r.handleAPIRequest(ctx, apiReq)
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
	var data models.DeploymentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO implement Delete
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}

func (r *DeploymentResource) handleAPIRequest(ctx context.Context, apiReq *http.Request) (bodyBytes []byte, err error) {
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
