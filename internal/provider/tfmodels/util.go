package tfmodels

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToPtr[T any](v T) *T {
	return &v
}

func parseUUID(value types.String) uuid.UUID {
	// TODO: [uuid.MustParse] will panic if it fails, we should return an error instead.
	return uuid.MustParse(value.ValueString())
}

func ParseOptionalUUID(value types.String) *uuid.UUID {
	if value.IsNull() || len(value.ValueString()) == 0 {
		return nil
	}

	// TODO: [uuid.MustParse] will panic if it fails, we should return an error instead.
	return ToPtr(uuid.MustParse(value.ValueString()))
}

func optionalUUIDToStringValue(value *uuid.UUID) types.String {
	if value == nil {
		return types.StringValue("")
	}
	return types.StringValue(value.String())
}

func parseOptionalBool(value types.Bool) *bool {
	if value.IsNull() {
		return nil
	}
	return ToPtr(value.ValueBool())
}

func optionalBoolToBoolValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}

func parseOptionalString(value types.String) *string {
	if value.IsNull() {
		return nil
	}
	return ToPtr(value.ValueString())
}

func optionalStringToStringValue(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}
