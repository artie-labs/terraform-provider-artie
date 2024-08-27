package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type DestinationType string

const (
	BigQuery  DestinationType = "bigquery"
	Redshift  DestinationType = "redshift"
	Snowflake DestinationType = "snowflake"
)

var AllDestinationTypes = []string{
	string(BigQuery),
	string(Redshift),
	string(Snowflake),
}

func DestinationTypeFromString(destType string) DestinationType {
	switch DestinationType(destType) {
	case Snowflake:
		return Snowflake
	case BigQuery:
		return BigQuery
	case Redshift:
		return Redshift
	default:
		panic(fmt.Sprintf("invalid destination type: %s", destType))
	}
}

type BaseDestination struct {
	Type          DestinationType         `json:"type"`
	Label         string                  `json:"label"`
	SSHTunnelUUID *uuid.UUID              `json:"sshTunnelUUID"`
	Config        DestinationSharedConfig `json:"sharedConfig"`
}
type Destination struct {
	BaseDestination
	UUID uuid.UUID `json:"uuid"`
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

func (dc DestinationClient) Create(ctx context.Context, destination BaseDestination) (Destination, error) {
	body := map[string]any{
		"type":          destination.Type,
		"label":         destination.Label,
		"sharedConfig":  destination.Config,
		"sshTunnelUUID": destination.SSHTunnelUUID,
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

func (dc DestinationClient) TestConnection(ctx context.Context, destination BaseDestination) error {
	path, err := url.JoinPath(dc.basePath(), "ping")
	if err != nil {
		return err
	}

	body := map[string]any{
		"type":          destination.Type,
		"sharedConfig":  destination.Config,
		"sshTunnelUUID": destination.SSHTunnelUUID,
	}

	response, err := makeRequest[validationResponse](ctx, dc.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("failed to connect to destination: %s", response.Error)
	}

	return nil
}

func (dc DestinationClient) Delete(ctx context.Context, destinationUUID string) error {
	path, err := url.JoinPath(dc.basePath(), destinationUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, dc.client, http.MethodDelete, path, nil)
	return err
}
