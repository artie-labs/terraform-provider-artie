package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDataPlaneChangeRequiresReplace(t *testing.T) {
	{
		// Moving between data planes: replaced.
		assert.True(t, dataPlaneChangeRequiresReplace(types.StringValue("a"), types.StringValue("b")))
	}
	{
		// No data plane yet: set in place.
		assert.False(t, dataPlaneChangeRequiresReplace(types.StringValue(""), types.StringValue("b")))
		assert.False(t, dataPlaneChangeRequiresReplace(types.StringNull(), types.StringValue("b")))
	}
	{
		// Unchanged, cleared, or only known at apply time: left alone.
		assert.False(t, dataPlaneChangeRequiresReplace(types.StringValue("a"), types.StringValue("a")))
		assert.False(t, dataPlaneChangeRequiresReplace(types.StringValue("a"), types.StringValue("")))
		assert.False(t, dataPlaneChangeRequiresReplace(types.StringValue("a"), types.StringUnknown()))
	}
}
