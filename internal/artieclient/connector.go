package artieclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type BaseConnector struct {
	Type          ConnectorType   `json:"type"`
	Label         string          `json:"label"`
	DataPlaneName string          `json:"dataPlaneName"`
	SSHTunnelUUID *uuid.UUID      `json:"sshTunnelUUID"`
	Config        ConnectorConfig `json:"sharedConfig"`
}
type Connector struct {
	BaseConnector
	UUID uuid.UUID `json:"uuid"`
}

type ConnectorConfig struct {
	Host         string `json:"host"`
	SnapshotHost string `json:"snapshotHost"`
	SnapshotPort int32  `json:"snapshotPort,omitempty"`
	Port         int32  `json:"port"`
	Endpoint     string `json:"endpoint"`
	User         string `json:"user"`
	Username     string `json:"username"`
	Password     string `json:"password"`

	// MySQL:
	MySQLTLSMode string `json:"tlsMode,omitempty"`

	// BigQuery:
	GCPProjectID       string `json:"projectID"`
	GCPLocation        string `json:"location"`
	GCPCredentialsData string `json:"credentialsData"`

	// Snowflake:
	SnowflakeAccountIdentifier string `json:"accountIdentifier"`
	SnowflakeAccountURL        string `json:"accountURL"`
	SnowflakeVirtualDWH        string `json:"virtualDWH"`
	SnowflakePrivateKey        string `json:"privateKey"`

	// Databricks:
	DatabricksHttpPath            string `json:"httpPath"`
	DatabricksPersonalAccessToken string `json:"personalAccessToken"`
	DatabricksVolume              string `json:"volume"`

	// Dynamo, S3, Iceberg:
	AWSAccessKeyID     string `json:"awsAccessKeyID"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey"`

	// Dynamo:
	DynamoStreamArn string `json:"streamsArn"`

	// S3
	AWSRegion string `json:"awsRegion"`

	// Iceberg:
	IcebergProvider  string `json:"provider"`
	IcebergBucketARN string `json:"bucketARN"`
	IcebergRegion    string `json:"region,omitempty"`

	// Iceberg REST Catalog:
	IcebergURI        string `json:"uri,omitempty"`
	IcebergToken      string `json:"token,omitempty"`
	IcebergCredential string `json:"credential,omitempty"`
	IcebergAuthURI    string `json:"authURI,omitempty"`
	IcebergScope      string `json:"scope,omitempty"`
	IcebergWarehouse  string `json:"warehouse,omitempty"`
	IcebergPrefix     string `json:"prefix,omitempty"`
}

type validationResponse struct {
	Error string `json:"error"`
}

type ConnectorClient struct {
	client Client
}

func (ConnectorClient) basePath() string {
	return "connectors"
}

func (c ConnectorClient) Get(ctx context.Context, connectorUUID string) (Connector, error) {
	path, err := url.JoinPath(c.basePath(), connectorUUID)
	if err != nil {
		return Connector{}, err
	}
	return makeRequest[Connector](ctx, c.client, http.MethodGet, path, nil)
}

func (c ConnectorClient) Create(ctx context.Context, connector BaseConnector) (Connector, error) {
	body := map[string]any{
		"type":          connector.Type,
		"label":         connector.Label,
		"sharedConfig":  connector.Config,
		"dataPlaneName": connector.DataPlaneName,
		"sshTunnelUUID": connector.SSHTunnelUUID,
	}
	return makeRequest[Connector](ctx, c.client, http.MethodPost, c.basePath(), body)
}

func (c ConnectorClient) Update(ctx context.Context, connector Connector) (Connector, error) {
	path, err := url.JoinPath(c.basePath(), connector.UUID.String())
	if err != nil {
		return Connector{}, err
	}

	return makeRequest[Connector](ctx, c.client, http.MethodPost, path, connector)
}

func (c ConnectorClient) TestConnection(ctx context.Context, connector BaseConnector) error {
	path, err := url.JoinPath(c.basePath(), "ping")
	if err != nil {
		return err
	}

	body := map[string]any{
		"type":          connector.Type,
		"sharedConfig":  connector.Config,
		"dataPlaneName": connector.DataPlaneName,
		"sshTunnelUUID": connector.SSHTunnelUUID,
	}

	response, err := makeRequest[validationResponse](ctx, c.client, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("failed to connect to destination: %s", response.Error)
	}

	return nil
}

func (c ConnectorClient) Delete(ctx context.Context, connectorUUID string) error {
	path, err := url.JoinPath(c.basePath(), connectorUUID)
	if err != nil {
		return err
	}

	_, err = makeRequest[any](ctx, c.client, http.MethodDelete, path, nil)
	return err
}
