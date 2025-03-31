package tfmodels

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentFlushConfigFromAPIModel(t *testing.T) {
	{
		// nil object
		var object types.Object
		flushConfig, diags := buildFlushConfig(t.Context(), object)
		assert.False(t, diags.HasError(), "expected no error when creating nil object")
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
