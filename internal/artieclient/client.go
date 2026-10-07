package artieclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"terraform-provider-artie/internal/openapi"
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

func BuildResponseError(statusCode int, body []byte) error {
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("artie-client: not found (HTTP %d), response: %q", statusCode, string(body))
	} else if statusCode >= 400 && statusCode < 500 {
		type errorBody struct {
			ErrorMsg string `json:"error"`
		}

		var errorResponse errorBody
		if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.ErrorMsg != "" {
			return HttpError{StatusCode: statusCode, message: errorResponse.ErrorMsg}
		}
	}
	return HttpError{StatusCode: statusCode}
}

type openAPIResponse interface {
	StatusCode() int
	GetBody() []byte
}

// JSON200 unwraps a generated OpenAPI client call, returning its 200 JSON body or an error built from the response.
func JSON200[T any, Resp interface {
	openAPIResponse
	GetJSON200() *T
}](resp Resp, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if body := resp.GetJSON200(); body != nil {
		return body, nil
	}
	if resp.StatusCode() == http.StatusOK {
		return nil, fmt.Errorf("artie-client: expected a JSON response body (HTTP 200), got: %q", resp.GetBody())
	}
	return nil, BuildResponseError(resp.StatusCode(), resp.GetBody())
}

// CheckResponse unwraps a generated OpenAPI client call that has no JSON body, returning an error unless it succeeded.
func CheckResponse(resp openAPIResponse, err error) error {
	if err != nil {
		return err
	}
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.GetBody())
	}
	return nil
}

// ValidationError unwraps a generated validate-unsaved call, returning the API's validation message as an error.
func ValidationError[Resp interface {
	openAPIResponse
	GetJSON200() *openapi.RouterValidateErrorResponse
}](resp Resp, err error) error {
	if err := CheckResponse(resp, err); err != nil {
		return err
	}
	body := resp.GetJSON200()
	if body == nil && resp.StatusCode() == http.StatusOK {
		return fmt.Errorf("artie-client: expected a JSON response body (HTTP 200), got: %q", resp.GetBody())
	}
	if body != nil && body.Error != nil && *body.Error != "" {
		return errors.New(*body.Error)
	}
	return nil
}
