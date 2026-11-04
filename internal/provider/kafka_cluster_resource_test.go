package provider

import (
	"terraform-provider-artie/internal/provider/tfmodels"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestShrinks(t *testing.T) {
	assert.True(t, shrinks(types.Int64Value(3), types.Int64Value(2)))
	assert.False(t, shrinks(types.Int64Value(3), types.Int64Value(3)))
	assert.False(t, shrinks(types.Int64Value(3), types.Int64Value(4)))
	// Unknown or null on either side (create, import, a value from another resource) isn't checked.
	assert.False(t, shrinks(types.Int64Null(), types.Int64Value(2)))
	assert.False(t, shrinks(types.Int64Value(3), types.Int64Unknown()))
}

func TestValidateKafkaStorageAutoscale(t *testing.T) {
	settings := func(bumpAt, target types.Int64) tfmodels.KafkaClusterStorageAutoscale {
		return tfmodels.KafkaClusterStorageAutoscale{Enabled: types.BoolNull(), BumpAtPercent: bumpAt, TargetPercent: target, MaxGiB: types.Int64Null()}
	}
	assert.False(t, validateKafkaStorageAutoscale(settings(types.Int64Value(80), types.Int64Value(70))).HasError())
	assert.True(t, validateKafkaStorageAutoscale(settings(types.Int64Value(70), types.Int64Value(70))).HasError())
	// Unset values use the defaults: a bump at 40% is below the default target of 55%.
	assert.True(t, validateKafkaStorageAutoscale(settings(types.Int64Value(40), types.Int64Null())).HasError())
	assert.False(t, validateKafkaStorageAutoscale(settings(types.Int64Null(), types.Int64Null())).HasError())
	// Unknown values (from another resource) aren't checked.
	assert.False(t, validateKafkaStorageAutoscale(settings(types.Int64Unknown(), types.Int64Value(90))).HasError())
}

func TestValidateKafkaTieredStorage(t *testing.T) {
	settings := func(enabled types.Bool, localRetentionMS, localRetentionBytes types.Int64) tfmodels.KafkaClusterTieredStorage {
		return tfmodels.KafkaClusterTieredStorage{Enabled: enabled, LocalRetentionMS: localRetentionMS, LocalRetentionBytes: localRetentionBytes}
	}
	assert.False(t, validateKafkaTieredStorage(settings(types.BoolValue(true), types.Int64Value(86400000), types.Int64Value(1073741824))).HasError())
	// Unset, enabled defaults to true.
	assert.False(t, validateKafkaTieredStorage(settings(types.BoolNull(), types.Int64Value(86400000), types.Int64Null())).HasError())
	assert.False(t, validateKafkaTieredStorage(settings(types.BoolValue(false), types.Int64Null(), types.Int64Null())).HasError())
	// With tiered storage off, neither applies.
	assert.True(t, validateKafkaTieredStorage(settings(types.BoolValue(false), types.Int64Value(86400000), types.Int64Null())).HasError())
	assert.True(t, validateKafkaTieredStorage(settings(types.BoolValue(false), types.Int64Null(), types.Int64Unknown())).HasError())
	// An unknown enabled (from another resource) isn't checked.
	assert.False(t, validateKafkaTieredStorage(settings(types.BoolUnknown(), types.Int64Value(86400000), types.Int64Null())).HasError())
}

func TestKafkaTieredStoragePlanModifier(t *testing.T) {
	tieredStorage := func(enabled bool, localRetentionMS, localRetentionBytes types.Int64) types.Object {
		return types.ObjectValueMust(tfmodels.KafkaClusterTieredStorageAttrTypes, map[string]attr.Value{
			"enabled":               types.BoolValue(enabled),
			"local_retention_ms":    localRetentionMS,
			"local_retention_bytes": localRetentionBytes,
		})
	}
	plan := func(config, state types.Object) types.Object {
		req := planmodifier.ObjectRequest{ConfigValue: config, StateValue: state, PlanValue: types.ObjectUnknown(tfmodels.KafkaClusterTieredStorageAttrTypes)}
		resp := &planmodifier.ObjectResponse{PlanValue: req.PlanValue}
		kafkaTieredStoragePlanModifier{}.PlanModifyObject(t.Context(), req, resp)
		return resp.PlanValue
	}
	unset := types.ObjectNull(tfmodels.KafkaClusterTieredStorageAttrTypes)
	prior := tieredStorage(false, types.Int64Value(86400000), types.Int64Value(1073741824))

	// Removing the block turns tiered storage back on and drops the disk cap, keeping the days on disk.
	assert.Equal(t, tieredStorage(true, types.Int64Value(86400000), types.Int64Null()), plan(unset, prior))
	// On create, the days on disk are Artie's to fill in.
	assert.Equal(t, tieredStorage(true, types.Int64Unknown(), types.Int64Null()), plan(unset, unset))
	// A block in the config is planned from its own attributes.
	assert.True(t, plan(prior, prior).IsUnknown())
}

func TestKafkaClusterModifyPlan(t *testing.T) {
	ctx := t.Context()
	schemaResp := &resource.SchemaResponse{}
	(&KafkaClusterResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	modifyPlan := func(priorRegion, plannedRegion string, priorSize, plannedSize int64) diag.Diagnostics {
		state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
		plan := tfsdk.Plan{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
		var diags diag.Diagnostics
		diags.Append(state.SetAttribute(ctx, path.Root("region"), priorRegion)...)
		diags.Append(plan.SetAttribute(ctx, path.Root("region"), plannedRegion)...)
		for _, attr := range []string{"brokers", "storage_gib"} {
			diags.Append(state.SetAttribute(ctx, path.Root(attr), priorSize)...)
			diags.Append(plan.SetAttribute(ctx, path.Root(attr), plannedSize)...)
		}
		assert.False(t, diags.HasError(), diags)
		resp := &resource.ModifyPlanResponse{Plan: plan}
		(&KafkaClusterResource{}).ModifyPlan(ctx, resource.ModifyPlanRequest{State: state, Plan: plan}, resp)
		return resp.Diagnostics
	}

	shrinkDiags := modifyPlan("us-east-1", "us-east-1", 80, 50)
	if assert.Len(t, shrinkDiags, 2) {
		for _, d := range shrinkDiags {
			assert.Equal(t, "Kafka cluster can't shrink", d.Summary())
		}
	}
	assert.Empty(t, modifyPlan("us-east-1", "us-east-1", 50, 80))
	// A new region replaces the cluster, so it can start smaller.
	assert.Empty(t, modifyPlan("us-east-1", "us-west-2", 80, 50))
}
