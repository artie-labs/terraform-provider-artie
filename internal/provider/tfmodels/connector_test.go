package tfmodels

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-artie/internal/openapi"
)

func TestConnectorAPIModelRoundTrip(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "postgresql",
		"label": "prod db",
		"dataPlaneName": "aws-us-east-1",
		"sshTunnelUUID": "7f3c2a10-1e5b-4c1d-8d2e-3b4a5c6d7e8f",
		"sharedConfig": {"host": "db.example.com", "port": 5432, "user": "artie", "password": "secret"},
		"companyUUID": "00000000-0000-0000-0000-000000000001",
		"environmentUUID": "00000000-0000-0000-0000-000000000002",
		"createdAt": "2026-01-01T00:00:00Z",
		"updatedAt": "2026-01-01T00:00:00Z",
		"isValid": true
	}`), &apiConnector))

	connector, diags := ConnectorFromAPIModel(apiConnector, Connector{})
	require.False(t, diags.HasError(), diags)
	{
		// Response fields and untyped sharedConfig land in the Terraform model
		assert.Equal(t, "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01", connector.UUID.ValueString())
		assert.Equal(t, "prod db", connector.Name.ValueString())
		assert.Equal(t, "aws-us-east-1", connector.DataPlaneName.ValueString())
		assert.Equal(t, "7f3c2a10-1e5b-4c1d-8d2e-3b4a5c6d7e8f", connector.SSHTunnelUUID.ValueString())
		require.NotNil(t, connector.PostgresConfig)
		assert.Equal(t, "db.example.com", connector.PostgresConfig.Host.ValueString())
		assert.Equal(t, int32(5432), connector.PostgresConfig.Port.ValueInt32())
		assert.Equal(t, "artie", connector.PostgresConfig.Username.ValueString())
		assert.Equal(t, "secret", connector.PostgresConfig.Password.ValueString())
	}
	{
		// Update body carries the UUID and the same sharedConfig keys as the old client
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01", *body.Uuid)
		assert.Equal(t, "postgresql", *body.Type)
		assert.Equal(t, "7f3c2a10-1e5b-4c1d-8d2e-3b4a5c6d7e8f", body.SshTunnelUUID.String())
		require.NotNil(t, body.SharedConfig)
		config := *body.SharedConfig
		assert.Equal(t, "db.example.com", config["host"])
		assert.InDelta(t, 5432, config["port"], 0)
		assert.Equal(t, "", config["snapshotHost"])
		assert.NotContains(t, config, "snapshotPort")
	}
	{
		// Create body omits the unknown UUID, an empty data plane, and a cleared SSH tunnel
		connector.UUID = types.StringUnknown()
		connector.DataPlaneName = types.StringValue("")
		connector.SSHTunnelUUID = types.StringValue("")
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Nil(t, body.Uuid)
		assert.Nil(t, body.DataPlaneName)
		assert.Nil(t, body.SshTunnelUUID)
	}
}

func TestConnectorFromAPIModelKeepsMaskedSecrets(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "snowflake",
		"label": "warehouse",
		"sharedConfig": {"accountURL": "https://acme.snowflakecomputing.com", "username": "artie", "password": "", "privateKey": "********"}
	}`), &apiConnector))
	prior := Connector{
		Type: types.StringValue("snowflake"),
		SnowflakeConfig: &SnowflakeSharedConfig{
			AccountURL: types.StringValue("https://acme.snowflakecomputing.com"),
			Username:   types.StringValue("artie"),
			Password:   types.StringValue(""),
			PrivateKey: types.StringValue("real-key"),
		},
	}

	connector, diags := ConnectorFromAPIModel(apiConnector, prior)
	require.False(t, diags.HasError(), diags)
	require.NotNil(t, connector.SnowflakeConfig)
	assert.Equal(t, "real-key", connector.SnowflakeConfig.PrivateKey.ValueString())
	assert.Equal(t, "https://acme.snowflakecomputing.com", connector.SnowflakeConfig.AccountURL.ValueString())
}

