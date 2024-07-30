package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type DeploymentResourceModel struct {
	UUID              types.String                      `tfsdk:"uuid"`
	CompanyUUID       types.String                      `tfsdk:"company_uuid"`
	Name              types.String                      `tfsdk:"name"`
	Status            types.String                      `tfsdk:"status"`
	DestinationUUID   types.String                      `tfsdk:"destination_uuid"`
	Source            *SourceModel                      `tfsdk:"source"`
	AdvancedSettings  *DeploymentAdvancedSettingsModel  `tfsdk:"advanced_settings"`
	DestinationConfig *DeploymentDestinationConfigModel `tfsdk:"destination_config"`
}

type SourceModel struct {
	Name   types.String      `tfsdk:"name"`
	Config SourceConfigModel `tfsdk:"config"`
	Tables []TableModel      `tfsdk:"tables"`
}

type SourceConfigModel struct {
	Host         types.String         `tfsdk:"host"`
	SnapshotHost types.String         `tfsdk:"snapshot_host"`
	Port         types.Int64          `tfsdk:"port"`
	User         types.String         `tfsdk:"user"`
	Database     types.String         `tfsdk:"database"`
	DynamoDB     *DynamoDBConfigModel `tfsdk:"dynamodb"`
	Password     types.String         `tfsdk:"password"`
}

type DynamoDBConfigModel struct {
	Region             types.String `tfsdk:"region"`
	TableName          types.String `tfsdk:"table_name"`
	StreamsArn         types.String `tfsdk:"streams_arn"`
	AwsAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
}

type TableModel struct {
	UUID                 types.String                `tfsdk:"uuid"`
	Name                 types.String                `tfsdk:"name"`
	Schema               types.String                `tfsdk:"schema"`
	EnableHistoryMode    types.Bool                  `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool                  `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool                  `tfsdk:"is_partitioned"`
	AdvancedSettings     *TableAdvancedSettingsModel `tfsdk:"advanced_settings"`
}

type TableAdvancedSettingsModel struct {
	Alias                types.String `tfsdk:"alias"`
	SkipDelete           types.Bool   `tfsdk:"skip_delete"`
	FlushIntervalSeconds types.Int64  `tfsdk:"flush_interval_seconds"`
	BufferRows           types.Int64  `tfsdk:"buffer_rows"`
	FlushSizeKB          types.Int64  `tfsdk:"flush_size_kb"`
	AutoscaleMaxReplicas types.Int64  `tfsdk:"autoscale_max_replicas"`
	AutoscaleTargetValue types.Int64  `tfsdk:"autoscale_target_value"`
	K8sRequestCPU        types.Int64  `tfsdk:"k8s_request_cpu"`
	K8sRequestMemoryMB   types.Int64  `tfsdk:"k8s_request_memory_mb"`
	// TODO BigQueryPartitionSettings, MergePredicates, ExcludeColumns
}

type DeploymentAdvancedSettingsModel struct {
	DropDeletedColumns             types.Bool   `tfsdk:"drop_deleted_columns"`
	IncludeArtieUpdatedAtColumn    types.Bool   `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool   `tfsdk:"include_database_updated_at_column"`
	EnableHeartbeats               types.Bool   `tfsdk:"enable_heartbeats"`
	EnableSoftDelete               types.Bool   `tfsdk:"enable_soft_delete"`
	FlushIntervalSeconds           types.Int64  `tfsdk:"flush_interval_seconds"`
	BufferRows                     types.Int64  `tfsdk:"buffer_rows"`
	FlushSizeKB                    types.Int64  `tfsdk:"flush_size_kb"`
	PublicationNameOverride        types.String `tfsdk:"publication_name_override"`
	ReplicationSlotOverride        types.String `tfsdk:"replication_slot_override"`
	PublicationAutoCreateMode      types.String `tfsdk:"publication_auto_create_mode"`
	// TODO PartitionRegex
}

type DeploymentDestinationConfigModel struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	SchemaOverride        types.String `tfsdk:"schema_override"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
	BucketName            types.String `tfsdk:"bucket_name"`
	OptionalPrefix        types.String `tfsdk:"optional_prefix"`
}
