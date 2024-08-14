package models

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
