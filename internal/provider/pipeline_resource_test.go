package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func TestBackfillMaySwitch(t *testing.T) {
	{
		// Turned on (from off or never set), turned off, or only known at apply: may switch
		assert.True(t, backfillMaySwitch(types.BoolValue(false), types.BoolValue(true)))
		assert.True(t, backfillMaySwitch(types.BoolNull(), types.BoolValue(true)))
		assert.True(t, backfillMaySwitch(types.BoolValue(true), types.BoolValue(false)))
		assert.True(t, backfillMaySwitch(types.BoolValue(false), types.BoolUnknown()))
	}
	{
		// Unchanged: no switch
		assert.False(t, backfillMaySwitch(types.BoolValue(true), types.BoolValue(true)))
		assert.False(t, backfillMaySwitch(types.BoolValue(false), types.BoolValue(false)))
		assert.False(t, backfillMaySwitch(types.BoolNull(), types.BoolValue(false)))
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
				DynamoDBBackfillConfig: &openapi.PayloadsDynamoDBSnapshotConfig{Enabled: lib.ToPtr(backfillEnabled), Bucket: lib.ToPtr("exports"), ExportArn: lib.ToPtr("arn")},
			},
		}
		pipeline, diags := tfmodels.PipelineFromAPIModel(ctx, apiModel)
		assert.False(t, diags.HasError(), diags)
		state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
		setDiags := state.Set(ctx, pipeline)
		assert.False(t, setDiags.HasError(), setDiags)
		return state.Raw
	}
	// modifyPlan plans a switch of the backfill with table "a" leaving skip_backfill out of the config and table "b"
	// setting it to configuredB.
	modifyPlan := func(priorEnabled, plannedEnabled bool, configuredB types.Bool) (types.Bool, types.Bool, diag.Diagnostics) {
		config := tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(plannedEnabled)}
		assert.False(t, config.SetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), types.BoolNull()).HasError())
		assert.False(t, config.SetAttribute(ctx, path.Root("tables").AtMapKey("b").AtName("skip_backfill"), configuredB).HasError())
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: config.Raw},
			State:  tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(priorEnabled)},
			Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: pipelineValue(plannedEnabled)},
		}
		resp := &resource.ModifyPlanResponse{Plan: req.Plan}
		(&PipelineResource{}).ModifyPlan(ctx, req, resp)
		var a, b types.Bool
		resp.Plan.GetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), &a)
		resp.Plan.GetAttribute(ctx, path.Root("tables").AtMapKey("b").AtName("skip_backfill"), &b)
		return a, b, resp.Diagnostics
	}
	{
		// Turning the backfill on: the API clears skip_backfill on existing tables, so a left-out value is unknown
		a, b, diags := modifyPlan(false, true, types.BoolValue(false))
		assert.False(t, diags.HasError(), diags)
		assert.True(t, a.IsUnknown())
		assert.Equal(t, types.BoolValue(true), b) // the test plan still carries state; ModifyPlan leaves configured values alone
	}
	{
		// Turning the backfill on while an existing table keeps skip_backfill = true: rejected, since the API clears it
		_, _, diags := modifyPlan(false, true, types.BoolValue(true))
		assert.True(t, diags.HasError())
	}
	{
		// Turning the backfill off: the API may set skip_backfill on tables waiting for a backfill, so a left-out value is unknown
		a, _, diags := modifyPlan(true, false, types.BoolValue(true))
		assert.False(t, diags.HasError(), diags)
		assert.True(t, a.IsUnknown())
	}
	{
		// Turning the backfill on while removing table "b" from the config: "b" stays out of the plan
		config := tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(true)}
		var tables types.Map
		assert.False(t, config.GetAttribute(ctx, path.Root("tables"), &tables).HasError())
		remaining := map[string]attr.Value{"a": tables.Elements()["a"]}
		assert.False(t, config.SetAttribute(ctx, path.Root("tables"), types.MapValueMust(tables.ElementType(ctx), remaining)).HasError())
		assert.False(t, config.SetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), types.BoolNull()).HasError())
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: config.Raw},
			State:  tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(false)},
			Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: config.Raw},
		}
		resp := &resource.ModifyPlanResponse{Plan: req.Plan}
		(&PipelineResource{}).ModifyPlan(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		var planned types.Map
		assert.False(t, resp.Plan.GetAttribute(ctx, path.Root("tables"), &planned).HasError())
		assert.Len(t, planned.Elements(), 1)
		_, hasB := planned.Elements()["b"]
		assert.False(t, hasB)
	}
	{
		// A backfill block only known at apply may turn the backfill on, so a left-out skip_backfill is unknown
		config := tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(false)}
		assert.False(t, config.SetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), types.BoolNull()).HasError())
		unknownBackfill := types.ObjectUnknown(tfmodels.DynamoDBBackfillConfigAttrTypes)
		assert.False(t, config.SetAttribute(ctx, path.Root("dynamodb_backfill_config"), unknownBackfill).HasError())
		req := resource.ModifyPlanRequest{
			Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: config.Raw},
			State:  tfsdk.State{Schema: schemaResp.Schema, Raw: pipelineValue(false)},
			Plan:   tfsdk.Plan{Schema: schemaResp.Schema, Raw: config.Raw},
		}
		resp := &resource.ModifyPlanResponse{Plan: req.Plan}
		(&PipelineResource{}).ModifyPlan(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
		var a types.Bool
		resp.Plan.GetAttribute(ctx, path.Root("tables").AtMapKey("a").AtName("skip_backfill"), &a)
		assert.True(t, a.IsUnknown())
	}
	{
		// Backfill already on: the plan is left alone
		a, b, diags := modifyPlan(true, true, types.BoolValue(true))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.BoolValue(true), a)
		assert.Equal(t, types.BoolValue(true), b)
	}
}

