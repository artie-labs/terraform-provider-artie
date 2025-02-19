package provider

import (
	"context"
	"fmt"
	"math"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
						MarkdownDescription: "The type of source database. This must be one of the following: `mysql`, `oracle`, `postgresql`.",
						Validators:          []validator.String{stringvalidator.OneOf(artieclient.AllSourceTypes...)},
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
					"oracle_config": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "This should be filled out if the source type is `oracle`.",
						Attributes: map[string]schema.Attribute{
							"host": schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the Oracle database. This must point to the primary host, not a read replica. This database must also have `ARCHIVELOG` mode and supplemental logging enabled."},
							"port": schema.Int32Attribute{
								Required:            true,
								MarkdownDescription: "The default port for Oracle is 1521.",
								Validators: []validator.Int32{
									int32validator.Between(1024, math.MaxUint16),
								},
							},
							"user":      schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the Oracle database."},
							"password":  schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
							"service":   schema.StringAttribute{Required: true, MarkdownDescription: "The name of the service in the Oracle server."},
							"container": schema.StringAttribute{Optional: true, MarkdownDescription: "The name of the container (pluggable database). Required if you are using a container database; otherwise this should be omitted."},
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
								"alias":                 schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional alias for the table. If set, this will be the name of the destination table."},
								"exclude_columns":       schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional list of columns to exclude from syncing to the destination."},
								"columns_to_hash":       schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional list of columns to hash in the destination. Values for these columns will be obscured with a one-way hash."},
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
					"bucket": schema.StringAttribute{
						MarkdownDescription: "The name of the S3 bucket that data should be synced to. This should be filled if the destination is S3.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"folder": schema.StringAttribute{
						MarkdownDescription: "If provided, all files will be stored under this folder inside the S3 bucket. This is optional and only applies if the destination is S3.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"drop_deleted_columns":               schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, when a column is dropped from the source it will also be dropped in the destination."},
			"soft_delete_rows":                   schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, a new boolean column called __artie_delete will be added to your destination to indicate if the row has been deleted."},
			"include_artie_updated_at_column":    schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will add a new column to your dataset called __artie_updated_at."},
			"include_database_updated_at_column": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will add a new column to your dataset called __artie_db_updated_at."},
			"one_topic_per_schema":               schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will write all incoming CDC events into a single Kafka topic per schema. This only works if your source is Oracle and your account has this feature enabled."},
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

func (r *DeploymentResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.Deployment
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *DeploymentResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.Deployment, bool) {
	var planData tfmodels.Deployment
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *DeploymentResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, deployment artieclient.Deployment) {
	// Translate API response type into Terraform model and save it into state
	diagnostics.Append(state.Set(ctx, tfmodels.DeploymentFromAPIModel(deployment))...)
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	// Validate config before creating the deployment
	deployment := planData.ToAPIBaseModel()
	if err := r.client.Deployments().ValidateSource(ctx, deployment); err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	if err := r.client.Deployments().ValidateDestination(ctx, deployment); err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	createdDeployment, err := r.client.Deployments().Create(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Deployment", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, createdDeployment)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	deploymentUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	deployment, err := r.client.Deployments().Get(ctx, deploymentUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Deployment", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, deployment)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	// Validate source & destination config before updating the deployment
	baseDeployment := planData.ToAPIBaseModel()
	if err := r.client.Deployments().ValidateSource(ctx, baseDeployment); err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}
	if err := r.client.Deployments().ValidateDestination(ctx, baseDeployment); err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	updatedDeployment, err := r.client.Deployments().Update(ctx, planData.ToAPIModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Deployment", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedDeployment)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deploymentUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.Deployments().Delete(ctx, deploymentUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Deployment", err.Error())
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
