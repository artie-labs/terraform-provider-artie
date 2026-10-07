package artieclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-artie/internal/openapi"
)

func TestJSON200(t *testing.T) {
	{
		// 200 with a JSON body
		tunnel := &openapi.PayloadsSSHTunnel{Name: "tunnel"}
		out, err := JSON200(&openapi.SshTunnelDetailResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: tunnel}, nil)
		require.NoError(t, err)
		assert.Equal(t, tunnel, out)
	}
	{
		// Client error surfaces the API's error message
		_, err := JSON200(&openapi.SshTunnelDetailResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"error":"bad host"}`)}, nil)
		assert.EqualError(t, err, "bad host (HTTP 400)")
	}
	{
		// 200 without a JSON body (e.g. a proxy's HTML page) is not reported as a non-200 status
		_, err := JSON200(&openapi.SshTunnelDetailResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, Body: []byte("<html>")}, nil)
		assert.EqualError(t, err, `artie-client: expected a JSON response body (HTTP 200), got: "<html>"`)
	}
	{
		// Transport error is returned as-is
		_, err := JSON200[openapi.PayloadsSSHTunnel, *openapi.SshTunnelDetailResponse](nil, errors.New("dial failed"))
		assert.EqualError(t, err, "dial failed")
	}
}

func TestCheckResponse(t *testing.T) {
	{
		// 204 succeeds
		assert.NoError(t, CheckResponse(&openapi.SshTunnelDeleteResponse{HTTPResponse: &http.Response{StatusCode: http.StatusNoContent}}, nil))
	}
	{
		// Client error surfaces the API's error message
		err := CheckResponse(&openapi.SshTunnelDeleteResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"error":"ssh tunnel is in use"}`)}, nil)
		assert.EqualError(t, err, "ssh tunnel is in use (HTTP 400)")
	}
	{
		// Transport error is returned as-is
		var resp *openapi.SshTunnelDeleteResponse
		assert.EqualError(t, CheckResponse(resp, errors.New("dial failed")), "dial failed")
	}
}

func TestValidationError(t *testing.T) {
	{
		// 204 means validation passed
		assert.NoError(t, ValidationError(&openapi.PipelineValidateUnsavedSourceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusNoContent}}, nil))
	}
	{
		// 200 with an empty error means validation passed
		assert.NoError(t, ValidationError(&openapi.PipelineValidateUnsavedSourceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &openapi.RouterValidateErrorResponse{}}, nil))
	}
	{
		// 200 with an error message means validation failed
		message := "table public.orders has no primary key"
		err := ValidationError(&openapi.PipelineValidateUnsavedSourceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &openapi.RouterValidateErrorResponse{Error: &message}}, nil)
		assert.EqualError(t, err, message)
	}
	{
		// 200 without a JSON body is not treated as a pass
		err := ValidationError(&openapi.PipelineValidateUnsavedSourceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, Body: []byte("<html>")}, nil)
		assert.EqualError(t, err, `artie-client: expected a JSON response body (HTTP 200), got: "<html>"`)
	}
	{
		// Client error surfaces the API's error message
		err := ValidationError(&openapi.PipelineValidateUnsavedSourceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"error":"source reader not found"}`)}, nil)
		assert.EqualError(t, err, "source reader not found (HTTP 400)")
	}
}
