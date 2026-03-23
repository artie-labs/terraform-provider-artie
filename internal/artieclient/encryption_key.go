package artieclient

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type BaseEncryptionKey struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	KMSKeyUUID  *uuid.UUID `json:"kmsKeyUUID,omitempty"`
}

type EncryptionKey struct {
	UUID        uuid.UUID  `json:"uuid"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	KMSKeyUUID  *uuid.UUID `json:"kmsKeyUUID"`
	Key         string     `json:"key"`
}

type encryptionKeyCreateResponse struct {
	EncryptionKey EncryptionKey `json:"encryptionKey"`
	Key           string        `json:"key"`
}

type EncryptionKeyClient struct {
	client Client
}

func (EncryptionKeyClient) basePath() string {
	return "encryption-keys"
}

func (ec EncryptionKeyClient) Get(ctx context.Context, encryptionKeyUUID string) (EncryptionKey, error) {
	path, err := url.JoinPath(ec.basePath(), encryptionKeyUUID)
	if err != nil {
		return EncryptionKey{}, err
	}

	return makeRequest[EncryptionKey](ctx, ec.client, http.MethodGet, path, nil)
}

func (ec EncryptionKeyClient) Create(ctx context.Context, encryptionKey BaseEncryptionKey) (EncryptionKey, error) {
	resp, err := makeRequest[encryptionKeyCreateResponse](ctx, ec.client, http.MethodPost, ec.basePath(), encryptionKey)
	if err != nil {
		return EncryptionKey{}, err
	}

	result := resp.EncryptionKey
	result.Key = resp.Key
	return result, nil
}

func (ec EncryptionKeyClient) Update(ctx context.Context, encryptionKeyUUID string, body map[string]any) (EncryptionKey, error) {
	path, err := url.JoinPath(ec.basePath(), encryptionKeyUUID)
	if err != nil {
		return EncryptionKey{}, err
	}

	return makeRequest[EncryptionKey](ctx, ec.client, http.MethodPost, path, body)
}

func (ec EncryptionKeyClient) Delete(ctx context.Context, encryptionKeyUUID string) error {
	path, err := url.JoinPath(ec.basePath(), encryptionKeyUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, ec.client, http.MethodDelete, path, nil)
	return err
}
