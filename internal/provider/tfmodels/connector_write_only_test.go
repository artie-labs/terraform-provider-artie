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
		"type": "postgresql",
		"sharedConfig": {"host": "db.example.com", "port": 5432, "user": "artie", "password": "********"}
	}`), &apiConnector))
	postgres := func(password string, version types.Int64) Connector {
		return Connector{
			UUID: types.StringValue("0b9e1c55-6a4f-4a8e-9c35-6d5f2f1b6a01"),
			Type: types.StringValue("postgresql"),
			PostgresConfig: &PostgresSharedConfig{
				Host:              types.StringValue("db.example.com"),
				Port:              types.Int32Value(5432),
				Username:          types.StringValue("artie"),
				Password:          types.StringValue(password),
				PasswordWOVersion: version,
			},
		}
	}
	{
		// The API request gets password_wo from config, while the plan it was built from keeps "" for state
		plan := postgres("", types.Int64Value(1))
		apiData := postgres("", types.Int64Value(1))
		config := postgres("", types.Int64Value(1))
		config.PostgresConfig.PasswordWO = types.StringValue("from-config")

		CopyWriteOnlySecrets(&apiData, config)
		body, diags := apiData.ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "from-config", (*body.SharedConfig)["password"])
		assert.True(t, plan.PostgresConfig.PasswordWO.IsNull())
	}
	{
		// Reading back a connector that uses password_wo keeps the secret out of state and keeps the version
		connector, diags := ConnectorFromAPIModel(apiConnector, postgres("", types.Int64Value(3)))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "", connector.PostgresConfig.Password.ValueString())
		assert.True(t, connector.PostgresConfig.PasswordWO.IsNull())
		assert.Equal(t, int64(3), connector.PostgresConfig.PasswordWOVersion.ValueInt64())
	}
	{
		// A connector that still uses password reads back as before, with no version
		connector, diags := ConnectorFromAPIModel(apiConnector, postgres("secret", types.Int64Null()))
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "secret", connector.PostgresConfig.Password.ValueString())
		assert.True(t, connector.PostgresConfig.PasswordWO.IsNull())
		assert.True(t, connector.PostgresConfig.PasswordWOVersion.IsNull())
	}
	{
		// Without password_wo, the request sends password as before
		body, diags := postgres("secret", types.Int64Null()).ToAPIModel()
		require.False(t, diags.HasError(), diags)
		assert.Equal(t, "secret", (*body.SharedConfig)["password"])
	}
}

func TestSecretMissingAndSet(t *testing.T) {
	{
		// With no write-only value, both match the checks they replaced
		for _, plain := range []types.String{types.StringNull(), types.StringValue(""), types.StringValue("x"), types.StringUnknown()} {
			assert.Equal(t, IsKnownAndEmpty(plain), SecretMissing(plain, types.StringNull()), plain)
			assert.Equal(t, IsKnownAndNonEmpty(plain), SecretSet(plain, types.StringNull()), plain)
		}
	}
	{
		// A write-only value counts
		assert.False(t, SecretMissing(types.StringValue(""), types.StringValue("x")))
		assert.True(t, SecretSet(types.StringValue(""), types.StringValue("x")))
	}
}
