package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type SourceType string

const (
	MySQL      SourceType = "mysql"
	Oracle     SourceType = "oracle"
	PostgreSQL SourceType = "postgresql"
)

var AllSourceTypes = []string{
	string(MySQL),
	string(Oracle),
	string(PostgreSQL),
}

func SourceTypeFromString(sourceType string) SourceType {
	switch SourceType(sourceType) {
	case MySQL:
		return MySQL
	case Oracle:
		return Oracle
	case PostgreSQL:
		return PostgreSQL
	default:
		panic(fmt.Sprintf("invalid source type: %s", sourceType))
	}
}

type BaseDeployment struct {
	Name                     string            `json:"name"`
	Source                   Source            `json:"source"`
	DestinationUUID          *uuid.UUID        `json:"destinationUUID"`
	DestinationConfig        DestinationConfig `json:"specificDestCfg"`
	SSHTunnelUUID            *uuid.UUID        `json:"sshTunnelUUID"`
	SnowflakeEcoScheduleUUID *uuid.UUID        `json:"snowflakeEcoScheduleUUID"`

	// Advanced settings
	DropDeletedColumns             *bool `json:"dropDeletedColumns"`
	EnableSoftDelete               *bool `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn    *bool `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn *bool `json:"includeDatabaseUpdatedAtColumn"`
	OneTopicPerSchema              *bool `json:"oneTopicPerSchema"`
}

type Deployment struct {
	BaseDeployment
	UUID   uuid.UUID `json:"uuid"`
	Status string    `json:"status"`
}

type deploymentWithAdvSettings struct {
	Deployment
	AdvancedSettings advancedSettings `json:"advancedSettings"`
}

type advancedSettings struct {
	DropDeletedColumns             bool `json:"dropDeletedColumns"`
	EnableSoftDelete               bool `json:"enableSoftDelete"`
	IncludeArtieUpdatedAtColumn    bool `json:"includeArtieUpdatedAtColumn"`
	IncludeDatabaseUpdatedAtColumn bool `json:"includeDatabaseUpdatedAtColumn"`
	OneTopicPerSchema              bool `json:"oneTopicPerSchema"`
}

func unnestAdvSettings(deployment deploymentWithAdvSettings) Deployment {
	deployment.DropDeletedColumns = &deployment.AdvancedSettings.DropDeletedColumns
	deployment.EnableSoftDelete = &deployment.AdvancedSettings.EnableSoftDelete
	deployment.IncludeArtieUpdatedAtColumn = &deployment.AdvancedSettings.IncludeArtieUpdatedAtColumn
	deployment.IncludeDatabaseUpdatedAtColumn = &deployment.AdvancedSettings.IncludeDatabaseUpdatedAtColumn
	deployment.OneTopicPerSchema = &deployment.AdvancedSettings.OneTopicPerSchema
	return deployment.Deployment
}

type Source struct {
	Type   SourceType   `json:"type"`
	Config SourceConfig `json:"config"`
	Tables []Table      `json:"tables"`
}

type SourceConfig struct {
	Host         string `json:"host"`
	SnapshotHost string `json:"snapshotHost"`
	Port         int32  `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Database     string `json:"database"`
	Container    string `json:"containerName,omitempty"`
}

type Table struct {
	UUID                 uuid.UUID `json:"uuid"`
	Name                 string    `json:"name"`
	Schema               string    `json:"schema"`
	EnableHistoryMode    bool      `json:"enableHistoryMode"`
	IndividualDeployment bool      `json:"individualDeployment"`
	IsPartitioned        bool      `json:"isPartitioned"`
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
	return unnestAdvSettings(response.Deployment), nil
}

func (dc DeploymentClient) Create(ctx context.Context, deployment BaseDeployment) (Deployment, error) {
	body := map[string]any{
		"deployment":      deployment,
		"startDeployment": true,
	}
	return makeRequest[Deployment](ctx, dc.client, http.MethodPost, dc.basePath(), body)
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
	return unnestAdvSettings(response.Deployment), nil
}

func (dc DeploymentClient) ValidateSource(ctx context.Context, deployment BaseDeployment) error {
	path, err := url.JoinPath(dc.basePath(), "validate-source")
	if err != nil {
		return err
	}

	body := map[string]any{
		"source":         deployment.Source,
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
