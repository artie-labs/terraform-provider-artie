package tfmodels

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type Connector struct {
	UUID              types.String             `tfsdk:"uuid"`
	SSHTunnelUUID     types.String             `tfsdk:"ssh_tunnel_uuid"`
	Type              types.String             `tfsdk:"type"`
	Name              types.String             `tfsdk:"name"`
	DataPlaneName     types.String             `tfsdk:"data_plane_name"`
	BigQueryConfig    *BigQuerySharedConfig    `tfsdk:"bigquery_config"`
	CockroachDBConfig *CockroachDBSharedConfig `tfsdk:"cockroach_config"`
	DynamoDBConfig    *DynamoDBConfig          `tfsdk:"dynamodb_config"`
	GCSConfig         *GCSSharedConfig         `tfsdk:"gcs_config"`
	IcebergConfig     *IcebergSharedConfig     `tfsdk:"iceberg_config"`
	MongoDBConfig     *MongoDBSharedConfig     `tfsdk:"mongodb_config"`
	MySQLConfig       *MySQLSharedConfig       `tfsdk:"mysql_config"`
	MSSQLConfig       *MSSQLSharedConfig       `tfsdk:"mssql_config"`
	OracleConfig      *OracleSharedConfig      `tfsdk:"oracle_config"`
	PostgresConfig    *PostgresSharedConfig    `tfsdk:"postgresql_config"`
	RedshiftConfig    *RedshiftSharedConfig    `tfsdk:"redshift_config"`
	S3Config          *S3SharedConfig          `tfsdk:"s3_config"`
	SnowflakeConfig   *SnowflakeSharedConfig   `tfsdk:"snowflake_config"`
	DatabricksConfig  *DatabricksSharedConfig  `tfsdk:"databricks_config"`
	KeyspacesConfig   *KeyspacesSharedConfig   `tfsdk:"keyspaces_config"`
}

func (c Connector) ToAPIModel() (openapi.PayloadsConnectorPayload, diag.Diagnostics) {
	var sharedConfig ConnectorConfig
	switch openapi.EnumsConnectorSlug(c.Type.ValueString()) {
	case openapi.EnumsConnectorSlugApi:
		// No config needed
	case openapi.EnumsConnectorSlugBigquery:
		sharedConfig = c.BigQueryConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugCockroach:
		sharedConfig = c.CockroachDBConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugDynamodb:
		sharedConfig = c.DynamoDBConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugGcs:
		sharedConfig = c.GCSConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugIceberg:
		sharedConfig = c.IcebergConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugMongodb:
		sharedConfig = c.MongoDBConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugMysql:
		sharedConfig = c.MySQLConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugMssql:
		sharedConfig = c.MSSQLConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugOracle:
		sharedConfig = c.OracleConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugPostgresql:
		sharedConfig = c.PostgresConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugRedshift:
		sharedConfig = c.RedshiftConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugS3:
		sharedConfig = c.S3Config.ToAPIModel()
	case openapi.EnumsConnectorSlugSnowflake:
		sharedConfig = c.SnowflakeConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugDatabricks:
		sharedConfig = c.DatabricksConfig.ToAPIModel()
	case openapi.EnumsConnectorSlugKeyspaces:
		sharedConfig = c.KeyspacesConfig.ToAPIModel()
	default:
		return openapi.PayloadsConnectorPayload{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Connector to API model", fmt.Sprintf("unhandled connector type: %s", c.Type.ValueString()),
		)}
	}

	apiSharedConfig, err := sharedConfig.toAPIModel()
	if err != nil {
		return openapi.PayloadsConnectorPayload{}, []diag.Diagnostic{diag.NewErrorDiagnostic("Unable to convert Connector to API model", err.Error())}
	}

	sshTunnelUUID, diags := parseOptionalUUID(c.SSHTunnelUUID)
	if diags.HasError() {
		return openapi.PayloadsConnectorPayload{}, diags
	}

	return openapi.PayloadsConnectorPayload{
		// UUID is unknown on create; on update the API uses it to resolve masked sensitive values.
		Uuid:          nonEmptyStringPointer(c.UUID),
		Type:          c.Type.ValueStringPointer(),
		DataPlaneName: nonEmptyStringPointer(c.DataPlaneName),
		Label:         c.Name.ValueStringPointer(),
		SharedConfig:  apiSharedConfig,
		SshTunnelUUID: sshTunnelUUID,
	}, diags
}

