package provider

import (
	"context"
	"fmt"
	"math"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithConfigure = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client artieclient.Client
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Deployment resource. This represents a connection that syncs data from a single source (e.g., Postgres) to a single destination (e.g., Snowflake).",
		Attributes: map[string]schema.Attribute{
			"uuid":                        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":                        schema.StringAttribute{Required: true, MarkdownDescription: "The human-readable name of the deployment. This is used only as a label and can contain any characters."},
			"status":                      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"destination_uuid":            schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This must point to an `artie_destination` resource."},
			"ssh_tunnel_uuid":             schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This can point to an `artie_ssh_tunnel` resource if you need us to use an SSH tunnel to connect to your source database."},
			"snowflake_eco_schedule_uuid": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"source": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "This contains configuration for this deployment's source database.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The type of source database. This must be one of the following: `mysql` or `postgresql`.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(models.MySQL),
								string(models.PostgreSQL),
							),
						},
					},
					"postgresql_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "This should be filled out if the source type is `postgresql`.",
						Attributes: map[string]schema.Attribute{
							"host": schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the PostgreSQL database. This must point to the primary host, not a read replica. This database must also have its `WAL_LEVEL` set to `logical`."},
							"port": schema.Int32Attribute{
								Required:            true,
								MarkdownDescription: "The default port for PostgreSQL is 5432.",
								Validators: []validator.Int32{
									int32validator.Between(1024, math.MaxUint16),
								},
							},
							"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the PostgreSQL database. This service account needs enough permissions to create and read from the replication slot."},
							"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
							"database": schema.StringAttribute{Required: true, MarkdownDescription: "The name of the database in the PostgreSQL server."},
						},
					},
					"mysql_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "This should be filled out if the source type is `mysql`.",
						Attributes: map[string]schema.Attribute{
							"host": schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the MySQL database. This must point to the primary host, not a read replica."},
							"port": schema.Int32Attribute{
								Required:            true,
								MarkdownDescription: "The default port for MySQL is 3306.",
								Validators: []validator.Int32{
									int32validator.Between(1024, math.MaxUint16),
								},
							},
							"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the MySQL database. This service account needs enough permissions to read from the server binlogs."},
							"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
							"database": schema.StringAttribute{Required: true, MarkdownDescription: "The name of the database in the MySQL server."},
						},
					},
					"tables": schema.MapNestedAttribute{
						Required:            true,
						MarkdownDescription: "A map of tables from the source database that you want to replicate to the destination. The key for each table should be formatted as `schema_name.table_name` if your source database uses schemas, otherwise just `table_name`.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"uuid":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
								"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "The name of the table in the source database."},
								"schema":                schema.StringAttribute{Optional: true, MarkdownDescription: "The name of the schema the table belongs to in the source database. This must be specified if your source database uses schemas (such as PostgreSQL), e.g. `public`."},
								"enable_history_mode":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, we will create an additional table in the destination (suffixed with `__history`) to store all changes to the source table over time."},
								"individual_deployment": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, we will spin up a separate Artie Transfer deployment to handle this table. This should only be used if this table has extremely high throughput (over 1M+ per hour) and has much higher throughput than other tables."},
								"is_partitioned":        schema.BoolAttribute{Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
							},
						},
					},
				},
			},
			"destination_config": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "This contains configuration that pertains to the destination database but is specific to this deployment. The basic connection settings for the destination, which can be shared by multiple deployments, are stored in the corresponding `artie_destination` resource.",
				Attributes: map[string]schema.Attribute{
					"database": schema.StringAttribute{
						MarkdownDescription: "The name of the database that data should be synced to in the destination. This should be filled if the destination is Snowflake, unless `use_same_schema_as_source` is set to true.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"schema": schema.StringAttribute{
						MarkdownDescription: "The name of the schema that data should be synced to in the destination. This should be filled if the destination is Snowflake or Redshift.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"dataset": schema.StringAttribute{
						MarkdownDescription: "The name of the dataset that data should be synced to in the destination. This should be filled if the destination is BigQuery.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"use_same_schema_as_source": schema.BoolAttribute{
						MarkdownDescription: "If set to true, each table from the source database will be synced to a schema with the same name as its source schema. This can only be used if the source database is PostgreSQL and the destination is Snowflake or Redshift.",
						Optional:            true,
						Computed:            true, Default: booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"schema_name_prefix": schema.StringAttribute{
						MarkdownDescription: "If `use_same_schema_as_source` is enabled, this prefix will be added to each schema name in the destination. This is useful if you want to namespace all of this deployment's schemas in the destination. This can only be used if the source database is PostgreSQL and the destination is Snowflake or Redshift.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
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

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate config before creating the deployment
	deployment := data.ToAPIBaseModel()
	if err := r.client.Deployments().ValidateSource(ctx, deployment); err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	if err := r.client.Deployments().ValidateDestination(ctx, deployment); err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	// Our API's create endpoint only accepts the source type, so we need to send two requests:
	// one to create the bare-bones deployment, then one to update it with the rest of the data
	createdDeployment, err := r.client.Deployments().Create(ctx, deployment.Source.Type)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	// Fill in computed fields from the API response of the newly created deployment
	fullDeployment := deployment.ToFullDeployment(createdDeployment.UUID, createdDeployment.Status)

	// Second API request: update the newly created deployment
	updatedDeployment, err := r.client.Deployments().Update(ctx, fullDeployment)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(updatedDeployment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.Deployments().Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(deployment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseDeployment := data.ToAPIBaseModel()
	if err := r.client.Deployments().ValidateSource(ctx, baseDeployment); err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	if err := r.client.Deployments().ValidateDestination(ctx, baseDeployment); err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	updatedDeployment, err := r.client.Deployments().Update(ctx, data.ToAPIModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	// Translate API response into Terraform state & save state
	data.UpdateFromAPIModel(updatedDeployment)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data models.DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Deployments().Delete(ctx, data.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Deployment", err.Error())
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
