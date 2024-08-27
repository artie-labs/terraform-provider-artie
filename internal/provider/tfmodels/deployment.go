package tfmodels

import (
	"fmt"

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

func (d Deployment) ToAPIBaseModel() artieclient.BaseDeployment {
	return artieclient.BaseDeployment{
		Name:                     d.Name.ValueString(),
		DestinationUUID:          ParseOptionalUUID(d.DestinationUUID),
		Source:                   d.Source.ToAPIModel(),
		DestinationConfig:        d.DestinationConfig.ToAPIModel(),
		SSHTunnelUUID:            ParseOptionalUUID(d.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: ParseOptionalUUID(d.SnowflakeEcoScheduleUUID),
	}
}

func (d Deployment) ToAPIModel() artieclient.Deployment {
	return artieclient.Deployment{
		UUID:           parseUUID(d.UUID),
		BaseDeployment: d.ToAPIBaseModel(),
	}
}

type DeploymentDestinationConfig struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
}

func (d DeploymentDestinationConfig) ToAPIModel() artieclient.DestinationConfig {
	return artieclient.DestinationConfig{
		Dataset:               d.Dataset.ValueString(),
		Database:              d.Database.ValueString(),
		Schema:                d.Schema.ValueString(),
		UseSameSchemaAsSource: d.UseSameSchemaAsSource.ValueBool(),
		SchemaNamePrefix:      d.SchemaNamePrefix.ValueString(),
	}
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
	case artieclient.MySQL:
		d.Source.MySQLConfig = &MySQLConfig{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int32Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	case artieclient.PostgreSQL:
		d.Source.PostgresConfig = &PostgresConfig{
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
