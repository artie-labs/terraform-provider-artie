package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias          types.String `tfsdk:"alias"`
	ExcludeColumns types.List   `tfsdk:"columns_to_exclude"`
	ColumnsToHash  types.List   `tfsdk:"columns_to_hash"`
	SkipDeletes    types.Bool   `tfsdk:"skip_deletes"`
}

func (t Table) ToAPIModel(ctx context.Context) (artieclient.Table, diag.Diagnostics) {
	tableUUID := uuid.Nil
	if t.UUID.ValueString() != "" {
		tableUUID = uuid.MustParse(t.UUID.ValueString())
	}

	colsToExclude, diags := parseOptionalStringList(ctx, t.ExcludeColumns)
	if diags.HasError() {
		return artieclient.Table{}, diags
	}
	colsToHash, diags := parseOptionalStringList(ctx, t.ColumnsToHash)
	if diags.HasError() {
		return artieclient.Table{}, diags
	}

	return artieclient.Table{
		UUID:                 tableUUID,
		Name:                 t.Name.ValueString(),
		Schema:               t.Schema.ValueString(),
		EnableHistoryMode:    t.EnableHistoryMode.ValueBool(),
		IndividualDeployment: t.IndividualDeployment.ValueBool(),
		IsPartitioned:        t.IsPartitioned.ValueBool(),
		Alias:                parseOptionalString(t.Alias),
		ExcludeColumns:       colsToExclude,
		ColumnsToHash:        colsToHash,
		SkipDeletes:          parseOptionalBool(t.SkipDeletes),
	}, nil
}

func TablesFromAPIModel(ctx context.Context, apiModelTables []artieclient.Table) (map[string]Table, diag.Diagnostics) {
	tables := map[string]Table{}
	for _, apiTable := range apiModelTables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}

		colsToExclude, diags := optionalStringListToStringValue(ctx, apiTable.ExcludeColumns)
		if diags.HasError() {
			return map[string]Table{}, diags
		}
		colsToHash, diags := optionalStringListToStringValue(ctx, apiTable.ColumnsToHash)
		if diags.HasError() {
			return map[string]Table{}, diags
		}

		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                optionalStringToStringValue(apiTable.Alias),
			ExcludeColumns:       colsToExclude,
			ColumnsToHash:        colsToHash,
			SkipDeletes:          optionalBoolToBoolValue(apiTable.SkipDeletes),
		}
	}

	return tables, nil
}
