package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type PipelineDestinationConfig struct {
	Dataset                 types.String `tfsdk:"dataset"`
	Database                types.String `tfsdk:"database"`
	Schema                  types.String `tfsdk:"schema"`
	UseSameSchemaAsSource   types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix        types.String `tfsdk:"schema_name_prefix"`
	Bucket                  types.String `tfsdk:"bucket"`
	TableNameSeparator      types.String `tfsdk:"table_name_separator"`
	Folder                  types.String `tfsdk:"folder"`
	CreateIcebergNamespaces types.Bool   `tfsdk:"create_iceberg_namespaces"`
}

func (d PipelineDestinationConfig) ToAPIModel() openapi.PayloadsSpecificConfig {
	return openapi.PayloadsSpecificConfig{
		Database:                    lib.ToPtr(d.Database.ValueString()),
		Schema:                      lib.ToPtr(d.Schema.ValueString()),
		UseSameSchemaAsSource:       lib.ToPtr(d.UseSameSchemaAsSource.ValueBool()),
		SchemaNamePrefix:            lib.ToPtr(d.SchemaNamePrefix.ValueString()),
		BucketName:                  lib.ToPtr(d.Bucket.ValueString()),
		TableNameSeparator:          lib.ToPtr(d.TableNameSeparator.ValueString()),
		FolderName:                  lib.ToPtr(d.Folder.ValueString()),
		DynamicallyCreateNamespaces: lib.ToPtr(d.CreateIcebergNamespaces.ValueBool()),
	}
}

func PipelineDestinationConfigFromAPIModel(apiModel openapi.PayloadsSpecificConfig) PipelineDestinationConfig {
	return PipelineDestinationConfig{
		// The API has no dataset field (BigQuery uses database), so this always reads back empty.
		Dataset:                 types.StringValue(""),
		Database:                types.StringValue(lib.RemovePtr(apiModel.Database)),
		Schema:                  types.StringValue(lib.RemovePtr(apiModel.Schema)),
		UseSameSchemaAsSource:   types.BoolValue(lib.RemovePtr(apiModel.UseSameSchemaAsSource)),
		SchemaNamePrefix:        types.StringValue(lib.RemovePtr(apiModel.SchemaNamePrefix)),
		Bucket:                  types.StringValue(lib.RemovePtr(apiModel.BucketName)),
		TableNameSeparator:      types.StringValue(lib.RemovePtr(apiModel.TableNameSeparator)),
		Folder:                  types.StringValue(lib.RemovePtr(apiModel.FolderName)),
		CreateIcebergNamespaces: types.BoolValue(lib.RemovePtr(apiModel.DynamicallyCreateNamespaces)),
	}
}

type FlushConfig struct {
	FlushIntervalSeconds types.Int64 `tfsdk:"flush_interval_seconds"`
	BufferRows           types.Int64 `tfsdk:"buffer_rows"`
	FlushSizeKB          types.Int64 `tfsdk:"flush_size_kb"`
}

var flushAttrTypes = map[string]attr.Type{
	"flush_interval_seconds": types.Int64Type,
	"buffer_rows":            types.Int64Type,
	"flush_size_kb":          types.Int64Type,
}

func buildFlushConfig(ctx context.Context, d types.Object) (*FlushConfig, diag.Diagnostics) {
	var flushConfig *FlushConfig
	flushConfigDiags := d.As(ctx, &flushConfig, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})

	if flushConfigDiags.HasError() {
		return nil, flushConfigDiags
	}

	return flushConfig, nil
}

type StaticColumn struct {
	Column types.String `tfsdk:"column"`
	Value  types.String `tfsdk:"value"`
}

var StaticColumnAttrTypes = map[string]attr.Type{
	"column": types.StringType,
	"value":  types.StringType,
}

