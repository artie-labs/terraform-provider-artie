package artieclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

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

type Client struct {
	endpoint string
	apiKey   string
	version  string
}

func New(endpoint string, apiKey string, version string) (Client, error) {
	if !strings.HasPrefix(apiKey, "arsk_") {
		return Client{}, fmt.Errorf("artie-client: api key is malformed (should start with arsk_)")
	}

	return Client{endpoint: endpoint, apiKey: apiKey, version: version}, nil
}

func buildError(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("artie-client: not found, request: %q, method: %q", resp.Request.URL.String(), resp.Request.Method)
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

func (c Client) makeRequest(ctx context.Context, method string, path string, body any, out any) error {
	_url, err := url.JoinPath(c.endpoint, path)
	if err != nil {
		return nil
	}

	tflog.Info(ctx, fmt.Sprintf("Making API request: %s %s", method, _url))

	var bodyReader io.Reader
	if body != nil {
		bodyBuff := new(bytes.Buffer)
		if err := json.NewEncoder(bodyBuff).Encode(body); err != nil {
			return fmt.Errorf("artie-client: failed to encode request body: %w", err)
		}
		bodyReader = bodyBuff
	}

	req, err := http.NewRequestWithContext(ctx, method, _url, bodyReader)
	if err != nil {
		return fmt.Errorf("artie-client: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("User-Agent", "terraform-provider-artie/"+c.version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return buildError(resp)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return fmt.Errorf("artie-client: failed to decode response body: %w", err)
		}
	}

	return nil
}

func makeRequest[Out any](ctx context.Context, client Client, method string, path string, body any) (Out, error) {
	respBody := new(Out)
	if err := client.makeRequest(ctx, method, path, body, respBody); err != nil {
		return *new(Out), err
	}
	return *respBody, nil
}

func (c Client) Connectors() ConnectorClient {
	return ConnectorClient{client: c}
}

func (c Client) SSHTunnels() SSHTunnelClient {
	return SSHTunnelClient{client: c}
}

func (c Client) SourceReaders() SourceReaderClient {
	return SourceReaderClient{client: c}
}

func (c Client) Pipelines() PipelineClient {
	return PipelineClient{client: c}
}
