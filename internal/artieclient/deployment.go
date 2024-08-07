package artieclient

import (
	"context"
	"net/http"
	"net/url"
)

type DeploymentAPIModel struct {
	UUID                     string                    `json:"uuid"`
	Name                     string                    `json:"name"`
	Status                   string                    `json:"status"`
	Source                   SourceAPIModel            `json:"source"`
	DestinationUUID          string                    `json:"destinationUUID"`
	DestinationConfig        DestinationConfigAPIModel `json:"uniqueConfig"`
	SSHTunnelUUID            *string                   `json:"sshTunnelUUID"`
	SnowflakeEcoScheduleUUID *string                   `json:"snowflakeEcoScheduleUUID"`
}

type SourceAPIModel struct {
	Type   string               `json:"name"`
	Config SourceConfigAPIModel `json:"config"`
	Tables []TableAPIModel      `json:"tables"`
}

type SourceConfigAPIModel struct {
	Host         string `json:"host"`
	SnapshotHost string `json:"snapshotHost"`
	Port         int32  `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	Database     string `json:"database"`
}

type TableAPIModel struct {
	UUID                 string `json:"uuid"`
	Name                 string `json:"name"`
	Schema               string `json:"schema"`
	EnableHistoryMode    bool   `json:"enableHistoryMode"`
	IndividualDeployment bool   `json:"individualDeployment"`
	IsPartitioned        bool   `json:"isPartitioned"`
}

type DestinationConfigAPIModel struct {
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
	Deployment DeploymentAPIModel `json:"deploy"`
}

func (dc DeploymentClient) Get(ctx context.Context, deploymentUUID string) (DeploymentAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), deploymentUUID)
	if err != nil {
		return DeploymentAPIModel{}, err
	}
	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodGet, path, nil)
	if err != nil {
		return DeploymentAPIModel{}, err
	}
	return response.Deployment, nil
}

func (dc DeploymentClient) Create(ctx context.Context, sourceType string) (DeploymentAPIModel, error) {
	body := map[string]any{"source": sourceType}
	return makeRequest[DeploymentAPIModel](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DeploymentClient) Update(ctx context.Context, deployment DeploymentAPIModel) (DeploymentAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), deployment.UUID)
	if err != nil {
		return DeploymentAPIModel{}, err
	}

	body := map[string]any{
		"deploy":           deployment,
		"updateDeployOnly": true,
	}

	response, err := makeRequest[deploymentAPIResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return DeploymentAPIModel{}, err
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