func ConnectorFromAPIModel(apiModel openapi.PayloadsFullConnector) (Connector, diag.Diagnostics) {
	config, err := connectorConfigFromAPIModel(apiModel.SharedConfig)
	if err != nil {
		return Connector{}, []diag.Diagnostic{diag.NewErrorDiagnostic("Unable to convert API model to Connector", err.Error())}
	}

	connector := Connector{
		UUID:          types.StringValue(apiModel.Uuid.String()),
		Type:          types.StringValue(string(apiModel.Type)),
		DataPlaneName: types.StringValue(lib.RemovePtr(apiModel.DataPlaneName)),
		Name:          types.StringValue(apiModel.Label),
		SSHTunnelUUID: optionalUUIDToStringValue(apiModel.SshTunnelUUID),
	}

	switch apiModel.Type {
	case openapi.EnumsConnectorSlugApi:
		// No config needed
	case openapi.EnumsConnectorSlugBigquery:
		connector.BigQueryConfig = BigQuerySharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugCockroach:
		connector.CockroachDBConfig = CockroachDBSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugDynamodb:
		connector.DynamoDBConfig = DynamoDBConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugGcs:
		connector.GCSConfig = GCSSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugIceberg:
		connector.IcebergConfig = IcebergSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugMongodb:
		connector.MongoDBConfig = MongoDBSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugMysql:
		connector.MySQLConfig = MySQLSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugMssql:
		connector.MSSQLConfig = MSSQLSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugOracle:
		connector.OracleConfig = OracleSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugPostgresql:
		connector.PostgresConfig = PostgresSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugRedshift:
		connector.RedshiftConfig = RedshiftSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugS3:
		connector.S3Config = S3SharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugSnowflake:
		connector.SnowflakeConfig = SnowflakeSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugDatabricks:
		connector.DatabricksConfig = DatabricksSharedConfigFromAPIModel(config)
	case openapi.EnumsConnectorSlugKeyspaces:
		connector.KeyspacesConfig = KeyspacesSharedConfigFromAPIModel(config)
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

func (b BigQuerySharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		GCPProjectID:       b.ProjectID.ValueString(),
		GCPLocation:        b.Location.ValueString(),
		GCPCredentialsData: b.CredentialsData.ValueString(),
	}
}

func BigQuerySharedConfigFromAPIModel(apiModel ConnectorConfig) *BigQuerySharedConfig {
	return &BigQuerySharedConfig{
		ProjectID:       types.StringValue(apiModel.GCPProjectID),
		Location:        types.StringValue(apiModel.GCPLocation),
		CredentialsData: types.StringValue(apiModel.GCPCredentialsData),
	}
}

type DynamoDBConfig struct {
	StreamArn          types.String `tfsdk:"stream_arn"`
	AwsAccessKeyID     types.String `tfsdk:"access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"secret_access_key"`
	RoleARN            types.String `tfsdk:"role_arn"`
	ExternalID         types.String `tfsdk:"external_id"`
}

func (d DynamoDBConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		DynamoStreamArn:    d.StreamArn.ValueString(),
		AWSAccessKeyID:     d.AwsAccessKeyID.ValueString(),
		AWSSecretAccessKey: d.AwsSecretAccessKey.ValueString(),
		AWSRoleARN:         d.RoleARN.ValueString(),
		AWSExternalID:      d.ExternalID.ValueString(),
	}
}

func DynamoDBConfigFromAPIModel(apiDynamoCfg ConnectorConfig) *DynamoDBConfig {
	return &DynamoDBConfig{
		StreamArn:          types.StringValue(apiDynamoCfg.DynamoStreamArn),
		AwsAccessKeyID:     types.StringValue(apiDynamoCfg.AWSAccessKeyID),
		AwsSecretAccessKey: types.StringValue(apiDynamoCfg.AWSSecretAccessKey),
		RoleARN:            types.StringValue(apiDynamoCfg.AWSRoleARN),
		ExternalID:         types.StringValue(apiDynamoCfg.AWSExternalID),
	}
}

type CockroachDBSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	SnapshotPort types.Int32  `tfsdk:"snapshot_port"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (c CockroachDBSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:         c.Host.ValueString(),
		SnapshotHost: c.SnapshotHost.ValueString(),
		SnapshotPort: c.SnapshotPort.ValueInt32(),
		Port:         c.Port.ValueInt32(),
		User:         c.Username.ValueString(),
		Password:     c.Password.ValueString(),
	}
}

func CockroachDBSharedConfigFromAPIModel(apiModel ConnectorConfig) *CockroachDBSharedConfig {
	return &CockroachDBSharedConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		SnapshotPort: types.Int32Value(apiModel.SnapshotPort),
		Port:         types.Int32Value(apiModel.Port),
		Username:     types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
	}
}

type MongoDBSharedConfig struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (m MongoDBSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:     m.Host.ValueString(),
		User:     m.Username.ValueString(),
		Password: m.Password.ValueString(),
	}
}

func MongoDBSharedConfigFromAPIModel(apiModel ConnectorConfig) *MongoDBSharedConfig {
	return &MongoDBSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Username: types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
	}
}

type MySQLSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	SnapshotPort types.Int32  `tfsdk:"snapshot_port"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	TLSMode      types.String `tfsdk:"tls_mode"`
}

func (m MySQLSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:         m.Host.ValueString(),
		SnapshotHost: m.SnapshotHost.ValueString(),
		SnapshotPort: m.SnapshotPort.ValueInt32(),
		Port:         m.Port.ValueInt32(),
		User:         m.Username.ValueString(),
		Password:     m.Password.ValueString(),
		MySQLTLSMode: m.TLSMode.ValueString(),
	}
}

func MySQLSharedConfigFromAPIModel(apiModel ConnectorConfig) *MySQLSharedConfig {
	return &MySQLSharedConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		SnapshotPort: types.Int32Value(apiModel.SnapshotPort),
		Port:         types.Int32Value(apiModel.Port),
		Username:     types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
		TLSMode:      types.StringValue(apiModel.MySQLTLSMode),
	}
}

type MSSQLSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (r MSSQLSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:         r.Host.ValueString(),
		SnapshotHost: r.SnapshotHost.ValueString(),
		Port:         r.Port.ValueInt32(),
		Username:     r.Username.ValueString(),
		Password:     r.Password.ValueString(),
	}
}

func MSSQLSharedConfigFromAPIModel(apiModel ConnectorConfig) *MSSQLSharedConfig {
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

func (o OracleSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:         o.Host.ValueString(),
		SnapshotHost: o.SnapshotHost.ValueString(),
		Port:         o.Port.ValueInt32(),
		User:         o.Username.ValueString(),
		Password:     o.Password.ValueString(),
	}
}

func OracleSharedConfigFromAPIModel(apiModel ConnectorConfig) *OracleSharedConfig {
	return &OracleSharedConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		Username:     types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
	}
}

type PostgresSharedConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (p PostgresSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:         p.Host.ValueString(),
		SnapshotHost: p.SnapshotHost.ValueString(),
		Port:         p.Port.ValueInt32(),
		User:         p.Username.ValueString(),
		Password:     p.Password.ValueString(),
	}
}

func PostgresSharedConfigFromAPIModel(apiModel ConnectorConfig) *PostgresSharedConfig {
	return &PostgresSharedConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		Username:     types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
	}
}

type RedshiftSharedConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r RedshiftSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Endpoint: r.Endpoint.ValueString(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func RedshiftSharedConfigFromAPIModel(apiModel ConnectorConfig) *RedshiftSharedConfig {
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
	RoleARN         types.String `tfsdk:"role_arn"`
	ExternalID      types.String `tfsdk:"external_id"`
}

func (s S3SharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		AWSAccessKeyID:     s.AccessKeyID.ValueString(),
		AWSSecretAccessKey: s.SecretAccessKey.ValueString(),
		AWSRegion:          s.Region.ValueString(),
		AWSRoleARN:         s.RoleARN.ValueString(),
		AWSExternalID:      s.ExternalID.ValueString(),
	}
}

