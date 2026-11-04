package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/provider/tfmodels"
)

func TestTieredStorageTransitionError(t *testing.T) {
	assert.NotEmpty(t, tieredStorageTransitionError(types.StringValue("on"), types.StringValue("off")))
	assert.Empty(t, tieredStorageTransitionError(types.StringValue("on"), types.StringValue("paused")))
	assert.Empty(t, tieredStorageTransitionError(types.StringValue("paused"), types.StringValue("off")))
	assert.Empty(t, tieredStorageTransitionError(types.StringValue("off"), types.StringValue("on")))
	// Left out of the config, the plan keeps state, so unknown isn't checked.
	assert.Empty(t, tieredStorageTransitionError(types.StringValue("on"), types.StringUnknown()))
}

func TestTieredRetentionError(t *testing.T) {
	topic := func(retentionMS types.Int64, tieredStorage types.String) tfmodels.KafkaTopic {
		return tfmodels.KafkaTopic{RetentionMS: retentionMS, TieredStorage: tieredStorage}
	}
	fifteenDays := types.Int64Value(15 * 86400000)
	assert.NotEmpty(t, tieredRetentionError(topic(fifteenDays, types.StringValue("on"))))
	assert.NotEmpty(t, tieredRetentionError(topic(fifteenDays, types.StringValue("paused"))))
	assert.Empty(t, tieredRetentionError(topic(fifteenDays, types.StringValue("off"))))
	assert.Empty(t, tieredRetentionError(topic(types.Int64Value(kafkaTopicMaxTieredRetention), types.StringValue("on"))))
	assert.Empty(t, tieredRetentionError(topic(fifteenDays, types.StringUnknown())))
}

func TestParseKafkaTopicImportID(t *testing.T) {
	{
		clusterUUID, topic, err := parseKafkaTopicImportID("5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10/orders.v1")
		assert.NoError(t, err)
		assert.Equal(t, "5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10", clusterUUID)
		assert.Equal(t, "orders.v1", topic)
	}
	for _, id := range []string{"orders", "/orders", "5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10/", "a/b/c"} {
		_, _, err := parseKafkaTopicImportID(id)
		assert.Error(t, err, id)
	}
}
