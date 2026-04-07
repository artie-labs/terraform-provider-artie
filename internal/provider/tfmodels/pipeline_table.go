package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type MergePredicate struct {
	PartitionField types.String `tfsdk:"partition_field"`
	PartitionType  types.String `tfsdk:"partition_type"`
}

var MergePredicateAttrTypes = map[string]attr.Type{
	"partition_field": types.StringType,
	"partition_type":  types.StringType,
}

func (m MergePredicate) ToAPIModel() openapi.PayloadsMergePredicates {
	return openapi.PayloadsMergePredicates{
		PartitionField: m.PartitionField.ValueStringPointer(),
		PartitionType:  m.PartitionType.ValueStringPointer(),
	}
}

func MergePredicatesFromAPIModel(ctx context.Context, apiMergePredicates *[]openapi.PayloadsMergePredicates) (types.List, diag.Diagnostics) {
	attrTypes := MergePredicateAttrTypes
	if apiMergePredicates == nil {
		return types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, []attr.Value{})
	}

	var diags diag.Diagnostics
	preds := []attr.Value{}
	for _, mp := range *apiMergePredicates {
		var partitionType types.String
		if mp.PartitionType == nil || *mp.PartitionType == "" {
			partitionType = types.StringNull()
		} else {
			partitionType = types.StringValue(*mp.PartitionType)
		}

		pred, predDiags := types.ObjectValueFrom(ctx, attrTypes, MergePredicate{
			PartitionField: types.StringValue(lib.RemovePtr(mp.PartitionField)),
			PartitionType:  partitionType,
		})
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

func (s SoftPartitioning) ToAPIModel() *openapi.PayloadsSoftPartitioning {
	return &openapi.PayloadsSoftPartitioning{
		Enabled:            s.Enabled.ValueBoolPointer(),
		PartitionFrequency: s.PartitionFrequency.ValueStringPointer(),
		PartitionColumn:    s.PartitionColumn.ValueStringPointer(),
		MaxPartitions:      lib.ToPtr(int(s.MaxPartitions.ValueInt32())),
	}
}

var SoftPartitioningAttrTypes = map[string]attr.Type{
	"enabled":             types.BoolType,
	"partition_frequency": types.StringType,
	"partition_column":    types.StringType,
	"max_partitions":      types.Int32Type,
}

func SoftPartitioningFromAPIModel(ctx context.Context, apiSoftPartitioning *openapi.PayloadsSoftPartitioning) (types.Object, diag.Diagnostics) {
	attrTypes := SoftPartitioningAttrTypes
	if apiSoftPartitioning == nil {
		return types.ObjectNull(attrTypes), nil
	}

	return types.ObjectValue(attrTypes, map[string]attr.Value{
		"enabled":             types.BoolValue(lib.RemovePtr(apiSoftPartitioning.Enabled)),
		"partition_frequency": types.StringValue(lib.RemovePtr(apiSoftPartitioning.PartitionFrequency)),
		"partition_column":    types.StringValue(lib.RemovePtr(apiSoftPartitioning.PartitionColumn)),
		"max_partitions":      types.Int32Value(int32(lib.RemovePtr(apiSoftPartitioning.MaxPartitions))),
	})
}

type Table struct {
	UUID               types.String `tfsdk:"uuid"`
	Name               types.String `tfsdk:"name"`
	Schema             types.String `tfsdk:"schema"`
	EnableHistoryMode  types.Bool   `tfsdk:"enable_history_mode"`
	DisableReplication types.Bool   `tfsdk:"disable_replication"`
	IsPartitioned      types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias                types.String `tfsdk:"alias"`
	ExcludeColumns       types.List   `tfsdk:"columns_to_exclude"`
	IncludeColumns       types.List   `tfsdk:"columns_to_include"`
	ColumnsToHash        types.List   `tfsdk:"columns_to_hash"`
	ColumnsToCompress    types.List   `tfsdk:"columns_to_compress"`
	ColumnsToEncrypt     types.List   `tfsdk:"columns_to_encrypt"`
	SkipDeletes          types.Bool   `tfsdk:"skip_deletes"`
	UnifyAcrossSchemas   types.Bool   `tfsdk:"unify_across_schemas"`
	UnifyAcrossDatabases types.Bool   `tfsdk:"unify_across_databases"`
	MergePredicates      types.List   `tfsdk:"merge_predicates"`
	SoftPartitioning     types.Object `tfsdk:"soft_partitioning"`
	BackfillHistoryTable types.Bool   `tfsdk:"backfill_history_table"`
	CTIDBackfill         types.Bool   `tfsdk:"ctid_backfill"`
	CTIDChunkSize        types.Int64  `tfsdk:"ctid_chunk_size"`
	CTIDMaxParallelism   types.Int64  `tfsdk:"ctid_max_parallelism"`
	SkipBackfill         types.Bool   `tfsdk:"skip_backfill"`
}

var TableAttrTypes = map[string]attr.Type{
	"uuid":                   types.StringType,
	"name":                   types.StringType,
	"schema":                 types.StringType,
	"enable_history_mode":    types.BoolType,
	"disable_replication":    types.BoolType,
	"is_partitioned":         types.BoolType,
	"alias":                  types.StringType,
	"columns_to_exclude":     types.ListType{ElemType: types.StringType},
	"columns_to_include":     types.ListType{ElemType: types.StringType},
	"columns_to_hash":        types.ListType{ElemType: types.StringType},
	"columns_to_compress":    types.ListType{ElemType: types.StringType},
	"columns_to_encrypt":     types.ListType{ElemType: types.StringType},
	"skip_deletes":           types.BoolType,
	"unify_across_schemas":   types.BoolType,
	"unify_across_databases": types.BoolType,
	"merge_predicates":       types.ListType{ElemType: types.ObjectType{AttrTypes: MergePredicateAttrTypes}},
	"soft_partitioning":      types.ObjectType{AttrTypes: SoftPartitioningAttrTypes},
	"backfill_history_table": types.BoolType,
	"ctid_backfill":          types.BoolType,
	"ctid_chunk_size":        types.Int64Type,
	"ctid_max_parallelism":   types.Int64Type,
	"skip_backfill":          types.BoolType,
}

func (t Table) parseAdvancedFields(ctx context.Context) (parsedAdvanced, diag.Diagnostics) {
	var diags diag.Diagnostics

	colsToExclude, excludeDiags := parseOptionalList[string](ctx, t.ExcludeColumns)
	diags.Append(excludeDiags...)

	colsToInclude, includeDiags := parseOptionalList[string](ctx, t.IncludeColumns)
	diags.Append(includeDiags...)

	colsToHash, hashDiags := parseOptionalList[string](ctx, t.ColumnsToHash)
	diags.Append(hashDiags...)

	colsToCompress, compressDiags := parseOptionalList[string](ctx, t.ColumnsToCompress)
	diags.Append(compressDiags...)

	colsToEncrypt, encryptDiags := parseOptionalList[string](ctx, t.ColumnsToEncrypt)
	diags.Append(encryptDiags...)

	mergePredicates, mergePredDiags := parseOptionalList[MergePredicate](ctx, t.MergePredicates)
	diags.Append(mergePredDiags...)
	var apiMergePreds *[]openapi.PayloadsMergePredicates
	if mergePredicates != nil && len(*mergePredicates) > 0 {
		var mps []openapi.PayloadsMergePredicates
		for _, mp := range *mergePredicates {
			mps = append(mps, mp.ToAPIModel())
		}
		apiMergePreds = &mps
	}

	softPartitioning, softPartitioningDiags := parseOptionalObject[SoftPartitioning](ctx, &t.SoftPartitioning)
	diags.Append(softPartitioningDiags...)
	var apiSoftPartitioning *openapi.PayloadsSoftPartitioning
	if softPartitioning != nil {
		apiSoftPartitioning = softPartitioning.ToAPIModel()
	}

	var apiCTIDSettings *openapi.PayloadsCTIDSettings
	if !t.CTIDBackfill.IsNull() && !t.CTIDBackfill.IsUnknown() {
		apiCTIDSettings = &openapi.PayloadsCTIDSettings{
			Enabled:        t.CTIDBackfill.ValueBoolPointer(),
			ChunkSize:      lib.ToPtr(int(t.CTIDChunkSize.ValueInt64())),
			MaxParallelism: lib.ToPtr(int(t.CTIDMaxParallelism.ValueInt64())),
		}
	}

	return parsedAdvanced{
		colsToExclude:   colsToExclude,
		colsToInclude:   colsToInclude,
		colsToHash:      colsToHash,
		colsToCompress:  colsToCompress,
		colsToEncrypt:   colsToEncrypt,
		mergePredicates: apiMergePreds,
		softPartitioning: apiSoftPartitioning,
		ctidSettings:    apiCTIDSettings,
	}, diags
}

type parsedAdvanced struct {
	colsToExclude    *[]string
	colsToInclude    *[]string
	colsToHash       *[]string
	colsToCompress   *[]string
	colsToEncrypt    *[]string
	mergePredicates  *[]openapi.PayloadsMergePredicates
	softPartitioning *openapi.PayloadsSoftPartitioning
	ctidSettings     *openapi.PayloadsCTIDSettings
}

// ToAPIPayload builds a PayloadsTablePayload for create/update requests.
func (t Table) ToAPIPayload(ctx context.Context) (openapi.PayloadsTablePayload, diag.Diagnostics) {
	var tableUUID *uuid.UUID
	var diags diag.Diagnostics
	if t.UUID.ValueString() != "" {
		u, uuidDiags := parseUUID(t.UUID)
		diags.Append(uuidDiags...)
		tableUUID = &u
	}

	adv, advDiags := t.parseAdvancedFields(ctx)
	diags.Append(advDiags...)
	if diags.HasError() {
		return openapi.PayloadsTablePayload{}, diags
	}

	return openapi.PayloadsTablePayload{
		Uuid:               tableUUID,
		Name:               lib.ToPtr(t.Name.ValueString()),
		Schema:             t.Schema.ValueStringPointer(),
		EnableHistoryMode:  t.EnableHistoryMode.ValueBoolPointer(),
		DisableReplication: t.DisableReplication.ValueBoolPointer(),
		IsPartitioned:      t.IsPartitioned.ValueBoolPointer(),
		AdvancedSettings: &openapi.PayloadsAdvancedTableSettingsPayload{
			Alias:                      t.Alias.ValueStringPointer(),
			ExcludeColumns:             adv.colsToExclude,
			IncludeColumns:             adv.colsToInclude,
			ColumnsToHash:              adv.colsToHash,
			ColumnsToCompress:          adv.colsToCompress,
			ColumnsToEncrypt:           adv.colsToEncrypt,
			SkipDelete:                 t.SkipDeletes.ValueBoolPointer(),
			UnifyAcrossSchemas:         t.UnifyAcrossSchemas.ValueBoolPointer(),
			UnifyAcrossDatabases:       t.UnifyAcrossDatabases.ValueBoolPointer(),
			MergePredicates:            adv.mergePredicates,
			SoftPartitioning:           adv.softPartitioning,
			ShouldBackfillHistoryTable: t.BackfillHistoryTable.ValueBoolPointer(),
			CtidSettings:               adv.ctidSettings,
		},
	}, diags
}

// ToAPITable builds a PayloadsTable for validation endpoints.
func (t Table) ToAPITable(ctx context.Context) (openapi.PayloadsTable, diag.Diagnostics) {
	var tableUUID *uuid.UUID
	var diags diag.Diagnostics
	if t.UUID.ValueString() != "" {
		u, uuidDiags := parseUUID(t.UUID)
		diags.Append(uuidDiags...)
		tableUUID = &u
	}

	adv, advDiags := t.parseAdvancedFields(ctx)
	diags.Append(advDiags...)
	if diags.HasError() {
		return openapi.PayloadsTable{}, diags
	}

	return openapi.PayloadsTable{
		Uuid:               tableUUID,
		Name:               lib.ToPtr(t.Name.ValueString()),
		Schema:             t.Schema.ValueStringPointer(),
		EnableHistoryMode:  t.EnableHistoryMode.ValueBoolPointer(),
		DisableReplication: t.DisableReplication.ValueBoolPointer(),
		IsPartitioned:      t.IsPartitioned.ValueBoolPointer(),
		AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
			Alias:                      t.Alias.ValueStringPointer(),
			ExcludeColumns:             adv.colsToExclude,
			IncludeColumns:             adv.colsToInclude,
			ColumnsToHash:              adv.colsToHash,
			ColumnsToCompress:          adv.colsToCompress,
			ColumnsToEncrypt:           adv.colsToEncrypt,
			SkipDelete:                 t.SkipDeletes.ValueBoolPointer(),
			UnifyAcrossSchemas:         t.UnifyAcrossSchemas.ValueBoolPointer(),
			UnifyAcrossDatabases:       t.UnifyAcrossDatabases.ValueBoolPointer(),
			MergePredicates:            adv.mergePredicates,
			SoftPartitioning:           adv.softPartitioning,
			ShouldBackfillHistoryTable: t.BackfillHistoryTable.ValueBoolPointer(),
			CtidSettings:               adv.ctidSettings,
		},
	}, diags
}

