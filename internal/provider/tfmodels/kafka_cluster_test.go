package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

func testTieredStorage(enabled bool, localRetentionMS, localRetentionBytes types.Int64) types.Object {
	return types.ObjectValueMust(KafkaClusterTieredStorageAttrTypes, map[string]attr.Value{
		"enabled":               types.BoolValue(enabled),
		"local_retention_ms":    localRetentionMS,
		"local_retention_bytes": localRetentionBytes,
	})
}

func testKafkaClusterPayload() openapi.PayloadsKafkaCluster {
	return openapi.PayloadsKafkaCluster{
		Uuid:                 uuid.New(),
		Name:                 "my-cluster",
		Region:               "us-east-1",
		Cloud:                "aws",
		Status:               openapi.EnumsKafkaStatusActive,
		ExternalBootstrapURL: "bootstrap.example.com:9092",
		IsManaged:            true,
		TierAllTopics:        true,
		LocalRetentionMS:     lib.ToPtr(259200000),
		Managed: &openapi.PayloadsKafkaClusterManaged{
			AllowedCIDRs:   []string{"10.0.0.0/16"},
			BrokerSettings: openapi.PayloadsKafkaClusterBrokerSettings{RetentionMS: 1209600000},
			Brokers:        3,
			InstanceType:   "m8g.xlarge",
			KafkaVersion:   "4.0.0",
			StorageGiB:     100,
		},
		StorageAutoscale: &openapi.PayloadsKafkaClusterStorageAutoscale{
			Enabled:       true,
			BumpAtPercent: 65,
			TargetPercent: 55,
		},
	}
}

func TestKafkaClusterFromAPIModel(t *testing.T) {
	cluster, diags := KafkaClusterFromAPIModel(t.Context(), testKafkaClusterPayload())
	assert.False(t, diags.HasError())
	assert.Equal(t, "m8g.xlarge", cluster.InstanceType.ValueString())
	assert.Equal(t, int64(3), cluster.Brokers.ValueInt64())
	assert.Equal(t, int64(100), cluster.StorageGiB.ValueInt64())
	assert.Equal(t, int64(100), cluster.CurrentStorageGiB.ValueInt64())
	assert.Equal(t, int64(1209600000), cluster.RetentionMS.ValueInt64())
	assert.Equal(t, testTieredStorage(true, types.Int64Value(259200000), types.Int64Null()), cluster.TieredStorage)
	assert.Equal(t, "active", cluster.Status.ValueString())

	var autoscale KafkaClusterStorageAutoscale
	assert.False(t, cluster.StorageAutoscale.As(t.Context(), &autoscale, basetypes.ObjectAsOptions{}).HasError())
	assert.True(t, autoscale.Enabled.ValueBool())
	assert.True(t, autoscale.MaxGiB.IsNull())
}

func TestKafkaClusterFromAPIModel_NotManaged(t *testing.T) {
	payload := testKafkaClusterPayload()
	payload.IsManaged = false
	payload.Managed = nil
	_, diags := KafkaClusterFromAPIModel(t.Context(), payload)
	assert.True(t, diags.HasError())
}

func TestKafkaClusterToAPICreateRequest_Defaults(t *testing.T) {
	cluster := KafkaCluster{
		Name:             types.StringValue("my-cluster"),
		Region:           types.StringValue("us-east-1"),
		InstanceType:     types.StringValue("m8g.xlarge"),
		Brokers:          types.Int64Value(3),
		StorageGiB:       types.Int64Value(100),
		AllowedCIDRs:     types.ListValueMust(types.StringType, []attr.Value{}),
		TieredStorage:    testTieredStorage(true, types.Int64Unknown(), types.Int64Null()),
		StorageAutoscale: types.ObjectUnknown(KafkaClusterStorageAutoscaleAttrTypes),
	}

	req, diags := cluster.ToAPICreateRequest(t.Context())
	assert.False(t, diags.HasError())
	// Left out of the config, these are omitted so the API applies its defaults.
	assert.True(t, req.TieredStorage)
	assert.Nil(t, req.LocalRetentionMS)
	assert.Nil(t, req.LocalRetentionBytes)
	assert.Nil(t, req.StorageAutoscale)
	assert.Nil(t, req.RetentionMS)
	// The API requires the list, so an empty one is sent rather than null.
	assert.Equal(t, []string{}, req.AllowedCIDRs)
	assert.Equal(t, openapi.PayloadsKafkaInstanceTypeM8gXlarge, req.InstanceType)
}

