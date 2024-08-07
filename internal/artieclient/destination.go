package artieclient

import (
	"context"
	"net/http"
	"net/url"
)

type DestinationAPIModel struct {
	UUID          string                          `json:"uuid"`
	Type          string                          `json:"name"`
	Label         string                          `json:"label"`
	SSHTunnelUUID *string                         `json:"sshTunnelUUID"`
	Config        DestinationSharedConfigAPIModel `json:"sharedConfig"`
}

type DestinationSharedConfigAPIModel struct {
	Host                string `json:"host"`
	Port                int32  `json:"port"`
	Endpoint            string `json:"endpoint"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	GCPProjectID        string `json:"projectID"`
	GCPLocation         string `json:"location"`
	GCPCredentialsData  string `json:"credentialsData"`
	SnowflakeAccountURL string `json:"accountURL"`
	SnowflakeVirtualDWH string `json:"virtualDWH"`
	SnowflakePrivateKey string `json:"privateKey"`
}

type DestinationClient struct {
	client Client
}

func (DestinationClient) basePath() string {
	return "destinations"
}

func (dc DestinationClient) Get(ctx context.Context, destinationUUID string) (DestinationAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return DestinationAPIModel{}, err
	}
	return makeRequest[DestinationAPIModel](ctx, dc.client, http.MethodGet, path, nil)
}

func (dc DestinationClient) Create(ctx context.Context, destination DestinationAPIModel) (DestinationAPIModel, error) {
	body := map[string]any{
		"name":         destination.Type,
		"label":        destination.Label,
		"sharedConfig": destination.Config,
	}
	if destination.SSHTunnelUUID != nil {
		body["sshTunnelUUID"] = *destination.SSHTunnelUUID
	}
	return makeRequest[DestinationAPIModel](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DestinationClient) Update(ctx context.Context, destination DestinationAPIModel) (DestinationAPIModel, error) {
	path, err := url.JoinPath(dc.basePath(), destination.UUID)
	if err != nil {
		return DestinationAPIModel{}, err
	}

	return makeRequest[DestinationAPIModel](ctx, dc.client, http.MethodPost, path, destination)
}

func (dc DestinationClient) Delete(ctx context.Context, destinationUUID string) error {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
