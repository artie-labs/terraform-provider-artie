package tfmodels

import (
	"context"
	"encoding/json"
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
	PartitionBy    types.String `tfsdk:"partition_by"`
}

var MergePredicateAttrTypes = map[string]attr.Type{
	"partition_field": types.StringType,
	"partition_type":  types.StringType,
	"partition_by":    types.StringType,
}

func (m MergePredicate) ToAPIModel() openapi.PayloadsMergePredicates {
	return openapi.PayloadsMergePredicates{PartitionField: lib.ToPtr(m.PartitionField.ValueString()), PartitionType: nonEmptyStringPointer(m.PartitionType), PartitionBy: nonEmptyStringPointer(m.PartitionBy)}
}

// Validate mirrors the API's check that only time predicates can have a partition_by.
func (m MergePredicate) Validate() error {
	if m.PartitionType.ValueString() == "integer" && m.PartitionBy.ValueString() != "" {
		return fmt.Errorf("merge predicate %q is an integer column, so it cannot have a `partition_by`", m.PartitionField.ValueString())
	}
	return nil
}

func MergePredicatesFromAPIModel(ctx context.Context, apiMergePredicates *[]openapi.PayloadsMergePredicates) (types.List, diag.Diagnostics) {
	attrTypes := MergePredicateAttrTypes
	if apiMergePredicates == nil {
		return types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, []attr.Value{})
	}

	var diags diag.Diagnostics
	preds := []attr.Value{}
	for _, mp := range *apiMergePredicates {
		pred, predDiags := types.ObjectValueFrom(ctx, attrTypes, MergePredicate{
			PartitionField: types.StringValue(lib.RemovePtr(mp.PartitionField)),
			PartitionType:  nonEmptyStringValue(mp.PartitionType),
			PartitionBy:    nonEmptyStringValue(mp.PartitionBy),
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
		Enabled:            lib.ToPtr(s.Enabled.ValueBool()),
		PartitionFrequency: lib.ToPtr(s.PartitionFrequency.ValueString()),
		PartitionColumn:    lib.ToPtr(s.PartitionColumn.ValueString()),
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

	// Advanced table settings
	Alias                types.String `tfsdk:"alias"`
	ExcludeColumns       types.List   `tfsdk:"columns_to_exclude"`
	IncludeColumns       types.List   `tfsdk:"columns_to_include"`
	PrimaryKeysOverride  types.List   `tfsdk:"primary_keys_override"`
	ColumnsToHash        types.List   `tfsdk:"columns_to_hash"`
	ColumnsToCompress    types.List   `tfsdk:"columns_to_compress"`
	ColumnsToEncrypt     types.List   `tfsdk:"columns_to_encrypt"`
	EncryptJSONBColumns  types.Bool   `tfsdk:"encrypt_jsonb_columns"`
	SkipDeletes          types.Bool   `tfsdk:"skip_deletes"`
	UnifyAcrossSchemas   types.Bool   `tfsdk:"unify_across_schemas"`
	UnifyAcrossDatabases types.Bool   `tfsdk:"unify_across_databases"`
	MergePredicates      types.List   `tfsdk:"merge_predicates"`
	SoftPartitioning     types.Object `tfsdk:"soft_partitioning"`
	BackfillHistoryTable types.Bool   `tfsdk:"backfill_history_table"`
	CTIDBackfill         types.Bool   `tfsdk:"ctid_backfill"`
	CTIDChunkSize        types.Int64  `tfsdk:"ctid_chunk_size"`
	CTIDMaxParallelism   types.Int64  `tfsdk:"ctid_max_parallelism"`
	RangeBackfill        types.Bool   `tfsdk:"range_backfill"`
	RangeChunkSize       types.Int64  `tfsdk:"range_chunk_size"`
	RangeMaxParallelism  types.Int64  `tfsdk:"range_max_parallelism"`
	RangeBatchSize       types.Int64  `tfsdk:"range_batch_size"`
	SkipBackfill         types.Bool   `tfsdk:"skip_backfill"`
	SkipNoOpUpdates      types.Bool   `tfsdk:"skip_no_op_updates"`

	BufferRows                   types.Int64  `tfsdk:"buffer_rows"`
	FlushIntervalSeconds         types.Int64  `tfsdk:"flush_interval_seconds"`
	FlushSizeKB                  types.Int64  `tfsdk:"flush_size_kb"`
	ColumnsToTrim                types.List   `tfsdk:"columns_to_trim"`
	FlattenDocuments             types.Bool   `tfsdk:"flatten_documents"`
	FlattenDocumentsMaxDepth     types.Int64  `tfsdk:"flatten_documents_max_depth"`
	K8sRequestCPU                types.Int64  `tfsdk:"k8s_request_cpu"`
	K8sRequestMemoryMB           types.Int64  `tfsdk:"k8s_request_memory_mb"`
	PartitionBackfill            types.Bool   `tfsdk:"partition_backfill"`
	PartitionRangeBackfill       types.Bool   `tfsdk:"partition_range_backfill"`
	PartitionRangeChunkSize      types.Int64  `tfsdk:"partition_range_chunk_size"`
	PartitionRangeMaxParallelism types.Int64  `tfsdk:"partition_range_max_parallelism"`
	StreamARN                    types.String `tfsdk:"stream_arn"`
}

var TableAttrTypes = map[string]attr.Type{
	"uuid":                   types.StringType,
	"name":                   types.StringType,
	"schema":                 types.StringType,
	"enable_history_mode":    types.BoolType,
	"disable_replication":    types.BoolType,
	"alias":                  types.StringType,
	"columns_to_exclude":     types.ListType{ElemType: types.StringType},
	"columns_to_include":     types.ListType{ElemType: types.StringType},
	"primary_keys_override":  types.ListType{ElemType: types.StringType},
	"columns_to_hash":        types.ListType{ElemType: types.StringType},
	"columns_to_compress":    types.ListType{ElemType: types.StringType},
	"columns_to_encrypt":     types.ListType{ElemType: types.StringType},
	"encrypt_jsonb_columns":  types.BoolType,
	"skip_deletes":           types.BoolType,
	"unify_across_schemas":   types.BoolType,
	"unify_across_databases": types.BoolType,
	"merge_predicates":       types.ListType{ElemType: types.ObjectType{AttrTypes: MergePredicateAttrTypes}},
	"soft_partitioning":      types.ObjectType{AttrTypes: SoftPartitioningAttrTypes},
	"backfill_history_table": types.BoolType,
	"ctid_backfill":          types.BoolType,
	"ctid_chunk_size":        types.Int64Type,
	"ctid_max_parallelism":   types.Int64Type,
	"range_backfill":         types.BoolType,
	"range_chunk_size":       types.Int64Type,
	"range_max_parallelism":  types.Int64Type,
	"range_batch_size":       types.Int64Type,
	"skip_backfill":          types.BoolType,
	"skip_no_op_updates":     types.BoolType,

	"buffer_rows":                     types.Int64Type,
	"flush_interval_seconds":          types.Int64Type,
	"flush_size_kb":                   types.Int64Type,
	"columns_to_trim":                 types.ListType{ElemType: types.StringType},
	"flatten_documents":               types.BoolType,
	"flatten_documents_max_depth":     types.Int64Type,
	"k8s_request_cpu":                 types.Int64Type,
	"k8s_request_memory_mb":           types.Int64Type,
	"partition_backfill":              types.BoolType,
	"partition_range_backfill":        types.BoolType,
	"partition_range_chunk_size":      types.Int64Type,
	"partition_range_max_parallelism": types.Int64Type,
	"stream_arn":                      types.StringType,
}

func (t Table) ToAPIModel(ctx context.Context) (openapi.PayloadsTablePayload, diag.Diagnostics) {
	tableUUID := uuid.Nil
	var diags diag.Diagnostics
	if t.UUID.ValueString() != "" {
		tableUUID, diags = parseUUID(t.UUID)
	}

	colsToExclude, excludeDiags := parseOptionalList[string](ctx, t.ExcludeColumns)
	diags.Append(excludeDiags...)

	colsToInclude, includeDiags := parseOptionalList[string](ctx, t.IncludeColumns)
	diags.Append(includeDiags...)

	primaryKeysOverride, primaryKeysOverrideDiags := parseOptionalList[string](ctx, t.PrimaryKeysOverride)
	diags.Append(primaryKeysOverrideDiags...)

	colsToHash, hashDiags := parseOptionalList[string](ctx, t.ColumnsToHash)
	diags.Append(hashDiags...)

	colsToCompress, compressDiags := parseOptionalList[string](ctx, t.ColumnsToCompress)
	diags.Append(compressDiags...)

	colsToEncrypt, encryptDiags := parseOptionalList[string](ctx, t.ColumnsToEncrypt)
	diags.Append(encryptDiags...)

	colsToTrim, trimDiags := parseOptionalList[string](ctx, t.ColumnsToTrim)
	diags.Append(trimDiags...)

	mergePredicates, mergePredDiags := parseOptionalList[MergePredicate](ctx, t.MergePredicates)
	diags.Append(mergePredDiags...)
	var clientMergePreds *[]openapi.PayloadsMergePredicates
	if mergePredicates != nil && len(*mergePredicates) > 0 {
		var clientMPs []openapi.PayloadsMergePredicates
		for _, mp := range *mergePredicates {
			clientMPs = append(clientMPs, mp.ToAPIModel())
		}

		clientMergePreds = &clientMPs
	}

	softPartitioning, softPartitioningDiags := parseOptionalObject[SoftPartitioning](ctx, &t.SoftPartitioning)
	var clientSoftPartitioning *openapi.PayloadsSoftPartitioning
	if softPartitioning != nil {
		clientSoftPartitioning = softPartitioning.ToAPIModel()
	}
	diags.Append(softPartitioningDiags...)

	var clientCTIDSettings *openapi.PayloadsCTIDSettings
	if IsKnown(t.CTIDBackfill) {
		clientCTIDSettings = &openapi.PayloadsCTIDSettings{
			Enabled:        lib.ToPtr(t.CTIDBackfill.ValueBool()),
			ChunkSize:      lib.ToPtr(int(t.CTIDChunkSize.ValueInt64())),
			MaxParallelism: lib.ToPtr(int(t.CTIDMaxParallelism.ValueInt64())),
		}
	}

	var clientRangeSettings *openapi.PayloadsRangeSettings
	if IsKnown(t.RangeBackfill) {
		clientRangeSettings = &openapi.PayloadsRangeSettings{
			Enabled:        lib.ToPtr(t.RangeBackfill.ValueBool()),
			ChunksSize:     lib.ToPtr(int(t.RangeChunkSize.ValueInt64())),
			MaxParallelism: lib.ToPtr(int(t.RangeMaxParallelism.ValueInt64())),
			BatchSize:      lib.ToPtr(int(t.RangeBatchSize.ValueInt64())),
		}
	}

	// The API replaces the whole partition range object when it's sent, so it's only sent when the switch is known.
	var clientPartitionRangeSettings *openapi.PayloadsPartitionRangeSettings
	if IsKnown(t.PartitionBackfill) {
		clientPartitionRangeSettings = &openapi.PayloadsPartitionRangeSettings{
			Enabled: lib.ToPtr(t.PartitionBackfill.ValueBool()),
			Range: &openapi.PayloadsRangeSettings{
				Enabled:        lib.ToPtr(t.PartitionRangeBackfill.ValueBool()),
				ChunksSize:     lib.ToPtr(int(t.PartitionRangeChunkSize.ValueInt64())),
				MaxParallelism: lib.ToPtr(int(t.PartitionRangeMaxParallelism.ValueInt64())),
			},
		}
	}

	if diags.HasError() {
		return openapi.PayloadsTablePayload{}, diags
	}

	return openapi.PayloadsTablePayload{
		// New tables send the nil UUID, as the hand-written client did.
		Uuid:               &tableUUID,
		Name:               lib.ToPtr(t.Name.ValueString()),
		Schema:             lib.ToPtr(t.Schema.ValueString()),
		EnableHistoryMode:  lib.ToPtr(t.EnableHistoryMode.ValueBool()),
		DisableReplication: lib.ToPtr(t.DisableReplication.ValueBool()),
		AdvancedSettings: &openapi.PayloadsAdvancedTableSettingsPayload{
			Alias:                      stringPointer(t.Alias),
			ExcludeColumns:             colsToExclude,
			IncludeColumns:             colsToInclude,
			PrimaryKeysOverride:        primaryKeysOverride,
			ColumnsToHash:              colsToHash,
			ColumnsToCompress:          colsToCompress,
			ColumnsToEncrypt:           colsToEncrypt,
			EncryptJSONBColumns:        boolPointer(t.EncryptJSONBColumns),
			SkipDelete:                 boolPointer(t.SkipDeletes),
			UnifyAcrossSchemas:         boolPointer(t.UnifyAcrossSchemas),
			UnifyAcrossDatabases:       boolPointer(t.UnifyAcrossDatabases),
			MergePredicates:            clientMergePreds,
			SoftPartitioning:           clientSoftPartitioning,
			ShouldBackfillHistoryTable: boolPointer(t.BackfillHistoryTable),
			CtidSettings:               clientCTIDSettings,
			RangeSettings:              clientRangeSettings,
			SkipBackfill:               boolPointer(t.SkipBackfill),
			SkipNoOpUpdates:            boolPointer(t.SkipNoOpUpdates),
			BufferRows:                 int64ToIntPointer(t.BufferRows),
			FlushIntervalSeconds:       int64ToIntPointer(t.FlushIntervalSeconds),
			FlushSizeKb:                int64ToIntPointer(t.FlushSizeKB),
			ColumnsToTrim:              colsToTrim,
			FlattenDocuments:           boolPointer(t.FlattenDocuments),
			FlattenDocumentsMaxDepth:   int64ToIntPointer(t.FlattenDocumentsMaxDepth),
			K8sRequestCPU:              int64ToIntPointer(t.K8sRequestCPU),
			K8sRequestMemoryMB:         int64ToIntPointer(t.K8sRequestMemoryMB),
			PartitionRangeSettings:     clientPartitionRangeSettings,
			StreamARN:                  stringPointer(t.StreamARN),
		},
	}, diags
}

func TablesFromAPIModel(ctx context.Context, apiModelTables []openapi.PayloadsTable) (map[string]Table, diag.Diagnostics) {
	tables := map[string]Table{}
	var diags diag.Diagnostics
	for _, apiTable := range apiModelTables {
		name := lib.RemovePtr(apiTable.Name)
		schema := lib.RemovePtr(apiTable.Schema)
		settings := lib.RemovePtr(apiTable.AdvancedSettings)
		tableKey := name
		if schema != "" {
			tableKey = fmt.Sprintf("%s.%s", schema, name)
		}

		colsToExclude, excludeDiags := optionalStringListToListValue(ctx, settings.ExcludeColumns)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := optionalStringListToListValue(ctx, settings.IncludeColumns)
		diags.Append(includeDiags...)

		primaryKeysOverride, primaryKeysOverrideDiags := optionalStringListToListValue(ctx, settings.PrimaryKeysOverride)
		diags.Append(primaryKeysOverrideDiags...)

		colsToHash, hashDiags := optionalStringListToListValue(ctx, settings.ColumnsToHash)
		diags.Append(hashDiags...)

		colsToCompress, compressDiags := optionalStringListToListValue(ctx, settings.ColumnsToCompress)
		diags.Append(compressDiags...)

		colsToEncrypt, encryptDiags := optionalStringListToListValue(ctx, settings.ColumnsToEncrypt)
		diags.Append(encryptDiags...)

		colsToTrim, trimDiags := optionalStringListToListValue(ctx, settings.ColumnsToTrim)
		diags.Append(trimDiags...)

		mergePredicates, mergePredDiags := MergePredicatesFromAPIModel(ctx, settings.MergePredicates)
		diags.Append(mergePredDiags...)

		softPartitioning, softPartitioningDiags := SoftPartitioningFromAPIModel(ctx, settings.SoftPartitioning)
		diags.Append(softPartitioningDiags...)

		// Extract CTID settings - initialize them to the zero-values (instead of null/unknown) because if
		// they're not in the api response, that means they're zero. This avoids extra noise in the plan output.
		ctidBackfill := types.BoolValue(false)
		ctidChunkSize := types.Int64Value(0)
		ctidMaxParallelism := types.Int64Value(0)
		if settings.CtidSettings != nil {
			ctidBackfill = types.BoolValue(lib.RemovePtr(settings.CtidSettings.Enabled))
			ctidChunkSize = types.Int64Value(int64(lib.RemovePtr(settings.CtidSettings.ChunkSize)))
			ctidMaxParallelism = types.Int64Value(int64(lib.RemovePtr(settings.CtidSettings.MaxParallelism)))
		}

		rangeBackfill := types.BoolValue(false)
		rangeChunkSize := types.Int64Value(0)
		rangeMaxParallelism := types.Int64Value(0)
		rangeBatchSize := types.Int64Value(0)
		if settings.RangeSettings != nil {
			rangeBackfill = types.BoolValue(lib.RemovePtr(settings.RangeSettings.Enabled))
			rangeChunkSize = types.Int64Value(int64(lib.RemovePtr(settings.RangeSettings.ChunksSize)))
			rangeMaxParallelism = types.Int64Value(int64(lib.RemovePtr(settings.RangeSettings.MaxParallelism)))
			rangeBatchSize = types.Int64Value(int64(lib.RemovePtr(settings.RangeSettings.BatchSize)))
		}

		// Null while the table has no partition settings, so ToAPIModel doesn't send (and save) a disabled object for
		// every table that never set one.
		partitionBackfill := types.BoolNull()
		partitionRangeBackfill := types.BoolValue(false)
		partitionRangeChunkSize := types.Int64Value(0)
		partitionRangeMaxParallelism := types.Int64Value(0)
		if settings.PartitionRangeSettings != nil {
			partitionRange := lib.RemovePtr(settings.PartitionRangeSettings.Range)
			partitionBackfill = types.BoolValue(lib.RemovePtr(settings.PartitionRangeSettings.Enabled))
			partitionRangeBackfill = types.BoolValue(lib.RemovePtr(partitionRange.Enabled))
			partitionRangeChunkSize = types.Int64Value(int64(lib.RemovePtr(partitionRange.ChunksSize)))
			partitionRangeMaxParallelism = types.Int64Value(int64(lib.RemovePtr(partitionRange.MaxParallelism)))
		}

		tables[tableKey] = Table{
			UUID:                types.StringValue(lib.RemovePtr(apiTable.Uuid).String()),
			Name:                types.StringValue(name),
			Schema:              types.StringValue(schema),
			EnableHistoryMode:   types.BoolValue(lib.RemovePtr(apiTable.EnableHistoryMode)),
			DisableReplication:  types.BoolValue(lib.RemovePtr(apiTable.DisableReplication)),
			Alias:               types.StringPointerValue(settings.Alias),
			ExcludeColumns:      colsToExclude,
			IncludeColumns:      colsToInclude,
			PrimaryKeysOverride: primaryKeysOverride,
			ColumnsToHash:       colsToHash,
			ColumnsToCompress:   colsToCompress,
			ColumnsToEncrypt:    colsToEncrypt,
			// The API stores these "absent means off" toggles as nil when false; coalesce nil to
			// false so an explicit `false` round-trips without a post-apply consistency error.
			EncryptJSONBColumns:  boolPointerValueOrFalse(settings.EncryptJSONBColumns),
			SkipDeletes:          boolPointerValueOrFalse(settings.SkipDelete),
			UnifyAcrossSchemas:   boolPointerValueOrFalse(settings.UnifyAcrossSchemas),
			UnifyAcrossDatabases: boolPointerValueOrFalse(settings.UnifyAcrossDatabases),
			MergePredicates:      mergePredicates,
			SoftPartitioning:     softPartitioning,
			BackfillHistoryTable: boolPointerValueOrFalse(settings.ShouldBackfillHistoryTable),
			CTIDBackfill:         ctidBackfill,
			CTIDChunkSize:        ctidChunkSize,
			CTIDMaxParallelism:   ctidMaxParallelism,
			RangeBackfill:        rangeBackfill,
			RangeChunkSize:       rangeChunkSize,
			RangeMaxParallelism:  rangeMaxParallelism,
			RangeBatchSize:       rangeBatchSize,
			SkipBackfill:         boolPointerValueOrFalse(settings.SkipBackfill),
			SkipNoOpUpdates:      boolPointerValueOrFalse(settings.SkipNoOpUpdates),
			// The API omits these when zero, false or empty, so they read back as 0/false/"" (not null).
			BufferRows:                   types.Int64Value(int64(lib.RemovePtr(settings.BufferRows))),
			FlushIntervalSeconds:         types.Int64Value(int64(lib.RemovePtr(settings.FlushIntervalSeconds))),
			FlushSizeKB:                  types.Int64Value(int64(lib.RemovePtr(settings.FlushSizeKb))),
			ColumnsToTrim:                colsToTrim,
			FlattenDocuments:             boolPointerValueOrFalse(settings.FlattenDocuments),
			FlattenDocumentsMaxDepth:     types.Int64Value(int64(lib.RemovePtr(settings.FlattenDocumentsMaxDepth))),
			K8sRequestCPU:                types.Int64Value(int64(lib.RemovePtr(settings.K8sRequestCPU))),
			K8sRequestMemoryMB:           types.Int64Value(int64(lib.RemovePtr(settings.K8sRequestMemoryMB))),
			PartitionBackfill:            partitionBackfill,
			PartitionRangeBackfill:       partitionRangeBackfill,
			PartitionRangeChunkSize:      partitionRangeChunkSize,
			PartitionRangeMaxParallelism: partitionRangeMaxParallelism,
			StreamARN:                    types.StringValue(lib.RemovePtr(settings.StreamARN)),
		}
	}

	if diags.HasError() {
		return map[string]Table{}, diags
	}

	return tables, diags
}

// ValidationTables converts table payloads into the table type the validate-unsaved endpoints accept.
// The two types share the same JSON, so the conversion is lossless.
func ValidationTables(tables *[]openapi.PayloadsTablePayload) ([]openapi.PayloadsTable, error) {
	out := []openapi.PayloadsTable{}
	if tables == nil {
		return out, nil
	}

	body, err := json.Marshal(*tables)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tables for validation: %w", err)
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to encode tables for validation: %w", err)
	}
	return out, nil
}
