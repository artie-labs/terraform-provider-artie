package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type MergePredicate struct {
	PartitionField string `tfsdk:"partition_field"`
}

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias           types.String      `tfsdk:"alias"`
	ExcludeColumns  types.List        `tfsdk:"columns_to_exclude"`
	ColumnsToHash   types.List        `tfsdk:"columns_to_hash"`
	SkipDeletes     types.Bool        `tfsdk:"skip_deletes"`
	MergePredicates *[]MergePredicate `tfsdk:"merge_predicates"`
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

	colsToHash, hashDiags := parseOptionalStringList(ctx, t.ColumnsToHash)
	diags.Append(hashDiags...)
	if diags.HasError() {
		return artieclient.Table{}, diags
	}

	var mergePredicates *[]artieclient.MergePredicate
	if t.MergePredicates != nil {
		preds := []artieclient.MergePredicate{}
		for _, mp := range *t.MergePredicates {
			preds = append(preds, artieclient.MergePredicate{PartitionField: mp.PartitionField})
		}
		mergePredicates = &preds
	}

	return artieclient.Table{
		UUID:                 tableUUID,
		Name:                 t.Name.ValueString(),
		Schema:               t.Schema.ValueString(),
		EnableHistoryMode:    t.EnableHistoryMode.ValueBool(),
		IndividualDeployment: t.IndividualDeployment.ValueBool(),
		IsPartitioned:        t.IsPartitioned.ValueBool(),
		Alias:                t.Alias.ValueStringPointer(),
		ExcludeColumns:       colsToExclude,
		ColumnsToHash:        colsToHash,
		SkipDeletes:          t.SkipDeletes.ValueBoolPointer(),
		MergePredicates:      mergePredicates,
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

		colsToHash, hashDiags := optionalStringListToStringValue(ctx, apiTable.ColumnsToHash)
		diags.Append(hashDiags...)
		if diags.HasError() {
			return map[string]Table{}, diags
		}

		var mergePredicates *[]MergePredicate
		if apiTable.MergePredicates != nil {
			preds := []MergePredicate{}
			for _, mp := range *apiTable.MergePredicates {
				preds = append(preds, MergePredicate{PartitionField: mp.PartitionField})
			}
			mergePredicates = &preds
		}

		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                types.StringPointerValue(apiTable.Alias),
			ExcludeColumns:       colsToExclude,
			ColumnsToHash:        colsToHash,
			SkipDeletes:          types.BoolPointerValue(apiTable.SkipDeletes),
			MergePredicates:      mergePredicates,
		}
	}

	return tables, nil
}
