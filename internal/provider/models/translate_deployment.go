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
	resourceModel.LastUpdatedAt = types.StringValue(apiModel.LastUpdatedAt)
	resourceModel.HasUndeployedChanges = types.BoolValue(apiModel.HasUndeployedChanges)
	resourceModel.DestinationUUID = types.StringValue(apiModel.DestinationUUID)

	tables := []TableModel{}
	for _, apiTable := range apiModel.Source.Tables {
		var advSettings *TableAdvancedSettingsModel
		if apiTable.AdvancedSettings != nil {
			advSettings = &TableAdvancedSettingsModel{
				Alias:                types.StringValue(apiTable.AdvancedSettings.Alias),
				SkipDelete:           types.BoolValue(apiTable.AdvancedSettings.SkipDelete),
				FlushIntervalSeconds: types.Int64Value(apiTable.AdvancedSettings.FlushIntervalSeconds),
				BufferRows:           types.Int64Value(apiTable.AdvancedSettings.BufferRows),
				FlushSizeKB:          types.Int64Value(apiTable.AdvancedSettings.FlushSizeKB),
				AutoscaleMaxReplicas: types.Int64Value(apiTable.AdvancedSettings.AutoscaleMaxReplicas),
				AutoscaleTargetValue: types.Int64Value(apiTable.AdvancedSettings.AutoscaleTargetValue),
				K8sRequestCPU:        types.Int64Value(apiTable.AdvancedSettings.K8sRequestCPU),
				K8sRequestMemoryMB:   types.Int64Value(apiTable.AdvancedSettings.K8sRequestMemoryMB),
				// TODO BigQueryPartitionSettings, MergePredicates, ExcludeColumns
			}
		}
		tables = append(tables, TableModel{
			UUID:                 types.StringValue(apiTable.UUID),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			AdvancedSettings:     advSettings,
		})
	}
	var dynamoDBConfig *DynamoDBConfigModel
	if apiModel.Source.Config.DynamoDB != nil {
		dynamoDBConfig = &DynamoDBConfigModel{
			Region:             types.StringValue(apiModel.Source.Config.DynamoDB.Region),
			TableName:          types.StringValue(apiModel.Source.Config.DynamoDB.TableName),
			StreamsArn:         types.StringValue(apiModel.Source.Config.DynamoDB.StreamsArn),
			AwsAccessKeyID:     types.StringValue(apiModel.Source.Config.DynamoDB.AwsAccessKeyID),
			AwsSecretAccessKey: types.StringValue(apiModel.Source.Config.DynamoDB.AwsSecretAccessKey),
		}
	}
	resourceModel.Source = &SourceModel{
		Name: types.StringValue(apiModel.Source.Name),
		Config: SourceConfigModel{
			Host:         types.StringValue(apiModel.Source.Config.Host),
			SnapshotHost: types.StringValue(apiModel.Source.Config.SnapshotHost),
			Port:         types.Int64Value(apiModel.Source.Config.Port),
			User:         types.StringValue(apiModel.Source.Config.User),
			Password:     types.StringValue(apiModel.Source.Config.Password),
			Database:     types.StringValue(apiModel.Source.Config.Database),
			DynamoDB:     dynamoDBConfig,
		},
		Tables: tables,
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

	var advSettings *DeploymentAdvancedSettingsModel
	if apiModel.AdvancedSettings != nil {
		advSettings = &DeploymentAdvancedSettingsModel{
			DropDeletedColumns:             types.BoolValue(apiModel.AdvancedSettings.DropDeletedColumns),
			IncludeArtieUpdatedAtColumn:    types.BoolValue(apiModel.AdvancedSettings.IncludeArtieUpdatedAtColumn),
			IncludeDatabaseUpdatedAtColumn: types.BoolValue(apiModel.AdvancedSettings.IncludeDatabaseUpdatedAtColumn),
			EnableHeartbeats:               types.BoolValue(apiModel.AdvancedSettings.EnableHeartbeats),
			EnableSoftDelete:               types.BoolValue(apiModel.AdvancedSettings.EnableSoftDelete),
			FlushIntervalSeconds:           types.Int64Value(apiModel.AdvancedSettings.FlushIntervalSeconds),
			BufferRows:                     types.Int64Value(apiModel.AdvancedSettings.BufferRows),
			FlushSizeKB:                    types.Int64Value(apiModel.AdvancedSettings.FlushSizeKB),
			PublicationNameOverride:        types.StringValue(apiModel.AdvancedSettings.PublicationNameOverride),
			ReplicationSlotOverride:        types.StringValue(apiModel.AdvancedSettings.ReplicationSlotOverride),
			PublicationAutoCreateMode:      types.StringValue(apiModel.AdvancedSettings.PublicationAutoCreateMode),
			// TODO PartitionRegex
		}
	}
	resourceModel.AdvancedSettings = advSettings
}

