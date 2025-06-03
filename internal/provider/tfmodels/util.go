package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func parseUUID(value types.String) (uuid.UUID, diag.Diagnostics) {
	if value.ValueString() == "" {
		return uuid.UUID{}, []diag.Diagnostic{diag.NewErrorDiagnostic("UUID is empty", "")}
	}

	u, err := uuid.Parse(value.ValueString())
	if err != nil {
		return uuid.UUID{}, []diag.Diagnostic{diag.NewErrorDiagnostic("Unable to parse UUID", fmt.Sprintf("value: %q", value.ValueString()))}
	}

	return u, nil
}

func parseOptionalUUID(value types.String) (*uuid.UUID, diag.Diagnostics) {
	if value.ValueString() == "" {
		return nil, nil
	}

	u, diags := parseUUID(value)
	if diags.HasError() {
		return nil, diags
	}

	return &u, diags
}

func optionalUUIDToStringValue(value *uuid.UUID) types.String {
	if value == nil {
		return types.StringValue("")
	}
	return types.StringValue(value.String())
}

func parseOptionalList[T any](ctx context.Context, value types.List) (*[]T, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	elements := make([]T, 0, len(value.Elements()))
	diags := value.ElementsAs(ctx, &elements, false)

	return &elements, diags
}

func parseList[T any](ctx context.Context, value types.List) ([]T, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	elements := make([]T, 0, len(value.Elements()))
	diags := value.ElementsAs(ctx, &elements, false)

	return elements, diags
}

// optionalStringListToListValue converts a pointer to a slice of strings to a Terraform List value.
// If the pointer is nil, it returns an empty list rather than a null list.
func optionalStringListToListValue(ctx context.Context, value *[]string) (types.List, diag.Diagnostics) {
	if value == nil {
		return types.ListValueFrom(ctx, types.StringType, []types.String{})
	}

	return types.ListValueFrom(ctx, types.StringType, *value)
}
