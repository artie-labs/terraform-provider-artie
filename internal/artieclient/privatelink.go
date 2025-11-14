package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type BasePrivateLinkConnection struct {
	VpcServiceName string `json:"vpcServiceName"`
	Region         string `json:"region"`
	VpcEndpointID  string `json:"vpcEndpointId"`
}

type PrivateLinkConnection struct {
	BasePrivateLinkConnection
	UUID          uuid.UUID `json:"uuid"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	DnsEntry      string    `json:"dnsEntry"`
	DataPlaneName string    `json:"dataPlaneName"`
}

type PrivateLinkClient struct {
	client Client
}

func (PrivateLinkClient) basePath() string {
	return "privatelink-connections"
}

func (pc PrivateLinkClient) Get(ctx context.Context, plUUID string) (PrivateLinkConnection, error) {
	path, err := url.JoinPath(pc.basePath(), plUUID)
	if err != nil {
		return PrivateLinkConnection{}, err
	}

	return makeRequest[PrivateLinkConnection](ctx, pc.client, http.MethodGet, path, nil)
}

func (pc PrivateLinkClient) Create(ctx context.Context, conn BasePrivateLinkConnection) (PrivateLinkConnection, error) {
	return makeRequest[PrivateLinkConnection](ctx, pc.client, http.MethodPost, pc.basePath(), conn)
}

func (pc PrivateLinkClient) Update(ctx context.Context, conn PrivateLinkConnection) (PrivateLinkConnection, error) {
	path, err := url.JoinPath(pc.basePath(), conn.UUID.String())
	if err != nil {
		return PrivateLinkConnection{}, err
	}

	return makeRequest[PrivateLinkConnection](ctx, pc.client, http.MethodPost, path, conn)
}

func (pc PrivateLinkClient) Delete(ctx context.Context, plUUID string) error {
	path, err := url.JoinPath(pc.basePath(), plUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, pc.client, http.MethodDelete, path, nil)
	return err
}