func S3SharedConfigFromAPIModel(apiModel ConnectorConfig) *S3SharedConfig {
	return &S3SharedConfig{
		AccessKeyID:     types.StringValue(apiModel.AWSAccessKeyID),
		SecretAccessKey: types.StringValue(apiModel.AWSSecretAccessKey),
		Region:          types.StringValue(apiModel.AWSRegion),
		RoleARN:         types.StringValue(apiModel.AWSRoleARN),
		ExternalID:      types.StringValue(apiModel.AWSExternalID),
	}
}

type SnowflakeSharedConfig struct {
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	AccountURL        types.String `tfsdk:"account_url"`
	VirtualDWH        types.String `tfsdk:"virtual_dwh"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	PrivateKey        types.String `tfsdk:"private_key"`
}

func (s SnowflakeSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		SnowflakeAccountIdentifier: s.AccountIdentifier.ValueString(),
		SnowflakeAccountURL:        s.AccountURL.ValueString(),
		SnowflakeVirtualDWH:        s.VirtualDWH.ValueString(),
		SnowflakePrivateKey:        s.PrivateKey.ValueString(),
		Username:                   s.Username.ValueString(),
		Password:                   s.Password.ValueString(),
	}
}

func SnowflakeSharedConfigFromAPIModel(apiModel ConnectorConfig) *SnowflakeSharedConfig {
	return &SnowflakeSharedConfig{
		AccountIdentifier: types.StringValue(apiModel.SnowflakeAccountIdentifier),
		AccountURL:        types.StringValue(apiModel.SnowflakeAccountURL),
		VirtualDWH:        types.StringValue(apiModel.SnowflakeVirtualDWH),
		PrivateKey:        types.StringValue(apiModel.SnowflakePrivateKey),
		Username:          types.StringValue(apiModel.Username),
		Password:          types.StringValue(apiModel.Password),
	}
}

type DatabricksSharedConfig struct {
	Host                types.String `tfsdk:"host"`
	HttpPath            types.String `tfsdk:"http_path"`
	PersonalAccessToken types.String `tfsdk:"personal_access_token"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	Volume              types.String `tfsdk:"volume"`
}

func (d DatabricksSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:                          d.Host.ValueString(),
		DatabricksHttpPath:            d.HttpPath.ValueString(),
		DatabricksPersonalAccessToken: d.PersonalAccessToken.ValueString(),
		DatabricksClientID:            d.ClientID.ValueString(),
		DatabricksClientSecret:        d.ClientSecret.ValueString(),
		DatabricksVolume:              d.Volume.ValueString(),
	}
}

func DatabricksSharedConfigFromAPIModel(apiModel ConnectorConfig) *DatabricksSharedConfig {
	return &DatabricksSharedConfig{
		Host:                types.StringValue(apiModel.Host),
		HttpPath:            types.StringValue(apiModel.DatabricksHttpPath),
		PersonalAccessToken: types.StringValue(apiModel.DatabricksPersonalAccessToken),
		ClientID:            types.StringValue(apiModel.DatabricksClientID),
		ClientSecret:        types.StringValue(apiModel.DatabricksClientSecret),
		Volume:              types.StringValue(apiModel.DatabricksVolume),
	}
}

type GCSSharedConfig struct {
	ProjectID       types.String `tfsdk:"project_id"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

func (g GCSSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		GCPProjectID:       g.ProjectID.ValueString(),
		GCPCredentialsData: g.CredentialsData.ValueString(),
	}
}

func GCSSharedConfigFromAPIModel(apiModel ConnectorConfig) *GCSSharedConfig {
	return &GCSSharedConfig{
		ProjectID:       types.StringValue(apiModel.GCPProjectID),
		CredentialsData: types.StringValue(apiModel.GCPCredentialsData),
	}
}

type IcebergSharedConfig struct {
	Provider types.String `tfsdk:"provider"`

	// S3 Tables fields:
	AwsAccessKeyID     types.String `tfsdk:"access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"secret_access_key"`
	BucketARN          types.String `tfsdk:"bucket_arn"`
	Region             types.String `tfsdk:"region"`

	// REST Catalog fields:
	URI        types.String `tfsdk:"uri"`
	Token      types.String `tfsdk:"token"`
	Credential types.String `tfsdk:"credential"`
	AuthURI    types.String `tfsdk:"auth_uri"`
	Scope      types.String `tfsdk:"scope"`
	Warehouse  types.String `tfsdk:"warehouse"`
	Prefix     types.String `tfsdk:"prefix"`
}

