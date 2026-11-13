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
		SMBShareConfig: &SMBShareConfig{Password: types.StringValue("real-password"), PrivateLinkUUID: types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")},
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
		connector.PrivateLinkUUID = types.StringValue("")
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
	{
		// The top-level private_link_uuid is sent when smb_share_config leaves it empty
		connector.SMBShareConfig.PrivateLinkUUID = types.StringValue("")
		connector.PrivateLinkUUID = types.StringValue("9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d")
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d", body.PrivateLinkUUID.String())
	}
	{
		// smb_share_config.private_link_uuid wins over the top-level one, which is unknown until apply when it follows it
		connector.SMBShareConfig.PrivateLinkUUID = types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")
		connector.PrivateLinkUUID = types.StringUnknown()
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", body.PrivateLinkUUID.String())
	}
}

func TestSMBShareConnectorPrivateLinkReadBack(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "smb_share",
		"privateLinkUUID": "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d",
		"sharedConfig": {"path": "\\\\fs\\share", "username": "u", "password": "********"}
	}`), &apiConnector))
	{
		// A config that sets smb_share_config.private_link_uuid reads it back there and at the top level
		prior := Connector{Type: types.StringValue("smb_share"), SMBShareConfig: &SMBShareConfig{PrivateLinkUUID: types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")}}
		connector, diags := ConnectorFromAPIModel(apiConnector, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.SMBShareConfig.PrivateLinkUUID.ValueString())
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.PrivateLinkUUID.ValueString())
	}
	{
		// An unknown smb_share_config.private_link_uuid (another resource's uuid on create) reads back the API's value
		prior := Connector{Type: types.StringValue("smb_share"), SMBShareConfig: &SMBShareConfig{PrivateLinkUUID: types.StringUnknown()}}
		connector, diags := ConnectorFromAPIModel(apiConnector, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.SMBShareConfig.PrivateLinkUUID.ValueString())
	}
	{
		// A config that uses the top-level private_link_uuid keeps smb_share_config's empty, so it plans no change
		prior := Connector{Type: types.StringValue("smb_share"), SMBShareConfig: &SMBShareConfig{PrivateLinkUUID: types.StringValue("")}}
		connector, diags := ConnectorFromAPIModel(apiConnector, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "", connector.SMBShareConfig.PrivateLinkUUID.ValueString())
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.PrivateLinkUUID.ValueString())
	}
	{
		// An import has no prior block, so both read back the PrivateLink, as smb_share_config did before
		connector, diags := ConnectorFromAPIModel(apiConnector, Connector{})
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.SMBShareConfig.PrivateLinkUUID.ValueString())
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.PrivateLinkUUID.ValueString())
	}
}

func TestConnectorPrivateLinksAndDefaultDatabase(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "postgresql",
		"privateLinkUUID": "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d",
		"snapshotPrivateLinkUUID": "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d",
		"defaultDatabase": "app",
		"sharedConfig": {"host": "db.example.com", "port": 5432, "user": "artie", "password": "secret"}
	}`), &apiConnector))

	connector, diags := ConnectorFromAPIModel(apiConnector, Connector{})
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", connector.PrivateLinkUUID.ValueString())
	assert.Equal(t, "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d", connector.SnapshotPrivateLinkUUID.ValueString())
	assert.Equal(t, "app", connector.DefaultDatabase.ValueString())
	{
		// The values read back are sent back, so an apply keeps ones set in the UI
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d", body.PrivateLinkUUID.String())
		assert.Equal(t, "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d", body.SnapshotPrivateLinkUUID.String())
		assert.Equal(t, "app", *body.DefaultDatabase)
	}
	{
		// Unknown (create) leaves them out. The API keeps the saved default database but removes the PrivateLinks.
		connector.PrivateLinkUUID = types.StringUnknown()
		connector.SnapshotPrivateLinkUUID = types.StringUnknown()
		connector.DefaultDatabase = types.StringUnknown()
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Nil(t, body.PrivateLinkUUID)
		assert.Nil(t, body.SnapshotPrivateLinkUUID)
		assert.Nil(t, body.DefaultDatabase)
	}
	{
		// "" removes the PrivateLinks and clears the default database
		connector.PrivateLinkUUID = types.StringValue("")
		connector.SnapshotPrivateLinkUUID = types.StringValue("")
		connector.DefaultDatabase = types.StringValue("")
		body, diags := connector.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Nil(t, body.PrivateLinkUUID)
		assert.Nil(t, body.SnapshotPrivateLinkUUID)
		require.NotNil(t, body.DefaultDatabase)
		assert.Equal(t, "", *body.DefaultDatabase)
	}
	{
		// Unset in the API reads back as "", not null
		connector, diags := ConnectorFromAPIModel(openapi.PayloadsFullConnector{Type: "postgresql", SharedConfig: apiConnector.SharedConfig}, Connector{})
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(""), connector.PrivateLinkUUID)
		assert.Equal(t, types.StringValue(""), connector.SnapshotPrivateLinkUUID)
		assert.Equal(t, types.StringValue(""), connector.DefaultDatabase)
	}
}

