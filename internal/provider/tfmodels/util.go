package tfmodels

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func parseUUID(value types.String) uuid.UUID {
	// TODO: [uuid.MustParse] will panic if it fails, we should return an error instead.
	return uuid.MustParse(value.ValueString())
}

func ParseOptionalUUID(value types.String) *uuid.UUID {
	if value.IsNull() || len(value.ValueString()) == 0 {
		return nil
	}

	// TODO: [uuid.MustParse] will panic if it fails, we should return an error instead.
	_uuid := uuid.MustParse(value.ValueString())
	return &_uuid
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
	_bool := value.ValueBool()
	return &_bool
}

func optionalBoolToBoolValue(value *bool) types.Bool {
	if value == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*value)
}

func parseOptionalString(value types.String) *string {
	if value.IsNull() {
		return nil
	}

	_str := value.ValueString()
	return &_str
}

func optionalStringToStringValue(value *string) types.String {
	if value == nil {
		return types.StringValue("")
	}
	return types.StringValue(*value)
}
