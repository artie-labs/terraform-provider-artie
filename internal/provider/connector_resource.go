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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ConnectorResource{}
var _ resource.ResourceWithConfigure = &ConnectorResource{}
var _ resource.ResourceWithImportState = &ConnectorResource{}

func NewConnectorResource() resource.Resource {
	return &ConnectorResource{}
}

type ConnectorResource struct {
	client artieclient.Client
}

func (r *ConnectorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *ConnectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Connector resource. This represents a database or data warehouse that you want to sync data from or to. Connectors are used by Deployments and Source Readers.",
		Attributes: map[string]schema.Attribute{
			"uuid":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssh_tunnel_uuid": schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This can point to an `artie_ssh_tunnel` resource if you need us to use an SSH tunnel to connect."},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of connector. This must be one of the following: `bigquery`, `dynamodb`, `mongodb`, `mysql`, `mssql`, `oracle`, `postgresql`, `redshift`, `s3`, `snowflake`.",
				Validators:          []validator.String{stringvalidator.OneOf(artieclient.AllConnectorTypes...)},
			},
			"name": schema.StringAttribute{Optional: true, MarkdownDescription: "An optional human-readable label for this connector."},
			"bigquery_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the connector type is `bigquery`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"project_id":       schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the Google Cloud project."},
					"location":         schema.StringAttribute{Required: true, MarkdownDescription: "The location of the BigQuery dataset. This must be either `US` or `EU`."},
					"credentials_data": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The credentials data for the Google Cloud service account that we should use to connect to BigQuery. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"dynamodb_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `dynamodb`.",
				Attributes: map[string]schema.Attribute{
					"stream_arn":        schema.StringAttribute{Required: true, MarkdownDescription: "The ARN (Amazon Resource Name) of the DynamoDB Stream."},
					"access_key_id":     schema.StringAttribute{Required: true, MarkdownDescription: "The AWS Access Key ID for the service account we should use to connect to DynamoDB."},
					"secret_access_key": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The AWS Secret Access Key for the service account we should use to connect to DynamoDB. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
					"backfill":          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "Whether or not we should backfill all existing data from DynamoDB to your destination."},
					"backfill_bucket":   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If backfill = true, specify the S3 bucket where the DynamoDB export should be stored."},
					"backfill_folder":   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If backfill = true, optionally specify the folder where the DynamoDB export should be stored within the specified S3 bucket."},
				},
			},
			"mongodb_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `mongodb`.",
				Attributes: map[string]schema.Attribute{
					"host":     schema.StringAttribute{Required: true, MarkdownDescription: "The connection string for the MongoDB server. This can be either SRV or standard format."},
					"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the MongoDB database."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account we will use to connect to the MongoDB database. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"mysql_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `mysql`.",
				Attributes: map[string]schema.Attribute{
					"host":          schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the MySQL database. This must point to the primary host, not a read replica."},
					"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The hostname of the MySQL database that we should use to snapshot the database. This can be a read replica and will only be used if this connector is being used as a source. If not provided, we will use the `host` value."},
					"port": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "The default port for MySQL is 3306.",
						Validators: []validator.Int32{
							int32validator.Between(1024, math.MaxUint16),
						},
					},
					"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the MySQL database. This service account needs enough permissions to read from the server binlogs."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"mssql_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `mssql`.",
				Attributes: map[string]schema.Attribute{
					"host":          schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the Microsoft SQL Server. This must point to the primary host, not a read replica."},
					"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The hostname of the Microsoft SQL Server that we should use to snapshot the database. This can be a read replica and will only be used if this connector is being used as a source. If not provided, we will use the `host` value."},
					"port": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "The default port for Microsoft SQL Server is 1433.",
						Validators: []validator.Int32{
							int32validator.Between(1024, math.MaxUint16),
						},
					},
					"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the database."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"oracle_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `oracle`.",
				Attributes: map[string]schema.Attribute{
					"host":          schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the Oracle database. This must point to the primary host, not a read replica. This database must also have `ARCHIVELOG` mode and supplemental logging enabled."},
					"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The hostname of the Oracle database that we should use to snapshot the database. This can be a read replica and will only be used if this connector is being used as a source. If not provided, we will use the `host` value."},
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
				MarkdownDescription: "This should be filled out if the connector type is `postgresql`.",
				Attributes: map[string]schema.Attribute{
					"host":          schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the PostgreSQL database. This must point to the primary host, not a read replica. This database must also have its `WAL_LEVEL` set to `logical`."},
					"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The hostname of the PostgreSQL database that we should use to snapshot the database. This can be a read replica and will only be used if this connector is being used as a source. If not provided, we will use the `host` value."},
					"port": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "The default port for PostgreSQL is 5432.",
						Validators: []validator.Int32{
							int32validator.Between(1024, math.MaxUint16),
						},
					},
					"user":     schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the PostgreSQL database. This service account needs enough permissions to create and read from the replication slot."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"redshift_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the connector type is `redshift`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{Required: true, MarkdownDescription: "The endpoint URL of your Redshift cluster. This should include both the host and port."},
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we should use to connect to Redshift."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password for the service account we should use to connect to Redshift. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"s3_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the connector type is `s3`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"access_key_id":     schema.StringAttribute{Required: true, MarkdownDescription: "The AWS Access Key ID for the service account we should use to connect to S3."},
					"secret_access_key": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The AWS Secret Access Key for the service account we should use to connect to S3. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
					"region":            schema.StringAttribute{Required: true, MarkdownDescription: "The AWS region where we should store your data in S3."},
				},
			},
			"snowflake_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the connector type is `snowflake`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"account_url": schema.StringAttribute{Required: true, MarkdownDescription: "The URL of your Snowflake account."},
					"virtual_dwh": schema.StringAttribute{Required: true, MarkdownDescription: "The name of your Snowflake virtual data warehouse."},
					"username":    schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we should use to connect to Snowflake."},
					"password":    schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The password for the service account we should use to connect to Snowflake. Either `password` or `private_key` must be provided."},
					"private_key": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The private key for the service account we should use to connect to Snowflake. Either `password` or `private_key` must be provided."},
				},
			},
		},
	}
}

func (r *ConnectorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConnectorResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.Destination
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *ConnectorResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.Destination, bool) {
	var planData tfmodels.Destination
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *ConnectorResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiConnector artieclient.Destination) {
	// Translate API response type into Terraform model and save it into state
	connector, diags := tfmodels.DestinationFromAPIModel(apiConnector)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, connector)...)
}

func (r *ConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	baseConnector, diags := planData.ToAPIBaseModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	connector, err := r.client.Destinations().Create(ctx, baseConnector)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Connector", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, connector)
}

func (r *ConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	connectorUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	connector, err := r.client.Destinations().Get(ctx, connectorUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Connector", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, connector)
}

func (r *ConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	apiBaseModel, diags := planData.ToAPIBaseModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if err := r.client.Destinations().TestConnection(ctx, apiBaseModel); err != nil {
		resp.Diagnostics.AddError("Unable to Update Connector", err.Error())
		return
	}

	apiModel, diags := planData.ToAPIModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	updatedConnector, err := r.client.Destinations().Update(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Connector", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedConnector)
}

func (r *ConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	connectorUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.Destinations().Delete(ctx, connectorUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Connector", err.Error())
		return
	}
}

func (r *ConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
