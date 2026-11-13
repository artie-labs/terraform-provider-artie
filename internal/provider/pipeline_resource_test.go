package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

func TestBackfillTurnsOn(t *testing.T) {
	{
		// Off or never set, then on: turned on
		assert.True(t, backfillTurnsOn(types.BoolValue(false), types.BoolValue(true)))
		assert.True(t, backfillTurnsOn(types.BoolNull(), types.BoolValue(true)))
	}
	{
		// Already on, staying off, turned off, or only known at apply: not turned on
		assert.False(t, backfillTurnsOn(types.BoolValue(true), types.BoolValue(true)))
		assert.False(t, backfillTurnsOn(types.BoolValue(false), types.BoolValue(false)))
		assert.False(t, backfillTurnsOn(types.BoolValue(true), types.BoolValue(false)))
		assert.False(t, backfillTurnsOn(types.BoolValue(false), types.BoolUnknown()))
	}
}

func TestSkipBackfillConfigured(t *testing.T) {
	attrTypes := map[string]attr.Type{"skip_backfill": types.BoolType}
	tables := types.MapValueMust(types.ObjectType{AttrTypes: attrTypes}, map[string]attr.Value{
		"set":   types.ObjectValueMust(attrTypes, map[string]attr.Value{"skip_backfill": types.BoolValue(true)}),
		"unset": types.ObjectValueMust(attrTypes, map[string]attr.Value{"skip_backfill": types.BoolNull()}),
	})
	{
		// Set in the config: kept as configured
		assert.True(t, skipBackfillConfigured(tables, "set"))
	}
	{
		// Left out, or a table only in state, or no tables in the config: planned as unknown
		assert.False(t, skipBackfillConfigured(tables, "unset"))
		assert.False(t, skipBackfillConfigured(tables, "missing"))
		assert.False(t, skipBackfillConfigured(types.MapNull(types.ObjectType{AttrTypes: attrTypes}), "set"))
	}
}

func TestPipelineModifyPlanBackfillSwitch(t *testing.T) {
	ctx := t.Context()
	schemaResp := &resource.SchemaResponse{}
	(&PipelineResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	// pipelineValue builds a pipeline with tables "a" and "b" (both skip_backfill = true) and the DynamoDB backfill on or off.
	pipelineValue := func(backfillEnabled bool) tftypes.Value {
		apiModel := openapi.PayloadsFullPipeline{
			Uuid: uuid.New(),
			Name: "test",
			Tables: []openapi.PayloadsTable{
				{Uuid: lib.ToPtr(uuid.New()), Name: lib.ToPtr("a"), AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{SkipBackfill: lib.ToPtr(true)}},
				{Uuid: lib.ToPtr(uuid.New()), Name: lib.ToPtr("b"), AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{SkipBackfill: lib.ToPtr(true)}},
			},
			AdvancedSettings: openapi.PayloadsPipelineAdvancedSettings{
				// The API always returns the flush rules.
				FlushIntervalSeconds:   lib.ToPtr(30),
				BufferRows:             lib.ToPtr(1000),
				FlushSizeKb:            lib.ToPtr(1000),
				DynamoDBBackfillConfig: &openapi.PayloadsDynamoDBSnapshotConfig{Enabled: lib.ToPtr(backfillEnabled), Bucket: lib.ToPtr("exports")},
			},
		}
		pipeline, diags := tfmodels.PipelineFromAPIModel(ctx, apiModel)
		assert.False(t, diags.HasError(), diags)
		state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
		setDiags := state.Set(ctx, pipeline)
		assert.False(t, setDiags.HasError(), setDiags)
		return state.Raw
	}
	modifyPlan := func(priorEnabled, plannedEnabled bool) (types.Bool, types.Bool) {
		// The config leaves skip_backfill out on table "a" and sets it on table "b".
		config := tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(plannedEnabled)}
		assert.False(t, config.SetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), types.BoolNull()).HasError())
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: config.Raw},
			State:  tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(priorEnabled)},
			Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: pipelineValue(plannedEnabled)},
		}
		resp := &resource.ModifyPlanResponse{Plan: req.Plan}
		(&PipelineResource{}).ModifyPlan(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		var a, b types.Bool
		resp.Plan.GetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), &a)
		resp.Plan.GetAttribute(ctx, path.Root("tables").AtMapKey("b").AtName("skip_backfill"), &b)
		return a, b
	}
	{
		// Turning the backfill on: the API clears skip_backfill, so a left-out value is unknown; a configured one is kept
		a, b := modifyPlan(false, true)
		assert.True(t, a.IsUnknown())
		assert.Equal(t, types.BoolValue(true), b)
	}
	{
		// Backfill already on: the plan is left alone
		a, b := modifyPlan(true, true)
		assert.Equal(t, types.BoolValue(true), a)
		assert.Equal(t, types.BoolValue(true), b)
	}
}
