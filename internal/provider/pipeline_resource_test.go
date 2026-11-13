package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
	"terraform-provider-artie/internal/provider/tfmodels"
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

func TestPipelineSchemaMatchesModel(t *testing.T) {
	// Setting state from a fully populated API model fails if a nested object's schema and model types disagree.
	ctx := t.Context()
	schemaResp := &resource.SchemaResponse{}
	(&PipelineResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	apiModel := openapi.PayloadsFullPipeline{
		Uuid: uuid.New(),
		Name: "test",
		Tables: []openapi.PayloadsTable{{
			Uuid: lib.ToPtr(uuid.New()),
			Name: lib.ToPtr("t"),
			AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
				ColumnsToTrim:          &[]string{"name"},
				PartitionRangeSettings: &openapi.PayloadsPartitionRangeSettings{Enabled: lib.ToPtr(true)},
			},
		}},
		SpecificDestCfg: openapi.PayloadsSpecificConfig{
			Compaction: &openapi.PayloadsCompactionConfig{QuietWindow: &openapi.PayloadsCompactionQuietWindow{}},
		},
		AdvancedSettings: openapi.PayloadsPipelineAdvancedSettings{
			// The API always returns the flush rules.
			FlushIntervalSeconds:   lib.ToPtr(30),
			BufferRows:             lib.ToPtr(1000),
			FlushSizeKb:            lib.ToPtr(1000),
			DynamoDBBackfillConfig: &openapi.PayloadsDynamoDBSnapshotConfig{Enabled: lib.ToPtr(true)},
		},
	}
	pipeline, diags := tfmodels.PipelineFromAPIModel(ctx, apiModel)
	assert.False(t, diags.HasError(), diags)

	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	setDiags := state.Set(ctx, pipeline)
	assert.False(t, setDiags.HasError(), setDiags)
	var readBack tfmodels.Pipeline
	getDiags := state.Get(ctx, &readBack)
	assert.False(t, getDiags.HasError(), getDiags)
	assert.True(t, readBack.DestinationConfig.Compaction.Equal(pipeline.DestinationConfig.Compaction))
	assert.True(t, readBack.DynamoDBBackfillConfig.Equal(pipeline.DynamoDBBackfillConfig))
}