func TestKafkaClusterToAPICreateRequest_RetentionMS(t *testing.T) {
	cluster := KafkaCluster{
		AllowedCIDRs:     types.ListValueMust(types.StringType, []attr.Value{}),
		RetentionMS:      types.Int64Value(-1),
		TieredStorage:    testTieredStorage(true, types.Int64Unknown(), types.Int64Null()),
		StorageAutoscale: types.ObjectUnknown(KafkaClusterStorageAutoscaleAttrTypes),
	}
	req, diags := cluster.ToAPICreateRequest(t.Context())
	assert.False(t, diags.HasError())
	assert.Equal(t, -1, *req.RetentionMS)
}

func TestKafkaClusterBrokerSettingsToAPIUpdate(t *testing.T) {
	state, diags := KafkaClusterFromAPIModel(t.Context(), testKafkaClusterPayload())
	assert.False(t, diags.HasError())

	{
		// Unchanged: nothing is sent.
		assert.Nil(t, state.BrokerSettingsToAPIUpdate(state))
	}
	{
		// Left out of the config with none in state (e.g. state from before the attribute existed): not a change.
		plan := state
		plan.RetentionMS = types.Int64Unknown()
		prior := state
		prior.RetentionMS = types.Int64Null()
		assert.Nil(t, plan.BrokerSettingsToAPIUpdate(prior))
	}
	{
		// Changed: the new retention is sent, and it isn't part of the cluster update.
		plan := state
		plan.RetentionMS = types.Int64Value(-1)
		req := plan.BrokerSettingsToAPIUpdate(state)
		if assert.NotNil(t, req) {
			assert.Equal(t, -1, *req.RetentionMS)
		}
		update, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.Equal(t, openapi.RouterUpdateKafkaClusterRequest{}, update)
	}
}

func TestAppliesToBrokers(t *testing.T) {
	assert.False(t, AppliesToBrokers(openapi.RouterUpdateKafkaClusterRequest{}))
	assert.False(t, AppliesToBrokers(openapi.RouterUpdateKafkaClusterRequest{Name: lib.ToPtr("renamed"), StorageAutoscale: &openapi.PayloadsKafkaClusterStorageAutoscale{}}))
	assert.True(t, AppliesToBrokers(openapi.RouterUpdateKafkaClusterRequest{Brokers: lib.ToPtr(4)}))
	assert.True(t, AppliesToBrokers(openapi.RouterUpdateKafkaClusterRequest{Name: lib.ToPtr("renamed"), AllowedCIDRs: &[]string{}}))
}

func TestKafkaClusterToAPIUpdateRequest(t *testing.T) {
	state, diags := KafkaClusterFromAPIModel(t.Context(), testKafkaClusterPayload())
	assert.False(t, diags.HasError())

	{
		// Nothing changed: nothing is sent.
		req, diags := state.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.Equal(t, openapi.RouterUpdateKafkaClusterRequest{}, req)
	}
	{
		// Only the changed settings are sent.
		plan := state
		plan.Brokers = types.Int64Value(4)
		plan.Name = types.StringValue("renamed")
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.Equal(t, 4, *req.Brokers)
		assert.Equal(t, "renamed", *req.Name)
		assert.Nil(t, req.StorageGiB)
		assert.Nil(t, req.InstanceType)
		assert.Nil(t, req.AllowedCIDRs)
		assert.Nil(t, req.TieredStorage)
		assert.Nil(t, req.StorageAutoscale)
	}
	{
		// Changing tiered storage keeps the days on disk from state when it's left out of the config.
		plan := state
		plan.TieredStorage = testTieredStorage(false, types.Int64Unknown(), types.Int64Null())
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.False(t, req.TieredStorage.TierAllTopics)
		assert.Equal(t, int64(259200000), req.TieredStorage.LocalRetentionMS)
		assert.Nil(t, req.TieredStorage.LocalRetentionBytes)
	}
	{
		// Setting a disk cap sends it with the other two.
		plan := state
		plan.TieredStorage = testTieredStorage(true, types.Int64Value(259200000), types.Int64Value(1073741824))
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.True(t, req.TieredStorage.TierAllTopics)
		assert.Equal(t, 1073741824, *req.TieredStorage.LocalRetentionBytes)
	}
	{
		// Clearing the CIDRs sends an empty list, which allows any IP.
		plan := state
		plan.AllowedCIDRs = types.ListValueMust(types.StringType, []attr.Value{})
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.Equal(t, []string{}, *req.AllowedCIDRs)
	}
	{
		// With no days on disk in state, leaving it out of the config isn't a change.
		noRetention := state
		noRetention.TieredStorage = testTieredStorage(true, types.Int64Null(), types.Int64Null())
		plan := noRetention
		plan.TieredStorage = testTieredStorage(true, types.Int64Unknown(), types.Int64Null())
		req, diags := plan.ToAPIUpdateRequest(t.Context(), noRetention)
		assert.False(t, diags.HasError())
		assert.Nil(t, req.TieredStorage)

		// Changing tiered storage with no days on disk in the config or state errors instead of sending 0.
		plan.TieredStorage = testTieredStorage(true, types.Int64Unknown(), types.Int64Value(1073741824))
		_, diags = plan.ToAPIUpdateRequest(t.Context(), noRetention)
		assert.True(t, diags.HasError())

		// Turning it off, where the days on disk can't be set, sends Artie's default.
		plan.TieredStorage = testTieredStorage(false, types.Int64Unknown(), types.Int64Null())
		req, diags = plan.ToAPIUpdateRequest(t.Context(), noRetention)
		assert.False(t, diags.HasError())
		assert.Equal(t, int64(DefaultKafkaLocalRetentionMS), req.TieredStorage.LocalRetentionMS)
	}
}