func TestSMBSharePrivateLinkPlan(t *testing.T) {
	smbShare := types.StringValue("smb_share")
	link := types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")
	state := types.StringValue("9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d")
	{
		// Other types keep the plan
		assert.Equal(t, state, SMBSharePrivateLinkPlan(types.StringValue("postgresql"), types.StringNull(), link, state))
	}
	{
		// A configured top-level value keeps the plan
		assert.Equal(t, state, SMBSharePrivateLinkPlan(smbShare, state, link, state))
	}
	{
		// Follows smb_share_config.private_link_uuid
		assert.Equal(t, link, SMBSharePrivateLinkPlan(smbShare, types.StringNull(), link, state))
	}
	{
		// Removing smb_share_config.private_link_uuid removes the top-level value too
		assert.Equal(t, types.StringValue(""), SMBSharePrivateLinkPlan(smbShare, types.StringNull(), types.StringNull(), state))
	}
	{
		// Unknown smb_share_config.private_link_uuid
		assert.Equal(t, types.StringUnknown(), SMBSharePrivateLinkPlan(smbShare, types.StringNull(), types.StringUnknown(), state))
	}
}

func TestConflictingSMBSharePrivateLinks(t *testing.T) {
	a := types.StringValue("5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d")
	b := types.StringValue("9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d")
	{
		// Different values conflict
		assert.True(t, ConflictingSMBSharePrivateLinks(a, b))
	}
	{
		// The same value doesn't
		assert.False(t, ConflictingSMBSharePrivateLinks(a, a))
	}
	{
		// Only one set doesn't
		assert.False(t, ConflictingSMBSharePrivateLinks(a, types.StringNull()))
		assert.False(t, ConflictingSMBSharePrivateLinks(types.StringNull(), b))
	}
	{
		// Unknown isn't checked yet
		assert.False(t, ConflictingSMBSharePrivateLinks(types.StringUnknown(), b))
	}
	{
		// An explicit "" conflicts with a link in smb_share_config
		assert.True(t, ConflictingSMBSharePrivateLinks(types.StringValue(""), b))
		assert.False(t, ConflictingSMBSharePrivateLinks(types.StringValue(""), types.StringValue("")))
	}
}

func TestConnectorFromAPIModelKeepsHostsRewrittenByPrivateLink(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "postgresql",
		"privateLinkUUID": "5d1f0a2b-8c3e-4f6a-9b7d-1e2f3a4b5c6d",
		"snapshotPrivateLinkUUID": "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d",
		"sharedConfig": {"host": "vpce-1.example.com", "snapshotHost": "vpce-2.example.com", "port": 5432, "user": "artie", "password": "secret"}
	}`), &apiConnector))
	prior := Connector{
		Type: types.StringValue("postgresql"),
		PostgresConfig: &PostgresSharedConfig{
			Host:         types.StringValue("db.example.com"),
			SnapshotHost: types.StringValue("replica.example.com"),
		},
	}
	{
		// Both links set: host and snapshot_host stay as configured
		connector, diags := ConnectorFromAPIModel(apiConnector, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "db.example.com", connector.PostgresConfig.Host.ValueString())
		assert.Equal(t, "replica.example.com", connector.PostgresConfig.SnapshotHost.ValueString())
	}
	{
		// Only the snapshot link set: host is the API's
		onlySnapshot := apiConnector
		onlySnapshot.PrivateLinkUUID = nil
		connector, diags := ConnectorFromAPIModel(onlySnapshot, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "vpce-1.example.com", connector.PostgresConfig.Host.ValueString())
		assert.Equal(t, "replica.example.com", connector.PostgresConfig.SnapshotHost.ValueString())
	}
	{
		// Only the primary link set: snapshot_host is the API's
		onlyPrimary := apiConnector
		onlyPrimary.SnapshotPrivateLinkUUID = nil
		connector, diags := ConnectorFromAPIModel(onlyPrimary, prior)
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "db.example.com", connector.PostgresConfig.Host.ValueString())
		assert.Equal(t, "vpce-2.example.com", connector.PostgresConfig.SnapshotHost.ValueString())
	}
	{
		// An import has no prior, so it keeps the API's hosts
		connector, diags := ConnectorFromAPIModel(apiConnector, Connector{})
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "vpce-1.example.com", connector.PostgresConfig.Host.ValueString())
		assert.Equal(t, "vpce-2.example.com", connector.PostgresConfig.SnapshotHost.ValueString())
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
