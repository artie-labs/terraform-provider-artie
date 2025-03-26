package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/artieclient"
)

type MergePredicate struct {
	PartitionField types.String `tfsdk:"partition_field"`
}

var MergePredicateAttrTypes = map[string]attr.Type{
	"partition_field": types.StringType,
}

func (m MergePredicate) ToAPIModel() artieclient.MergePredicate {
	return artieclient.MergePredicate{PartitionField: m.PartitionField.ValueString()}
}

func MergePredicatesFromAPIModel(ctx context.Context, apiMergePredicates *[]artieclient.MergePredicate) (types.List, diag.Diagnostics) {
	attrTypes := MergePredicateAttrTypes
	if apiMergePredicates == nil {
		return types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, []attr.Value{})
	}

	var diags diag.Diagnostics
	preds := []attr.Value{}
	for _, mp := range *apiMergePredicates {
		pred, predDiags := types.ObjectValueFrom(ctx, attrTypes, MergePredicate{PartitionField: types.StringValue(mp.PartitionField)})
		diags.Append(predDiags...)
		preds = append(preds, pred)
	}

	mergePredicates, listDiags := types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, preds)
	diags.Append(listDiags...)

	return mergePredicates, diags
}

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias           types.String `tfsdk:"alias"`
	ExcludeColumns  types.List   `tfsdk:"columns_to_exclude"`
	IncludeColumns  types.List   `tfsdk:"columns_to_include"`
	ColumnsToHash   types.List   `tfsdk:"columns_to_hash"`
	SkipDeletes     types.Bool   `tfsdk:"skip_deletes"`
	MergePredicates types.List   `tfsdk:"merge_predicates"`
}

var TableAttrTypes = map[string]attr.Type{
	"uuid":                  types.StringType,
	"name":                  types.StringType,
	"schema":                types.StringType,
	"enable_history_mode":   types.BoolType,
	"individual_deployment": types.BoolType,
	"is_partitioned":        types.BoolType,
	"alias":                 types.StringType,
	"columns_to_exclude":    types.ListType{ElemType: types.StringType},
	"columns_to_include":    types.ListType{ElemType: types.StringType},
	"columns_to_hash":       types.ListType{ElemType: types.StringType},
	"skip_deletes":          types.BoolType,
	"merge_predicates":      types.ListType{ElemType: types.ObjectType{AttrTypes: MergePredicateAttrTypes}},
}

func (t Table) ToAPIModel(ctx context.Context) (artieclient.Table, diag.Diagnostics) {
	tableUUID := uuid.Nil
	var diags diag.Diagnostics
	if t.UUID.ValueString() != "" {
		tableUUID, diags = parseUUID(t.UUID)
	}

	colsToExclude, excludeDiags := parseOptionalList[string](ctx, t.ExcludeColumns)
	diags.Append(excludeDiags...)

	colsToInclude, includeDiags := parseOptionalList[string](ctx, t.IncludeColumns)
	diags.Append(includeDiags...)

	colsToHash, hashDiags := parseOptionalList[string](ctx, t.ColumnsToHash)
	diags.Append(hashDiags...)

	mergePredicates, mergePredDiags := parseOptionalList[MergePredicate](ctx, t.MergePredicates)
	diags.Append(mergePredDiags...)
	var clientMergePreds *[]artieclient.MergePredicate
	if mergePredicates != nil && len(*mergePredicates) > 0 {
		clientMPs := []artieclient.MergePredicate{}
		for _, mp := range *mergePredicates {
			clientMPs = append(clientMPs, mp.ToAPIModel())
		}
		clientMergePreds = &clientMPs
	}

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
		Alias:                t.Alias.ValueStringPointer(),
		ExcludeColumns:       colsToExclude,
		IncludeColumns:       colsToInclude,
		ColumnsToHash:        colsToHash,
		SkipDeletes:          t.SkipDeletes.ValueBoolPointer(),
		MergePredicates:      clientMergePreds,
	}, diags
}

func TablesFromAPIModel(ctx context.Context, apiModelTables []artieclient.Table) (map[string]Table, diag.Diagnostics) {
	tables := map[string]Table{}
	var diags diag.Diagnostics
	for _, apiTable := range apiModelTables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}

		colsToExclude, excludeDiags := optionalStringListToStringValue(ctx, apiTable.ExcludeColumns)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := optionalStringListToStringValue(ctx, apiTable.IncludeColumns)
		diags.Append(includeDiags...)

		colsToHash, hashDiags := optionalStringListToStringValue(ctx, apiTable.ColumnsToHash)
		diags.Append(hashDiags...)

		mergePredicates, mergePredDiags := MergePredicatesFromAPIModel(ctx, apiTable.MergePredicates)
		diags.Append(mergePredDiags...)

		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                types.StringPointerValue(apiTable.Alias),
			ExcludeColumns:       colsToExclude,
			IncludeColumns:       colsToInclude,
			ColumnsToHash:        colsToHash,
			SkipDeletes:          types.BoolPointerValue(apiTable.SkipDeletes),
			MergePredicates:      mergePredicates,
		}
	}

	if diags.HasError() {
		return map[string]Table{}, diags
	}

	return tables, diags
}
