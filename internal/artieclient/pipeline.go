package artieclient

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/openapi"
)

type PipelineClient struct {
	client *openapi.ClientWithResponses
}

func NewPipelineClient(client *openapi.ClientWithResponses) PipelineClient {
	return PipelineClient{client: client}
}

func (pc PipelineClient) Get(ctx context.Context, pipelineUUID string) (*openapi.PayloadsFullPipeline, error) {
	resp, err := pc.client.GetPipelinesUuidWithResponse(ctx, pipelineUUID, nil)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return &resp.JSON200.Pipeline, nil
}

func (pc PipelineClient) ValidateSource(ctx context.Context, req openapi.RouterPipelineValidateUnsavedSourceRequest) error {
	resp, err := pc.client.PostPipelinesValidateUnsavedSourceWithResponse(ctx, req)
	if err != nil {
		return err
	}
	if resp.JSON200 != nil {
		if resp.JSON200.Error != nil && *resp.JSON200.Error != "" {
			return fmt.Errorf("source validation failed: %s", *resp.JSON200.Error)
		}
		return nil
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}
	return BuildResponseError(resp.StatusCode(), resp.Body)
}

func (pc PipelineClient) ValidateDestination(ctx context.Context, req openapi.RouterPipelineValidateUnsavedDestinationRequest) error {
	resp, err := pc.client.PostPipelinesValidateUnsavedDestinationWithResponse(ctx, req)
	if err != nil {
		return err
	}
	if resp.JSON200 != nil {
		if resp.JSON200.Error != nil && *resp.JSON200.Error != "" {
			return fmt.Errorf("destination validation failed: %s", *resp.JSON200.Error)
		}
		return nil
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}
	return BuildResponseError(resp.StatusCode(), resp.Body)
}

func (pc PipelineClient) Create(ctx context.Context, req openapi.RouterPipelineCreateRequest) (*openapi.PayloadsFullPipeline, error) {
	resp, err := pc.client.PostPipelinesWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

func (pc PipelineClient) Update(ctx context.Context, uuid string, req openapi.RouterPipelineUpdateRequest) (*openapi.PayloadsFullPipeline, error) {
	resp, err := pc.client.PostPipelinesUuidWithResponse(ctx, uuid, req)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200, nil
}

func (pc PipelineClient) Delete(ctx context.Context, pipelineUUID string) error {
	resp, err := pc.client.DeletePipelinesUuidWithResponse(ctx, pipelineUUID)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}

func (pc PipelineClient) StartPipeline(ctx context.Context, pipelineUUID string) error {
	resp, err := pc.client.PostPipelinesUuidStartWithResponse(ctx, pipelineUUID, openapi.RouterPipelineStartRequest{})
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}
