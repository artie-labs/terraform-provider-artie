package tfmodels

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Deployment struct {
	UUID                     types.String                 `tfsdk:"uuid"`
	Name                     types.String                 `tfsdk:"name"`
	Status                   types.String                 `tfsdk:"status"`
	Source                   *Source                      `tfsdk:"source"`
	DestinationUUID          types.String                 `tfsdk:"destination_uuid"`
	DestinationConfig        *DeploymentDestinationConfig `tfsdk:"destination_config"`
	SSHTunnelUUID            types.String                 `tfsdk:"ssh_tunnel_uuid"`
	SnowflakeEcoScheduleUUID types.String                 `tfsdk:"snowflake_eco_schedule_uuid"`
}

type Source struct {
	Type           types.String     `tfsdk:"type"`
	Tables         map[string]Table `tfsdk:"tables"`
	PostgresConfig *PostgresConfig  `tfsdk:"postgresql_config"`
	MySQLConfig    *MySQLConfig     `tfsdk:"mysql_config"`
}

type PostgresConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

type MySQLConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`
}

type DeploymentDestinationConfig struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
}

func (d *Deployment) UpdateFromAPIModel(apiModel artieclient.Deployment) {
	d.UUID = types.StringValue(apiModel.UUID.String())
	d.Name = types.StringValue(apiModel.Name)
	d.Status = types.StringValue(apiModel.Status)
	d.DestinationUUID = optionalUUIDToStringValue(apiModel.DestinationUUID)
	d.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)
	d.SnowflakeEcoScheduleUUID = optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID)

	tables := map[string]Table{}
	for _, apiTable := range apiModel.Source.Tables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}
		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
		}
	}
	d.Source = &Source{
		Type:   types.StringValue(string(apiModel.Source.Type)),
		Tables: tables,
	}
	switch apiModel.Source.Type {
	case artieclient.PostgreSQL:
		d.Source.PostgresConfig = &PostgresConfig{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int32Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	case artieclient.MySQL:
		d.Source.MySQLConfig = &MySQLConfig{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int32Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	default:
		panic(fmt.Sprintf("invalid source type: %s", apiModel.Source.Type))
	}

	d.DestinationConfig = &DeploymentDestinationConfig{
		Dataset:               types.StringValue(apiModel.DestinationConfig.Dataset),
		Database:              types.StringValue(apiModel.DestinationConfig.Database),
		Schema:                types.StringValue(apiModel.DestinationConfig.Schema),
		UseSameSchemaAsSource: types.BoolValue(apiModel.DestinationConfig.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.DestinationConfig.SchemaNamePrefix),
	}
}

func (d Deployment) ToAPIBaseModel() artieclient.BaseDeployment {
	tables := []artieclient.Table{}
	for _, table := range d.Source.Tables {
		tableUUID := table.UUID.ValueString()
		if tableUUID == "" {
			tableUUID = uuid.Nil.String()
		}
		tables = append(tables, artieclient.Table{
			UUID:                 uuid.MustParse(tableUUID),
			Name:                 table.Name.ValueString(),
			Schema:               table.Schema.ValueString(),
			EnableHistoryMode:    table.EnableHistoryMode.ValueBool(),
			IndividualDeployment: table.IndividualDeployment.ValueBool(),
			IsPartitioned:        table.IsPartitioned.ValueBool(),
		})
	}

	sourceType := artieclient.SourceTypeFromString(d.Source.Type.ValueString())
	baseDeployment := artieclient.BaseDeployment{
		Name:            d.Name.ValueString(),
		DestinationUUID: ParseOptionalUUID(d.DestinationUUID),
		Source: artieclient.Source{
			Type:   sourceType,
			Tables: tables,
		},
		DestinationConfig: artieclient.DestinationConfig{
			Dataset:               d.DestinationConfig.Dataset.ValueString(),
			Database:              d.DestinationConfig.Database.ValueString(),
			Schema:                d.DestinationConfig.Schema.ValueString(),
			UseSameSchemaAsSource: d.DestinationConfig.UseSameSchemaAsSource.ValueBool(),
			SchemaNamePrefix:      d.DestinationConfig.SchemaNamePrefix.ValueString(),
		},
		SSHTunnelUUID:            ParseOptionalUUID(d.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: ParseOptionalUUID(d.SnowflakeEcoScheduleUUID),
	}

	switch sourceType {
	case artieclient.PostgreSQL:
		baseDeployment.Source.Config = artieclient.SourceConfig{
			Host:     d.Source.PostgresConfig.Host.ValueString(),
			Port:     d.Source.PostgresConfig.Port.ValueInt32(),
			User:     d.Source.PostgresConfig.User.ValueString(),
			Password: d.Source.PostgresConfig.Password.ValueString(),
			Database: d.Source.PostgresConfig.Database.ValueString(),
		}
	case artieclient.MySQL:
		baseDeployment.Source.Config = artieclient.SourceConfig{
			Host:     d.Source.MySQLConfig.Host.ValueString(),
			Port:     d.Source.MySQLConfig.Port.ValueInt32(),
			User:     d.Source.MySQLConfig.User.ValueString(),
			Password: d.Source.MySQLConfig.Password.ValueString(),
			Database: d.Source.MySQLConfig.Database.ValueString(),
		}
	default:
		panic(fmt.Sprintf("invalid source type: %s", d.Source.Type.ValueString()))
	}

	return baseDeployment
}

func (d Deployment) ToAPIModel() artieclient.Deployment {
	return artieclient.Deployment{
		UUID:           parseUUID(d.UUID),
		BaseDeployment: d.ToAPIBaseModel(),
	}
}
