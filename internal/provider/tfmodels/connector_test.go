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
		// Unknown (create) leaves them out, so the API keeps the saved default database
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
}
