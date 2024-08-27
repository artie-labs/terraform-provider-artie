package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type BaseSSHTunnel struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int32  `json:"port"`
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
}

type SSHTunnel struct {
	BaseSSHTunnel
	UUID uuid.UUID `json:"uuid"`
}

type SSHTunnelClient struct {
	client Client
}

func (SSHTunnelClient) basePath() string {
	return "ssh-tunnels"
}

func (sc SSHTunnelClient) Get(ctx context.Context, sshTunnelUUID string) (SSHTunnel, error) {
	path, err := url.JoinPath(sc.basePath(), sshTunnelUUID)
	if err != nil {
		return SSHTunnel{}, err
	}

	return makeRequest[SSHTunnel](ctx, sc.client, http.MethodGet, path, nil)
}

func (sc SSHTunnelClient) Create(ctx context.Context, sshTunnel BaseSSHTunnel) (SSHTunnel, error) {
	return makeRequest[SSHTunnel](ctx, sc.client, http.MethodPost, sc.basePath(), sshTunnel)
}

func (sc SSHTunnelClient) Update(ctx context.Context, sshTunnel SSHTunnel) (SSHTunnel, error) {
	path, err := url.JoinPath(sc.basePath(), sshTunnel.UUID.String())
	if err != nil {
		return SSHTunnel{}, err
	}

	return makeRequest[SSHTunnel](ctx, sc.client, http.MethodPost, path, sshTunnel)
}

func (sc SSHTunnelClient) Delete(ctx context.Context, sshTunnelUUID string) error {
	path, err := url.JoinPath(sc.basePath(), sshTunnelUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, sc.client, http.MethodDelete, path, nil)
	return err
}
