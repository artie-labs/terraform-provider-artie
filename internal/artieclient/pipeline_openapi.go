package artieclient

import (
	"context"

	"terraform-provider-artie/internal/openapi"
)

type PipelineOpenAPIClient struct {
	client *openapi.ClientWithResponses
}

func NewPipelineOpenAPIClient(client *openapi.ClientWithResponses) PipelineOpenAPIClient {
	return PipelineOpenAPIClient{client: client}
}

func (pc PipelineOpenAPIClient) List(ctx context.Context) ([]openapi.PayloadsLightPipeline, error) {
	resp, err := pc.client.GetPipelinesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return resp.JSON200.Items, nil
}

func (pc PipelineOpenAPIClient) StartPipeline(ctx context.Context, pipelineUUID string) error {
	resp, err := pc.client.PostPipelinesUuidStartWithResponse(ctx, pipelineUUID, openapi.RouterPipelineStartRequest{})
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return BuildResponseError(resp.StatusCode(), resp.Body)
	}
	return nil
}
