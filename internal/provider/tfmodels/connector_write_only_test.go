package tfmodels

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-artie/internal/openapi"
)

func TestConnectorWriteOnlySecrets(t *testing.T) {
	var apiConnector openapi.PayloadsFullConnector
	require.NoError(t, json.Unmarshal([]byte(`{
		"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
		"type": "oracle",
		"sharedConfig": {"host": "db.example.com", "port": 1521, "user": "artie", "password": "********"}
	}`), &apiConnector))
	oracle := func(password string, version types.Int64) Connector {
		return Connector{
			UUID: types.StringValue("0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01"),
			Type: types.StringValue("oracle"),
			OracleConfig: &OracleSharedConfig{
				Host:              types.StringValue("db.example.com"),
				Port:              types.Int32Value(1521),
				Username:          types.StringValue("artie"),
				Password:          types.StringValue(password),
				PasswordWOVersion: version,
			},
		}
	}
	{
		// The API request gets password_wo from config
		apiData := oracle("", types.Int64Value(1))
		config := oracle("", types.Int64Value(1))
		config.OracleConfig.PasswordWO = types.StringValue("from-config")

		CopyWriteOnlySecrets(&apiData, config)
		body, diags := apiData.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "from-config", (*body.SharedConfig)["password"])
	}
	{
		// Reading back a connector that uses password_wo keeps the secret out of state and keeps the version
		connector, diags := ConnectorFromAPIModel(apiConnector, oracle("", types.Int64Value(3)))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "", connector.OracleConfig.Password.ValueString())
		assert.True(t, connector.OracleConfig.PasswordWO.IsNull())
		assert.Equal(t, int64(3), connector.OracleConfig.PasswordWOVersion.ValueInt64())
	}
	{
		// The API returns plaintext secrets to Terraform API keys that can edit connectors, and they stay out of state too
		var plaintext openapi.PayloadsFullConnector
		require.NoError(t, json.Unmarshal([]byte(`{
			"uuid": "0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01",
			"type": "oracle",
			"sharedConfig": {"host": "db.example.com", "port": 1521, "user": "artie", "password": "from-config"}
		}`), &plaintext))
		connector, diags := ConnectorFromAPIModel(plaintext, oracle("", types.Int64Value(3)))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "", connector.OracleConfig.Password.ValueString())

		// A connector that still uses password reads back the API's value
		connector, diags = ConnectorFromAPIModel(plaintext, oracle("secret", types.Int64Null()))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "from-config", connector.OracleConfig.Password.ValueString())
	}
	{
		// A connector that still uses password reads back as before, with no version
		connector, diags := ConnectorFromAPIModel(apiConnector, oracle("secret", types.Int64Null()))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "secret", connector.OracleConfig.Password.ValueString())
		assert.True(t, connector.OracleConfig.PasswordWO.IsNull())
		assert.True(t, connector.OracleConfig.PasswordWOVersion.IsNull())
	}
	{
		// Without password_wo, the request sends password as before
		body, diags := oracle("secret", types.Int64Null()).ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "secret", (*body.SharedConfig)["password"])
	}
}
