package models

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeploymentAPIToResourceModel(apiModel DeploymentAPIModel, resourceModel *DeploymentResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID)
	resourceModel.CompanyUUID = types.StringValue(apiModel.CompanyUUID)
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Status = types.StringValue(apiModel.Status)
	resourceModel.DestinationUUID = types.StringValue(apiModel.DestinationUUID)

	tables := []TableModel{}
	for _, apiTable := range apiModel.Source.Tables {
		tables = append(tables, TableModel{
			UUID:                 types.StringValue(apiTable.UUID),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
		})
	}
	resourceModel.Source = &SourceModel{
		Type:   types.StringValue(apiModel.Source.Type),
		Tables: tables,
	}
	switch resourceModel.Source.Type.ValueString() {
	case "PostgreSQL":
		resourceModel.Source.PostgresConfig = &PostgresConfigModel{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int64Value(apiModel.Source.Config.Port),
			User:     types.StringValue(apiModel.Source.Config.User),
			Password: types.StringValue(apiModel.Source.Config.Password),
			Database: types.StringValue(apiModel.Source.Config.Database),
		}
	case "MySQL":
		resourceModel.Source.MySQLConfig = &MySQLConfigModel{
			Host:     types.StringValue(apiModel.Source.Config.Host),
			Port:     types.Int64Value(apiModel.Source.Config.Port),
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
		BucketName:            types.StringValue(apiModel.DestinationConfig.BucketName),
		OptionalPrefix:        types.StringValue(apiModel.DestinationConfig.OptionalPrefix),
	}
}

func DeploymentResourceToAPIModel(resourceModel DeploymentResourceModel) DeploymentAPIModel {
	tables := []TableAPIModel{}
	for _, table := range resourceModel.Source.Tables {
		tableUUID := table.UUID.ValueString()
		if tableUUID == "" {
			tableUUID = uuid.Nil.String()
		}
		tables = append(tables, TableAPIModel{
			UUID:                 tableUUID,
			Name:                 table.Name.ValueString(),
			Schema:               table.Schema.ValueString(),
			EnableHistoryMode:    table.EnableHistoryMode.ValueBool(),
			IndividualDeployment: table.IndividualDeployment.ValueBool(),
			IsPartitioned:        table.IsPartitioned.ValueBool(),
		})
	}

	apiModel := DeploymentAPIModel{
		UUID:            resourceModel.UUID.ValueString(),
		CompanyUUID:     resourceModel.CompanyUUID.ValueString(),
		Name:            resourceModel.Name.ValueString(),
		Status:          resourceModel.Status.ValueString(),
		DestinationUUID: resourceModel.DestinationUUID.ValueString(),
		Source: SourceAPIModel{
			Type:   resourceModel.Source.Type.ValueString(),
			Tables: tables,
		},
		DestinationConfig: DestinationConfigAPIModel{
			Dataset:               resourceModel.DestinationConfig.Dataset.ValueString(),
			Database:              resourceModel.DestinationConfig.Database.ValueString(),
			Schema:                resourceModel.DestinationConfig.Schema.ValueString(),
			SchemaOverride:        resourceModel.DestinationConfig.SchemaOverride.ValueString(),
			UseSameSchemaAsSource: resourceModel.DestinationConfig.UseSameSchemaAsSource.ValueBool(),
			SchemaNamePrefix:      resourceModel.DestinationConfig.SchemaNamePrefix.ValueString(),
			BucketName:            resourceModel.DestinationConfig.BucketName.ValueString(),
			OptionalPrefix:        resourceModel.DestinationConfig.OptionalPrefix.ValueString(),
		},
	}

	switch resourceModel.Source.Type.ValueString() {
	case "PostgreSQL":
		apiModel.Source.Config = SourceConfigAPIModel{
			Host:     resourceModel.Source.PostgresConfig.Host.ValueString(),
			Port:     resourceModel.Source.PostgresConfig.Port.ValueInt64(),
			User:     resourceModel.Source.PostgresConfig.User.ValueString(),
			Password: resourceModel.Source.PostgresConfig.Password.ValueString(),
			Database: resourceModel.Source.PostgresConfig.Database.ValueString(),
		}
	case "MySQL":
		apiModel.Source.Config = SourceConfigAPIModel{
			Host:     resourceModel.Source.MySQLConfig.Host.ValueString(),
			Port:     resourceModel.Source.MySQLConfig.Port.ValueInt64(),
			User:     resourceModel.Source.MySQLConfig.User.ValueString(),
			Password: resourceModel.Source.MySQLConfig.Password.ValueString(),
			Database: resourceModel.Source.MySQLConfig.Database.ValueString(),
		}
	}

	return apiModel
}
