package artieclient

import (
	"context"
	"fmt"
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

	pipeline, err := makeRequest[Pipeline](ctx, pc.client, http.MethodGet, path, nil)
	if err != nil {
		return Pipeline{}, err
	}

	return pipeline, nil
}

func (pc PipelineClient) Create(ctx context.Context, pipeline BasePipeline) (Pipeline, error) {
	body := map[string]any{
		"pipeline":       pipeline,
		"deployPipeline": true,
	}
	createdPipeline, err := makeRequest[Pipeline](ctx, pc.client, http.MethodPost, pc.basePath(), body)
	if err != nil {
		return Pipeline{}, err
	}

	return createdPipeline, nil
}

func (pc PipelineClient) Update(ctx context.Context, pipeline Pipeline) (Pipeline, error) {
	path, err := url.JoinPath(pc.basePath(), pipeline.UUID.String())
	if err != nil {
		return Pipeline{}, err
	}

	body := map[string]any{
		"pipeline":       pipeline,
		"deployPipeline": true,
	}
	updatedPipeline, err := makeRequest[Pipeline](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return Pipeline{}, err
	}

	return updatedPipeline, nil
}

func (pc PipelineClient) Delete(ctx context.Context, pipelineUUID string) error {
	path, err := url.JoinPath(pc.basePath(), pipelineUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, pc.client, http.MethodDelete, path, nil)
	return err
}

func (pc PipelineClient) ValidateSource(ctx context.Context, pipeline BasePipeline) error {
	path, err := url.JoinPath(pc.basePath(), "validate-source")
	if err != nil {
		return err
	}

	body := map[string]any{
		"sourceReaderUUID": pipeline.SourceReaderUUID,
		"validateTables":   true,
		"tables":           pipeline.Tables,
	}

	response, err := makeRequest[validationResponse](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("source validation failed: %s", response.Error)
	}

	return nil
}

func (pc PipelineClient) ValidateDestination(ctx context.Context, pipeline BasePipeline) error {
	path, err := url.JoinPath(pc.basePath(), "validate-destination")
	if err != nil {
		return err
	}

	body := map[string]any{
		"sourceReaderUUID": pipeline.SourceReaderUUID,
		"destinationUUID":  pipeline.DestinationUUID,
		"specificCfg":      pipeline.DestinationConfig,
		"tables":           pipeline.Tables,
	}

	response, err := makeRequest[validationResponse](ctx, pc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("destination validation failed: %s", response.Error)
	}

	return nil
}
