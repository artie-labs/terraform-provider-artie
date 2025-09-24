package provider

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/artie-labs/transfer/lib/maputil"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"
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
	connectorTypes := maputil.NewSortedStringsMap[bool]()
	for _, connectorType := range artieclient.AllConnectorTypes {
		connectorTypes.Add(fmt.Sprintf("`%s`", connectorType), true)
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Connector resource. This represents a database or data warehouse that you want to sync data from or to. Connectors are used by Pipelines and Source Readers.",
		Attributes: map[string]schema.Attribute{
			"uuid":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"ssh_tunnel_uuid": schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This can point to an `artie_ssh_tunnel` resource if you need us to use an SSH tunnel to connect."},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: fmt.Sprintf("The type of connector. This must be one of the following: %s.", strings.Join(connectorTypes.Keys(), ", ")),
				Validators:          []validator.String{stringvalidator.OneOf(artieclient.AllConnectorTypes...)},
			},
			"name": schema.StringAttribute{Optional: true, MarkdownDescription: "An optional human-readable label for this connector."},
			"data_plane_name": schema.StringAttribute{
				MarkdownDescription: "The name of the data plane this connector is in (if applicable; this does not apply to cloud-based connectors like BigQuery and Snowflake). If this is not set, we will use the default data plane for your account. To see the full list of supported data planes on your account, click on 'New pipeline' in our UI.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
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
				},
			},
			"mongodb_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `mongodb`.",
				Attributes: map[string]schema.Attribute{
					"host":     schema.StringAttribute{Required: true, MarkdownDescription: "The connection string for the MongoDB server. This can be either SRV or standard format."},
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the MongoDB database."},
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
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the MySQL database. This service account needs enough permissions to read from the server binlogs."},
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
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the database."},
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
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the Oracle database."},
					"password": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file."},
				},
			},
			"postgresql_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "This should be filled out if the connector type is `postgresql`.",
				Attributes: map[string]schema.Attribute{
					"host":          schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the PostgreSQL database. This can point to a read replica if you are using PostgreSQL 16 or higher, not on Amazon Aurora, and `hot_standby_feedback` is enabled; otherwise it must point to the primary host. This database must also have its `WAL_LEVEL` set to `logical`."},
					"snapshot_host": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The hostname of the PostgreSQL database that we should use to snapshot the database. This can be a read replica and will only be used if this connector is being used as a source. If not provided, we will use the `host` value."},
					"port": schema.Int32Attribute{
						Required:            true,
						MarkdownDescription: "The default port for PostgreSQL is 5432.",
						Validators: []validator.Int32{
							int32validator.Between(1024, math.MaxUint16),
						},
					},
					"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we will use to connect to the PostgreSQL database. This service account needs enough permissions to create and read from the replication slot."},
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
					"account_identifier": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The [account identifier](https://docs.snowflake.com/user-guide/admin-account-identifier) of your Snowflake account. We recommend using this instead of `account_url`."},
					"account_url":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "(Legacy) The [URL](https://docs.snowflake.com/user-guide/admin-account-identifier) of your Snowflake account. We recommend using `account_identifier` instead."},
					"virtual_dwh":        schema.StringAttribute{Required: true, MarkdownDescription: "The name of your Snowflake virtual data warehouse."},
					"username":           schema.StringAttribute{Required: true, MarkdownDescription: "The username of the service account we should use to connect to Snowflake."},
					"password":           schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "(Legacy) The password for the service account we should use to connect to Snowflake. We recommend using `private_key` instead."},
					"private_key":        schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""), MarkdownDescription: "The private key for the service account we should use to connect to Snowflake. We recommend using this instead of `password`."},
				},
			},
			"databricks_config": schema.SingleNestedAttribute{
				MarkdownDescription: "This should be filled out if the connector type is `databricks`.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"host":                  schema.StringAttribute{Required: true, MarkdownDescription: "The hostname of the Databricks cluster."},
					"http_path":             schema.StringAttribute{Required: true, MarkdownDescription: "The HTTP path of the Databricks cluster."},
					"personal_access_token": schema.StringAttribute{Required: true, Sensitive: true, MarkdownDescription: "The personal access token for the service account we should use to connect to Databricks."},
					"volume":                schema.StringAttribute{Required: true, MarkdownDescription: "The volume of the Databricks cluster."},
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
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData))
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
	var stateData tfmodels.Connector
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *ConnectorResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.Connector, bool) {
	var planData tfmodels.Connector
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *ConnectorResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiConnector artieclient.Connector) {
	// Translate API response type into Terraform model and save it into state
	connector, diags := tfmodels.ConnectorFromAPIModel(apiConnector)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, connector)...)
}

