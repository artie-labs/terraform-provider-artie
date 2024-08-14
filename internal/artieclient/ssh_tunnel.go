package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type SSHTunnel struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Host      string    `json:"address"`
	Port      int32     `json:"port"`
	Username  string    `json:"username"`
	PublicKey string    `json:"publicKey"`
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

func (sc SSHTunnelClient) Create(ctx context.Context, name, host, username string, port int32) (SSHTunnel, error) {
	body := map[string]any{
		"name":     name,
		"address":  host,
		"port":     port,
		"username": username,
	}
	return makeRequest[SSHTunnel](ctx, sc.client, http.MethodPost, sc.basePath(), body)
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
