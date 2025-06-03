package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type AdvancedSettings struct {
	DropDeletedColumns             *bool   `json:"dropDeletedColumns"`
	EnableSoftDelete               *bool   `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn    *bool   `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn *bool   `json:"includeDatabaseUpdatedAtColumn"`
	OneTopicPerSchema              *bool   `json:"oneTopicPerSchema"`
	PublicationNameOverride        *string `json:"publicationNameOverride"`
	ReplicationSlotOverride        *string `json:"replicationSlotOverride"`
	FlushIntervalSeconds           *int64  `json:"flushIntervalSeconds"`
	BufferRows                     *int64  `json:"bufferRows"`
	FlushSizeKB                    *int64  `json:"flushSizeKb"`
}

type BasePipeline struct {
	Name                     string            `json:"name"`
	DataPlaneName            string            `json:"dataPlaneName"`
	SourceReaderUUID         *uuid.UUID        `json:"sourceReaderUUID"`
	Tables                   []Table           `json:"tables"`
	DestinationUUID          *uuid.UUID        `json:"destinationUUID"`
	DestinationConfig        DestinationConfig `json:"specificDestCfg"`
	SnowflakeEcoScheduleUUID *uuid.UUID        `json:"snowflakeEcoScheduleUUID"`
	AdvancedSettings         *AdvancedSettings `json:"advancedSettings"`
}

type Pipeline struct {
	BasePipeline
	UUID   uuid.UUID `json:"uuid"`
	Status string    `json:"status"`
}

type apiPipeline struct {
	Pipeline
	Tables []APITable `json:"tables"`
}

func (p apiPipeline) toPipeline() Pipeline {
	tables := []Table{}
	for _, table := range p.Tables {
		tables = append(tables, table.unnestTableAdvSettings())
	}

	p.Pipeline.Tables = tables
	return p.Pipeline
}

type Table struct {
	UUID              uuid.UUID `json:"uuid"`
	Name              string    `json:"name"`
	Schema            string    `json:"schema"`
	EnableHistoryMode bool      `json:"enableHistoryMode"`
	IsPartitioned     bool      `json:"isPartitioned"`

	// Advanced table settings - these must all be nullable
	Alias           *string           `json:"alias"`
	ExcludeColumns  *[]string         `json:"excludeColumns"`
	IncludeColumns  *[]string         `json:"includeColumns"`
	ColumnsToHash   *[]string         `json:"columnsToHash"`
	SkipDeletes     *bool             `json:"skipDelete"`
	MergePredicates *[]MergePredicate `json:"mergePredicates"`
}

type MergePredicate struct {
	PartitionField string `json:"partitionField"`
}

type advancedTableSettings struct {
	Alias           string           `json:"alias"`
	ExcludeColumns  []string         `json:"excludeColumns"`
	IncludeColumns  []string         `json:"includeColumns"`
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
	t.IncludeColumns = toSlicePtr(t.AdvancedSettings.IncludeColumns)
	t.ColumnsToHash = toSlicePtr(t.AdvancedSettings.ColumnsToHash)
	t.MergePredicates = toSlicePtr(t.AdvancedSettings.MergePredicates)

	return t.Table
}

type FlushConfig struct {
	FlushIntervalSeconds int64 `json:"flushIntervalSeconds"`
	BufferRows           int64 `json:"bufferRows"`
	FlushSizeKB          int64 `json:"flushSizeKB"`
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

type PipelineClient struct {
	client Client
}

func (PipelineClient) basePath() string {
	return "pipelines"
}

func (pc PipelineClient) Get(ctx context.Context, pipelineUUID string) (Pipeline, error) {
	path, err := url.JoinPath(pc.basePath(), pipelineUUID)
	if err != nil {
		return Pipeline{}, err
	}

	pipeline, err := makeRequest[apiPipeline](ctx, pc.client, http.MethodGet, path, nil)
	if err != nil {
		return Pipeline{}, err
	}

	return pipeline.toPipeline(), nil
}

func (pc PipelineClient) Create(ctx context.Context, pipeline BasePipeline) (Pipeline, error) {
	body := map[string]any{
		"pipeline": pipeline,
	}
	createdPipeline, err := makeRequest[apiPipeline](ctx, pc.client, http.MethodPost, pc.basePath(), body)
	if err != nil {
		return Pipeline{}, err
	}

	return createdPipeline.toPipeline(), nil
}

func (pc PipelineClient) Update(ctx context.Context, pipeline Pipeline) (Pipeline, error) {
	path, err := url.JoinPath(pc.basePath(), pipeline.UUID.String())
	if err != nil {
		return Pipeline{}, err
	}

	body := map[string]any{
		"pipeline": pipeline,
	}
	updatedPipeline, err := makeRequest[apiPipeline](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return Pipeline{}, err
	}

	return updatedPipeline.toPipeline(), nil
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
