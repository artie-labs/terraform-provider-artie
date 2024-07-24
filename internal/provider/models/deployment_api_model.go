package models

type DeploymentAPIResponse struct {
	Deployment DeploymentAPIModel `json:"deploy"`
}

type DeploymentAPIModel struct {
	UUID                 string                             `json:"uuid"`
	Name                 string                             `json:"name"`
	Status               string                             `json:"status"`
	LastUpdatedAt        string                             `json:"lastUpdatedAt"`
	DestinationUUID      string                             `json:"destinationUUID"`
	HasUndeployedChanges bool                               `json:"hasUndeployedChanges"`
	Source               SourceAPIModel                     `json:"source"`
	AdvancedSettings     DeploymentAdvancedSettingsAPIModel `json:"advancedSettings"`
	UniqueConfig         map[string]any                     `json:"uniqueConfig"`
}

type SourceAPIModel struct {
	Name   string               `json:"name"`
	Config SourceConfigAPIModel `json:"config"`
	Tables []TableAPIModel      `json:"tables"`
}

type SourceConfigAPIModel struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
	// Password
	// DynamoDBConfig
	// SnapshotHost
}

type TableAPIModel struct {
	UUID                 string                        `json:"uuid"`
	Name                 string                        `json:"name"`
	Schema               string                        `json:"schema"`
	EnableHistoryMode    bool                          `json:"enableHistoryMode"`
	IndividualDeployment bool                          `json:"individualDeployment"`
	IsPartitioned        bool                          `json:"isPartitioned"`
	AdvancedSettings     TableAdvancedSettingsAPIModel `json:"advancedSettings"`
}

type TableAdvancedSettingsAPIModel struct {
	Alias                string `json:"alias"`
	SkipDelete           bool   `json:"skip_delete"`
	FlushIntervalSeconds int64  `json:"flush_interval_seconds"`
	BufferRows           int64  `json:"buffer_rows"`
	FlushSizeKB          int64  `json:"flush_size_kb"`
	// BigQueryPartitionSettings
	// MergePredicates
	// AutoscaleMaxReplicas
	// AutoscaleTargetValue
	// K8sRequestCPU
	// K8sRequestMemoryMB
	// ExcludeColumns
}

type DeploymentAdvancedSettingsAPIModel struct {
	DropDeletedColumns             bool  `json:"drop_deleted_columns"`
	IncludeArtieUpdatedAtColumn    bool  `json:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn bool  `json:"include_database_updated_at_column"`
	EnableHeartbeats               bool  `json:"enable_heartbeats"`
	EnableSoftDelete               bool  `json:"enable_soft_delete"`
	FlushIntervalSeconds           int64 `json:"flush_interval_seconds"`
	BufferRows                     int64 `json:"buffer_rows"`
	FlushSizeKB                    int64 `json:"flush_size_kb"`
	// PublicationNameOverride
	// ReplicationSlotOverride
	// PublicationAutoCreateMode
	// PartitionRegex
}
