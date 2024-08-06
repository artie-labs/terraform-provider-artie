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

type ArtieClient struct {
	endpoint string
	apiKey   string
}

func NewClient(endpoint string, apiKey string) (ArtieClient, error) {
	if !strings.HasPrefix(apiKey, "arsk_") {
		return ArtieClient{}, fmt.Errorf("artie client: api key is malformed (should start with arsk_)")
	}

	return ArtieClient{endpoint: endpoint, apiKey: apiKey}, nil
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

	return http.DefaultClient.Do(req)
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

	if resp.StatusCode != http.StatusOK {
		return *new(Out), fmt.Errorf("server returned a non-200 status code (%d)", resp.StatusCode)
	}

	respBody := new(Out)
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return *new(Out), fmt.Errorf("failed to decode response body: %w", err)
	}

	return *respBody, nil
}

type DeploymentClient struct {
	client ArtieClient
}

func (ac ArtieClient) Deployments() DeploymentClient {
	return DeploymentClient{client: ac}
}

func (dc DeploymentClient) Get(ctx context.Context, deploymentUUID string) (models.DeploymentAPIResponse, error) {
	path, err := url.JoinPath("deployments", deploymentUUID)
	if err != nil {
		return models.DeploymentAPIResponse{}, err
	}
	return makeRequest[models.DeploymentAPIResponse](ctx, dc.client, http.MethodGet, path, nil)
}
