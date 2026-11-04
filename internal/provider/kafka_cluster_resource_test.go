package provider

import (
	"terraform-provider-artie/internal/provider/tfmodels"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