func (i IcebergSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		IcebergProvider:    i.Provider.ValueString(),
		AWSAccessKeyID:     i.AwsAccessKeyID.ValueString(),
		AWSSecretAccessKey: i.AwsSecretAccessKey.ValueString(),
		IcebergBucketARN:   i.BucketARN.ValueString(),
		IcebergRegion:      i.Region.ValueString(),
		IcebergURI:         i.URI.ValueString(),
		IcebergToken:       i.Token.ValueString(),
		IcebergCredential:  i.Credential.ValueString(),
		IcebergAuthURI:     i.AuthURI.ValueString(),
		IcebergScope:       i.Scope.ValueString(),
		IcebergWarehouse:   i.Warehouse.ValueString(),
		IcebergPrefix:      i.Prefix.ValueString(),
	}
}

func IcebergSharedConfigFromAPIModel(apiModel ConnectorConfig) *IcebergSharedConfig {
	return &IcebergSharedConfig{
		Provider:           types.StringValue(apiModel.IcebergProvider),
		AwsAccessKeyID:     types.StringValue(apiModel.AWSAccessKeyID),
		AwsSecretAccessKey: types.StringValue(apiModel.AWSSecretAccessKey),
		BucketARN:          types.StringValue(apiModel.IcebergBucketARN),
		Region:             types.StringValue(apiModel.IcebergRegion),
		URI:                types.StringValue(apiModel.IcebergURI),
		Token:              types.StringValue(apiModel.IcebergToken),
		Credential:         types.StringValue(apiModel.IcebergCredential),
		AuthURI:            types.StringValue(apiModel.IcebergAuthURI),
		Scope:              types.StringValue(apiModel.IcebergScope),
		Warehouse:          types.StringValue(apiModel.IcebergWarehouse),
		Prefix:             types.StringValue(apiModel.IcebergPrefix),
	}
}

type KeyspacesSharedConfig struct {
	Host               types.String `tfsdk:"host"`
	Port               types.Int32  `tfsdk:"port"`
	Region             types.String `tfsdk:"region"`
	AwsAccessKeyID     types.String `tfsdk:"access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"secret_access_key"`
	RoleARN            types.String `tfsdk:"role_arn"`
	ExternalID         types.String `tfsdk:"external_id"`
}

func (k KeyspacesSharedConfig) ToAPIModel() ConnectorConfig {
	return ConnectorConfig{
		Host:               k.Host.ValueString(),
		Port:               k.Port.ValueInt32(),
		AWSRegion:          k.Region.ValueString(),
		AWSAccessKeyID:     k.AwsAccessKeyID.ValueString(),
		AWSSecretAccessKey: k.AwsSecretAccessKey.ValueString(),
		AWSRoleARN:         k.RoleARN.ValueString(),
		AWSExternalID:      k.ExternalID.ValueString(),
	}
}

func KeyspacesSharedConfigFromAPIModel(apiModel ConnectorConfig) *KeyspacesSharedConfig {
	return &KeyspacesSharedConfig{
		Host:               types.StringValue(apiModel.Host),
		Port:               types.Int32Value(apiModel.Port),
		Region:             types.StringValue(apiModel.AWSRegion),
		AwsAccessKeyID:     types.StringValue(apiModel.AWSAccessKeyID),
		AwsSecretAccessKey: types.StringValue(apiModel.AWSSecretAccessKey),
		RoleARN:            types.StringValue(apiModel.AWSRoleARN),
		ExternalID:         types.StringValue(apiModel.AWSExternalID),
	}
}
