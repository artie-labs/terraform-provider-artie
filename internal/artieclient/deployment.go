package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"terraform-provider-artie/internal/lib"
)

type BaseDeployment struct {
	Name                     string            `json:"name"`
	Source                   Source            `json:"source"`
	DestinationUUID          *uuid.UUID        `json:"destinationUUID"`
	DestinationConfig        DestinationConfig `json:"specificDestCfg"`
	SSHTunnelUUID            *uuid.UUID        `json:"sshTunnelUUID"`
	SnowflakeEcoScheduleUUID *uuid.UUID        `json:"snowflakeEcoScheduleUUID"`
	DataPlaneName            string            `json:"dataPlaneName"`

	// Advanced settings - these must all be nullable
	DropDeletedColumns             *bool   `json:"dropDeletedColumns"`
	EnableSoftDelete               *bool   `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn    *bool   `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn *bool   `json:"includeDatabaseUpdatedAtColumn"`
	OneTopicPerSchema              *bool   `json:"oneTopicPerSchema"`
	PublicationNameOverride        *string `json:"publicationNameOverride"`
	ReplicationSlotOverride        *string `json:"replicationSlotOverride"`
}

type Deployment struct {
	BaseDeployment
	UUID   uuid.UUID `json:"uuid"`
	Status string    `json:"status"`
}

type advancedSettings struct {
	DropDeletedColumns             bool   `json:"dropDeletedColumns"`
	EnableSoftDelete               bool   `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn    bool   `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn bool   `json:"includeDatabaseUpdatedAtColumn"`
	OneTopicPerSchema              bool   `json:"oneTopicPerSchema"`
	PublicationNameOverride        string `json:"publicationNameOverride"`
	ReplicationSlotOverride        string `json:"replicationSlotOverride"`
}

type deploymentWithAdvSettings struct {
	Deployment
	AdvancedSettings advancedSettings           `json:"advancedSettings"`
	Source           sourceWithAdvTableSettings `json:"source"`
}

func (deployment deploymentWithAdvSettings) unnestAdvSettings() Deployment {
	deployment.DropDeletedColumns = &deployment.AdvancedSettings.DropDeletedColumns
	deployment.EnableSoftDelete = &deployment.AdvancedSettings.EnableSoftDelete
	deployment.IncludeArtieUpdatedAtColumn = &deployment.AdvancedSettings.IncludeArtieUpdatedAtColumn
	deployment.IncludeDatabaseUpdatedAtColumn = &deployment.AdvancedSettings.IncludeDatabaseUpdatedAtColumn
	deployment.OneTopicPerSchema = &deployment.AdvancedSettings.OneTopicPerSchema
	deployment.PublicationNameOverride = &deployment.AdvancedSettings.PublicationNameOverride
	deployment.ReplicationSlotOverride = &deployment.AdvancedSettings.ReplicationSlotOverride

	deployment.Deployment.Source.Type = deployment.Source.Type
	deployment.Deployment.Source.Config = deployment.Source.Config
	tables := []Table{}
	for _, table := range deployment.Source.Tables {
		tables = append(tables, table.unnestTableAdvSettings())
	}
	deployment.Deployment.Source.Tables = tables

	return deployment.Deployment
}

type Source struct {
	Type   ConnectorType `json:"type"`
	Config SourceConfig  `json:"config"`
	Tables []Table       `json:"tables"`
}

func (s Source) BuildAPISource() APISource {
	apiSource := APISource{
		Type:   s.Type,
		Config: s.Config,
		Tables: []APITable{},
	}

	for _, table := range s.Tables {
		apiSource.Tables = append(apiSource.Tables, table.BuildAPITable())
	}

	return apiSource
}

type APISource struct {
	Type   ConnectorType `json:"type"`
	Config SourceConfig  `json:"config"`
	Tables []APITable    `json:"tables"`
}

type sourceWithAdvTableSettings struct {
	Source
	Tables []APITable `json:"tables"`
}

type SourceConfig struct {
	Host         string          `json:"host"`
	SnapshotHost string          `json:"snapshotHost"`
	Port         int32           `json:"port"`
	User         string          `json:"user"`
	Password     string          `json:"password"`
	Database     string          `json:"database"`
	Container    string          `json:"containerName,omitempty"`
	DynamoDB     *DynamoDBConfig `json:"dynamodb,omitempty"`
}

type DynamoDBConfig struct {
	StreamsArn         string                 `json:"streamsArn"`
	AwsAccessKeyID     string                 `json:"awsAccessKeyId"`
	AwsSecretAccessKey string                 `json:"awsSecretAccessKey"`
	SnapshotConfig     DynamoDBSnapshotConfig `json:"snapshotConfig"`
}

type DynamoDBSnapshotConfig struct {
	Enabled        bool   `json:"enabled"`
	Bucket         string `json:"bucket"`
	OptionalFolder string `json:"optionalFolder"`
}

type MergePredicate struct {
	PartitionField string `json:"partitionField"`
}

type Table struct {
	UUID                 uuid.UUID `json:"uuid"`
	Name                 string    `json:"name"`
	Schema               string    `json:"schema"`
	EnableHistoryMode    bool      `json:"enableHistoryMode"`
	IndividualDeployment bool      `json:"individualDeployment"`
	IsPartitioned        bool      `json:"isPartitioned"`

	// Advanced table settings - these must all be nullable
	Alias           *string           `json:"alias"`
	ExcludeColumns  *[]string         `json:"excludeColumns"`
	ColumnsToHash   *[]string         `json:"columnsToHash"`
	SkipDeletes     *bool             `json:"skipDelete"`
	MergePredicates *[]MergePredicate `json:"mergePredicates"`
}

