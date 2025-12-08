package tfmodels

import (
	"context"
	"fmt"

	"github.com/artie-labs/transfer/lib/kafkalib"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/artieclient"
)

type MergePredicate struct {
	PartitionField types.String `tfsdk:"partition_field"`
	PartitionType  types.String `tfsdk:"partition_type"`
}

var MergePredicateAttrTypes = map[string]attr.Type{
	"partition_field": types.StringType,
	"partition_type":  types.StringType,
}

func (m MergePredicate) ToAPIModel() artieclient.MergePredicate {
	return artieclient.MergePredicate{PartitionField: m.PartitionField.ValueString(), PartitionType: m.PartitionType.ValueString()}
}

func MergePredicatesFromAPIModel(ctx context.Context, apiMergePredicates *[]artieclient.MergePredicate) (types.List, diag.Diagnostics) {
	attrTypes := MergePredicateAttrTypes
	if apiMergePredicates == nil {
		return types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, []attr.Value{})
	}

	var diags diag.Diagnostics
	preds := []attr.Value{}
	for _, mp := range *apiMergePredicates {
		var partitionType types.String
		if mp.PartitionType == "" {
			partitionType = types.StringNull()
		} else {
			partitionType = types.StringValue(mp.PartitionType)
		}

		pred, predDiags := types.ObjectValueFrom(ctx, attrTypes, MergePredicate{PartitionField: types.StringValue(mp.PartitionField), PartitionType: partitionType})
		diags.Append(predDiags...)
		preds = append(preds, pred)
	}

	mergePredicates, listDiags := types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, preds)
	diags.Append(listDiags...)

	return mergePredicates, diags
}

type SoftPartitioning struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	PartitionFrequency types.String `tfsdk:"partition_frequency"`
	PartitionColumn    types.String `tfsdk:"partition_column"`
	MaxPartitions      types.Int32  `tfsdk:"max_partitions"`
}

func (s SoftPartitioning) ToAPIModel() *artieclient.SoftPartitioning {
	return &artieclient.SoftPartitioning{
		Enabled:            s.Enabled.ValueBool(),
		PartitionFrequency: kafkalib.PartitionFrequency(s.PartitionFrequency.ValueString()),
		PartitionColumn:    s.PartitionColumn.ValueString(),
		MaxPartitions:      int(s.MaxPartitions.ValueInt32()),
	}
}

var SoftPartitioningAttrTypes = map[string]attr.Type{
	"enabled":             types.BoolType,
	"partition_frequency": types.StringType,
	"partition_column":    types.StringType,
	"max_partitions":      types.Int32Type,
}

func SoftPartitioningFromAPIModel(ctx context.Context, apiSoftPartitioning *artieclient.SoftPartitioning) (types.Object, diag.Diagnostics) {
	attrTypes := SoftPartitioningAttrTypes
	if apiSoftPartitioning == nil {
		return types.ObjectNull(attrTypes), nil
	}

	return types.ObjectValue(attrTypes, map[string]attr.Value{
		"enabled":             types.BoolValue(apiSoftPartitioning.Enabled),
		"partition_frequency": types.StringValue(string(apiSoftPartitioning.PartitionFrequency)),
		"partition_column":    types.StringValue(apiSoftPartitioning.PartitionColumn),
		"max_partitions":      types.Int32Value(int32(apiSoftPartitioning.MaxPartitions)),
	})
}

type Table struct {
	UUID              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	Schema            types.String `tfsdk:"schema"`
	EnableHistoryMode types.Bool   `tfsdk:"enable_history_mode"`
	IsPartitioned     types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias                types.String `tfsdk:"alias"`
	ExcludeColumns       types.List   `tfsdk:"columns_to_exclude"`
	IncludeColumns       types.List   `tfsdk:"columns_to_include"`
	ColumnsToHash        types.List   `tfsdk:"columns_to_hash"`
	SkipDeletes          types.Bool   `tfsdk:"skip_deletes"`
	UnifyAcrossSchemas   types.Bool   `tfsdk:"unify_across_schemas"`
	UnifyAcrossDatabases types.Bool   `tfsdk:"unify_across_databases"`
	MergePredicates      types.List   `tfsdk:"merge_predicates"`
	SoftPartitioning     types.Object `tfsdk:"soft_partitioning"`
	BackfillHistoryTable types.Bool   `tfsdk:"backfill_history_table"`
}

