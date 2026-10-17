package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestClearedRegexPlan(t *testing.T) {
	{
		// Left out while the other is set: cleared, so the old value in state isn't sent alongside the new one.
		planned, ok := clearedRegexPlan(types.StringNull(), types.StringValue("^billing\\."))
		assert.True(t, ok)
		assert.Equal(t, types.StringValue(""), planned)
	}
	{
		// Left out while the other is only known at apply time: unknown, so the plan can't change at apply.
		planned, ok := clearedRegexPlan(types.StringNull(), types.StringUnknown())
		assert.True(t, ok)
		assert.True(t, planned.IsUnknown())
	}
	{
		// Left out and the other isn't set or is cleared: the plan is left alone.
		_, ok := clearedRegexPlan(types.StringNull(), types.StringNull())
		assert.False(t, ok)
		_, ok = clearedRegexPlan(types.StringNull(), types.StringValue(""))
		assert.False(t, ok)
	}
	{
		// Set in the config: never overridden.
		_, ok := clearedRegexPlan(types.StringValue("^tmp_"), types.StringValue("^billing\\."))
		assert.False(t, ok)
		_, ok = clearedRegexPlan(types.StringValue(""), types.StringValue("^billing\\."))
		assert.False(t, ok)
	}
}

func TestPlannedPartitionBy(t *testing.T) {
	{
		// Left out on a time predicate: the granularity from state (e.g. set in the dashboard) is kept.
		assert.Equal(t, types.StringValue("monthly"), plannedPartitionBy(types.StringValue("time"), types.StringValue("monthly")))
	}
	{
		// No prior state: left to the API's default.
		assert.True(t, plannedPartitionBy(types.StringValue("time"), types.StringNull()).IsUnknown())
	}
	{
		// Switched to an integer predicate: dropped, since the API rejects partitionBy there.
		assert.True(t, plannedPartitionBy(types.StringValue("integer"), types.StringValue("monthly")).IsNull())
	}
	{
		// Partition type only known at apply time: the granularity from state is kept.
		assert.Equal(t, types.StringValue("monthly"), plannedPartitionBy(types.StringUnknown(), types.StringValue("monthly")))
	}
}
