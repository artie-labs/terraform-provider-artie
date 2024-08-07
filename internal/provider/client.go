package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"terraform-provider-artie/internal/provider/models"
)

var ErrNotFound = fmt.Errorf("artie-client: not found")

type HttpError struct {
	StatusCode int
	message    string
}

func (he HttpError) Error() string {
	message := he.message
	if len(message) == 0 {
		message = "server returned a non-200 status code"
	}
	return fmt.Sprintf("%s (HTTP %d)", message, he.StatusCode)
}

type ArtieClient struct {
	endpoint string
	apiKey   string
}

func NewClient(endpoint string, apiKey string) (ArtieClient, error) {
	if !strings.HasPrefix(apiKey, "arsk_") {
		return ArtieClient{}, fmt.Errorf("artie-client: api key is malformed (should start with arsk_)")
	}

	return ArtieClient{endpoint: endpoint, apiKey: apiKey}, nil
}

func buildError(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 { // Client errors
		type errorBody struct {
			ErrorMsg string `json:"error"`
		}
		errorResponse := errorBody{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return HttpError{StatusCode: resp.StatusCode, message: errorResponse.ErrorMsg}
		}
	}
	return HttpError{StatusCode: resp.StatusCode}
}

func (ac ArtieClient) makeRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	_url, err := url.JoinPath(ac.endpoint, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, _url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ac.apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, buildError(resp)
	}

	return resp, nil
}

func makeRequest[Out any](ctx context.Context, client ArtieClient, method string, path string, body any) (Out, error) {
	bodyBuf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(bodyBuf).Encode(body); err != nil {
			return *new(Out), fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	resp, err := client.makeRequest(ctx, method, path, bodyBuf)
	if err != nil {
		return *new(Out), fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody := new(Out)
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return *new(Out), fmt.Errorf("failed to decode response body: %w", err)
	}

	return *respBody, nil
}

type DeploymentClient struct {
	client ArtieClient
}

func (DeploymentClient) basePath() string {
	return "deployments"
}

func (ac ArtieClient) Deployments() DeploymentClient {
	return DeploymentClient{client: ac}
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

func (dc DeploymentClient) Update(ctx context.Context, deployment models.DeploymentAPIModel, updateDeploymentOnly bool) (models.DeploymentAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), deployment.UUID)
	if err != nil {
		return models.DeploymentAPIModel{}, err
	}

	body := map[string]any{
		"deploy":           deployment,
		"updateDeployOnly": updateDeploymentOnly,
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
