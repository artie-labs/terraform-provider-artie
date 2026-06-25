package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"terraform-provider-artie/internal/openapi"
)

type StaticColumn struct {
	Column string `json:"column"`
	Value  string `json:"value"`
}

type AdvancedSettings struct {
	DropDeletedColumns                           *bool           `json:"dropDeletedColumns"`
	EnableSoftDelete                             *bool           `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn                  *bool           `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn               *bool           `json:"includeDatabaseUpdatedAtColumn"`
	IncludeArtieOperationColumn                  *bool           `json:"includeArtieOperationColumn"`
	IncludeFullSourceTableNameColumn             *bool           `json:"includeFullSourceTableNameColumn"`
	IncludeFullSourceTableNameColumnAsPrimaryKey *bool           `json:"includeFullSourceTableNameColumnAsPrimaryKey"`
	IncludeSourceMetadataColumn                  *bool           `json:"includeSourceMetadataColumn"`
	FlushIntervalSeconds                         *int64          `json:"flushIntervalSeconds"`
	BufferRows                                   *int64          `json:"bufferRows"`
	FlushSizeKB                                  *int64          `json:"flushSizeKb"`
	DefaultSourceSchema                          *string         `json:"defaultSourceSchema"`
	SplitEventsByType                            *bool           `json:"splitEventsByType"`
	StaticColumns                                *[]StaticColumn `json:"staticColumns,omitempty"`
	AutoReplicateNewTables                       *bool           `json:"autoReplicateNewTables"`
	AppendOnly                                   *bool           `json:"appendOnly"`
	StagingSchema                                *string         `json:"stagingSchema"`
	ForceUTCTimezone                             *bool           `json:"forceUTCTimezone"`
	WriteRawBinaryValues                         *bool           `json:"writeRawBinaryValues"`
	DisableAlerts                                *bool           `json:"disableAlerts"`
}

type BasePipeline struct {
	Name                     string            `json:"name"`
	DataPlaneName            string            `json:"dataPlaneName"`
	SourceReaderUUID         *uuid.UUID        `json:"sourceReaderUUID"`
	Tables                   []Table           `json:"tables"`
	DestinationUUID          *uuid.UUID        `json:"destinationUUID"`
	DestinationConfig        DestinationConfig `json:"specificDestCfg"`
	SnowflakeEcoScheduleUUID *uuid.UUID        `json:"snowflakeEcoScheduleUUID"`
	EncryptionKeyUUID        *uuid.UUID        `json:"encryptionKeyUUID"`
	ColumnHashingSaltUUID    *uuid.UUID        `json:"columnHashingSaltUUID"`
	AdvancedSettings         *AdvancedSettings `json:"advancedSettings"`
}

type Pipeline struct {
	BasePipeline
	UUID uuid.UUID `json:"uuid"`
}

type Table struct {
	UUID               uuid.UUID             `json:"uuid"`
	Name               string                `json:"name"`
	Schema             string                `json:"schema"`
	EnableHistoryMode  bool                  `json:"enableHistoryMode"`
	DisableReplication bool                  `json:"disableReplication"`
	AdvancedSettings   AdvancedTableSettings `json:"advancedSettings"`
}

type MergePredicate struct {
	PartitionField string `json:"partitionField"`
	PartitionType  string `json:"partitionType,omitempty"`
}

type SoftPartitioning struct {
	Enabled            bool               `json:"enabled"`
	PartitionFrequency PartitionFrequency `json:"partitionFrequency"`
	PartitionColumn    string             `json:"partitionColumn"`
	MaxPartitions      int                `json:"maxPartitions"`
}

type CTIDSettings struct {
	ChunkSize      uint `json:"chunkSize"`
	Enabled        bool `json:"enabled"`
	MaxParallelism uint `json:"maxParallelism"`
}

type RangeSettings struct {
	Enabled        bool `json:"enabled"`
	ChunkSize      int  `json:"chunksSize"`
	MaxParallelism int  `json:"maxParallelism"`
	BatchSize      int  `json:"batchSize"`
}

type AdvancedTableSettings struct {
	Alias                      *string           `json:"alias"`
	ExcludeColumns             *[]string         `json:"excludeColumns"`
	IncludeColumns             *[]string         `json:"includeColumns"`
	ColumnsToHash              *[]string         `json:"columnsToHash"`
	ColumnsToCompress          *[]string         `json:"columnsToCompress"`
	ColumnsToEncrypt           *[]string         `json:"columnsToEncrypt"`
	EncryptJSONBColumns        *bool             `json:"encryptJSONBColumns"`
	SkipDeletes                *bool             `json:"skipDelete"`
	UnifyAcrossSchemas         *bool             `json:"unifyAcrossSchemas"`
	UnifyAcrossDatabases       *bool             `json:"unifyAcrossDatabases"`
	MergePredicates            *[]MergePredicate `json:"mergePredicates"`
	SoftPartitioning           *SoftPartitioning `json:"softPartitioning,omitempty"`
	ShouldBackfillHistoryTable *bool             `json:"shouldBackfillHistoryTable"`
	CTIDSettings               *CTIDSettings     `json:"ctidSettings,omitempty"`
	RangeSettings              *RangeSettings    `json:"rangeSettings,omitempty"`
	SkipBackfill               *bool             `json:"skipBackfill"`
	SkipNoOpUpdates            *bool             `json:"skipNoOpUpdates"`
}

