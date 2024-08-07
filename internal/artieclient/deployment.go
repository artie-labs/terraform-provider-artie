package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type Deployment struct {
	UUID                     uuid.UUID         `json:"uuid"`
	Name                     string            `json:"name"`
	Status                   string            `json:"status"`
	Source                   Source            `json:"source"`
	DestinationUUID          *uuid.UUID        `json:"destinationUUID"`
	DestinationConfig        DestinationConfig `json:"uniqueConfig"`
	SSHTunnelUUID            *uuid.UUID        `json:"sshTunnelUUID"`
	SnowflakeEcoScheduleUUID *uuid.UUID        `json:"snowflakeEcoScheduleUUID"`
}

type Source struct {
	Type   string       `json:"name"`
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
	SchemaOverride        string `json:"schemaOverride"`
	UseSameSchemaAsSource bool   `json:"useSameSchemaAsSource"`
	SchemaNamePrefix      string `json:"schemaNamePrefix"`
}

type DeploymentClient struct {
	client Client
}

func (DeploymentClient) basePath() string {
	return "deployments"
}

type deploymentAPIResponse struct {
	Deployment Deployment `json:"deploy"`
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
	return response.Deployment, nil
}

func (dc DeploymentClient) Create(ctx context.Context, sourceType string) (Deployment, error) {
	body := map[string]any{"source": sourceType}
	return makeRequest[Deployment](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DeploymentClient) Update(ctx context.Context, deployment Deployment) (Deployment, error) {
	path, err := url.JoinPath(dc.basePath(), deployment.UUID.String())
	if err != nil {
		return Deployment{}, err
	}

	body := map[string]any{
		"deploy":           deployment,
		"updateDeployOnly": true,
	}

	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return Deployment{}, err
	}
	return response.Deployment, nil
}

func (dc DeploymentClient) Delete(ctx context.Context, deploymentUUID string) error {
	path, err := url.JoinPath(dc.basePath(), deploymentUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
