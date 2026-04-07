package tfmodels

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

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

		assert.Equal(t, int64(100), flushConfig.FlushIntervalSeconds.ValueInt64())
		assert.Equal(t, int64(5000), flushConfig.BufferRows.ValueInt64())
		assert.Equal(t, int64(1000), flushConfig.FlushSizeKB.ValueInt64())
	}
}
