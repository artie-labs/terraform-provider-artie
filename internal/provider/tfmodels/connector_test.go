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

	connector, diags := ConnectorFromAPIModel(apiConnector)
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
