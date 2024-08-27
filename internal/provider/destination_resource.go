package provider

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DestinationResource{}
var _ resource.ResourceWithConfigure = &DestinationResource{}
var _ resource.ResourceWithImportState = &DestinationResource{}

func NewDestinationResource() resource.Resource {
	return &DestinationResource{}
}

type DestinationResource struct {
	client artieclient.Client
}

func (r *DestinationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Destination resource. This represents a database or data warehouse that you want to replicate data to. Destinations are used by Deployments.",
		Attributes: map[string]schema.Attribute{
			"uuid":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssh_tunnel_uuid": schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This can point to an `artie_ssh_tunnel` resource if you need us to use an SSH tunnel to connect to your destination database. This can only be used if the destination is Redshift."},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of destination database. This must be one of the following: `bigquery`, `redshift`, or `snowflake`.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(tfmodels.BigQuery),
						string(tfmodels.Redshift),
						string(tfmodels.Snowflake),
					),
				},
			},
			"label": schema.StringAttribute{Optional: true, MarkdownDescription: "An optional human-readable label for this destination."},
			"snowflake_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the destination type is `snowflake`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"account_url": schema.StringAttribute{Required: true, MarkdownDescription: "The URL of your Snowflake account."},
					"virtual_dwh": schema.StringAttribute{Required: true, MarkdownDescription: "The name of your Snowflake virtual data warehouse."},
					"username":    schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we should use to connect to Snowflake."},
					"password":    schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The password for the service account we should use to connect to Snowflake. Either `password` or `private_key` must be provided."},
					"private_key": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The private key for the service account we should use to connect to Snowflake. Either `password` or `private_key` must be provided."},
				},
			},
			"bigquery_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the destination type is `bigquery`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"project_id":       schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the Google Cloud project."},
					"location":         schema.StringAttribute{Required: true, MarkdownDescription: "The location of the BigQuery dataset. This must be either `US` or `EU`."},
					"credentials_data": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The credentials data for the Google Cloud service account that we should use to connect to BigQuery. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"redshift_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the destination type is `redshift`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{Required: true, MarkdownDescription: "The endpoint URL of your Redshift cluster. This should include both the host and port."},
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we should use to connect to Redshift."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password for the service account we should use to connect to Redshift."},
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

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data tfmodels.DestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destination, err := r.client.Destinations().Create(ctx, data.ToAPIBaseModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(destination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data tfmodels.DestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	destinationResp, err := r.client.Destinations().Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(destinationResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data tfmodels.DestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Destinations().TestConnection(ctx, data.ToAPIBaseModel()); err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	fullDestination := data.ToAPIModel()
	updatedDestination, err := r.client.Destinations().Update(ctx, fullDestination)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(updatedDestination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data tfmodels.DestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Destinations().Delete(ctx, data.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Destination", err.Error())
		return
	}
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
