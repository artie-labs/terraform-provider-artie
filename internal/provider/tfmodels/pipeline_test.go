package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/artieclient"
)

func ptr[T any](v T) *T { return &v }

func TestPipelineFromAPIModel_DisableAlertsReadsBackAsFalse(t *testing.T) {
	// disable_alerts is an "absent means off" toggle. The backend persists false as
	// nil, so both an omitted value and an explicit false must read back as a stable
	// false (never null), otherwise Terraform's post-apply consistency check errors:
	// ".disable_alerts: was cty.False, but now null".
	base := artieclient.Pipeline{
		UUID: uuid.New(),
		BasePipeline: artieclient.BasePipeline{
			Name:             "test",
			Tables:           []artieclient.Table{},
			AdvancedSettings: &artieclient.AdvancedSettings{},
		},
	}

	{
		// DisableAlerts omitted (nil) -> false
		pipeline, diags := PipelineFromAPIModel(t.Context(), base)
		assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
		assert.False(t, pipeline.DisableAlerts.IsNull(), "disable_alerts should not be null when omitted")
		assert.False(t, pipeline.DisableAlerts.ValueBool())
	}
	{
		// DisableAlerts explicitly false -> false
		base.AdvancedSettings.DisableAlerts = ptr(false)
		pipeline, diags := PipelineFromAPIModel(t.Context(), base)
		assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
		assert.False(t, pipeline.DisableAlerts.IsNull(), "disable_alerts should not be null when false")
		assert.False(t, pipeline.DisableAlerts.ValueBool())
	}
	{
		// DisableAlerts explicitly true -> true
		base.AdvancedSettings.DisableAlerts = ptr(true)
		pipeline, diags := PipelineFromAPIModel(t.Context(), base)
		assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
		assert.True(t, pipeline.DisableAlerts.ValueBool())
	}
}

func TestPipelineToAPIBaseModel_TurboSettings(t *testing.T) {
	tablesMap, mapDiags := types.MapValueFrom(t.Context(), types.ObjectType{AttrTypes: TableAttrTypes}, map[string]Table{})
	assert.False(t, mapDiags.HasError(), "unexpected diags: %v", mapDiags)

	pipeline := Pipeline{
		Name:                         types.StringValue("test"),
		Tables:                       tablesMap,
		DestinationConfig:            &PipelineDestinationConfig{},
		TurboWarehouse:               types.StringValue("ARTIE_WEB_WH_LARGE"),
		TurboRowThreshold:            types.Int64Value(500000),
		TurboLatencyThresholdMinutes: types.Int64Value(30),
	}

	apiModel, diags := pipeline.ToAPIBaseModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
	assert.NotNil(t, apiModel.AdvancedSettings)
	assert.Equal(t, "ARTIE_WEB_WH_LARGE", *apiModel.AdvancedSettings.TurboWarehouse)
	assert.Equal(t, int64(500000), *apiModel.AdvancedSettings.TurboRowThreshold)
	assert.Equal(t, int64(30), *apiModel.AdvancedSettings.TurboLatencyThresholdMinutes)
}

func TestPipelineToAPIBaseModel_OmittedTurboSettings(t *testing.T) {
	tablesMap, mapDiags := types.MapValueFrom(t.Context(), types.ObjectType{AttrTypes: TableAttrTypes}, map[string]Table{})
	assert.False(t, mapDiags.HasError(), "unexpected diags: %v", mapDiags)

	pipeline := Pipeline{
		Name:              types.StringValue("test"),
		Tables:            tablesMap,
		DestinationConfig: &PipelineDestinationConfig{},
	}

	apiModel, diags := pipeline.ToAPIBaseModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
	assert.NotNil(t, apiModel.AdvancedSettings)
	assert.Nil(t, apiModel.AdvancedSettings.TurboWarehouse)
	assert.Nil(t, apiModel.AdvancedSettings.TurboRowThreshold)
	assert.Nil(t, apiModel.AdvancedSettings.TurboLatencyThresholdMinutes)
}

func TestPipelineFromAPIModel_TurboSettings(t *testing.T) {
	apiModel := artieclient.Pipeline{
		UUID: uuid.New(),
		BasePipeline: artieclient.BasePipeline{
			Name:   "test",
			Tables: []artieclient.Table{},
			AdvancedSettings: &artieclient.AdvancedSettings{
				TurboWarehouse:               ptr("ARTIE_WEB_WH_LARGE"),
				TurboRowThreshold:            ptr[int64](500000),
				TurboLatencyThresholdMinutes: ptr[int64](30),
			},
		},
	}

	pipeline, diags := PipelineFromAPIModel(t.Context(), apiModel)
	assert.False(t, diags.HasError(), "unexpected diags: %v", diags)
	assert.Equal(t, "ARTIE_WEB_WH_LARGE", pipeline.TurboWarehouse.ValueString())
	assert.Equal(t, int64(500000), pipeline.TurboRowThreshold.ValueInt64())
	assert.Equal(t, int64(30), pipeline.TurboLatencyThresholdMinutes.ValueInt64())
}

func TestFlushConfigFromAPIModel(t *testing.T) {
	{
		// zero object
		var object types.Object
		flushConfig, diags := buildFlushConfig(t.Context(), object)
		assert.False(t, diags.HasError(), "expected no error when creating nil object")
		assert.Nil(t, flushConfig)
	}
	{
		// null object
		nullObject := types.ObjectNull(flushAttrTypes)
		flushConfig, diags := buildFlushConfig(t.Context(), nullObject)
		assert.False(t, diags.HasError(), "expected no error when creating null object")
		assert.Nil(t, flushConfig)
	}
	{
		// unknown object
		unknownObject := types.ObjectUnknown(flushAttrTypes)
		flushConfig, diags := buildFlushConfig(t.Context(), unknownObject)
		assert.False(t, diags.HasError(), "expected no error when creating unknown object")
		assert.Nil(t, flushConfig)
	}
	{
		// object is set
		flushObject, diags := types.ObjectValue(flushAttrTypes, map[string]attr.Value{
			"flush_interval_seconds": types.Int64Value(100),
			"buffer_rows":            types.Int64Value(5000),
			"flush_size_kb":          types.Int64Value(1000),
		})

		assert.False(t, diags.HasError(), "failed to create flush object")

		flushConfig, diags := buildFlushConfig(t.Context(), flushObject)
		assert.False(t, diags.HasError(), "failed to create flush config")
		assert.NotNil(t, flushConfig)

		assert.Equal(t, flushConfig.FlushIntervalSeconds.ValueInt64(), int64(100))
		assert.Equal(t, flushConfig.BufferRows.ValueInt64(), int64(5000))
		assert.Equal(t, flushConfig.FlushSizeKB.ValueInt64(), int64(1000))

		apiFlushConfig := flushConfig.ToAPIModel()
		assert.Equal(t, apiFlushConfig.FlushIntervalSeconds, int64(100))
		assert.Equal(t, apiFlushConfig.BufferRows, int64(5000))
		assert.Equal(t, apiFlushConfig.FlushSizeKB, int64(1000))
	}
}
