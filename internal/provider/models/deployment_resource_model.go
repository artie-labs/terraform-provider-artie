package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SourceType string

const (
	PostgreSQL SourceType = "PostgreSQL"
	MySQL      SourceType = "MySQL"
)

type DeploymentResourceModel struct {
	UUID                     types.String                      `tfsdk:"uuid"`
	Name                     types.String                      `tfsdk:"name"`
	Status                   types.String                      `tfsdk:"status"`
	DestinationUUID          types.String                      `tfsdk:"destination_uuid"`
	SSHTunnelUUID            types.String                      `tfsdk:"ssh_tunnel_uuid"`
	SnowflakeEcoScheduleUUID types.String                      `tfsdk:"snowflake_eco_schedule_uuid"`
	Source                   *SourceModel                      `tfsdk:"source"`
	DestinationConfig        *DeploymentDestinationConfigModel `tfsdk:"destination_config"`
}

type SourceModel struct {
	Type           types.String         `tfsdk:"type"`
	Tables         []TableModel         `tfsdk:"tables"`
	PostgresConfig *PostgresConfigModel `tfsdk:"postgres_config"`
	MySQLConfig    *MySQLConfigModel    `tfsdk:"mysql_config"`
}

type PostgresConfigModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

type MySQLConfigModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

type TableModel struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`
}

type DeploymentDestinationConfigModel struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	SchemaOverride        types.String `tfsdk:"schema_override"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
}