func staticColumnsToAPI(ctx context.Context, staticColumnsList types.List) (*[]openapi.PayloadsStaticColumn, diag.Diagnostics) {
	staticColumns, diags := parseOptionalList[StaticColumn](ctx, staticColumnsList)
	if staticColumns == nil {
		return nil, diags
	}

	var apiStaticColumns []openapi.PayloadsStaticColumn
	for _, sc := range *staticColumns {
		apiStaticColumns = append(apiStaticColumns, openapi.PayloadsStaticColumn{
			Column: lib.ToPtr(sc.Column.ValueString()),
			Value:  lib.ToPtr(sc.Value.ValueString()),
		})
	}

	return &apiStaticColumns, diags
}

func staticColumnsFromAPI(ctx context.Context, apiStaticColumns *[]openapi.PayloadsStaticColumn) (types.List, diag.Diagnostics) {
	if apiStaticColumns == nil || len(*apiStaticColumns) == 0 {
		// Return an empty list instead of null to avoid perpetual diffs when
		// the user explicitly specifies `static_columns = []`
		return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: StaticColumnAttrTypes}, []StaticColumn{})
	}

	var staticColumns []StaticColumn
	for _, sc := range *apiStaticColumns {
		staticColumns = append(staticColumns, StaticColumn{
			Column: types.StringValue(lib.RemovePtr(sc.Column)),
			Value:  types.StringValue(lib.RemovePtr(sc.Value)),
		})
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: StaticColumnAttrTypes}, staticColumns)
}

