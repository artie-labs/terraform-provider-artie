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
			AllowedCIDRs: []string{"10.0.0.0/16"},
			Brokers:      3,
			InstanceType: "m8g.xlarge",
			KafkaVersion: "4.0.0",
			StorageGiB:   100,
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
	assert.Equal(t, int64(259200000), cluster.LocalRetentionMS.ValueInt64())
	assert.True(t, cluster.LocalRetentionBytes.IsNull())
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
		Name:                types.StringValue("my-cluster"),
		Region:              types.StringValue("us-east-1"),
		InstanceType:        types.StringValue("m8g.xlarge"),
		Brokers:             types.Int64Value(3),
		StorageGiB:          types.Int64Value(100),
		AllowedCIDRs:        types.ListValueMust(types.StringType, []attr.Value{}),
		TieredStorage:       types.BoolValue(true),
		LocalRetentionMS:    types.Int64Unknown(),
		LocalRetentionBytes: types.Int64Null(),
		StorageAutoscale:    types.ObjectUnknown(KafkaClusterStorageAutoscaleAttrTypes),
	}

	req, diags := cluster.ToAPICreateRequest(t.Context())
	assert.False(t, diags.HasError())
	// Left out of the config, these are omitted so the API applies its defaults.
	assert.Nil(t, req.LocalRetentionMS)
	assert.Nil(t, req.LocalRetentionBytes)
	assert.Nil(t, req.StorageAutoscale)
	// The API requires the list, so an empty one is sent rather than null.
	assert.Equal(t, []string{}, req.AllowedCIDRs)
	assert.Equal(t, openapi.PayloadsKafkaInstanceTypeM8gXlarge, req.InstanceType)
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
		plan.TieredStorage = types.BoolValue(false)
		plan.LocalRetentionMS = types.Int64Unknown()
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.False(t, req.TieredStorage.TierAllTopics)
		assert.Equal(t, int64(259200000), req.TieredStorage.LocalRetentionMS)
		assert.Nil(t, req.TieredStorage.LocalRetentionBytes)
	}
	{
		// Clearing the CIDRs sends an empty list, which allows any IP.
		plan := state
		plan.AllowedCIDRs = types.ListValueMust(types.StringType, []attr.Value{})
		req, diags := plan.ToAPIUpdateRequest(t.Context(), state)
		assert.False(t, diags.HasError())
		assert.Equal(t, []string{}, *req.AllowedCIDRs)
	}
}
