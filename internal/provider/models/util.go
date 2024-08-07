package models

import "github.com/google/uuid"

func parseOptionalUUID(value string) *uuid.UUID {
	if len(value) == 0 {
		return nil
	}

	// TODO: [uuid.MustParse] will panic if it fails, we should return an error instead.
	_uuid := uuid.MustParse(value)
	return &_uuid
}

func optionalUUIDToString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}
