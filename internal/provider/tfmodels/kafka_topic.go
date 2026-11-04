package tfmodels

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type KafkaTopic struct {
	ClusterUUID       types.String `tfsdk:"cluster_uuid"`
	Topic             types.String `tfsdk:"topic"`
	PartitionCount    types.Int64  `tfsdk:"partition_count"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
	ConfigEntries     types.Map    `tfsdk:"config_entries"`
	RetentionMS       types.Int64  `tfsdk:"retention_ms"`
	LocalRetentionMS  types.Int64  `tfsdk:"local_retention_ms"`
	TieredStorage     types.String `tfsdk:"tiered_storage"`
}

func (k KafkaTopic) ToAPICreateRequest(ctx context.Context) (openapi.RouterCreateKafkaTopicRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var configEntries *map[string]string
	if IsKnown(k.ConfigEntries) {
		entries := map[string]string{}
		diags.Append(k.ConfigEntries.ElementsAs(ctx, &entries, false)...)
		configEntries = &entries
	}

	return openapi.RouterCreateKafkaTopicRequest{
		Topic:             k.Topic.ValueString(),
		PartitionCount:    int(k.PartitionCount.ValueInt64()),
		ReplicationFactor: int(k.ReplicationFactor.ValueInt64()),
		ConfigEntries:     configEntries,
	}, diags
}

// ToAPISettingsRequest only sends the settings that are set in the config and differ from current, which is state on
// update and the freshly created topic on create. Settings left out of the config keep whatever the topic has.
func (k KafkaTopic) ToAPISettingsRequest(current KafkaTopic) openapi.RouterUpdateKafkaTopicSettingsRequest {
	var req openapi.RouterUpdateKafkaTopicSettingsRequest
	if IsKnown(k.RetentionMS) && !k.RetentionMS.Equal(current.RetentionMS) {
		req.RetentionMS = int64ToIntPointer(k.RetentionMS)
	}
	if IsKnown(k.LocalRetentionMS) && !k.LocalRetentionMS.Equal(current.LocalRetentionMS) {
		req.LocalRetentionMS = int64ToIntPointer(k.LocalRetentionMS)
	}
	if IsKnown(k.TieredStorage) && !k.TieredStorage.Equal(current.TieredStorage) {
		req.TieredStorage = lib.ToPtr(openapi.RouterUpdateKafkaTopicSettingsRequestTieredStorage(k.TieredStorage.ValueString()))
	}
	return req
}

// KafkaTopicFromAPIModel builds the topic from Get a Kafka topic. The API doesn't return the config entries a topic was
// created with, or its replication factor outside Managed Kafka clusters, so those are kept from prior.
func KafkaTopicFromAPIModel(clusterUUID string, apiModel openapi.PayloadsKafkaTopicDetail, prior KafkaTopic) (KafkaTopic, diag.Diagnostics) {
	var diags diag.Diagnostics
	topic := apiModel.Topic
	if topic.ArtieManaged {
		diags.AddError("Unsupported Kafka topic",
			fmt.Sprintf("Kafka topic %q is managed by Artie (a CDC or Kafka-internal topic) and can't be managed with the artie_kafka_topic resource.", topic.Topic))
		return KafkaTopic{}, diags
	}

	retentionMS, err := strconv.ParseInt(topic.RetentionMS, 10, 64)
	if err != nil {
		diags.AddError("Unable to parse Kafka topic retention", fmt.Sprintf("retentionMS %q: %s", topic.RetentionMS, err))
	}
	localRetentionMS := types.Int64Null()
	if topic.LocalRetentionMS != nil {
		value, err := strconv.ParseInt(*topic.LocalRetentionMS, 10, 64)
		if err != nil {
			diags.AddError("Unable to parse Kafka topic local retention", fmt.Sprintf("localRetentionMS %q: %s", *topic.LocalRetentionMS, err))
		}
		localRetentionMS = types.Int64Value(value)
	}

	replicationFactor := prior.ReplicationFactor
	if apiModel.Managed != nil {
		replicationFactor = types.Int64Value(int64(apiModel.Managed.ReplicationFactor))
	}
	configEntries := prior.ConfigEntries
	if configEntries.IsUnknown() {
		configEntries = types.MapNull(types.StringType)
	}

	return KafkaTopic{
		ClusterUUID:       types.StringValue(clusterUUID),
		Topic:             types.StringValue(topic.Topic),
		PartitionCount:    types.Int64Value(int64(topic.PartitionCount)),
		ReplicationFactor: replicationFactor,
		ConfigEntries:     configEntries,
		RetentionMS:       types.Int64Value(retentionMS),
		LocalRetentionMS:  localRetentionMS,
		TieredStorage:     types.StringValue(string(topic.TieredStorage.State)),
	}, diags
}
