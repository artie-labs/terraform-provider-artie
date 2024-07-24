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
	DestinationUUID      types.String `tfsdk:"destination_uuid"`
	HasUndeployedChanges types.Bool   `tfsdk:"has_undeployed_changes"`
	Source               *SourceModel `tfsdk:"source"`
	AdvancedSettings     types.Map    `tfsdk:"advanced_settings"`
	UniqueConfig         types.Map    `tfsdk:"unique_config"`
}

type SourceModel struct {
	Name   types.String      `tfsdk:"name"`
	Config SourceConfigModel `tfsdk:"config"`
	Tables []TableModel      `tfsdk:"tables"`
}

type SourceConfigModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	// Password
	// DynamoDBConfig
	// SnapshotHost
}

type TableModel struct {
	UUID                 types.String               `tfsdk:"uuid"`
	Name                 types.String               `tfsdk:"name"`
	Schema               types.String               `tfsdk:"schema"`
	EnableHistoryMode    types.Bool                 `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool                 `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool                 `tfsdk:"is_partitioned"`
	AdvancedSettings     TableAdvancedSettingsModel `tfsdk:"advanced_settings"`
}

type TableAdvancedSettingsModel struct {
	Alias                types.String `tfsdk:"alias"`
	SkipDelete           types.Bool   `tfsdk:"skip_delete"`
	FlushIntervalSeconds types.Int64  `tfsdk:"flush_interval_seconds"`
	BufferRows           types.Int64  `tfsdk:"buffer_rows"`
	FlushSizeKB          types.Int64  `tfsdk:"flush_size_kb"`
	// BigQueryPartitionSettings
	// MergePredicates
	// AutoscaleMaxReplicas
	// AutoscaleTargetValue
	// K8sRequestCPU
	// K8sRequestMemoryMB
	// ExcludeColumns
}

type DeploymentResponse struct {
	Deployment DeploymentResourceAPIModel `json:"deploy"`
}

type DeploymentResourceAPIModel struct {
	UUID                 string         `json:"uuid"`
	Name                 string         `json:"name"`
	Status               string         `json:"status"`
	LastUpdatedAt        string         `json:"lastUpdatedAt"`
	DestinationUUID      string         `json:"destinationUUID"`
	HasUndeployedChanges bool           `json:"hasUndeployedChanges"`
	Source               SourceAPIModel `json:"source"`
	AdvancedSettings     map[string]any `json:"advancedSettings"`
	UniqueConfig         map[string]any `json:"uniqueConfig"`
}

type SourceAPIModel struct {
	Name   string               `json:"name"`
	Config SourceConfigAPIModel `json:"config"`
	Tables []TableAPIModel      `json:"tables"`
}

type SourceConfigAPIModel struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
}

type TableAPIModel struct {
	UUID                 string                        `json:"uuid"`
	Name                 string                        `json:"name"`
	Schema               string                        `json:"schema"`
	EnableHistoryMode    bool                          `json:"enableHistoryMode"`
	IndividualDeployment bool                          `json:"individualDeployment"`
	IsPartitioned        bool                          `json:"isPartitioned"`
	AdvancedSettings     TableAdvancedSettingsAPIModel `json:"advancedSettings"`
}

type TableAdvancedSettingsAPIModel struct {
	Alias                string `json:"alias"`
	SkipDelete           bool   `json:"skip_delete"`
	FlushIntervalSeconds int64  `json:"flush_interval_seconds"`
	BufferRows           int64  `json:"buffer_rows"`
	FlushSizeKB          int64  `json:"flush_size_kb"`
	// BigQueryPartitionSettings
	// MergePredicates
	// AutoscaleMaxReplicas
	// AutoscaleTargetValue
	// K8sRequestCPU
	// K8sRequestMemoryMB
	// ExcludeColumns
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
			"destination_uuid":       schema.StringAttribute{Computed: true},
			"has_undeployed_changes": schema.BoolAttribute{Computed: true},
			"source": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Required: true},
					"config": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"host":     schema.StringAttribute{Required: true},
							"port":     schema.NumberAttribute{Required: true},
							"user":     schema.StringAttribute{Required: true},
							"database": schema.StringAttribute{Required: true},
						},
					},
					"tables": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"uuid":                  schema.StringAttribute{Computed: true},
								"name":                  schema.StringAttribute{Required: true},
								"schema":                schema.StringAttribute{Required: true},
								"enable_history_mode":   schema.BoolAttribute{Optional: true},
								"individual_deployment": schema.BoolAttribute{Optional: true},
								"is_partitioned":        schema.BoolAttribute{Optional: true},
								"advanced_settings": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"alias":                  schema.StringAttribute{Optional: true},
										"skip_delete":            schema.BoolAttribute{Optional: true},
										"flush_interval_seconds": schema.NumberAttribute{Optional: true},
										"buffer_rows":            schema.NumberAttribute{Optional: true},
										"flush_size_kb":          schema.NumberAttribute{Optional: true},
									},
								},
							},
						},
					},
				},
			},
			"advanced_settings": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"unique_config": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
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

	data.Name = types.StringValue(deploymentResp.Deployment.Name)
	data.Status = types.StringValue(deploymentResp.Deployment.Status)
	data.LastUpdatedAt = types.StringValue(deploymentResp.Deployment.LastUpdatedAt)
	data.HasUndeployedChanges = types.BoolValue(deploymentResp.Deployment.HasUndeployedChanges)
	data.DestinationUUID = types.StringValue(deploymentResp.Deployment.DestinationUUID)

	tables := []TableModel{}
	for _, apiTable := range deploymentResp.Deployment.Source.Tables {
		tables = append(tables, TableModel{
			UUID:                 types.StringValue(apiTable.UUID),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			AdvancedSettings: TableAdvancedSettingsModel{
				Alias:                types.StringValue(apiTable.AdvancedSettings.Alias),
				SkipDelete:           types.BoolValue(apiTable.AdvancedSettings.SkipDelete),
				FlushIntervalSeconds: types.Int64Value(apiTable.AdvancedSettings.FlushIntervalSeconds),
				BufferRows:           types.Int64Value(apiTable.AdvancedSettings.BufferRows),
				FlushSizeKB:          types.Int64Value(apiTable.AdvancedSettings.FlushSizeKB),
			},
		})
	}
	data.Source = &SourceModel{
		Name: types.StringValue(deploymentResp.Deployment.Source.Name),
		Config: SourceConfigModel{
			Host:     types.StringValue(deploymentResp.Deployment.Source.Config.Host),
			Port:     types.Int64Value(deploymentResp.Deployment.Source.Config.Port),
			User:     types.StringValue(deploymentResp.Deployment.Source.Config.User),
			Database: types.StringValue(deploymentResp.Deployment.Source.Config.Database),
		},
		Tables: tables,
	}

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
