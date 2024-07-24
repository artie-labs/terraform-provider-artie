package models

import "github.com/hashicorp/terraform-plugin-framework/types"

func DeploymentAPIToResourceModel(apiModel DeploymentAPIModel, resourceModel *DeploymentResourceModel) {
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Status = types.StringValue(apiModel.Status)
	resourceModel.LastUpdatedAt = types.StringValue(apiModel.LastUpdatedAt)
	resourceModel.HasUndeployedChanges = types.BoolValue(apiModel.HasUndeployedChanges)
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
			AdvancedSettings: TableAdvancedSettingsModel{
				Alias:                types.StringValue(apiTable.AdvancedSettings.Alias),
				SkipDelete:           types.BoolValue(apiTable.AdvancedSettings.SkipDelete),
				FlushIntervalSeconds: types.Int64Value(apiTable.AdvancedSettings.FlushIntervalSeconds),
				BufferRows:           types.Int64Value(apiTable.AdvancedSettings.BufferRows),
				FlushSizeKB:          types.Int64Value(apiTable.AdvancedSettings.FlushSizeKB),
				AutoscaleMaxReplicas: types.Int64Value(apiTable.AdvancedSettings.AutoscaleMaxReplicas),
				AutoscaleTargetValue: types.Int64Value(apiTable.AdvancedSettings.AutoscaleTargetValue),
				K8sRequestCPU:        types.Int64Value(apiTable.AdvancedSettings.K8sRequestCPU),
				K8sRequestMemoryMB:   types.Int64Value(apiTable.AdvancedSettings.K8sRequestMemoryMB),
			},
		})
	}
	resourceModel.Source = &SourceModel{
		Name: types.StringValue(apiModel.Source.Name),
		Config: SourceConfigModel{
			Host:         types.StringValue(apiModel.Source.Config.Host),
			SnapshotHost: types.StringValue(apiModel.Source.Config.SnapshotHost),
			Port:         types.Int64Value(apiModel.Source.Config.Port),
			User:         types.StringValue(apiModel.Source.Config.User),
			Database:     types.StringValue(apiModel.Source.Config.Database),
			DynamoDB: &DynamoDBConfigModel{
				Region:             types.StringValue(apiModel.Source.Config.DynamoDB.Region),
				TableName:          types.StringValue(apiModel.Source.Config.DynamoDB.TableName),
				StreamsArn:         types.StringValue(apiModel.Source.Config.DynamoDB.StreamsArn),
				AwsAccessKeyID:     types.StringValue(apiModel.Source.Config.DynamoDB.AwsAccessKeyID),
				AwsSecretAccessKey: types.StringValue(apiModel.Source.Config.DynamoDB.AwsSecretAccessKey),
			},
		},
		Tables: tables,
	}
	resourceModel.DestinationConfig = &DestinationConfigModel{
		Dataset:               types.StringValue(apiModel.DestinationConfig.Dataset),
		Database:              types.StringValue(apiModel.DestinationConfig.Database),
		Schema:                types.StringValue(apiModel.DestinationConfig.Schema),
		SchemaOverride:        types.StringValue(apiModel.DestinationConfig.SchemaOverride),
		UseSameSchemaAsSource: types.BoolValue(apiModel.DestinationConfig.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.DestinationConfig.SchemaNamePrefix),
		BucketName:            types.StringValue(apiModel.DestinationConfig.BucketName),
		OptionalPrefix:        types.StringValue(apiModel.DestinationConfig.OptionalPrefix),
	}
	resourceModel.AdvancedSettings = &DeploymentAdvancedSettingsModel{
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
	}
}
