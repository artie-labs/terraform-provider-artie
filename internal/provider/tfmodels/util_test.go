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
		assert.Equal(t, "a", listValue.Elements()[0].(types.String).ValueString(), "expected list value to have 'a' as first element")
		assert.Equal(t, "b", listValue.Elements()[1].(types.String).ValueString(), "expected list value to have 'b' as second element")
		assert.Equal(t, "c", listValue.Elements()[2].(types.String).ValueString(), "expected list value to have 'c' as third element")
	}
}