func TestSMBShareConnectorAPIModelRoundTrip(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "smb_share",
		"label": "log backups",
		"privateLinkUUID": "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d",
		"sharedConfig": {"path": "\\\\fs-0123.corp.example.com\\share\\backups", "username": "CORP\\svc-artie", "password": "********"}
	}`), &apiConnector))
	prior := Connector{
		Type:           types.StringValue("smb_share"),
		SMBShareConfig: &SMBShareConfig{Password: types.StringValue("real-password")},
	}

	connector, diags := ConnectorFromAPIModel(apiConnector, prior)
	require.False(t, diags.HasError(), diags)
	require.NotNil(t, connector.SMBShareConfig)
	assert.Equal(t, `\\fs-0123.corp.example.com\share\backups`, connector.SMBShareConfig.Path.ValueString())
	assert.Equal(t, `CORP\svc-artie`, connector.SMBShareConfig.Username.ValueString())
	assert.Equal(t, "real-password", connector.SMBShareConfig.Password.ValueString())
	assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.SMBShareConfig.PrivateLinkUUID.ValueString())

	{
		// The PrivateLink goes on the connector, not in sharedConfig
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "smb_share", *body.Type)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", body.PrivateLinkUUID.String())
		require.NotNil(t, body.SharedConfig)
		config := *body.SharedConfig
		assert.Equal(t, `\\fs-0123.corp.example.com\share\backups`, config["path"])
		assert.Equal(t, `CORP\svc-artie`, config["username"])
		assert.Equal(t, "real-password", config["password"])
	}
	{
		// No PrivateLink
		connector.SMBShareConfig.PrivateLinkUUID = types.StringValue("")
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Nil(t, body.PrivateLinkUUID)
	}
	{
		// A stray smb_share_config block on another connector type doesn't send its PrivateLink
		other := connector
		other.Type = types.StringValue("postgresql")
		other.PostgresConfig = &PostgresSharedConfig{}
		other.SMBShareConfig = &SMBShareConfig{PrivateLinkUUID: types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")}
		body, diags := other.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Nil(t, body.PrivateLinkUUID)
	}
}

// roundTripSharedConfig reads sharedConfig into a connector of connectorType, then returns that connector and the
// sharedConfig it sends back.
func roundTripSharedConfig(t *testing.T, connectorType string, sharedConfig string, prior Connector) (Connector, map[string]any) {
	t.Helper()
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "`+connectorType+`",
		"label": "test",
		"sharedConfig": `+sharedConfig+`
	}`), &apiConnector))

	connector, diags := ConnectorFromAPIModel(apiConnector, prior)
	require.False(t, diags.HasError(), diags)

	body, diags := connector.ToAPIModel()
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, connectorType, *body.Type)
	require.NotNil(t, body.SharedConfig)
	return connector, *body.SharedConfig
}

func TestIcebergRESTTokenAPIModelRoundTrip(t *testing.T) {
	// A masked token is kept from prior and sent back under the same key
	prior := Connector{Type: types.StringValue("iceberg"), IcebergConfig: &IcebergSharedConfig{Token: types.StringValue("real-token")}}
	connector, config := roundTripSharedConfig(t, "iceberg", `{"provider": "rest", "uri": "https://catalog.example.com", "token": "********", "warehouse": "wh"}`, prior)
	require.NotNil(t, connector.IcebergConfig)
	assert.Equal(t, "real-token", connector.IcebergConfig.Token.ValueString())
	assert.Equal(t, "real-token", config["token"])
}