func TestKafkaClusterWithCreatedCluster(t *testing.T) {
	payload := testKafkaClusterPayload()
	plan := KafkaCluster{
		UUID:                   types.StringUnknown(),
		Name:                   types.StringValue("my-cluster"),
		TieredStorage:          testTieredStorage(true, types.Int64Unknown(), types.Int64Null()),
		StorageAutoscale:       types.ObjectUnknown(KafkaClusterStorageAutoscaleAttrTypes),
		Cloud:                  types.StringUnknown(),
		Status:                 types.StringUnknown(),
		ExternalBootstrapURL:   types.StringUnknown(),
		CurrentStorageGiB:      types.Int64Unknown(),
		KafkaVersion:           types.StringUnknown(),
		TieredStorageAvailable: types.BoolUnknown(),
		RetentionMS:            types.Int64Unknown(),
	}
	cluster := plan.WithCreatedCluster(payload)
	// Left out of the config, the retention comes from the response; a configured one is kept.
	assert.Equal(t, int64(1209600000), cluster.RetentionMS.ValueInt64())
	plan.RetentionMS = types.Int64Value(-1)
	assert.Equal(t, int64(-1), plan.WithCreatedCluster(payload).RetentionMS.ValueInt64())
	assert.Equal(t, payload.Uuid.String(), cluster.UUID.ValueString())
	assert.Equal(t, "my-cluster", cluster.Name.ValueString())
	assert.Equal(t, "active", cluster.Status.ValueString())
	assert.Equal(t, testTieredStorage(true, types.Int64Value(259200000), types.Int64Null()), cluster.TieredStorage)
	assert.True(t, cluster.StorageAutoscale.IsNull())
	assert.True(t, cluster.CurrentStorageGiB.IsNull())
	assert.True(t, cluster.KafkaVersion.IsNull())
}

func TestKafkaClusterWithRequestedStorage(t *testing.T) {
	// Autoscaling grew the disks from 100 to 150 GiB.
	payload := testKafkaClusterPayload()
	payload.Managed.StorageGiB = 150
	cluster, diags := KafkaClusterFromAPIModel(t.Context(), payload)
	assert.False(t, diags.HasError())

	{
		// The config asked for less than the brokers have, so it's kept.
		kept := cluster.WithRequestedStorage(types.Int64Value(100))
		assert.Equal(t, int64(100), kept.StorageGiB.ValueInt64())
		assert.Equal(t, int64(150), kept.CurrentStorageGiB.ValueInt64())
	}
	{
		// The config asked for more than the brokers have (the grow hasn't finished), so the API's size shows the drift.
		assert.Equal(t, int64(150), cluster.WithRequestedStorage(types.Int64Value(200)).StorageGiB.ValueInt64())
	}
	{
		// Import has nothing requested yet.
		assert.Equal(t, int64(150), cluster.WithRequestedStorage(types.Int64Null()).StorageGiB.ValueInt64())
	}
}
