package provider

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
				MarkdownDescription: "The type of destination database. This must be one of the following: `bigquery`, `redshift`, `snowflake`, `s3`.",
				Validators:          []validator.String{stringvalidator.OneOf(artieclient.AllDestinationTypes...)},
			},
			"label": schema.StringAttribute{Optional: true, MarkdownDescription: "An optional human-readable label for this destination."},
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
			"s3_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the destination type is `s3`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"access_key_id":     schema.StringAttribute{Required: true, MarkdownDescription: "The AWS Access Key ID for the service account we should use to connect to S3."},
					"secret_access_key": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The AWS Secret Access Key for the service account we should use to connect to S3. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
					"region":            schema.StringAttribute{Required: true, MarkdownDescription: "The AWS region where we should store your data in S3."},
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

func (r *DestinationResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.Destination
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *DestinationResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.Destination, bool) {
	var planData tfmodels.Destination
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *DestinationResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, destination artieclient.Destination) {
	// Translate API response type into Terraform model and save it into state
	diagnostics.Append(state.Set(ctx, tfmodels.DestinationFromAPIModel(destination))...)
}

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	baseDestination := planData.ToAPIBaseModel()
	if err := r.client.Destinations().TestConnection(ctx, baseDestination); err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	destination, err := r.client.Destinations().Create(ctx, baseDestination)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Destination", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, destination)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	destinationUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	destination, err := r.client.Destinations().Get(ctx, destinationUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Destination", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, destination)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.Destinations().TestConnection(ctx, planData.ToAPIBaseModel()); err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	updatedDestination, err := r.client.Destinations().Update(ctx, planData.ToAPIModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Destination", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedDestination)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	destinationUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.Destinations().Delete(ctx, destinationUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Destination", err.Error())
		return
	}
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
