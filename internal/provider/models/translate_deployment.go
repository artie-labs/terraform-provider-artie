package models

import (
	"fmt"
	"terraform-provider-artie/internal/artieclient"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentAPIToResourceModel(apiModel artieclient.DeploymentAPIModel, resourceModel *DeploymentResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID)
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Status = types.StringValue(apiModel.Status)
	resourceModel.DestinationUUID = types.StringValue(apiModel.DestinationUUID)

	sshTunnelUUID := ""
	if apiModel.SSHTunnelUUID != nil {
		sshTunnelUUID = *apiModel.SSHTunnelUUID
	}
	resourceModel.SSHTunnelUUID = types.StringValue(sshTunnelUUID)

	snowflakeEcoScheduleUUID := ""
	if apiModel.SnowflakeEcoScheduleUUID != nil {
		snowflakeEcoScheduleUUID = *apiModel.SnowflakeEcoScheduleUUID
	}
	resourceModel.SnowflakeEcoScheduleUUID = types.StringValue(snowflakeEcoScheduleUUID)

	tables := map[string]TableModel{}
	for _, apiTable := range apiModel.Source.Tables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}
		tables[tableKey] = TableModel{
			UUID:                 types.StringValue(apiTable.UUID),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
		}
	}
	resourceModel.Source = &SourceModel{
		Type:   types.StringValue(apiModel.Source.Type),
		Tables: tables,
	}
	switch resourceModel.Source.Type.ValueString() {
	case string(PostgreSQL):
		resourceModel.Source.PostgresConfig = &PostgresConfigModel{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int32Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	case string(MySQL):
		resourceModel.Source.MySQLConfig = &MySQLConfigModel{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int32Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	}

	resourceModel.DestinationConfig = &DeploymentDestinationConfigModel{
		Dataset:               types.StringValue(apiModel.DestinationConfig.Dataset),
		Database:              types.StringValue(apiModel.DestinationConfig.Database),
		Schema:                types.StringValue(apiModel.DestinationConfig.Schema),
		SchemaOverride:        types.StringValue(apiModel.DestinationConfig.SchemaOverride),
		UseSameSchemaAsSource: types.BoolValue(apiModel.DestinationConfig.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.DestinationConfig.SchemaNamePrefix),
	}
}

func DeploymentResourceToAPIModel(resourceModel DeploymentResourceModel) artieclient.DeploymentAPIModel {
	tables := []artieclient.TableAPIModel{}
	for _, table := range resourceModel.Source.Tables {
		tableUUID := table.UUID.ValueString()
		if tableUUID == "" {
			tableUUID = uuid.Nil.String()
		}
		tables = append(tables, artieclient.TableAPIModel{
			UUID:                 tableUUID,
			Name:                 table.Name.ValueString(),
			Schema:               table.Schema.ValueString(),
			EnableHistoryMode:    table.EnableHistoryMode.ValueBool(),
			IndividualDeployment: table.IndividualDeployment.ValueBool(),
			IsPartitioned:        table.IsPartitioned.ValueBool(),
		})
	}

	apiModel := artieclient.DeploymentAPIModel{
		UUID:            resourceModel.UUID.ValueString(),
		Name:            resourceModel.Name.ValueString(),
		Status:          resourceModel.Status.ValueString(),
		DestinationUUID: resourceModel.DestinationUUID.ValueString(),
		Source: artieclient.SourceAPIModel{
			Type:   resourceModel.Source.Type.ValueString(),
			Tables: tables,
		},
		DestinationConfig: artieclient.DestinationConfigAPIModel{
			Dataset:               resourceModel.DestinationConfig.Dataset.ValueString(),
			Database:              resourceModel.DestinationConfig.Database.ValueString(),
			Schema:                resourceModel.DestinationConfig.Schema.ValueString(),
			SchemaOverride:        resourceModel.DestinationConfig.SchemaOverride.ValueString(),
			UseSameSchemaAsSource: resourceModel.DestinationConfig.UseSameSchemaAsSource.ValueBool(),
			SchemaNamePrefix:      resourceModel.DestinationConfig.SchemaNamePrefix.ValueString(),
		},
	}

	sshTunnelUUID := resourceModel.SSHTunnelUUID.ValueString()
	if sshTunnelUUID != "" {
		apiModel.SSHTunnelUUID = &sshTunnelUUID
	}

	snowflakeEcoScheduleUUID := resourceModel.SnowflakeEcoScheduleUUID.ValueString()
	if snowflakeEcoScheduleUUID != "" {
		apiModel.SnowflakeEcoScheduleUUID = &snowflakeEcoScheduleUUID
	}

	switch resourceModel.Source.Type.ValueString() {
	case string(PostgreSQL):
		apiModel.Source.Config = artieclient.SourceConfigAPIModel{
			Host:     resourceModel.Source.PostgresConfig.Host.ValueString(),
			Port:     resourceModel.Source.PostgresConfig.Port.ValueInt32(),
			User:     resourceModel.Source.PostgresConfig.User.ValueString(),
			Password: resourceModel.Source.PostgresConfig.Password.ValueString(),
			Database: resourceModel.Source.PostgresConfig.Database.ValueString(),
		}
	case string(MySQL):
		apiModel.Source.Config = artieclient.SourceConfigAPIModel{
			Host:     resourceModel.Source.MySQLConfig.Host.ValueString(),
			Port:     resourceModel.Source.MySQLConfig.Port.ValueInt32(),
			User:     resourceModel.Source.MySQLConfig.User.ValueString(),
			Password: resourceModel.Source.MySQLConfig.Password.ValueString(),
			Database: resourceModel.Source.MySQLConfig.Database.ValueString(),
		}
	}

	return apiModel
}