// This is used for just the validate-source endpoint for now
// We'll spin up a separate workstream to deprecate the need to have [Table] in general (which will require us to work on nullability of advanced settings)
func (t Table) BuildAPITable() APITable {
	apiTable := APITable{
		Table: t,
		AdvancedSettings: advancedTableSettings{
			Alias:           lib.RemovePtr(t.Alias),
			ExcludeColumns:  lib.RemovePtr(t.ExcludeColumns),
			ColumnsToHash:   lib.RemovePtr(t.ColumnsToHash),
			SkipDeletes:     lib.RemovePtr(t.SkipDeletes),
			MergePredicates: lib.RemovePtr(t.MergePredicates),
		},
	}

	return apiTable
}

type advancedTableSettings struct {
	Alias           string           `json:"alias"`
	ExcludeColumns  []string         `json:"excludeColumns"`
	ColumnsToHash   []string         `json:"columnsToHash"`
	SkipDeletes     bool             `json:"skipDelete"`
	MergePredicates []MergePredicate `json:"mergePredicates"`
}

type APITable struct {
	Table
	AdvancedSettings advancedTableSettings `json:"advancedSettings"`
}

func toSlicePtr[T any](slice []T) *[]T {
	if len(slice) == 0 {
		return &[]T{}
	}
	return &slice
}

func (t APITable) unnestTableAdvSettings() Table {
	t.Alias = &t.AdvancedSettings.Alias
	t.SkipDeletes = &t.AdvancedSettings.SkipDeletes

	// These arrays are omitted from the api response if empty; fallback to empty slices
	// so terraform doesn't think a change is needed if the tf config specifies empty slices
	t.ExcludeColumns = toSlicePtr(t.AdvancedSettings.ExcludeColumns)
	t.ColumnsToHash = toSlicePtr(t.AdvancedSettings.ColumnsToHash)
	t.MergePredicates = toSlicePtr[MergePredicate](t.AdvancedSettings.MergePredicates)

	return t.Table
}

type DestinationConfig struct {
	Dataset               string `json:"dataset"`
	Database              string `json:"database"`
	Schema                string `json:"schema"`
	UseSameSchemaAsSource bool   `json:"useSameSchemaAsSource"`
	SchemaNamePrefix      string `json:"schemaNamePrefix"`
	Bucket                string `json:"bucketName"`
	Folder                string `json:"folderName"`
}

type DeploymentClient struct {
	client Client
}

func (DeploymentClient) basePath() string {
	return "deployments"
}

type deploymentAPIResponse struct {
	Deployment deploymentWithAdvSettings `json:"deployment"`
}

type validationResponse struct {
	Error string `json:"error"`
}

func (dc DeploymentClient) Get(ctx context.Context, deploymentUUID string) (Deployment, error) {
	path, err := url.JoinPath(dc.basePath(), deploymentUUID)
	if err != nil {
		return Deployment{}, err
	}
	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodGet, path, nil)
	if err != nil {
		return Deployment{}, err
	}
	return response.Deployment.unnestAdvSettings(), nil
}

func (dc DeploymentClient) Create(ctx context.Context, deployment BaseDeployment) (Deployment, error) {
	body := map[string]any{
		"deployment":      deployment,
		"startDeployment": true,
	}
	deploymentResp, err := makeRequest[deploymentWithAdvSettings](ctx, dc.client, http.MethodPost, dc.basePath(), body)
	if err != nil {
		return Deployment{}, err
	}
	return deploymentResp.unnestAdvSettings(), nil
}

func (dc DeploymentClient) Update(ctx context.Context, deployment Deployment) (Deployment, error) {
	path, err := url.JoinPath(dc.basePath(), deployment.UUID.String())
	if err != nil {
		return Deployment{}, err
	}

	body := map[string]any{
		"deployment":      deployment,
		"startDeployment": true,
	}

	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return Deployment{}, err
	}
	return response.Deployment.unnestAdvSettings(), nil
}

func (dc DeploymentClient) ValidateSource(ctx context.Context, deployment BaseDeployment) error {
	path, err := url.JoinPath(dc.basePath(), "validate-source")
	if err != nil {
		return err
	}

	body := map[string]any{
		"source":         deployment.Source.BuildAPISource(),
		"sshTunnelUUID":  deployment.SSHTunnelUUID,
		"validateTables": true,
	}

	response, err := makeRequest[validationResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("source validation failed: %s", response.Error)
	}

	return nil
}

func (dc DeploymentClient) ValidateDestination(ctx context.Context, deployment BaseDeployment) error {
	path, err := url.JoinPath(dc.basePath(), "validate-destination")
	if err != nil {
		return err
	}

	body := map[string]any{
		"destinationUUID": deployment.DestinationUUID,
		"specificCfg":     deployment.DestinationConfig,
		"tables":          deployment.Source.Tables,
		"sourceType":      deployment.Source.Type,
	}

	response, err := makeRequest[validationResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("destination validation failed: %s", response.Error)
	}

	return nil
}

func (dc DeploymentClient) Delete(ctx context.Context, deploymentUUID string) error {
	path, err := url.JoinPath(dc.basePath(), deploymentUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
