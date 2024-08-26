package models

import (
	"fmt"
	"strings"
	"terraform-provider-artie/internal/artieclient"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentAPIToResourceModel(apiModel artieclient.Deployment, resourceModel *DeploymentResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID.String())
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Status = types.StringValue(apiModel.Status)
	resourceModel.DestinationUUID = optionalUUIDToStringValue(apiModel.DestinationUUID)
	resourceModel.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)
	resourceModel.SnowflakeEcoScheduleUUID = optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID)

	tables := map[string]TableModel{}
	for _, apiTable := range apiModel.Source.Tables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}
		tables[tableKey] = TableModel{
			UUID:                 types.StringValue(apiTable.UUID.String()),
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
	switch strings.ToLower(resourceModel.Source.Type.ValueString()) {
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
		UseSameSchemaAsSource: types.BoolValue(apiModel.DestinationConfig.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.DestinationConfig.SchemaNamePrefix),
	}
}

func DeploymentResourceToBaseAPIModel(resourceModel DeploymentResourceModel) artieclient.BaseDeployment {
	tables := []artieclient.Table{}
	for _, table := range resourceModel.Source.Tables {
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

	baseDeployment := artieclient.BaseDeployment{
		Name:            resourceModel.Name.ValueString(),
		DestinationUUID: ParseOptionalUUID(resourceModel.DestinationUUID),
		Source: artieclient.Source{
			Type:   resourceModel.Source.Type.ValueString(),
			Tables: tables,
		},
		DestinationConfig: artieclient.DestinationConfig{
			Dataset:               resourceModel.DestinationConfig.Dataset.ValueString(),
			Database:              resourceModel.DestinationConfig.Database.ValueString(),
			Schema:                resourceModel.DestinationConfig.Schema.ValueString(),
			UseSameSchemaAsSource: resourceModel.DestinationConfig.UseSameSchemaAsSource.ValueBool(),
			SchemaNamePrefix:      resourceModel.DestinationConfig.SchemaNamePrefix.ValueString(),
		},
		SSHTunnelUUID:            ParseOptionalUUID(resourceModel.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: ParseOptionalUUID(resourceModel.SnowflakeEcoScheduleUUID),
	}

	switch resourceModel.Source.Type.ValueString() {
	case string(PostgreSQL):
		baseDeployment.Source.Config = artieclient.SourceConfig{
			Host:     resourceModel.Source.PostgresConfig.Host.ValueString(),
			Port:     resourceModel.Source.PostgresConfig.Port.ValueInt32(),
			User:     resourceModel.Source.PostgresConfig.User.ValueString(),
			Password: resourceModel.Source.PostgresConfig.Password.ValueString(),
			Database: resourceModel.Source.PostgresConfig.Database.ValueString(),
		}
	case string(MySQL):
		baseDeployment.Source.Config = artieclient.SourceConfig{
			Host:     resourceModel.Source.MySQLConfig.Host.ValueString(),
			Port:     resourceModel.Source.MySQLConfig.Port.ValueInt32(),
			User:     resourceModel.Source.MySQLConfig.User.ValueString(),
			Password: resourceModel.Source.MySQLConfig.Password.ValueString(),
			Database: resourceModel.Source.MySQLConfig.Database.ValueString(),
		}
	}

	return baseDeployment
}

func DeploymentResourceToAPIModel(resourceModel DeploymentResourceModel) artieclient.Deployment {
	return artieclient.Deployment{
		UUID:           parseUUID(resourceModel.UUID),
		BaseDeployment: DeploymentResourceToBaseAPIModel(resourceModel),
	}
}

func BaseDeploymentAPIModelToDeploymentAPIModel(baseDeployment artieclient.BaseDeployment, _uuid uuid.UUID, status string) artieclient.Deployment {
	return artieclient.Deployment{
		UUID:           _uuid,
		Status:         status,
		BaseDeployment: baseDeployment,
	}
}
