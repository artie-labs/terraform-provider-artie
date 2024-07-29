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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DestinationResource{}
var _ resource.ResourceWithConfigure = &DestinationResource{}
var _ resource.ResourceWithImportState = &DestinationResource{}

func NewDestinationResource() resource.Resource {
	return &DestinationResource{}
}

type DestinationResource struct {
	endpoint string
	apiKey   string
}

func (r *DestinationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Destination resource",
		Attributes: map[string]schema.Attribute{
			"uuid":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"company_uuid":    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssh_tunnel_uuid": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":            schema.StringAttribute{Required: true},
			"label":           schema.StringAttribute{Optional: true},
			"last_updated_at": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"config": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"host":                  schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"port":                  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
					"endpoint":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"username":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"gcp_project_id":        schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"gcp_location":          schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"aws_access_key_id":     schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"aws_region":            schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"snowflake_account_url": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
					"snowflake_virtual_dwh": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")},
				},
			},
		},
	}
}

func (r *DestinationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data models.DestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destModel := models.DestinationResourceToAPIModel(data)
	payload := map[string]any{
		"name":         destModel.Name,
		"label":        destModel.Label,
		"sharedConfig": destModel.Config,
	}
	if destModel.SSHTunnelUUID != "" {
		payload["sshTunnelUUID"] = destModel.SSHTunnelUUID
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	url := fmt.Sprintf("%s/destinations", r.endpoint)
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "payload", string(payloadBytes))
	tflog.Info(ctx, "Creating destination via API")

	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	ctx = tflog.SetField(ctx, "response", string(bodyBytes))
	tflog.Info(ctx, "Created destination")

	var destination models.DestinationAPIModel
	err = json.Unmarshal(bodyBytes, &destination)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DestinationAPIToResourceModel(destination, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data models.DestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("%s/destinations/%s", r.endpoint, data.UUID.ValueString())
	ctx = tflog.SetField(ctx, "url", url)
	tflog.Info(ctx, "Reading destination from API")
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	var destinationResp models.DestinationAPIModel
	err = json.Unmarshal(bodyBytes, &destinationResp)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DestinationAPIToResourceModel(destinationResp, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data models.DestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payloadBytes, err := json.Marshal(models.DestinationResourceToAPIModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	url := fmt.Sprintf("%s/destinations/%s", r.endpoint, data.UUID.ValueString())
	ctx = tflog.SetField(ctx, "url", url)
	ctx = tflog.SetField(ctx, "payload", string(payloadBytes))
	tflog.Info(ctx, "Updating destination via API")

	apiReq, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	bodyBytes, err := r.handleAPIRequest(apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	var destination models.DestinationAPIModel
	err = json.Unmarshal(bodyBytes, &destination)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.DestinationAPIToResourceModel(destination, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data models.DestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO implement Delete
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}

func (r *DestinationResource) handleAPIRequest(apiReq *http.Request) (bodyBytes []byte, err error) {
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
