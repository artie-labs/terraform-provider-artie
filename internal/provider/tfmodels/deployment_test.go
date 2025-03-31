package tfmodels

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentFlushConfigFromAPIModel(t *testing.T) {
	flushObject, diags := types.ObjectValue(flushAttrTypes, map[string]attr.Value{
		"flush_interval_seconds": types.Int64Value(100),
		"buffer_rows":            types.Int64Value(5000),
		"flush_size_kb":          types.Int64Value(1000),
	})

	assert.False(t, diags.HasError(), "failed to create flush config")

	var flushConfig *DeploymentFlushConfig
	flushConfigDiags := flushObject.As(t.Context(), &flushConfig, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})

	assert.False(t, flushConfigDiags.HasError(), "failed to convert flush config")
	assert.Equal(t, flushConfig.FlushIntervalSeconds.ValueInt64(), int64(100))
	assert.Equal(t, flushConfig.BufferRows.ValueInt64(), int64(5000))
	assert.Equal(t, flushConfig.FlushSizeKB.ValueInt64(), int64(1000))

	apiFlushConfig := flushConfig.ToAPIModel()
	assert.Equal(t, apiFlushConfig.FlushIntervalSeconds, int64(100))
	assert.Equal(t, apiFlushConfig.BufferRows, int64(5000))
	assert.Equal(t, apiFlushConfig.FlushSizeKB, int64(1000))
}
