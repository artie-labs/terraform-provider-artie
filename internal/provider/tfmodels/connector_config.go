package tfmodels

import (
	"encoding/json"
	"fmt"

	"terraform-provider-artie/internal/openapi"
)

// ConnectorConfig is the shape of a connector's `sharedConfig`, which the OpenAPI spec leaves untyped.
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
	DatabricksClientID            string `json:"clientID"`
	DatabricksClientSecret        string `json:"clientSecret"`
	DatabricksVolume              string `json:"volume"`

	// Dynamo, S3, Iceberg, Keyspaces:
	AWSAccessKeyID     string `json:"awsAccessKeyID"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey"`
	AWSRoleARN         string `json:"awsRoleARN"`
	AWSExternalID      string `json:"awsExternalID"`

	// Dynamo:
	DynamoStreamArn string `json:"streamsArn"`

	// S3, Keyspaces:
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

func (c ConnectorConfig) toAPIModel() (*map[string]any, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to encode connector config: %w", err)
	}

	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to encode connector config: %w", err)
	}
	return &out, nil
}

func connectorConfigFromAPIModel(value *openapi.PayloadsJSONValue) (ConnectorConfig, error) {
	var out ConnectorConfig
	if value == nil {
		return out, nil
	}

	body, err := value.MarshalJSON()
	if err != nil {
		return out, fmt.Errorf("failed to decode connector config: %w", err)
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return out, fmt.Errorf("failed to decode connector config: %w", err)
	}
	return out, nil
}
