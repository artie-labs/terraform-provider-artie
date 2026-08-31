package artieclient

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/openapi"
)

type SourceReaderClient struct {
	client *openapi.ClientWithResponses
}

func NewSourceReaderClient(client *openapi.ClientWithResponses) SourceReaderClient {
	return SourceReaderClient{client: client}
}

func (sc SourceReaderClient) Get(ctx context.Context, sourceReaderUUID string) (*openapi.PayloadsSourceReader, error) {
	resp, err := sc.client.SourceReaderDetailWithResponse(ctx, sourceReaderUUID)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

func (sc SourceReaderClient) Validate(ctx context.Context, sourceReader openapi.PayloadsSourceReader) error {
	resp, err := sc.client.SourceReaderValidateUnsavedWithResponse(ctx, openapi.RouterSourceReaderValidateUnsavedRequest{
		SourceReader: sourceReader,
	})
	if err != nil {
		return err
	}
	if resp.JSON200 != nil {
		if resp.JSON200.Error != "" {
			return fmt.Errorf("source reader validation failed: %s", resp.JSON200.Error)
		}
		return nil
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}
	return BuildResponseError(resp.StatusCode(), resp.Body)
}

func (sc SourceReaderClient) Create(ctx context.Context, req openapi.RouterSourceReaderCreateRequest) (*openapi.PayloadsSourceReader, error) {
	resp, err := sc.client.SourceReaderCreateWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

func (sc SourceReaderClient) Update(ctx context.Context, uuid string, sourceReader openapi.PayloadsSourceReader) (*openapi.PayloadsSourceReader, error) {
	resp, err := sc.client.SourceReaderUpdateWithResponse(ctx, uuid, sourceReader)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

func (sc SourceReaderClient) Delete(ctx context.Context, sourceReaderUUID string) error {
	resp, err := sc.client.SourceReaderDeleteWithResponse(ctx, sourceReaderUUID)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}

func (sc SourceReaderClient) Deploy(ctx context.Context, sourceReaderUUID string) error {
	resp, err := sc.client.SourceReaderDeployWithResponse(ctx, sourceReaderUUID)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}

func (sc SourceReaderClient) UpdateStatus(ctx context.Context, sourceReaderUUID string, status string) error {
	resp, err := sc.client.SourceReaderUpdateStatusWithResponse(ctx, sourceReaderUUID, openapi.RouterSourceReaderUpdateStatusRequest{
		Status: openapi.EnumsSourceReaderStatus(status),
	})
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}
