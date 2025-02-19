package tfmodels

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func parseOptionalStringList(ctx context.Context, value types.List) (*[]string, diag.Diagnostics) {
	if value.IsNull() {
		return nil, nil
	}

	elements := make([]types.String, 0, len(value.Elements()))
	diags := value.ElementsAs(ctx, &elements, false)
	if diags.HasError() {
		return nil, diags
	}

	out := []string{}
	for _, el := range elements {
		if !el.IsNull() {
			out = append(out, el.ValueString())
		}
	}

	return &out, nil
}

func optionalStringListToStringValue(ctx context.Context, value *[]string) (types.List, diag.Diagnostics) {
	if value == nil {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, *value)
}
