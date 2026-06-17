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
