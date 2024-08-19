package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type DestinationType string

const (
	Snowflake DestinationType = "snowflake"
	BigQuery  DestinationType = "bigquery"
	Redshift  DestinationType = "redshift"
)

type DestinationResourceModel struct {
	UUID            types.String                `tfsdk:"uuid"`
	SSHTunnelUUID   types.String                `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String                `tfsdk:"type"`
	Label           types.String                `tfsdk:"label"`
	SnowflakeConfig *SnowflakeSharedConfigModel `tfsdk:"snowflake_config"`
	BigQueryConfig  *BigQuerySharedConfigModel  `tfsdk:"bigquery_config"`
	RedshiftConfig  *RedshiftSharedConfigModel  `tfsdk:"redshift_config"`
}

type SnowflakeSharedConfigModel struct {
	AccountURL types.String `tfsdk:"account_url"`
	VirtualDWH types.String `tfsdk:"virtual_dwh"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

type BigQuerySharedConfigModel struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

type RedshiftSharedConfigModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}