type FlushConfig struct {
	FlushIntervalSeconds int64 `json:"flushIntervalSeconds"`
	BufferRows           int64 `json:"bufferRows"`
	FlushSizeKB          int64 `json:"flushSizeKB"`
}

type DestinationConfig struct {
	Dataset                 string `json:"dataset"`
	Database                string `json:"database"`
	Schema                  string `json:"schema"`
	UseSameSchemaAsSource   bool   `json:"useSameSchemaAsSource"`
	SchemaNamePrefix        string `json:"schemaNamePrefix"`
	Bucket                  string `json:"bucketName"`
	TableNameSeparator      string `json:"tableNameSeparator"`
	Folder                  string `json:"folderName"`
	CreateIcebergNamespaces bool   `json:"dynamicallyCreateNamespaces"`
}

type PipelineClient struct {
	client       Client
	openAPICient *openapi.ClientWithResponses
}

func (PipelineClient) basePath() string {
	return "pipelines"
}

func (pc PipelineClient) Get(ctx context.Context, pipelineUUID string) (Pipeline, error) {
	path, err := url.JoinPath(pc.basePath(), pipelineUUID)
	if err != nil {
		return Pipeline{}, err
	}

	return makeRequest[Pipeline](ctx, pc.client, http.MethodGet, path, nil)
}

func (pc PipelineClient) ValidateSource(ctx context.Context, pipeline BasePipeline) error {
	body := map[string]any{
		"sourceReaderUUID": pipeline.SourceReaderUUID,
		"validateTables":   true,
		"tables":           pipeline.Tables,
		"dataPlaneName":    pipeline.DataPlaneName,
	}
	path, err := url.JoinPath(pc.basePath(), "validate-unsaved-source")
	if err != nil {
		return err
	}

	response, err := makeRequest[validationResponse](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("source validation failed: %s", response.Error)
	}

	return err
}

func (pc PipelineClient) ValidateDestination(ctx context.Context, pipeline BasePipeline) error {
	body := map[string]any{
		"destinationUUID":  pipeline.DestinationUUID,
		"sourceReaderUUID": pipeline.SourceReaderUUID,
		"specificCfg":      pipeline.DestinationConfig,
		"tables":           pipeline.Tables,
		"advancedSettings": pipeline.AdvancedSettings,
	}

	path, err := url.JoinPath(pc.basePath(), "validate-unsaved-destination")
	if err != nil {
		return err
	}

	response, err := makeRequest[validationResponse](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("destination validation failed: %s", response.Error)
	}

	return err
}

func (pc PipelineClient) Create(ctx context.Context, pipeline BasePipeline) (Pipeline, error) {
	body := map[string]any{
		"pipeline": pipeline,
	}

	return makeRequest[Pipeline](ctx, pc.client, http.MethodPost, pc.basePath(), body)
}

func (pc PipelineClient) Update(ctx context.Context, pipeline Pipeline) (Pipeline, error) {
	path, err := url.JoinPath(pc.basePath(), pipeline.UUID.String())
	if err != nil {
		return Pipeline{}, err
	}

	body := map[string]any{
		"pipeline": pipeline,
	}

	return makeRequest[Pipeline](ctx, pc.client, http.MethodPost, path, body)
}

func (pc PipelineClient) Delete(ctx context.Context, pipelineUUID string) error {
	path, err := url.JoinPath(pc.basePath(), pipelineUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, pc.client, http.MethodDelete, path, nil)
	return err
}

func (pc PipelineClient) StartPipeline(ctx context.Context, pipelineUUID string) error {
	path, err := url.JoinPath(pc.basePath(), pipelineUUID, "start")
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, pc.client, http.MethodPost, path, nil)
	return err
}

func (pc PipelineClient) UpdateStatus(ctx context.Context, pipelineUUID string, status string) error {
	resp, err := pc.openAPICient.PostPipelinesUuidStatusWithResponse(ctx, pipelineUUID, openapi.RouterPipelineUpdateStatusRequest{
		Status: openapi.EnumsPipelineStatus(status),
	})
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}