type Pipeline struct {
	UUID                     types.String               `tfsdk:"uuid"`
	Name                     types.String               `tfsdk:"name"`
	SourceReaderUUID         types.String               `tfsdk:"source_reader_uuid"`
	DestinationUUID          types.String               `tfsdk:"destination_connector_uuid"`
	DestinationConfig        *PipelineDestinationConfig `tfsdk:"destination_config"`
	SnowflakeEcoScheduleUUID types.String               `tfsdk:"snowflake_eco_schedule_uuid"`
	EncryptionKeyUUID        types.String               `tfsdk:"encryption_key_uuid"`
	ColumnHashingSaltUUID    types.String               `tfsdk:"column_hashing_salt_uuid"`
	DataPlaneName            types.String               `tfsdk:"data_plane_name"`
	Tables                   types.Map                  `tfsdk:"tables"`
	StatusOverride           types.String               `tfsdk:"status_override"`

	// Advanced settings
	FlushConfig                                  types.Object `tfsdk:"flush_rules"`
	DropDeletedColumns                           types.Bool   `tfsdk:"drop_deleted_columns"`
	SoftDeleteRows                               types.Bool   `tfsdk:"soft_delete_rows"`
	IncludeArtieUpdatedAtColumn                  types.Bool   `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn               types.Bool   `tfsdk:"include_database_updated_at_column"`
	IncludeArtieOperationColumn                  types.Bool   `tfsdk:"include_artie_operation_column"`
	IncludeFullSourceTableNameColumn             types.Bool   `tfsdk:"include_full_source_table_name_column"`
	IncludeFullSourceTableNameColumnAsPrimaryKey types.Bool   `tfsdk:"include_full_source_table_name_column_as_primary_key"`
	DefaultSourceSchema                          types.String `tfsdk:"default_source_schema"`
	SplitEventsByType                            types.Bool   `tfsdk:"split_events_by_type"`
	IncludeSourceMetadataColumn                  types.Bool   `tfsdk:"include_source_metadata_column"`
	AutoEnableHistoryForNewTables                types.Bool   `tfsdk:"auto_enable_history_for_new_tables"`
	AutoEnableHistoryIgnoreRegex                 types.String `tfsdk:"auto_enable_history_ignore_regex"`
	AutoReplicateIgnoreRegex                     types.String `tfsdk:"auto_replicate_ignore_regex"`
	AutoReplicateNewTables                       types.Bool   `tfsdk:"auto_replicate_new_tables"`
	AppendOnly                                   types.Bool   `tfsdk:"append_only"`
	StaticColumns                                types.List   `tfsdk:"static_columns"`
	StagingSchema                                types.String `tfsdk:"staging_schema"`
	ForceUTCTimezone                             types.Bool   `tfsdk:"force_utc_timezone"`
	WriteRawBinaryValues                         types.Bool   `tfsdk:"write_raw_binary_values"`
	DisableAlerts                                types.Bool   `tfsdk:"disable_alerts"`
	DatabricksAutoLiquidClustering               types.Bool   `tfsdk:"databricks_auto_liquid_clustering"`
	MaxConcurrentSnapshots                       types.Int64  `tfsdk:"max_concurrent_snapshots"`
	TurboWarehouse                               types.String `tfsdk:"turbo_warehouse"`
	TurboRowThreshold                            types.Int64  `tfsdk:"turbo_row_threshold"`
	TurboLatencyThresholdMinutes                 types.Int64  `tfsdk:"turbo_latency_threshold_minutes"`
}

func (p Pipeline) ToAPIModel(ctx context.Context) (openapi.PayloadsPipelinePayload, diag.Diagnostics) {
	tables := map[string]Table{}
	diags := p.Tables.ElementsAs(ctx, &tables, false)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	apiTables := []openapi.PayloadsTablePayload{}
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return openapi.PayloadsPipelinePayload{}, diags
		}
		apiTables = append(apiTables, apiTable)
	}

	sourceReaderUUID, sourceReaderDiags := parseOptionalUUID(p.SourceReaderUUID)
	diags.Append(sourceReaderDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	destinationUUID, destDiags := parseOptionalUUID(p.DestinationUUID)
	diags.Append(destDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	snowflakeEcoScheduleUUID, snowflakeDiags := parseOptionalUUID(p.SnowflakeEcoScheduleUUID)
	diags.Append(snowflakeDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	encryptionKeyUUID, encryptionKeyDiags := parseOptionalUUID(p.EncryptionKeyUUID)
	diags.Append(encryptionKeyDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	columnHashingSaltUUID, columnHashingSaltDiags := parseOptionalUUID(p.ColumnHashingSaltUUID)
	diags.Append(columnHashingSaltDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	flushConfig, flushConfigDiags := buildFlushConfig(ctx, p.FlushConfig)
	diags.Append(flushConfigDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	staticColumns, staticColumnsDiags := staticColumnsToAPI(ctx, p.StaticColumns)
	diags.Append(staticColumnsDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	advancedSettings := openapi.PayloadsAdvancedPipelineSettingsPayload{
		DropDeletedColumns:                           p.DropDeletedColumns.ValueBoolPointer(),
		EnableSoftDelete:                             p.SoftDeleteRows.ValueBoolPointer(),
		IncludeArtieUpdatedAtColumn:                  p.IncludeArtieUpdatedAtColumn.ValueBoolPointer(),
		IncludeDatabaseUpdatedAtColumn:               p.IncludeDatabaseUpdatedAtColumn.ValueBoolPointer(),
		IncludeArtieOperationColumn:                  p.IncludeArtieOperationColumn.ValueBoolPointer(),
		IncludeFullSourceTableNameColumn:             p.IncludeFullSourceTableNameColumn.ValueBoolPointer(),
		IncludeFullSourceTableNameColumnAsPrimaryKey: p.IncludeFullSourceTableNameColumnAsPrimaryKey.ValueBoolPointer(),
		DefaultSourceSchema:                          p.DefaultSourceSchema.ValueStringPointer(),
		SplitEventsByType:                            p.SplitEventsByType.ValueBoolPointer(),
		IncludeSourceMetadataColumn:                  p.IncludeSourceMetadataColumn.ValueBoolPointer(),
		AutoEnableHistoryForNewTables:                p.AutoEnableHistoryForNewTables.ValueBoolPointer(),
		AutoEnableHistoryIgnoreRegex:                 p.AutoEnableHistoryIgnoreRegex.ValueStringPointer(),
		AutoReplicateIgnoreRegex:                     p.AutoReplicateIgnoreRegex.ValueStringPointer(),
		AutoReplicateNewTables:                       p.AutoReplicateNewTables.ValueBoolPointer(),
		AppendOnly:                                   p.AppendOnly.ValueBoolPointer(),
		StaticColumns:                                staticColumns,
		StagingSchema:                                p.StagingSchema.ValueStringPointer(),
		ForceUTCTimezone:                             p.ForceUTCTimezone.ValueBoolPointer(),
		WriteRawBinaryValues:                         p.WriteRawBinaryValues.ValueBoolPointer(),
		DisableAlerts:                                p.DisableAlerts.ValueBoolPointer(),
		DatabricksAutoLiquidClustering:               p.DatabricksAutoLiquidClustering.ValueBoolPointer(),
		MaxConcurrentSnapshots:                       int64ToIntPointer(p.MaxConcurrentSnapshots),
		TurboWarehouse:                               p.TurboWarehouse.ValueStringPointer(),
		TurboRowThreshold:                            int64ToIntPointer(p.TurboRowThreshold),
		TurboLatencyThresholdMinutes:                 int64ToIntPointer(p.TurboLatencyThresholdMinutes),
	}
	if flushConfig != nil {
		advancedSettings.FlushIntervalSeconds = int64ToIntPointer(flushConfig.FlushIntervalSeconds)
		advancedSettings.BufferRows = int64ToIntPointer(flushConfig.BufferRows)
		advancedSettings.FlushSizeKb = int64ToIntPointer(flushConfig.FlushSizeKB)
	}

	destinationConfig := p.DestinationConfig.ToAPIModel()
	return openapi.PayloadsPipelinePayload{
		Name:                     lib.ToPtr(p.Name.ValueString()),
		SourceReaderUUID:         sourceReaderUUID,
		Tables:                   &apiTables,
		DestinationUUID:          destinationUUID,
		SpecificDestCfg:          &destinationConfig,
		SnowflakeEcoScheduleUUID: snowflakeEcoScheduleUUID,
		EncryptionKeyUUID:        encryptionKeyUUID,
		ColumnHashingSaltUUID:    columnHashingSaltUUID,
		DataPlaneName:            lib.ToPtr(p.DataPlaneName.ValueString()),
		AdvancedSettings:         &advancedSettings,
	}, diags
}

func PipelineFromAPIModel(ctx context.Context, apiModel openapi.PayloadsFullPipeline) (Pipeline, diag.Diagnostics) {
	tables, diags := TablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TableAttrTypes}, tables)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	destinationConfig := PipelineDestinationConfigFromAPIModel(apiModel.SpecificDestCfg)

	var flushConfig types.Object
	var dropDeletedColumns types.Bool
	var softDeleteRows types.Bool
	var includeArtieUpdatedAtColumn types.Bool
	var includeDatabaseUpdatedAtColumn types.Bool
	var includeArtieOperationColumn types.Bool
	var includeFullSourceTableNameColumn types.Bool
	var includeFullSourceTableNameColumnAsPrimaryKey types.Bool
	var defaultSourceSchema types.String
	var splitEventsByType types.Bool
	var includeSourceMetadataColumn types.Bool
	var autoEnableHistoryIgnoreRegex types.String
	var autoReplicateIgnoreRegex types.String
	// These should default to false even if they're omitted from the API response.
	autoEnableHistoryForNewTables := types.BoolValue(false)
	var appendOnly types.Bool
	var stagingSchema types.String
	var forceUTCTimezone types.Bool
	var writeRawBinaryValues types.Bool
	var maxConcurrentSnapshots types.Int64
	var turboWarehouse types.String
	var turboRowThreshold types.Int64
	var turboLatencyThresholdMinutes types.Int64

	autoReplicateNewTables := types.BoolValue(false)

	settings := apiModel.AdvancedSettings
	if settings.DropDeletedColumns != nil {
		dropDeletedColumns = types.BoolValue(*settings.DropDeletedColumns)
	}
	if settings.EnableSoftDelete != nil {
		softDeleteRows = types.BoolValue(*settings.EnableSoftDelete)
	}
	if settings.IncludeArtieUpdatedAtColumn != nil {
		includeArtieUpdatedAtColumn = types.BoolValue(*settings.IncludeArtieUpdatedAtColumn)
	}
	if settings.IncludeDatabaseUpdatedAtColumn != nil {
		includeDatabaseUpdatedAtColumn = types.BoolValue(*settings.IncludeDatabaseUpdatedAtColumn)
	}
	if settings.IncludeArtieOperationColumn != nil {
		includeArtieOperationColumn = types.BoolValue(*settings.IncludeArtieOperationColumn)
	}
	if settings.IncludeFullSourceTableNameColumn != nil {
		includeFullSourceTableNameColumn = types.BoolValue(*settings.IncludeFullSourceTableNameColumn)
	}
	if settings.IncludeFullSourceTableNameColumnAsPrimaryKey != nil {
		includeFullSourceTableNameColumnAsPrimaryKey = types.BoolValue(*settings.IncludeFullSourceTableNameColumnAsPrimaryKey)
	}
	if settings.DefaultSourceSchema != nil {
		defaultSourceSchema = types.StringValue(*settings.DefaultSourceSchema)
	}
	if settings.SplitEventsByType != nil {
		splitEventsByType = types.BoolValue(*settings.SplitEventsByType)
	}
	if settings.IncludeSourceMetadataColumn != nil {
		includeSourceMetadataColumn = types.BoolValue(*settings.IncludeSourceMetadataColumn)
	}
	if settings.AutoEnableHistoryForNewTables != nil {
		autoEnableHistoryForNewTables = types.BoolValue(*settings.AutoEnableHistoryForNewTables)
	}
	if settings.AutoEnableHistoryIgnoreRegex != nil {
		autoEnableHistoryIgnoreRegex = types.StringValue(*settings.AutoEnableHistoryIgnoreRegex)
	}
	if settings.AutoReplicateIgnoreRegex != nil {
		autoReplicateIgnoreRegex = types.StringValue(*settings.AutoReplicateIgnoreRegex)
	}
	if settings.AutoReplicateNewTables != nil {
		autoReplicateNewTables = types.BoolValue(*settings.AutoReplicateNewTables)
	}
	if settings.AppendOnly != nil {
		appendOnly = types.BoolValue(*settings.AppendOnly)
	}
	if settings.StagingSchema != nil {
		stagingSchema = types.StringValue(*settings.StagingSchema)
	}
	if settings.ForceUTCTimezone != nil {
		forceUTCTimezone = types.BoolValue(*settings.ForceUTCTimezone)
	}
	if settings.WriteRawBinaryValues != nil {
		writeRawBinaryValues = types.BoolValue(*settings.WriteRawBinaryValues)
	}
	if settings.TurboWarehouse != nil {
		turboWarehouse = types.StringValue(*settings.TurboWarehouse)
	}
	if settings.TurboRowThreshold != nil {
		turboRowThreshold = types.Int64Value(int64(*settings.TurboRowThreshold))
	}
	if settings.TurboLatencyThresholdMinutes != nil {
		turboLatencyThresholdMinutes = types.Int64Value(int64(*settings.TurboLatencyThresholdMinutes))
	}
	if settings.MaxConcurrentSnapshots != nil {
		maxConcurrentSnapshots = types.Int64Value(int64(*settings.MaxConcurrentSnapshots))
	}
	disableAlerts := boolPointerValueOrFalse(settings.DisableAlerts)
	databricksAutoLiquidClustering := boolPointerValueOrFalse(settings.DatabricksAutoLiquidClustering)
	flushConfigMap := map[string]attr.Value{}
	if settings.FlushIntervalSeconds != nil {
		flushConfigMap["flush_interval_seconds"] = types.Int64Value(int64(*settings.FlushIntervalSeconds))
	}
	if settings.BufferRows != nil {
		flushConfigMap["buffer_rows"] = types.Int64Value(int64(*settings.BufferRows))
	}
	if settings.FlushSizeKb != nil {
		flushConfigMap["flush_size_kb"] = types.Int64Value(int64(*settings.FlushSizeKb))
	}
	if len(flushConfigMap) > 0 {
		var flushConfigDiags diag.Diagnostics
		flushConfig, flushConfigDiags = types.ObjectValue(flushAttrTypes, flushConfigMap)
		diags.Append(flushConfigDiags...)
		if diags.HasError() {
			return Pipeline{}, diags
		}
	}

	staticColumns, staticColumnsDiags := staticColumnsFromAPI(ctx, settings.StaticColumns)
	diags.Append(staticColumnsDiags...)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	return Pipeline{
		UUID:                     types.StringValue(apiModel.Uuid.String()),
		Name:                     types.StringValue(apiModel.Name),
		Tables:                   tablesMap,
		SourceReaderUUID:         optionalUUIDToStringValue(apiModel.SourceReaderUUID),
		DestinationUUID:          optionalUUIDToStringValue(apiModel.DestinationUUID),
		DestinationConfig:        &destinationConfig,
		SnowflakeEcoScheduleUUID: optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID),
		EncryptionKeyUUID:        optionalUUIDToStringValue(apiModel.EncryptionKeyUUID),
		ColumnHashingSaltUUID:    optionalUUIDToStringValue(apiModel.ColumnHashingSaltUUID),
		DataPlaneName:            types.StringValue(apiModel.DataPlaneName),

		// Advanced settings:
		DropDeletedColumns:                           dropDeletedColumns,
		SoftDeleteRows:                               softDeleteRows,
		IncludeArtieUpdatedAtColumn:                  includeArtieUpdatedAtColumn,
		IncludeDatabaseUpdatedAtColumn:               includeDatabaseUpdatedAtColumn,
		IncludeArtieOperationColumn:                  includeArtieOperationColumn,
		IncludeFullSourceTableNameColumn:             includeFullSourceTableNameColumn,
		IncludeFullSourceTableNameColumnAsPrimaryKey: includeFullSourceTableNameColumnAsPrimaryKey,
		FlushConfig:                                  flushConfig,
		DefaultSourceSchema:                          defaultSourceSchema,
		SplitEventsByType:                            splitEventsByType,
		IncludeSourceMetadataColumn:                  includeSourceMetadataColumn,
		AutoEnableHistoryForNewTables:                autoEnableHistoryForNewTables,
		AutoEnableHistoryIgnoreRegex:                 autoEnableHistoryIgnoreRegex,
		AutoReplicateIgnoreRegex:                     autoReplicateIgnoreRegex,
		AutoReplicateNewTables:                       autoReplicateNewTables,
		AppendOnly:                                   appendOnly,
		StaticColumns:                                staticColumns,
		StagingSchema:                                stagingSchema,
		ForceUTCTimezone:                             forceUTCTimezone,
		WriteRawBinaryValues:                         writeRawBinaryValues,
		DisableAlerts:                                disableAlerts,
		DatabricksAutoLiquidClustering:               databricksAutoLiquidClustering,
		MaxConcurrentSnapshots:                       maxConcurrentSnapshots,
		TurboWarehouse:                               turboWarehouse,
		TurboRowThreshold:                            turboRowThreshold,
		TurboLatencyThresholdMinutes:                 turboLatencyThresholdMinutes,
	}, diags
}