func DeploymentResourceToAPIModel(resourceModel DeploymentResourceModel) DeploymentAPIModel {
	tables := []TableAPIModel{}
	for _, table := range resourceModel.Source.Tables {
		tableUUID := table.UUID.ValueString()
		if tableUUID == "" {
			tableUUID = uuid.Nil.String()
		}
		var advSettings *TableAdvancedSettingsAPIModel
		if table.AdvancedSettings != nil {
			advSettings = &TableAdvancedSettingsAPIModel{
				Alias:                table.AdvancedSettings.Alias.ValueString(),
				SkipDelete:           table.AdvancedSettings.SkipDelete.ValueBool(),
				FlushIntervalSeconds: table.AdvancedSettings.FlushIntervalSeconds.ValueInt64(),
				BufferRows:           table.AdvancedSettings.BufferRows.ValueInt64(),
				FlushSizeKB:          table.AdvancedSettings.FlushSizeKB.ValueInt64(),
				AutoscaleMaxReplicas: table.AdvancedSettings.AutoscaleMaxReplicas.ValueInt64(),
				AutoscaleTargetValue: table.AdvancedSettings.AutoscaleTargetValue.ValueInt64(),
				K8sRequestCPU:        table.AdvancedSettings.K8sRequestCPU.ValueInt64(),
				K8sRequestMemoryMB:   table.AdvancedSettings.K8sRequestMemoryMB.ValueInt64(),
				// TODO BigQueryPartitionSettings, MergePredicates, ExcludeColumns
			}
		}
		tables = append(tables, TableAPIModel{
			UUID:                 tableUUID,
			Name:                 table.Name.ValueString(),
			Schema:               table.Schema.ValueString(),
			EnableHistoryMode:    table.EnableHistoryMode.ValueBool(),
			IndividualDeployment: table.IndividualDeployment.ValueBool(),
			IsPartitioned:        table.IsPartitioned.ValueBool(),
			AdvancedSettings:     advSettings,
		})
	}

	var dynamoDBConfig *DynamoDBConfigAPIModel
	if resourceModel.Source.Config.DynamoDB != nil {
		dynamoDBConfig = &DynamoDBConfigAPIModel{
			Region:             resourceModel.Source.Config.DynamoDB.Region.ValueString(),
			TableName:          resourceModel.Source.Config.DynamoDB.TableName.ValueString(),
			StreamsArn:         resourceModel.Source.Config.DynamoDB.StreamsArn.ValueString(),
			AwsAccessKeyID:     resourceModel.Source.Config.DynamoDB.AwsAccessKeyID.ValueString(),
			AwsSecretAccessKey: resourceModel.Source.Config.DynamoDB.AwsSecretAccessKey.ValueString(),
		}
	}

	var advSettings *DeploymentAdvancedSettingsAPIModel
	if resourceModel.AdvancedSettings != nil {
		advSettings = &DeploymentAdvancedSettingsAPIModel{
			DropDeletedColumns:             resourceModel.AdvancedSettings.DropDeletedColumns.ValueBool(),
			IncludeArtieUpdatedAtColumn:    resourceModel.AdvancedSettings.IncludeArtieUpdatedAtColumn.ValueBool(),
			IncludeDatabaseUpdatedAtColumn: resourceModel.AdvancedSettings.IncludeDatabaseUpdatedAtColumn.ValueBool(),
			EnableHeartbeats:               resourceModel.AdvancedSettings.EnableHeartbeats.ValueBool(),
			EnableSoftDelete:               resourceModel.AdvancedSettings.EnableSoftDelete.ValueBool(),
			FlushIntervalSeconds:           resourceModel.AdvancedSettings.FlushIntervalSeconds.ValueInt64(),
			BufferRows:                     resourceModel.AdvancedSettings.BufferRows.ValueInt64(),
			FlushSizeKB:                    resourceModel.AdvancedSettings.FlushSizeKB.ValueInt64(),
			PublicationNameOverride:        resourceModel.AdvancedSettings.PublicationNameOverride.ValueString(),
			ReplicationSlotOverride:        resourceModel.AdvancedSettings.ReplicationSlotOverride.ValueString(),
			PublicationAutoCreateMode:      resourceModel.AdvancedSettings.PublicationAutoCreateMode.ValueString(),
			// TODO PartitionRegex
		}
	}

	return DeploymentAPIModel{
		UUID:                 resourceModel.UUID.ValueString(),
		CompanyUUID:          resourceModel.CompanyUUID.ValueString(),
		Name:                 resourceModel.Name.ValueString(),
		Status:               resourceModel.Status.ValueString(),
		LastUpdatedAt:        resourceModel.LastUpdatedAt.ValueString(),
		HasUndeployedChanges: resourceModel.HasUndeployedChanges.ValueBool(),
		DestinationUUID:      resourceModel.DestinationUUID.ValueString(),
		Source: SourceAPIModel{
			Name: resourceModel.Source.Name.ValueString(),
			Config: SourceConfigAPIModel{
				Host:         resourceModel.Source.Config.Host.ValueString(),
				SnapshotHost: resourceModel.Source.Config.SnapshotHost.ValueString(),
				Port:         resourceModel.Source.Config.Port.ValueInt64(),
				User:         resourceModel.Source.Config.User.ValueString(),
				Password:     resourceModel.Source.Config.Password.ValueString(),
				Database:     resourceModel.Source.Config.Database.ValueString(),
				DynamoDB:     dynamoDBConfig,
			},
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
		AdvancedSettings: advSettings,
	}
}
