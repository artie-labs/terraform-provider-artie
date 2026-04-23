package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type BaseColumnHashingSalt struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Salt        string `json:"salt,omitempty"`
}

type ColumnHashingSalt struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Salt        string    `json:"salt"`
}

type UpdateColumnHashingSaltRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type columnHashingSaltCreateResponse struct {
	ColumnHashingSalt ColumnHashingSalt `json:"columnHashingSalt"`
	Salt              string            `json:"salt"`
}

type ColumnHashingSaltClient struct {
	client Client
}

func (ColumnHashingSaltClient) basePath() string {
	return "column-hashing-salts"
}

func (cc ColumnHashingSaltClient) Get(ctx context.Context, saltUUID string) (ColumnHashingSalt, error) {
	path, err := url.JoinPath(cc.basePath(), saltUUID)
	if err != nil {
		return ColumnHashingSalt{}, err
	}

	return makeRequest[ColumnHashingSalt](ctx, cc.client, http.MethodGet, path, nil)
}

func (cc ColumnHashingSaltClient) Create(ctx context.Context, salt BaseColumnHashingSalt) (ColumnHashingSalt, error) {
	resp, err := makeRequest[columnHashingSaltCreateResponse](ctx, cc.client, http.MethodPost, cc.basePath(), salt)
	if err != nil {
		return ColumnHashingSalt{}, err
	}

	result := resp.ColumnHashingSalt
	result.Salt = resp.Salt
	return result, nil
}

func (cc ColumnHashingSaltClient) Update(ctx context.Context, saltUUID string, req UpdateColumnHashingSaltRequest) (ColumnHashingSalt, error) {
	path, err := url.JoinPath(cc.basePath(), saltUUID)
	if err != nil {
		return ColumnHashingSalt{}, err
	}

	return makeRequest[ColumnHashingSalt](ctx, cc.client, http.MethodPost, path, req)
}

func (cc ColumnHashingSaltClient) Delete(ctx context.Context, saltUUID string) error {
	path, err := url.JoinPath(cc.basePath(), saltUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, cc.client, http.MethodDelete, path, nil)
	return err
}
