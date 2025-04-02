package tfmodels

import (
	"fmt"
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Destination struct {
	UUID            types.String           `tfsdk:"uuid"`
	SSHTunnelUUID   types.String           `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String           `tfsdk:"type"`
	Label           types.String           `tfsdk:"label"`
	DataPlaneName   types.String           `tfsdk:"data_plane_name"`
	BigQueryConfig  *BigQuerySharedConfig  `tfsdk:"bigquery_config"`
	MSSQLConfig     *MSSQLDestSharedConfig `tfsdk:"mssql_config"`
	RedshiftConfig  *RedshiftSharedConfig  `tfsdk:"redshift_config"`
	S3Config        *S3SharedConfig        `tfsdk:"s3_config"`
	SnowflakeConfig *SnowflakeSharedConfig `tfsdk:"snowflake_config"`
}

func (d Destination) ToAPIBaseModel() (artieclient.BaseConnector, diag.Diagnostics) {
	var sharedConfig artieclient.ConnectorConfig
	destinationType, err := artieclient.ConnectorTypeFromString(d.Type.ValueString())
	if err != nil {
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", err.Error(),
		)}
	}

	switch destinationType {
	case artieclient.BigQuery:
		sharedConfig = d.BigQueryConfig.ToAPIModel()
	case artieclient.MSSQL:
		sharedConfig = d.MSSQLConfig.ToAPIModel()
	case artieclient.Redshift:
		sharedConfig = d.RedshiftConfig.ToAPIModel()
	case artieclient.S3:
		sharedConfig = d.S3Config.ToAPIModel()
	case artieclient.Snowflake:
		sharedConfig = d.SnowflakeConfig.ToAPIModel()
	default:
		return artieclient.BaseConnector{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", fmt.Sprintf("unhandled destination type: %s", d.Type.ValueString()),
		)}
	}

	sshTunnelUUID, diags := parseOptionalUUID(d.SSHTunnelUUID)
	if diags.HasError() {
		return artieclient.BaseConnector{}, diags
	}

	return artieclient.BaseConnector{
		Type:          destinationType,
		DataPlaneName: d.DataPlaneName.ValueString(),
		Label:         d.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: sshTunnelUUID,
	}, diags
}

func (d Destination) ToAPIModel() (artieclient.Connector, diag.Diagnostics) {
	baseModel, diags := d.ToAPIBaseModel()
	if diags.HasError() {
		return artieclient.Connector{}, diags
	}

	uuid, uuidDiags := parseUUID(d.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Connector{}, diags
	}

	return artieclient.Connector{
		UUID:          uuid,
		BaseConnector: baseModel,
	}, diags
}

func DestinationFromAPIModel(apiModel artieclient.Connector) (Destination, diag.Diagnostics) {
	destination := Destination{
		UUID:          types.StringValue(apiModel.UUID.String()),
		Type:          types.StringValue(string(apiModel.Type)),
		DataPlaneName: types.StringValue(apiModel.DataPlaneName),
		Label:         types.StringValue(apiModel.Label),
		SSHTunnelUUID: optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
	}

	switch apiModel.Type {
	case artieclient.BigQuery:
		destination.BigQueryConfig = BigQuerySharedConfigFromAPIModel(apiModel.Config)
	case artieclient.MSSQL:
		destination.MSSQLConfig = MSSQLDestSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Redshift:
		destination.RedshiftConfig = RedshiftSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.S3:
		destination.S3Config = S3SharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Snowflake:
		destination.SnowflakeConfig = SnowflakeSharedConfigFromAPIModel(apiModel.Config)
	default:
		return Destination{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert API model to Destination", fmt.Sprintf("invalid destination type: %s", apiModel.Type),
		)}
	}

	return destination, nil
}

type MSSQLDestSharedConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r MSSQLDestSharedConfig) ToAPIModel() artieclient.ConnectorConfig {
	return artieclient.ConnectorConfig{
		Host:     r.Host.ValueString(),
		Port:     r.Port.ValueInt32(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func MSSQLDestSharedConfigFromAPIModel(apiModel artieclient.ConnectorConfig) *MSSQLDestSharedConfig {
	return &MSSQLDestSharedConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		Username: types.StringValue(apiModel.Username),
		Password: types.StringValue(apiModel.Password),
	}
}
