package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type SourceReaderSettings struct {
	OneTopicPerSchema               bool   `json:"oneTopicPerSchema"`
	PostgresPublicationNameOverride string `json:"publicationNameOverride"`
	PostgresReplicationSlotOverride string `json:"replicationSlotOverride"`
}

type SourceReaderTable struct {
	Name                     string   `json:"name"`
	Schema                   string   `json:"schema"`
	IsPartitioned            bool     `json:"isPartitioned"`
	ColumnsToExclude         []string `json:"excludeColumns"`
	ColumnsToInclude         []string `json:"includeColumns"`
	ChildPartitionSchemaName string   `json:"childPartitionSchemaName,omitempty"`
}

type BaseSourceReader struct {
	Name          string                       `json:"name"`
	DataPlaneName string                       `json:"dataPlaneName"`
	ConnectorUUID uuid.UUID                    `json:"connectorUUID"`
	IsShared      bool                         `json:"isShared"`
	DatabaseName  string                       `json:"database"`
	ContainerName string                       `json:"containerName"`
	Settings      SourceReaderSettings         `json:"settings"`
	Tables        map[string]SourceReaderTable `json:"tablesConfig"`
}

type SourceReader struct {
	BaseSourceReader
	UUID uuid.UUID `json:"uuid"`
}

type SourceReaderClient struct {
	client Client
}

func (SourceReaderClient) basePath() string {
	return "source-readers"
}

func (sc SourceReaderClient) Get(ctx context.Context, sourceReaderUUID string) (SourceReader, error) {
	path, err := url.JoinPath(sc.basePath(), sourceReaderUUID)
	if err != nil {
		return SourceReader{}, err
	}

	return makeRequest[SourceReader](ctx, sc.client, http.MethodGet, path, nil)
}

func (sc SourceReaderClient) Create(ctx context.Context, sourceReader BaseSourceReader) (SourceReader, error) {
	return makeRequest[SourceReader](ctx, sc.client, http.MethodPost, sc.basePath(), sourceReader)
}

func (sc SourceReaderClient) Update(ctx context.Context, sourceReader SourceReader) (SourceReader, error) {
	path, err := url.JoinPath(sc.basePath(), sourceReader.UUID.String())
	if err != nil {
		return SourceReader{}, err
	}

	return makeRequest[SourceReader](ctx, sc.client, http.MethodPost, path, sourceReader)
}

func (sc SourceReaderClient) Delete(ctx context.Context, sourceReaderUUID string) error {
	path, err := url.JoinPath(sc.basePath(), sourceReaderUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, sc.client, http.MethodDelete, path, nil)
	return err
}

func (sc SourceReaderClient) Validate(ctx context.Context, sourceReader BaseSourceReader) error {
	path, err := url.JoinPath(sc.basePath(), "validate")
	if err != nil {
		return err
	}

	response, err := makeRequest[validationResponse](ctx, sc.client, http.MethodPost, path, sourceReader)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("source reader validation failed: %s", response.Error)
	}

	return nil
}