var TableAttrTypes = map[string]attr.Type{
	"uuid":                   types.StringType,
	"name":                   types.StringType,
	"schema":                 types.StringType,
	"enable_history_mode":    types.BoolType,
	"is_partitioned":         types.BoolType,
	"alias":                  types.StringType,
	"columns_to_exclude":     types.ListType{ElemType: types.StringType},
	"columns_to_include":     types.ListType{ElemType: types.StringType},
	"columns_to_hash":        types.ListType{ElemType: types.StringType},
	"skip_deletes":           types.BoolType,
	"unify_across_schemas":   types.BoolType,
	"unify_across_databases": types.BoolType,
	"merge_predicates":       types.ListType{ElemType: types.ObjectType{AttrTypes: MergePredicateAttrTypes}},
	"soft_partitioning":      types.ObjectType{AttrTypes: SoftPartitioningAttrTypes},
	"backfill_history_table": types.BoolType,
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
		var clientMPs []artieclient.MergePredicate
		for _, mp := range *mergePredicates {
			clientMPs = append(clientMPs, mp.ToAPIModel())
		}

		clientMergePreds = &clientMPs
	}

	softPartitioning, softPartitioningDiags := parseOptionalObject[SoftPartitioning](ctx, &t.SoftPartitioning)
	var clientSoftPartitioning *artieclient.SoftPartitioning
	if softPartitioning != nil {
		clientSoftPartitioning = softPartitioning.ToAPIModel()
	}
	diags.Append(softPartitioningDiags...)

	if diags.HasError() {
		return artieclient.Table{}, diags
	}

	return artieclient.Table{
		UUID:              tableUUID,
		Name:              t.Name.ValueString(),
		Schema:            t.Schema.ValueString(),
		EnableHistoryMode: t.EnableHistoryMode.ValueBool(),
		IsPartitioned:     t.IsPartitioned.ValueBool(),
		AdvancedSettings: artieclient.AdvancedTableSettings{
			Alias:                      t.Alias.ValueStringPointer(),
			ExcludeColumns:             colsToExclude,
			IncludeColumns:             colsToInclude,
			ColumnsToHash:              colsToHash,
			SkipDeletes:                t.SkipDeletes.ValueBoolPointer(),
			UnifyAcrossSchemas:         t.UnifyAcrossSchemas.ValueBoolPointer(),
			UnifyAcrossDatabases:       t.UnifyAcrossDatabases.ValueBoolPointer(),
			MergePredicates:            clientMergePreds,
			SoftPartitioning:           clientSoftPartitioning,
			ShouldBackfillHistoryTable: t.BackfillHistoryTable.ValueBoolPointer(),
		},
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

		colsToExclude, excludeDiags := optionalStringListToListValue(ctx, apiTable.AdvancedSettings.ExcludeColumns)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := optionalStringListToListValue(ctx, apiTable.AdvancedSettings.IncludeColumns)
		diags.Append(includeDiags...)

		colsToHash, hashDiags := optionalStringListToListValue(ctx, apiTable.AdvancedSettings.ColumnsToHash)
		diags.Append(hashDiags...)

		mergePredicates, mergePredDiags := MergePredicatesFromAPIModel(ctx, apiTable.AdvancedSettings.MergePredicates)
		diags.Append(mergePredDiags...)

		softPartitioning, softPartitioningDiags := SoftPartitioningFromAPIModel(ctx, apiTable.AdvancedSettings.SoftPartitioning)
		diags.Append(softPartitioningDiags...)

		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                types.StringPointerValue(apiTable.AdvancedSettings.Alias),
			ExcludeColumns:       colsToExclude,
			IncludeColumns:       colsToInclude,
			ColumnsToHash:        colsToHash,
			SkipDeletes:          types.BoolPointerValue(apiTable.AdvancedSettings.SkipDeletes),
			UnifyAcrossSchemas:   types.BoolPointerValue(apiTable.AdvancedSettings.UnifyAcrossSchemas),
			UnifyAcrossDatabases: types.BoolPointerValue(apiTable.AdvancedSettings.UnifyAcrossDatabases),
			MergePredicates:      mergePredicates,
			SoftPartitioning:     softPartitioning,
			BackfillHistoryTable: types.BoolPointerValue(apiTable.AdvancedSettings.ShouldBackfillHistoryTable),
		}
	}

	if diags.HasError() {
		return map[string]Table{}, diags
	}

	return tables, diags
}
