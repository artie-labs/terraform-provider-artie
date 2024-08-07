package artieclient

import (
	"context"
	"net/http"
	"net/url"
	"terraform-provider-artie/internal/provider/models"
)

type DeploymentClient struct {
	client Client
}

func (DeploymentClient) basePath() string {
	return "deployments"
}

func (dc DeploymentClient) Get(ctx context.Context, deploymentUUID string) (models.DeploymentAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), deploymentUUID)
	if err != nil {
		return models.DeploymentAPIModel{}, err
	}
	response, err := makeRequest[models.DeploymentAPIResponse](ctx, dc.client, http.MethodGet, path, nil)
	if err != nil {
		return models.DeploymentAPIModel{}, err
	}
	return response.Deployment, nil
}

func (dc DeploymentClient) Create(ctx context.Context, sourceType string) (models.DeploymentAPIModel, error) {
	body := map[string]any{"source": sourceType}
	return makeRequest[models.DeploymentAPIModel](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DeploymentClient) Update(ctx context.Context, deployment models.DeploymentAPIModel) (models.DeploymentAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), deployment.UUID)
	if err != nil {
		return models.DeploymentAPIModel{}, err
	}

	body := map[string]any{
		"deploy":           deployment,
		"updateDeployOnly": true,
	}

	response, err := makeRequest[models.DeploymentAPIResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return models.DeploymentAPIModel{}, err
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
