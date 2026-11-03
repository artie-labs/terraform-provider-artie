package provider

import (
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