func TestValidateDynamoDBBackfillPlan(t *testing.T) {
	ctx := t.Context()
	schemaResp := &resource.SchemaResponse{}
	(&PipelineResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	// run checks a plan whose backfill block is planned, with export_arn configured as given.
	run := func(planned types.Object, configuredARN types.String) diag.Diagnostics {
		pipeline, diags := tfmodels.PipelineFromAPIModel(ctx, openapi.PayloadsFullPipeline{
			Uuid:   uuid.New(),
			Name:   "test",
			Tables: []openapi.PayloadsTable{},
			// The API always returns the flush rules.
			AdvancedSettings: openapi.PayloadsPipelineAdvancedSettings{FlushIntervalSeconds: lib.ToPtr(30), BufferRows: lib.ToPtr(1000), FlushSizeKb: lib.ToPtr(1000)},
		})
		assert.False(t, diags.HasError(), diags)
		value := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
		setDiags := value.Set(ctx, pipeline)
		assert.False(t, setDiags.HasError(), setDiags)
		setDiags = value.SetAttribute(ctx, path.Root("dynamodb_backfill_config"), planned)
		assert.False(t, setDiags.HasError(), setDiags)
		plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: value.Raw}
		assert.False(t, value.SetAttribute(ctx, path.Root("dynamodb_backfill_config").AtName("export_arn"), configuredARN).HasError())
		req := resource.ModifyPlanRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: value.Raw}, Plan: plan}
		resp := &resource.ModifyPlanResponse{Plan: plan}
		validateDynamoDBBackfillPlan(ctx, req, resp)
		return resp.Diagnostics
	}
	backfill := func(enabled bool, exportARN types.String) types.Object {
		return types.ObjectValueMust(tfmodels.DynamoDBBackfillConfigAttrTypes, map[string]attr.Value{
			"enabled":         types.BoolValue(enabled),
			"bucket":          types.StringValue("exports"),
			"optional_folder": types.StringUnknown(),
			"export_arn":      exportARN,
			"max_parallelism": types.Int64Unknown(),
		})
	}
	{
		// Enabled with the ARN left out and nothing saved (new pipeline or new block): rejected
		diags := run(backfill(true, types.StringUnknown()), types.StringNull())
		assert.True(t, diags.HasError())
	}
	{
		// Enabled on an update whose saved ARN is empty: rejected
		diags := run(backfill(true, types.StringValue("")), types.StringNull())
		assert.True(t, diags.HasError())
	}
	{
		// Enabled with a saved ARN carried into the plan: fine
		diags := run(backfill(true, types.StringValue("arn")), types.StringNull())
		assert.False(t, diags.HasError(), diags)
	}
	{
		// Enabled with an ARN only known at apply: fine
		diags := run(backfill(true, types.StringUnknown()), types.StringUnknown())
		assert.False(t, diags.HasError(), diags)
	}
	{
		// Disabled, or the whole block only known at apply: fine
		assert.False(t, run(backfill(false, types.StringUnknown()), types.StringNull()).HasError())
		assert.False(t, run(types.ObjectUnknown(tfmodels.DynamoDBBackfillConfigAttrTypes), types.StringNull()).HasError())
	}
}
