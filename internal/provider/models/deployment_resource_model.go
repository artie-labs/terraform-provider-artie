package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type DeploymentResourceModel struct {
	UUID                 types.String                     `tfsdk:"uuid"`
	Name                 types.String                     `tfsdk:"name"`
	Status               types.String                     `tfsdk:"status"`
	LastUpdatedAt        types.String                     `tfsdk:"last_updated_at"`
	DestinationUUID      types.String                     `tfsdk:"destination_uuid"`
	HasUndeployedChanges types.Bool                       `tfsdk:"has_undeployed_changes"`
	Source               *SourceModel                     `tfsdk:"source"`
	AdvancedSettings     *DeploymentAdvancedSettingsModel `tfsdk:"advanced_settings"`
	UniqueConfig         types.Map                        `tfsdk:"unique_config"`
}

type SourceModel struct {
	Name   types.String      `tfsdk:"name"`
	Config SourceConfigModel `tfsdk:"config"`
	Tables []TableModel      `tfsdk:"tables"`
}

type SourceConfigModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	// Password
	// DynamoDBConfig
	// SnapshotHost
}

type TableModel struct {
	UUID                 types.String               `tfsdk:"uuid"`
	Name                 types.String               `tfsdk:"name"`
	Schema               types.String               `tfsdk:"schema"`
	EnableHistoryMode    types.Bool                 `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool                 `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool                 `tfsdk:"is_partitioned"`
	AdvancedSettings     TableAdvancedSettingsModel `tfsdk:"advanced_settings"`
}

type TableAdvancedSettingsModel struct {
	Alias                types.String `tfsdk:"alias"`
	SkipDelete           types.Bool   `tfsdk:"skip_delete"`
	FlushIntervalSeconds types.Int64  `tfsdk:"flush_interval_seconds"`
	BufferRows           types.Int64  `tfsdk:"buffer_rows"`
	FlushSizeKB          types.Int64  `tfsdk:"flush_size_kb"`
	// BigQueryPartitionSettings
	// MergePredicates
	// AutoscaleMaxReplicas
	// AutoscaleTargetValue
	// K8sRequestCPU
	// K8sRequestMemoryMB
	// ExcludeColumns
}

type DeploymentAdvancedSettingsModel struct {
	DropDeletedColumns             types.Bool  `tfsdk:"drop_deleted_columns"`
	IncludeArtieUpdatedAtColumn    types.Bool  `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool  `tfsdk:"include_database_updated_at_column"`
	EnableHeartbeats               types.Bool  `tfsdk:"enable_heartbeats"`
	EnableSoftDelete               types.Bool  `tfsdk:"enable_soft_delete"`
	FlushIntervalSeconds           types.Int64 `tfsdk:"flush_interval_seconds"`
	BufferRows                     types.Int64 `tfsdk:"buffer_rows"`
	FlushSizeKB                    types.Int64 `tfsdk:"flush_size_kb"`
	// PublicationNameOverride
	// ReplicationSlotOverride
	// PublicationAutoCreateMode
	// PartitionRegex
}