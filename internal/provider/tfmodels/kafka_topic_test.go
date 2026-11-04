package tfmodels

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

const testKafkaClusterUUID = "5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10"

func testKafkaTopicPayload() openapi.PayloadsKafkaTopicDetail {
	return openapi.PayloadsKafkaTopicDetail{
		Topic: openapi.PayloadsKafkaTopic{
			Topic:            "orders",
			PartitionCount:   6,
			RetentionMS:      "604800000",
			LocalRetentionMS: lib.ToPtr("-2"),
			TieredStorage:    openapi.PayloadsKafkaTopicTieredStorage{State: openapi.EnumsKafkaTopicTieredStorageStateOn, CanTier: true, SwitchOn: true},
		},
		Managed: &openapi.PayloadsKafkaTopicDetailManaged{ReplicationFactor: 3},
	}
}

func TestKafkaTopicFromAPIModel(t *testing.T) {
	prior := KafkaTopic{
		ReplicationFactor: types.Int64Value(2),
		ConfigEntries:     types.MapValueMust(types.StringType, map[string]attr.Value{"cleanup.policy": types.StringValue("delete")}),
	}
	topic, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, testKafkaTopicPayload(), prior)
	assert.False(t, diags.HasError())
	assert.Equal(t, testKafkaClusterUUID, topic.ClusterUUID.ValueString())
	assert.Equal(t, "orders", topic.Topic.ValueString())
	assert.Equal(t, int64(6), topic.PartitionCount.ValueInt64())
	assert.Equal(t, int64(604800000), topic.RetentionMS.ValueInt64())
	assert.Equal(t, int64(-2), topic.LocalRetentionMS.ValueInt64())
	assert.Equal(t, "on", topic.TieredStorage.ValueString())
	// The replication factor comes from the API when it returns one; config entries always come from prior.
	assert.Equal(t, int64(3), topic.ReplicationFactor.ValueInt64())
	assert.Equal(t, prior.ConfigEntries, topic.ConfigEntries)
}

func TestKafkaTopicFromAPIModel_KeepsPrior(t *testing.T) {
	payload := testKafkaTopicPayload()
	payload.Managed = nil
	payload.Topic.LocalRetentionMS = nil

	{
		// Outside Managed Kafka clusters the replication factor is kept from prior.
		topic, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, payload, KafkaTopic{ReplicationFactor: types.Int64Value(2), ConfigEntries: types.MapUnknown(types.StringType)})
		assert.False(t, diags.HasError())
		assert.Equal(t, int64(2), topic.ReplicationFactor.ValueInt64())
		assert.True(t, topic.LocalRetentionMS.IsNull())
		assert.True(t, topic.ConfigEntries.IsNull())
	}
	{
		// An imported topic has neither yet.
		topic, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, payload, KafkaTopic{ReplicationFactor: types.Int64Null(), ConfigEntries: types.MapNull(types.StringType)})
		assert.False(t, diags.HasError())
		assert.True(t, topic.ReplicationFactor.IsNull())
		assert.True(t, topic.ConfigEntries.IsNull())
	}
}

func TestKafkaTopicFromAPIModel_Errors(t *testing.T) {
	{
		payload := testKafkaTopicPayload()
		payload.Topic.ArtieManaged = true
		_, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, payload, KafkaTopic{})
		assert.True(t, diags.HasError())
	}
	{
		payload := testKafkaTopicPayload()
		payload.Topic.RetentionMS = "forever"
		_, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, payload, KafkaTopic{})
		assert.True(t, diags.HasError())
	}
}

func TestKafkaTopicToAPICreateRequest(t *testing.T) {
	topic := KafkaTopic{
		Topic:             types.StringValue("orders"),
		PartitionCount:    types.Int64Value(6),
		ReplicationFactor: types.Int64Value(3),
		ConfigEntries:     types.MapNull(types.StringType),
	}
	{
		req, diags := topic.ToAPICreateRequest(t.Context())
		assert.False(t, diags.HasError())
		assert.Equal(t, openapi.RouterCreateKafkaTopicRequest{Topic: "orders", PartitionCount: 6, ReplicationFactor: 3}, req)
	}
	{
		topic.ConfigEntries = types.MapValueMust(types.StringType, map[string]attr.Value{"cleanup.policy": types.StringValue("compact")})
		req, diags := topic.ToAPICreateRequest(t.Context())
		assert.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"cleanup.policy": "compact"}, *req.ConfigEntries)
	}
}

func TestKafkaTopicToAPISettingsRequest(t *testing.T) {
	state, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, testKafkaTopicPayload(), KafkaTopic{ConfigEntries: types.MapNull(types.StringType)})
	assert.False(t, diags.HasError())

	{
		// Nothing changed: nothing is sent.
		assert.Equal(t, openapi.RouterUpdateKafkaTopicSettingsRequest{}, state.ToAPISettingsRequest(state))
	}
	{
		// Settings left out of the config are unknown in the plan and aren't sent.
		plan := state
		plan.RetentionMS = types.Int64Unknown()
		plan.LocalRetentionMS = types.Int64Unknown()
		plan.TieredStorage = types.StringUnknown()
		assert.Equal(t, openapi.RouterUpdateKafkaTopicSettingsRequest{}, plan.ToAPISettingsRequest(state))
	}
	{
		// Only the changed settings are sent.
		plan := state
		plan.RetentionMS = types.Int64Value(86400000)
		plan.TieredStorage = types.StringValue("paused")
		req := plan.ToAPISettingsRequest(state)
		assert.Equal(t, 86400000, *req.RetentionMS)
		assert.Equal(t, openapi.RouterUpdateKafkaTopicSettingsRequestTieredStoragePaused, *req.TieredStorage)
		assert.Nil(t, req.LocalRetentionMS)
	}
}

func TestKafkaTopicToAPICreateSettingsRequest(t *testing.T) {
	created, diags := KafkaTopicFromAPIModel(testKafkaClusterUUID, testKafkaTopicPayload(), KafkaTopic{ConfigEntries: types.MapNull(types.StringType)})
	assert.False(t, diags.HasError())
	created.LocalRetentionMS = types.Int64Value(259200000) // the cluster's days on disk

	{
		// Retentions set in the config are sent even when they match the cluster's defaults, so the topic keeps them.
		plan := created
		req := plan.ToAPICreateSettingsRequest(created)
		assert.Equal(t, 604800000, *req.RetentionMS)
		assert.Equal(t, 259200000, *req.LocalRetentionMS)
		assert.Nil(t, req.TieredStorage)
	}
	{
		// Settings left out of the config aren't sent, and tiered storage is only sent when it changes.
		plan := created
		plan.RetentionMS = types.Int64Unknown()
		plan.LocalRetentionMS = types.Int64Unknown()
		plan.TieredStorage = types.StringValue("paused")
		req := plan.ToAPICreateSettingsRequest(created)
		assert.Nil(t, req.RetentionMS)
		assert.Nil(t, req.LocalRetentionMS)
		assert.Equal(t, openapi.RouterUpdateKafkaTopicSettingsRequestTieredStoragePaused, *req.TieredStorage)
	}
}
