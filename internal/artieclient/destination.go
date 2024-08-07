package artieclient

import (
	"context"
	"net/http"
	"net/url"
	"terraform-provider-artie/internal/provider/models"
)

type DestinationClient struct {
	client Client
}

func (DestinationClient) basePath() string {
	return "destinations"
}

func (dc DestinationClient) Get(ctx context.Context, destinationUUID string) (models.DestinationAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return models.DestinationAPIModel{}, err
	}
	return makeRequest[models.DestinationAPIModel](ctx, dc.client, http.MethodGet, path, nil)
}

func (dc DestinationClient) Create(ctx context.Context, destination models.DestinationAPIModel) (models.DestinationAPIModel, error) {
	body := map[string]any{
		"name":         destination.Type,
		"label":        destination.Label,
		"sharedConfig": destination.Config,
	}
	if destination.SSHTunnelUUID != nil {
		body["sshTunnelUUID"] = *destination.SSHTunnelUUID
	}
	return makeRequest[models.DestinationAPIModel](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DestinationClient) Update(ctx context.Context, destination models.DestinationAPIModel) (models.DestinationAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), destination.UUID)
	if err != nil {
		return models.DestinationAPIModel{}, err
	}

	return makeRequest[models.DestinationAPIModel](ctx, dc.client, http.MethodPost, path, destination)
}

func (dc DestinationClient) Delete(ctx context.Context, destinationUUID string) error {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
