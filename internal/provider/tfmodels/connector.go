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
	Label           types.String           `tfsdk:"label"`
	BigQueryConfig  *BigQuerySharedConfig  `tfsdk:"bigquery_config"`
	MSSQLConfig     *MSSQLSharedConfig     `tfsdk:"mssql_config"`
	RedshiftConfig  *RedshiftSharedConfig  `tfsdk:"redshift_config"`
	S3Config        *S3SharedConfig        `tfsdk:"s3_config"`
	SnowflakeConfig *SnowflakeSharedConfig `tfsdk:"snowflake_config"`
}

func (c Connector) ToAPIBaseModel() (artieclient.BaseConnector, diag.Diagnostics) {
	var sharedConfig artieclient.DestinationSharedConfig
	destinationType, err := artieclient.ConnectorTypeFromString(c.Type.ValueString())
	if err != nil {
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", err.Error(),
		)}
	}

	switch destinationType {
	case artieclient.BigQuery:
		sharedConfig = c.BigQueryConfig.ToAPIModel()
	case artieclient.MSSQL:
		sharedConfig = c.MSSQLConfig.ToAPIModel()
	case artieclient.Redshift:
		sharedConfig = c.RedshiftConfig.ToAPIModel()
	case artieclient.S3:
		sharedConfig = c.S3Config.ToAPIModel()
	case artieclient.Snowflake:
		sharedConfig = c.SnowflakeConfig.ToAPIModel()
	default:
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", fmt.Sprintf("unhandled destination type: %s", c.Type.ValueString()),
		)}
	}

	sshTunnelUUID, diags := parseOptionalUUID(c.SSHTunnelUUID)
	if diags.HasError() {
		return artieclient.BaseConnector{}, diags
	}

	return artieclient.BaseConnector{
		Type:          destinationType,
		Label:         c.Label.ValueString(),
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

func DestinationFromAPIModel(apiModel artieclient.Connector) (Connector, diag.Diagnostics) {
	destination := Connector{
		UUID:          types.StringValue(apiModel.UUID.String()),
		Type:          types.StringValue(string(apiModel.Type)),
		Label:         types.StringValue(apiModel.Label),
		SSHTunnelUUID: optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
	}

	switch apiModel.Type {
	case artieclient.BigQuery:
		destination.BigQueryConfig = BigQuerySharedConfigFromAPIModel(apiModel.Config)
	case artieclient.MSSQL:
		destination.MSSQLConfig = MSSQLSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Redshift:
		destination.RedshiftConfig = RedshiftSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.S3:
		destination.S3Config = S3SharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Snowflake:
		destination.SnowflakeConfig = SnowflakeSharedConfigFromAPIModel(apiModel.Config)
	default:
		return Connector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert API model to Destination", fmt.Sprintf("invalid destination type: %s", apiModel.Type),
		)}
	}

	return destination, nil
}

type BigQuerySharedConfig struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

func (b BigQuerySharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		GCPProjectID:       b.ProjectID.ValueString(),
		GCPLocation:        b.Location.ValueString(),
		GCPCredentialsData: b.CredentialsData.ValueString(),
	}
}

func BigQuerySharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *BigQuerySharedConfig {
	return &BigQuerySharedConfig{
		ProjectID:       types.StringValue(apiModel.GCPProjectID),
		Location:        types.StringValue(apiModel.GCPLocation),
		CredentialsData: types.StringValue(apiModel.GCPCredentialsData),
	}
}

type MSSQLSharedConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r MSSQLSharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		Host:     r.Host.ValueString(),
		Port:     r.Port.ValueInt32(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func MSSQLSharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *MSSQLSharedConfig {
	return &MSSQLSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		Username: types.StringValue(apiModel.Username),
		Password: types.StringValue(apiModel.Password),
	}
}

type RedshiftSharedConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r RedshiftSharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		Endpoint: r.Endpoint.ValueString(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func RedshiftSharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *RedshiftSharedConfig {
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

func (s S3SharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		AWSAccessKeyID:     s.AccessKeyID.ValueString(),
		AWSSecretAccessKey: s.SecretAccessKey.ValueString(),
		AWSRegion:          s.Region.ValueString(),
	}
}

func S3SharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *S3SharedConfig {
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

func (s SnowflakeSharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		SnowflakeAccountURL: s.AccountURL.ValueString(),
		SnowflakeVirtualDWH: s.VirtualDWH.ValueString(),
		SnowflakePrivateKey: s.PrivateKey.ValueString(),
		Username:            s.Username.ValueString(),
		Password:            s.Password.ValueString(),
	}
}

func SnowflakeSharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *SnowflakeSharedConfig {
	return &SnowflakeSharedConfig{
		AccountURL: types.StringValue(apiModel.SnowflakeAccountURL),
		VirtualDWH: types.StringValue(apiModel.SnowflakeVirtualDWH),
		PrivateKey: types.StringValue(apiModel.SnowflakePrivateKey),
		Username:   types.StringValue(apiModel.Username),
		Password:   types.StringValue(apiModel.Password),
	}
}
