package tfmodels

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Connector struct {
	UUID            types.String           `tfsdk:"uuid"`
	SSHTunnelUUID   types.String           `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String           `tfsdk:"type"`
	Name            types.String           `tfsdk:"name"`
	DataPlaneName   types.String           `tfsdk:"data_plane_name"`
	BigQueryConfig  *BigQuerySharedConfig  `tfsdk:"bigquery_config"`
	DynamoDBConfig  *DynamoDBConfig        `tfsdk:"dynamodb_config"`
	MongoDBConfig   *MongoDBSharedConfig   `tfsdk:"mongodb_config"`
	MySQLConfig     *MySQLSharedConfig     `tfsdk:"mysql_config"`
	MSSQLConfig     *MSSQLSharedConfig     `tfsdk:"mssql_config"`
	OracleConfig    *OracleSharedConfig    `tfsdk:"oracle_config"`
	PostgresConfig  *PostgresSharedConfig  `tfsdk:"postgresql_config"`
	RedshiftConfig  *RedshiftSharedConfig  `tfsdk:"redshift_config"`
	S3Config        *S3SharedConfig        `tfsdk:"s3_config"`
	SnowflakeConfig *SnowflakeSharedConfig `tfsdk:"snowflake_config"`
}

func (c Connector) ToAPIBaseModel() (artieclient.BaseConnector, diag.Diagnostics) {
	var sharedConfig artieclient.ConnectorConfig
	connectorType, err := artieclient.ConnectorTypeFromString(c.Type.ValueString())
	if err != nil {
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Connector to API model", err.Error(),
		)}
	}

	switch connectorType {
	case artieclient.BigQuery:
		sharedConfig = c.BigQueryConfig.ToAPIModel()
	case artieclient.DynamoDB:
		sharedConfig = c.DynamoDBConfig.ToAPIModel()
	case artieclient.MongoDB:
		sharedConfig = c.MongoDBConfig.ToAPIModel()
	case artieclient.MySQL:
		sharedConfig = c.MySQLConfig.ToAPIModel()
	case artieclient.MSSQL:
		sharedConfig = c.MSSQLConfig.ToAPIModel()
	case artieclient.Oracle:
		sharedConfig = c.OracleConfig.ToAPIModel()
	case artieclient.PostgreSQL:
		sharedConfig = c.PostgresConfig.ToAPIModel()
	case artieclient.Redshift:
		sharedConfig = c.RedshiftConfig.ToAPIModel()
	case artieclient.S3:
		sharedConfig = c.S3Config.ToAPIModel()
	case artieclient.Snowflake:
		sharedConfig = c.SnowflakeConfig.ToAPIModel()
	default:
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Connector to API model", fmt.Sprintf("unhandled connector type: %s", c.Type.ValueString()),
		)}
	}

	sshTunnelUUID, diags := parseOptionalUUID(c.SSHTunnelUUID)
	if diags.HasError() {
		return artieclient.BaseConnector{}, diags
	}

	return artieclient.BaseConnector{
		Type:          connectorType,
		DataPlaneName: c.DataPlaneName.ValueString(),
		Label:         c.Name.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: sshTunnelUUID,
	}, diags
}

func (c Connector) ToAPIModel() (artieclient.Connector, diag.Diagnostics) {
	baseModel, diags := c.ToAPIBaseModel()
	if diags.HasError() {
		return artieclient.Connector{}, diags
	}

	uuid, uuidDiags := parseUUID(c.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Connector{}, diags
	}

	return artieclient.Connector{
		UUID:          uuid,
		BaseConnector: baseModel,
	}, diags
}

func ConnectorFromAPIModel(apiModel artieclient.Connector) (Connector, diag.Diagnostics) {
	connector := Connector{
		UUID:          types.StringValue(apiModel.UUID.String()),
		Type:          types.StringValue(string(apiModel.Type)),
		DataPlaneName: types.StringValue(apiModel.DataPlaneName),
		Name:          types.StringValue(apiModel.Label),
		SSHTunnelUUID: optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
	}

	switch apiModel.Type {
	case artieclient.BigQuery:
		connector.BigQueryConfig = BigQuerySharedConfigFromAPIModel(apiModel.Config)
	case artieclient.DynamoDB:
		connector.DynamoDBConfig = DynamoDBConfigFromAPIModel(apiModel.Config)
	case artieclient.MongoDB:
		connector.MongoDBConfig = MongoDBSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.MySQL:
		connector.MySQLConfig = MySQLSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.MSSQL:
		connector.MSSQLConfig = MSSQLSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Oracle:
		connector.OracleConfig = OracleSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.PostgreSQL:
		connector.PostgresConfig = PostgresSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Redshift:
		connector.RedshiftConfig = RedshiftSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.S3:
		connector.S3Config = S3SharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Snowflake:
		connector.SnowflakeConfig = SnowflakeSharedConfigFromAPIModel(apiModel.Config)
	default:
		return Connector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert API model to Connector", fmt.Sprintf("invalid connector type: %s", apiModel.Type),
		)}
	}

	return connector, nil
}

type BigQuerySharedConfig struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

func (b BigQuerySharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		GCPProjectID:       b.ProjectID.ValueString(),
		GCPLocation:        b.Location.ValueString(),
		GCPCredentialsData: b.CredentialsData.ValueString(),
	}
}

func BigQuerySharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *BigQuerySharedConfig {
	return &BigQuerySharedConfig{
		ProjectID:       types.StringValue(apiModel.GCPProjectID),
		Location:        types.StringValue(apiModel.GCPLocation),
		CredentialsData: types.StringValue(apiModel.GCPCredentialsData),
	}
}

type MongoDBSharedConfig struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (m MongoDBSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:     m.Host.ValueString(),
		User:     m.Username.ValueString(),
		Password: m.Password.ValueString(),
	}
}

func MongoDBSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *MongoDBSharedConfig {
	return &MongoDBSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Username: types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
	}
}

type MySQLSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (m MySQLSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:         m.Host.ValueString(),
		SnapshotHost: m.SnapshotHost.ValueString(),
		Port:         m.Port.ValueInt32(),
		User:         m.Username.ValueString(),
		Password:     m.Password.ValueString(),
	}
}

func MySQLSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *MySQLSharedConfig {
	return &MySQLSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		Username: types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
	}
}

type MSSQLSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (r MSSQLSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:         r.Host.ValueString(),
		SnapshotHost: r.SnapshotHost.ValueString(),
		Port:         r.Port.ValueInt32(),
		Username:     r.Username.ValueString(),
		Password:     r.Password.ValueString(),
	}
}

func MSSQLSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *MSSQLSharedConfig {
	return &MSSQLSharedConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		Username:     types.StringValue(apiModel.Username),
		Password:     types.StringValue(apiModel.Password),
	}
}

type OracleSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (o OracleSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:         o.Host.ValueString(),
		SnapshotHost: o.SnapshotHost.ValueString(),
		Port:         o.Port.ValueInt32(),
		User:         o.Username.ValueString(),
		Password:     o.Password.ValueString(),
	}
}

func OracleSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *OracleSharedConfig {
	return &OracleSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		Username: types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
	}
}

type PostgresSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (p PostgresSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:         p.Host.ValueString(),
		SnapshotHost: p.SnapshotHost.ValueString(),
		Port:         p.Port.ValueInt32(),
		User:         p.Username.ValueString(),
		Password:     p.Password.ValueString(),
	}
}

func PostgresSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *PostgresSharedConfig {
	return &PostgresSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		Username: types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
	}
}

type RedshiftSharedConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r RedshiftSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Endpoint: r.Endpoint.ValueString(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func RedshiftSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *RedshiftSharedConfig {
	return &RedshiftSharedConfig{
		Endpoint: types.StringValue(apiModel.Endpoint),
		Username: types.StringValue(apiModel.Username),
		Password: types.StringValue(apiModel.Password),
	}
}

type S3SharedConfig struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Region          types.String `tfsdk:"region"`
}

func (s S3SharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		AWSAccessKeyID:     s.AccessKeyID.ValueString(),
		AWSSecretAccessKey: s.SecretAccessKey.ValueString(),
		AWSRegion:          s.Region.ValueString(),
	}
}

func S3SharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *S3SharedConfig {
	return &S3SharedConfig{
		AccessKeyID:     types.StringValue(apiModel.AWSAccessKeyID),
		SecretAccessKey: types.StringValue(apiModel.AWSSecretAccessKey),
		Region:          types.StringValue(apiModel.AWSRegion),
	}
}

type SnowflakeSharedConfig struct {
	AccountURL types.String `tfsdk:"account_url"`
	VirtualDWH types.String `tfsdk:"virtual_dwh"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

func (s SnowflakeSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		SnowflakeAccountURL: s.AccountURL.ValueString(),
		SnowflakeVirtualDWH: s.VirtualDWH.ValueString(),
		SnowflakePrivateKey: s.PrivateKey.ValueString(),
		Username:            s.Username.ValueString(),
		Password:            s.Password.ValueString(),
	}
}

func SnowflakeSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *SnowflakeSharedConfig {
	return &SnowflakeSharedConfig{
		AccountURL: types.StringValue(apiModel.SnowflakeAccountURL),
		VirtualDWH: types.StringValue(apiModel.SnowflakeVirtualDWH),
		PrivateKey: types.StringValue(apiModel.SnowflakePrivateKey),
		Username:   types.StringValue(apiModel.Username),
		Password:   types.StringValue(apiModel.Password),
	}
}
