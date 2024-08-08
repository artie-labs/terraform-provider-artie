package artieclient

import (
	"context"
	"fmt"
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
		"startDeploy":      true,
	}

	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return Deployment{}, err
	}
	return response.Deployment, nil
}

func (dc DeploymentClient) ValidateSource(ctx context.Context, deployment Deployment) error {
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

func (dc DeploymentClient) ValidateDestination(ctx context.Context, deployment Deployment) error {
	path, err := url.JoinPath(dc.basePath(), "validate-destination")
	if err != nil {
		return err
	}

	body := map[string]any{
		"destinationUUID": deployment.DestinationUUID,
		"uniqueCfg":       deployment.DestinationConfig,
		"tables":          deployment.Source.Tables,
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
