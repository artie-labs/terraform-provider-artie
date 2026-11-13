package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/provider/tfmodels"
)

func TestValidateKafkaConfig(t *testing.T) {
	kafkaConfig := func(mechanism, username, password, clientCert, clientKey string) tfmodels.KafkaSharedConfig {
		return tfmodels.KafkaSharedConfig{
			SASLMechanism: types.StringValue(mechanism),
			Username:      types.StringValue(username),
			Password:      types.StringValue(password),
			ClientCert:    types.StringValue(clientCert),
			ClientKey:     types.StringValue(clientKey),
		}
	}
	{
		// SCRAM with a username and password
		assert.False(t, validateKafkaConfig(kafkaConfig("SCRAM-SHA-512", "artie", "secret", "", "")).HasError())
	}
	{
		// No SASL and no credentials
		assert.False(t, validateKafkaConfig(kafkaConfig("", "", "", "", "")).HasError())
	}
	{
		// A password without a mechanism would be dropped by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "", "secret", "", "")).HasError())
	}
	{
		// A username without a mechanism is rejected by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "artie", "", "", "")).HasError())
	}
	{
		// AWS-MSK-IAM doesn't use a username or password, so the API drops them
		assert.True(t, validateKafkaConfig(kafkaConfig("AWS-MSK-IAM", "artie", "secret", "", "")).HasError())
	}
	{
		// A client key without a client cert would be dropped by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "", "", "", "key")).HasError())
	}
	{
		// Mutual TLS
		assert.False(t, validateKafkaConfig(kafkaConfig("", "", "", "cert", "key")).HasError())
	}
	{
		// An unknown mechanism, e.g. from another resource, is checked once it's known
		config := kafkaConfig("", "artie", "secret", "", "")
		config.SASLMechanism = types.StringUnknown()
		assert.False(t, validateKafkaConfig(config).HasError())
	}
}
