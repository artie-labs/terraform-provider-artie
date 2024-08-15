package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type Destination struct {
	UUID          uuid.UUID               `json:"uuid"`
	Type          string                  `json:"name"`
	Label         string                  `json:"label"`
	SSHTunnelUUID *uuid.UUID              `json:"sshTunnelUUID"`
	Config        DestinationSharedConfig `json:"sharedConfig"`
}

type DestinationSharedConfig struct {
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

func (dc DestinationClient) Get(ctx context.Context, destinationUUID string) (Destination, error) {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return Destination{}, err
	}
	return makeRequest[Destination](ctx, dc.client, http.MethodGet, path, nil)
}

func (dc DestinationClient) Create(ctx context.Context, type_, label string, config DestinationSharedConfig, sshTunnelUUID *uuid.UUID) (Destination, error) {
	body := map[string]any{
		"type":         type_,
		"label":        label,
		"sharedConfig": config,
	}
	if sshTunnelUUID != nil {
		body["sshTunnelUUID"] = *sshTunnelUUID
	}
	return makeRequest[Destination](ctx, dc.client, http.MethodPost, dc.basePath(), body)
}

func (dc DestinationClient) Update(ctx context.Context, destination Destination) (Destination, error) {
	path, err := url.JoinPath(dc.basePath(), destination.UUID.String())
	if err != nil {
		return Destination{}, err
	}

	return makeRequest[Destination](ctx, dc.client, http.MethodPost, path, destination)
}

func (dc DestinationClient) Delete(ctx context.Context, destinationUUID string) error {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
