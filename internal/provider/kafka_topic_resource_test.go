package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-artie/internal/openapi"
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

	forever := types.Int64Value(kafkaTopicRetentionForever)
	assert.Empty(t, tieredRetentionError(topic(forever, types.StringValue("off"))))
	assert.Empty(t, tieredRetentionError(topic(forever, types.StringValue("on"))))
	assert.Empty(t, tieredRetentionError(topic(forever, types.StringValue("paused"))))
}

func TestKafkaTopicRetentionMSValidator(t *testing.T) {
	var resp resource.SchemaResponse
	(&KafkaTopicResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	retentionAttr, ok := resp.Schema.Attributes["retention_ms"].(schema.Int64Attribute)
	require.True(t, ok)

	validate := func(value int64) bool {
		req := validator.Int64Request{Path: path.Root("retention_ms"), ConfigValue: types.Int64Value(value)}
		var res validator.Int64Response
		for _, v := range retentionAttr.Validators {
			v.ValidateInt64(context.Background(), req, &res)
		}
		return !res.Diagnostics.HasError()
	}
	assert.True(t, validate(kafkaTopicRetentionForever))
	assert.True(t, validate(kafkaTopicMinRetentionMS))
	assert.False(t, validate(0))
	assert.False(t, validate(-2))
	assert.False(t, validate(kafkaTopicMinRetentionMS-1))
}

func TestLocalRetentionError(t *testing.T) {
	topic := func(retentionMS, localRetentionMS types.Int64, tieredStorage types.String) tfmodels.KafkaTopic {
		return tfmodels.KafkaTopic{RetentionMS: retentionMS, LocalRetentionMS: localRetentionMS, TieredStorage: tieredStorage}
	}
	oneDay, twoDays := types.Int64Value(86400000), types.Int64Value(2*86400000)
	assert.NotEmpty(t, localRetentionError(topic(oneDay, twoDays, types.StringValue("on"))))
	// Left out of the config, a new topic follows the cluster's tiered storage, which can be on.
	assert.NotEmpty(t, localRetentionError(topic(oneDay, twoDays, types.StringUnknown())))
	assert.Empty(t, localRetentionError(topic(oneDay, twoDays, types.StringValue("off"))))
	assert.Empty(t, localRetentionError(topic(twoDays, twoDays, types.StringValue("on"))))
	assert.Empty(t, localRetentionError(topic(oneDay, types.Int64Value(-2), types.StringValue("on"))))
	assert.Empty(t, localRetentionError(topic(types.Int64Unknown(), twoDays, types.StringValue("on"))))
}

func TestCreateTopicWaitsForDeployingCluster(t *testing.T) {
	kafkaClusterDeployPollInterval = time.Millisecond
	t.Cleanup(func() { kafkaClusterDeployPollInterval = 30 * time.Second })

	newResource := func(t *testing.T, clusterStatuses ...string) (*KafkaTopicResource, *int) {
		creates := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch {
			case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/topics"):
				creates++
				if creates <= len(clusterStatuses) {
					w.WriteHeader(http.StatusConflict)
					_, _ = w.Write([]byte(`{"error": "cluster isn't active"}`))
					return
				}
				_, _ = w.Write([]byte(`{}`))
			case req.Method == http.MethodGet:
				_, _ = fmt.Fprintf(w, `{"status": %q}`, clusterStatuses[creates-1])
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(server.Close)
		client, err := openapi.NewClientWithResponses(server.URL)
		assert.NoError(t, err)
		return &KafkaTopicResource{client: client}, &creates
	}
	body := openapi.RouterCreateKafkaTopicRequest{Topic: "orders", PartitionCount: 1, ReplicationFactor: 3}
	clusterUUID := "5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10"
	{
		r, creates := newResource(t, "deploying", "deploying")
		assert.NoError(t, r.createTopic(context.Background(), clusterUUID, body))
		assert.Equal(t, 3, *creates)
	}
	{
		// The deploy finished between the create and the cluster read: the create is retried once.
		r, creates := newResource(t, "deploying", "active")
		assert.NoError(t, r.createTopic(context.Background(), clusterUUID, body))
		assert.Equal(t, 3, *creates)
	}
	{
		// A second conflict on an active cluster is a real conflict.
		r, creates := newResource(t, "active", "active")
		err := r.createTopic(context.Background(), clusterUUID, body)
		assert.ErrorContains(t, err, "cluster isn't active (HTTP 409)")
		assert.Equal(t, 2, *creates)
	}
}

func TestConfigEntriesChangeReplaces(t *testing.T) {
	null := types.MapNull(types.StringType)
	compact := types.MapValueMust(types.StringType, map[string]attr.Value{"cleanup.policy": types.StringValue("compact")})
	del := types.MapValueMust(types.StringType, map[string]attr.Value{"cleanup.policy": types.StringValue("delete")})

	assert.False(t, configEntriesChangeReplaces(compact, compact, false))
	assert.True(t, configEntriesChangeReplaces(compact, del, false))
	assert.True(t, configEntriesChangeReplaces(compact, null, false))
	// A topic created without config entries can't be given them in place.
	assert.True(t, configEntriesChangeReplaces(null, compact, false))
	// An imported topic adopts its first config entries, and changes after that replace it.
	assert.False(t, configEntriesChangeReplaces(null, compact, true))
	assert.True(t, configEntriesChangeReplaces(compact, del, true))
}

func TestReplacesTopic(t *testing.T) {
	state := tfmodels.KafkaTopic{
		ClusterUUID:       types.StringValue("5b5d8d1e-7a39-4ef0-9a3b-5c1f8f1d2a10"),
		Topic:             types.StringValue("orders"),
		ReplicationFactor: types.Int64Value(3),
		ConfigEntries:     types.MapNull(types.StringType),
	}
	assert.False(t, replacesTopic(state, state, false))
	{
		plan := state
		plan.Topic = types.StringValue("orders.v2")
		assert.True(t, replacesTopic(state, plan, false))
	}
	{
		plan := state
		plan.ReplicationFactor = types.Int64Value(2)
		assert.True(t, replacesTopic(state, plan, false))
		// An imported topic's replication factor can be null, and is adopted.
		imported := state
		imported.ReplicationFactor = types.Int64Null()
		assert.False(t, replacesTopic(imported, plan, true))
	}
	{
		plan := state
		plan.ConfigEntries = types.MapValueMust(types.StringType, map[string]attr.Value{"cleanup.policy": types.StringValue("compact")})
		assert.True(t, replacesTopic(state, plan, false))
		assert.False(t, replacesTopic(state, plan, true))
	}
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