func TestNewConnectorTypesAPIModelRoundTrip(t *testing.T) {
	{
		// Delta takes the same keys as S3
		prior := Connector{Type: types.StringValue("delta"), DeltaConfig: &S3SharedConfig{SecretAccessKey: types.StringValue("real-secret")}}
		connector, config := roundTripSharedConfig(t, "delta", `{"awsAccessKeyID": "AKIA123", "awsSecretAccessKey": "********", "awsRegion": "us-west-2", "awsRoleARN": "", "awsExternalID": ""}`, prior)
		require.NotNil(t, connector.DeltaConfig)
		assert.Equal(t, "AKIA123", connector.DeltaConfig.AccessKeyID.ValueString())
		assert.Equal(t, "real-secret", connector.DeltaConfig.SecretAccessKey.ValueString())
		assert.Equal(t, "us-west-2", connector.DeltaConfig.Region.ValueString())

		assert.Equal(t, "AKIA123", config["awsAccessKeyID"])
		assert.Equal(t, "real-secret", config["awsSecretAccessKey"])
		assert.Equal(t, "us-west-2", config["awsRegion"])
	}
	{
		// MotherDuck
		prior := Connector{Type: types.StringValue("motherduck"), MotherDuckConfig: &MotherDuckSharedConfig{Token: types.StringValue("real-token")}}
		connector, config := roundTripSharedConfig(t, "motherduck", `{"token": "********"}`, prior)
		require.NotNil(t, connector.MotherDuckConfig)
		assert.Equal(t, "real-token", connector.MotherDuckConfig.Token.ValueString())
		assert.Equal(t, "real-token", config["token"])
	}
	{
		// PlanetScale: the API forces tlsMode, so it's neither read nor sent
		connector, config := roundTripSharedConfig(t, "planetscale", `{"host": "aws.connect.psdb.cloud", "port": 3306, "user": "artie", "password": "secret", "tlsMode": "true"}`, Connector{})
		require.NotNil(t, connector.PlanetScaleConfig)
		assert.Equal(t, "aws.connect.psdb.cloud", connector.PlanetScaleConfig.Host.ValueString())
		assert.Equal(t, int32(3306), connector.PlanetScaleConfig.Port.ValueInt32())
		assert.Equal(t, "artie", connector.PlanetScaleConfig.Username.ValueString())
		assert.Equal(t, "secret", connector.PlanetScaleConfig.Password.ValueString())

		assert.Equal(t, "aws.connect.psdb.cloud", config["host"])
		assert.InDelta(t, 3306, config["port"], 0)
		assert.Equal(t, "artie", config["user"])
		assert.Equal(t, "secret", config["password"])
		assert.NotContains(t, config, "tlsMode")
	}
	{
		// Kafka: masked password and client key are kept from prior
		prior := Connector{Type: types.StringValue("kafka"), KafkaConfig: &KafkaSharedConfig{Password: types.StringValue("real-password"), ClientKey: types.StringValue("real-key")}}
		connector, config := roundTripSharedConfig(t, "kafka", `{"bootstrapServers": "b1:9092,b2:9092", "topicPrefix": "artie", "saslMechanism": "SCRAM-SHA-512", "username": "artie", "password": "********", "caCert": "ca", "clientCert": "cert", "clientKey": "********", "createTopics": true}`, prior)
		require.NotNil(t, connector.KafkaConfig)
		assert.Equal(t, "b1:9092,b2:9092", connector.KafkaConfig.BootstrapServers.ValueString())
		assert.Equal(t, "artie", connector.KafkaConfig.TopicPrefix.ValueString())
		assert.Equal(t, "SCRAM-SHA-512", connector.KafkaConfig.SASLMechanism.ValueString())
		assert.Equal(t, "artie", connector.KafkaConfig.Username.ValueString())
		assert.Equal(t, "real-password", connector.KafkaConfig.Password.ValueString())
		assert.Equal(t, "", connector.KafkaConfig.RoleARN.ValueString())
		assert.False(t, connector.KafkaConfig.DisableTLS.ValueBool())
		assert.Equal(t, "ca", connector.KafkaConfig.CACert.ValueString())
		assert.Equal(t, "cert", connector.KafkaConfig.ClientCert.ValueString())
		assert.Equal(t, "real-key", connector.KafkaConfig.ClientKey.ValueString())
		assert.True(t, connector.KafkaConfig.CreateTopics.ValueBool())

		assert.Equal(t, "b1:9092,b2:9092", config["bootstrapServers"])
		assert.Equal(t, "artie", config["topicPrefix"])
		assert.Equal(t, "SCRAM-SHA-512", config["saslMechanism"])
		assert.Equal(t, "artie", config["username"])
		assert.Equal(t, "real-password", config["password"])
		assert.Equal(t, "ca", config["caCert"])
		assert.Equal(t, "cert", config["clientCert"])
		assert.Equal(t, "real-key", config["clientKey"])
		assert.Equal(t, true, config["createTopics"])
		assert.NotContains(t, config, "disableTLS")
	}
	{
		// Kafka with AWS-MSK-IAM
		connector, config := roundTripSharedConfig(t, "kafka", `{"bootstrapServers": "b1:9098", "topicPrefix": "artie", "saslMechanism": "AWS-MSK-IAM", "awsRoleARN": "arn:aws:iam::123:role/artie", "awsExternalID": "ext"}`, Connector{})
		require.NotNil(t, connector.KafkaConfig)
		assert.Equal(t, "arn:aws:iam::123:role/artie", connector.KafkaConfig.RoleARN.ValueString())
		assert.Equal(t, "ext", connector.KafkaConfig.ExternalID.ValueString())
		assert.Equal(t, "", connector.KafkaConfig.Username.ValueString())

		assert.Equal(t, "arn:aws:iam::123:role/artie", config["awsRoleARN"])
		assert.Equal(t, "ext", config["awsExternalID"])
	}
	{
		// Redis
		connector, config := roundTripSharedConfig(t, "redis", `{"host": "redis.example.com", "port": 6379, "username": "artie", "password": "secret", "database": 2, "tls": true, "clusterMode": false, "maxStreamLength": 10000}`, Connector{})
		require.NotNil(t, connector.RedisConfig)
		assert.Equal(t, "redis.example.com", connector.RedisConfig.Host.ValueString())
		assert.Equal(t, int32(6379), connector.RedisConfig.Port.ValueInt32())
		assert.Equal(t, "artie", connector.RedisConfig.Username.ValueString())
		assert.Equal(t, "secret", connector.RedisConfig.Password.ValueString())
		assert.Equal(t, int32(2), connector.RedisConfig.Database.ValueInt32())
		assert.True(t, connector.RedisConfig.TLS.ValueBool())
		assert.False(t, connector.RedisConfig.ClusterMode.ValueBool())
		assert.Equal(t, int64(10000), connector.RedisConfig.MaxStreamLength.ValueInt64())

		assert.Equal(t, "redis.example.com", config["host"])
		assert.InDelta(t, 6379, config["port"], 0)
		assert.Equal(t, "artie", config["username"])
		assert.Equal(t, "secret", config["password"])
		assert.InDelta(t, 2, config["database"], 0)
		assert.Equal(t, true, config["tls"])
		assert.NotContains(t, config, "clusterMode")
		assert.InDelta(t, 10000, config["maxStreamLength"], 0)
	}
	{
		// Redis database 0 is left out, which the API reads as 0
		connector, config := roundTripSharedConfig(t, "redis", `{"host": "redis.example.com", "port": 6379, "database": 0}`, Connector{})
		require.NotNil(t, connector.RedisConfig)
		assert.Equal(t, int32(0), connector.RedisConfig.Database.ValueInt32())
		assert.NotContains(t, config, "database")
	}
	{
		// DocumentDB
		connector, config := roundTripSharedConfig(t, "documentdb", `{"host": "docdb.example.com", "port": 27017, "user": "artie", "password": "secret", "disableSSL": true}`, Connector{})
		require.NotNil(t, connector.DocumentDBConfig)
		assert.Equal(t, "docdb.example.com", connector.DocumentDBConfig.Host.ValueString())
		assert.Equal(t, int32(27017), connector.DocumentDBConfig.Port.ValueInt32())
		assert.Equal(t, "artie", connector.DocumentDBConfig.Username.ValueString())
		assert.Equal(t, "secret", connector.DocumentDBConfig.Password.ValueString())
		assert.True(t, connector.DocumentDBConfig.DisableSSL.ValueBool())

		assert.Equal(t, "docdb.example.com", config["host"])
		assert.InDelta(t, 27017, config["port"], 0)
		assert.Equal(t, "artie", config["user"])
		assert.Equal(t, "secret", config["password"])
		assert.Equal(t, true, config["disableSSL"])
	}
	{
		// DocumentDB: a masked password is kept from prior, and SSL left on omits disableSSL, which the API reads as false
		prior := Connector{Type: types.StringValue("documentdb"), DocumentDBConfig: &DocumentDBSharedConfig{Password: types.StringValue("real-password")}}
		connector, config := roundTripSharedConfig(t, "documentdb", `{"host": "docdb.example.com", "port": 27017, "user": "artie", "password": "********"}`, prior)
		require.NotNil(t, connector.DocumentDBConfig)
		assert.Equal(t, "real-password", connector.DocumentDBConfig.Password.ValueString())
		assert.False(t, connector.DocumentDBConfig.DisableSSL.ValueBool())

		assert.Equal(t, "real-password", config["password"])
		assert.NotContains(t, config, "disableSSL")
	}
}