func (r *ConnectorResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var configData tfmodels.Connector
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch configData.Type.ValueString() {
	case string(artieclient.BigQuery):
		if configData.BigQueryConfig == nil {
			resp.Diagnostics.AddError("bigquery_config is required", "Please provide `bigquery_config` inside `connector`.")
			return
		}
	case string(artieclient.DynamoDB):
		if configData.DynamoDBConfig == nil {
			resp.Diagnostics.AddError("dynamodb_config is required", "Please provide `dynamodb_config` inside `connector`.")
			return
		}
	case string(artieclient.MongoDB):
		if configData.MongoDBConfig == nil {
			resp.Diagnostics.AddError("mongodb_config is required", "Please provide `mongodb_config` inside `connector`.")
			return
		}
	case string(artieclient.MySQL):
		if configData.MySQLConfig == nil {
			resp.Diagnostics.AddError("mysql_config is required", "Please provide `mysql_config` inside `connector`.")
			return
		}
	case string(artieclient.MSSQL):
		if configData.MSSQLConfig == nil {
			resp.Diagnostics.AddError("mssql_config is required", "Please provide `mssql_config` inside `connector`.")
			return
		}
	case string(artieclient.Oracle):
		if configData.OracleConfig == nil {
			resp.Diagnostics.AddError("oracle_config is required", "Please provide `oracle_config` inside `connector`.")
			return
		}
	case string(artieclient.PostgreSQL):
		if configData.PostgresConfig == nil {
			resp.Diagnostics.AddError("postgresql_config is required", "Please provide `postgresql_config` inside `connector`.")
			return
		}
	case string(artieclient.Redshift):
		if configData.RedshiftConfig == nil {
			resp.Diagnostics.AddError("redshift_config is required", "Please provide `redshift_config` inside `connector`.")
			return
		}
	case string(artieclient.S3):
		if configData.S3Config == nil {
			resp.Diagnostics.AddError("s3_config is required", "Please provide `s3_config` inside `connector`.")
			return
		}
	case string(artieclient.Snowflake):
		if configData.SnowflakeConfig == nil {
			resp.Diagnostics.AddError("snowflake_config is required", "Please provide `snowflake_config` inside `connector`.")
			return
		}

		// For snowflake, either account_identifier or account_url must be provided
		if configData.SnowflakeConfig.AccountIdentifier.IsNull() && configData.SnowflakeConfig.AccountURL.IsNull() {
			resp.Diagnostics.AddError("Either account_identifier or account_url must be provided", "Please provide either `account_identifier` or `account_url` inside `snowflake_config`. We recommend using `account_identifier`.")
		}

		// Either password or private_key must be provided
		if configData.SnowflakeConfig.Password.IsNull() && configData.SnowflakeConfig.PrivateKey.IsNull() {
			resp.Diagnostics.AddError("Either password or private_key must be provided", "Please provide either `password` or `private_key` inside `snowflake_config`. We recommend using `private_key`.")
		}
	case string(artieclient.Databricks):
		if configData.DatabricksConfig == nil {
			resp.Diagnostics.AddError("databricks_config is required", "Please provide `databricks_config` inside `connector`.")
			return
		}
	}
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

	connector, err := r.client.Connectors().Create(ctx, baseConnector)
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

	connector, err := r.client.Connectors().Get(ctx, connectorUUID)
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

	apiModel, diags := planData.ToAPIModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	updatedConnector, err := r.client.Connectors().Update(ctx, apiModel)
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

	if err := r.client.Connectors().Delete(ctx, connectorUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Connector", err.Error())
		return
	}
}

func (r *ConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
