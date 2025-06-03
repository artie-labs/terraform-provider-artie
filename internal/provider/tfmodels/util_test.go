package tfmodels

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestOptionalStringListToListValue(t *testing.T) {
	{
		// nil slice
		var value *[]string
		listValue, diags := optionalStringListToListValue(t.Context(), value)
		assert.False(t, diags.HasError(), "expected no error when creating list value from nil slice")
		assert.False(t, listValue.IsNull(), "expected list value to be not null")
		assert.False(t, listValue.IsUnknown(), "expected list value to be not unknown")
		assert.Equal(t, 0, len(listValue.Elements()), "expected list value to be empty")
	}
	{
		// empty slice
		value := []string{}
		listValue, diags := optionalStringListToListValue(t.Context(), &value)
		assert.False(t, diags.HasError(), "expected no error when creating list value from empty slice")
		assert.False(t, listValue.IsNull(), "expected list value to be not null")
		assert.False(t, listValue.IsUnknown(), "expected list value to be not unknown")
		assert.Equal(t, 0, len(listValue.Elements()), "expected list value to be empty")
	}
	{
		// non-empty slice
		value := []string{"a", "b", "c"}
		listValue, diags := optionalStringListToListValue(t.Context(), &value)
		assert.False(t, diags.HasError(), "expected no error when creating list value from non-empty slice")
		assert.False(t, listValue.IsNull(), "expected list value to be not null")
		assert.False(t, listValue.IsUnknown(), "expected list value to be not unknown")
		assert.Equal(t, 3, len(listValue.Elements()), "expected list value to have 3 elements")

		val1, ok := listValue.Elements()[0].(types.String)
		assert.True(t, ok, "expected first element to be a string")
		assert.Equal(t, "a", val1.ValueString(), "expected list value to have 'a' as first element")

		val2, ok := listValue.Elements()[1].(types.String)
		assert.True(t, ok, "expected second element to be a string")
		assert.Equal(t, "b", val2.ValueString(), "expected list value to have 'b' as second element")

		val3, ok := listValue.Elements()[2].(types.String)
		assert.True(t, ok, "expected third element to be a string")
		assert.Equal(t, "c", val3.ValueString(), "expected list value to have 'c' as third element")
	}
}