func TablesFromAPIModel(ctx context.Context, apiModelTables []openapi.PayloadsTable) (map[string]Table, diag.Diagnostics) {
	tables := map[string]Table{}
	var diags diag.Diagnostics
	for _, apiTable := range apiModelTables {
		name := lib.RemovePtr(apiTable.Name)
		schema := lib.RemovePtr(apiTable.Schema)
		tableKey := name
		if schema != "" {
			tableKey = fmt.Sprintf("%s.%s", schema, name)
		}

		var advSettings openapi.PayloadsTableAdvancedSettings
		if apiTable.AdvancedSettings != nil {
			advSettings = *apiTable.AdvancedSettings
		}

		colsToExclude, excludeDiags := optionalStringListToListValue(ctx, advSettings.ExcludeColumns)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := optionalStringListToListValue(ctx, advSettings.IncludeColumns)
		diags.Append(includeDiags...)

		colsToHash, hashDiags := optionalStringListToListValue(ctx, advSettings.ColumnsToHash)
		diags.Append(hashDiags...)

		colsToCompress, compressDiags := optionalStringListToListValue(ctx, advSettings.ColumnsToCompress)
		diags.Append(compressDiags...)

		colsToEncrypt, encryptDiags := optionalStringListToListValue(ctx, advSettings.ColumnsToEncrypt)
		diags.Append(encryptDiags...)

		mergePredicates, mergePredDiags := MergePredicatesFromAPIModel(ctx, advSettings.MergePredicates)
		diags.Append(mergePredDiags...)

		softPartitioning, softPartitioningDiags := SoftPartitioningFromAPIModel(ctx, advSettings.SoftPartitioning)
		diags.Append(softPartitioningDiags...)

		ctidBackfill := types.BoolValue(false)
		ctidChunkSize := types.Int64Value(0)
		ctidMaxParallelism := types.Int64Value(0)
		if advSettings.CtidSettings != nil {
			ctidBackfill = types.BoolValue(lib.RemovePtr(advSettings.CtidSettings.Enabled))
			ctidChunkSize = types.Int64Value(int64(lib.RemovePtr(advSettings.CtidSettings.ChunkSize)))
			ctidMaxParallelism = types.Int64Value(int64(lib.RemovePtr(advSettings.CtidSettings.MaxParallelism)))
		}

		tableUUID := ""
		if apiTable.Uuid != nil {
			tableUUID = apiTable.Uuid.String()
		}

		tables[tableKey] = Table{
			UUID:                 types.StringValue(tableUUID),
			Name:                 types.StringValue(name),
			Schema:               types.StringValue(schema),
			EnableHistoryMode:    types.BoolValue(lib.RemovePtr(apiTable.EnableHistoryMode)),
			DisableReplication:   types.BoolValue(lib.RemovePtr(apiTable.DisableReplication)),
			IsPartitioned:        types.BoolValue(lib.RemovePtr(apiTable.IsPartitioned)),
			Alias:                types.StringPointerValue(advSettings.Alias),
			ExcludeColumns:       colsToExclude,
			IncludeColumns:       colsToInclude,
			ColumnsToHash:        colsToHash,
			ColumnsToCompress:    colsToCompress,
			ColumnsToEncrypt:     colsToEncrypt,
			SkipDeletes:          types.BoolPointerValue(advSettings.SkipDelete),
			UnifyAcrossSchemas:   types.BoolPointerValue(advSettings.UnifyAcrossSchemas),
			UnifyAcrossDatabases: types.BoolPointerValue(advSettings.UnifyAcrossDatabases),
			MergePredicates:      mergePredicates,
			SoftPartitioning:     softPartitioning,
			BackfillHistoryTable: types.BoolPointerValue(advSettings.ShouldBackfillHistoryTable),
			CTIDBackfill:         ctidBackfill,
			CTIDChunkSize:        ctidChunkSize,
			CTIDMaxParallelism:   ctidMaxParallelism,
			SkipBackfill:         types.BoolValue(false),
		}
	}

	if diags.HasError() {
		return map[string]Table{}, diags
	}

	return tables, diags
}
