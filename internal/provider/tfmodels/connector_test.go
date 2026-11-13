package tfmodels

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		// ClickHouse: the API returns a masked password and a database name under the key Redis uses for its index
		prior := Connector{Type: types.StringValue("clickhouse"), ClickHouseConfig: &ClickHouseSharedConfig{Password: types.StringValue("real-password")}}
		connector, config := roundTripSharedConfig(t, "clickhouse", `{"addresses": ["ch1.example.com:9440", "ch2.example.com:9440"], "username": "artie", "password": "********", "isInsecure": true, "database": ""}`, prior)
		require.NotNil(t, connector.ClickHouseConfig)
		assert.Equal(t, types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ch1.example.com:9440"), types.StringValue("ch2.example.com:9440")}), connector.ClickHouseConfig.Addresses)
		assert.Equal(t, "artie", connector.ClickHouseConfig.Username.ValueString())
		assert.Equal(t, "real-password", connector.ClickHouseConfig.Password.ValueString())
		assert.True(t, connector.ClickHouseConfig.IsInsecure.ValueBool())

		assert.Equal(t, []any{"ch1.example.com:9440", "ch2.example.com:9440"}, config["addresses"])
		assert.Equal(t, "artie", config["username"])
		assert.Equal(t, "real-password", config["password"])
		assert.Equal(t, true, config["isInsecure"])
		// A number here would fail to decode into ClickHouse's string database field
		assert.NotContains(t, config, "database")
	}
	{
		// An unknown address list, e.g. from another resource's output, converts without panicking
		config := ClickHouseSharedConfig{Addresses: types.ListUnknown(types.StringType)}.ToAPIModel()
		assert.Empty(t, config.ClickHouseAddresses)
	}
}
