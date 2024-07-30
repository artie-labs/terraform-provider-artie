package models

type DeploymentAPIResponse struct {
	Deployment DeploymentAPIModel `json:"deploy"`
}

type DeploymentAPIModel struct {
	UUID                 string                             `json:"uuid"`
	CompanyUUID          string                             `json:"companyUUID"`
	Name                 string                             `json:"name"`
	Status               string                             `json:"status"`
	LastUpdatedAt        string                             `json:"lastUpdatedAt"`
	DestinationUUID      string                             `json:"destinationUUID"`
	HasUndeployedChanges bool                               `json:"hasUndeployedChanges"`
	Source               SourceAPIModel                     `json:"source"`
	AdvancedSettings     DeploymentAdvancedSettingsAPIModel `json:"advancedSettings"`
	DestinationConfig    DestinationConfigAPIModel          `json:"uniqueConfig"`
}

type SourceAPIModel struct {
	Name   string               `json:"name"`
	Config SourceConfigAPIModel `json:"config"`
	Tables []TableAPIModel      `json:"tables"`
}

type SourceConfigAPIModel struct {
	Host         string                  `json:"host"`
	SnapshotHost string                  `json:"snapshotHost"`
	Port         int64                   `json:"port"`
	User         string                  `json:"user"`
	Password     string                  `json:"password"`
	Database     string                  `json:"database"`
	DynamoDB     *DynamoDBConfigAPIModel `json:"dynamodb"`
}

type DynamoDBConfigAPIModel struct {
	Region             string `json:"region"`
	TableName          string `json:"tableName"`
	StreamsArn         string `json:"streamsArn"`
	AwsAccessKeyID     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
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
	SkipDelete           bool   `json:"skipDelete"`
	FlushIntervalSeconds int64  `json:"flushIntervalSeconds"`
	BufferRows           int64  `json:"bufferRows"`
	FlushSizeKB          int64  `json:"flushSizeKb"`
	AutoscaleMaxReplicas int64  `json:"autoscaleMaxReplicas"`
	AutoscaleTargetValue int64  `json:"autoscaleTargetValue"`
	K8sRequestCPU        int64  `json:"k8sRequestCPU"`
	K8sRequestMemoryMB   int64  `json:"k8sRequestMemoryMB"`
	// TODO BigQueryPartitionSettings, MergePredicates, ExcludeColumns
}

type DeploymentAdvancedSettingsAPIModel struct {
	DropDeletedColumns             bool   `json:"dropDeletedColumns"`
	IncludeArtieUpdatedAtColumn    bool   `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn bool   `json:"includeDatabaseUpdatedAtColumn"`
	EnableHeartbeats               bool   `json:"enableHeartbeats"`
	EnableSoftDelete               bool   `json:"enableSoftDelete"`
	FlushIntervalSeconds           int64  `json:"flushIntervalSeconds"`
	BufferRows                     int64  `json:"bufferRows"`
	FlushSizeKB                    int64  `json:"flushSizeKb"`
	PublicationNameOverride        string `json:"publicationNameOverride"`
	ReplicationSlotOverride        string `json:"replicationSlotOverride"`
	PublicationAutoCreateMode      string `json:"publicationAutoCreateMode"`
	// TODO PartitionRegex
}

type DestinationConfigAPIModel struct {
	Dataset               string `json:"dataset"`
	Database              string `json:"database"`
	Schema                string `json:"schema"`
	SchemaOverride        string `json:"schemaOverride"`
	UseSameSchemaAsSource bool   `json:"useSameSchemaAsSource"`
	SchemaNamePrefix      string `json:"schemaNamePrefix"`
	BucketName            string `json:"bucketName"`
	OptionalPrefix        string `json:"optionalPrefix"`
}
