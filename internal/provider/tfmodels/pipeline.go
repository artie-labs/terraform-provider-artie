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
	// The API uses a single "database" field for both Snowflake's database and BigQuery's dataset.
	database := d.Database.ValueString()
	if database == "" {
		database = d.Dataset.ValueString()
	}

	return openapi.PayloadsSpecificConfig{
		Database:                    lib.ToPtr(database),
		Schema:                      d.Schema.ValueStringPointer(),
		UseSameSchemaAsSource:       d.UseSameSchemaAsSource.ValueBoolPointer(),
		SchemaNamePrefix:            d.SchemaNamePrefix.ValueStringPointer(),
		BucketName:                  d.Bucket.ValueStringPointer(),
		TableNameSeparator:          d.TableNameSeparator.ValueStringPointer(),
		FolderName:                  d.Folder.ValueStringPointer(),
		DynamicallyCreateNamespaces: d.CreateIcebergNamespaces.ValueBoolPointer(),
	}
}

func PipelineDestinationConfigFromAPIModel(apiModel openapi.PayloadsSpecificConfig) PipelineDestinationConfig {
	database := lib.RemovePtr(apiModel.Database)
	return PipelineDestinationConfig{
		Dataset:                 types.StringValue(database),
		Database:                types.StringValue(database),
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
			Column: sc.Column.ValueStringPointer(),
			Value:  sc.Value.ValueStringPointer(),
		})
	}

	return &apiStaticColumns, diags
}

func staticColumnsFromAPI(ctx context.Context, apiStaticColumns *[]openapi.PayloadsStaticColumn) (types.List, diag.Diagnostics) {
	if apiStaticColumns == nil || len(*apiStaticColumns) == 0 {
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
	DataPlaneName            types.String               `tfsdk:"data_plane_name"`
	Tables                   types.Map                  `tfsdk:"tables"`

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
	AutoReplicateNewTables                       types.Bool   `tfsdk:"auto_replicate_new_tables"`
	AppendOnly                                   types.Bool   `tfsdk:"append_only"`
	StaticColumns                                types.List   `tfsdk:"static_columns"`
	StagingSchema                                types.String `tfsdk:"staging_schema"`
	ForceUTCTimezone                             types.Bool   `tfsdk:"force_utc_timezone"`
	WriteRawBinaryValues                         types.Bool   `tfsdk:"write_raw_binary_values"`
}

func (p Pipeline) toAdvancedSettings(ctx context.Context) (*openapi.PayloadsAdvancedPipelineSettingsPayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	flushConfig, flushConfigDiags := buildFlushConfig(ctx, p.FlushConfig)
	diags.Append(flushConfigDiags...)
	if diags.HasError() {
		return nil, diags
	}

	staticColumns, staticColumnsDiags := staticColumnsToAPI(ctx, p.StaticColumns)
	diags.Append(staticColumnsDiags...)
	if diags.HasError() {
		return nil, diags
	}

	settings := &openapi.PayloadsAdvancedPipelineSettingsPayload{
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
		AutoReplicateNewTables:                       p.AutoReplicateNewTables.ValueBoolPointer(),
		AppendOnly:                                   p.AppendOnly.ValueBoolPointer(),
		StaticColumns:                                staticColumns,
		StagingSchema:                                p.StagingSchema.ValueStringPointer(),
		ForceUTCTimezone:                             p.ForceUTCTimezone.ValueBoolPointer(),
		WriteRawBinaryValues:                         p.WriteRawBinaryValues.ValueBoolPointer(),
	}
	if flushConfig != nil {
		settings.FlushIntervalSeconds = intPtrFromInt64(flushConfig.FlushIntervalSeconds)
		settings.BufferRows = intPtrFromInt64(flushConfig.BufferRows)
		settings.FlushSizeKb = intPtrFromInt64(flushConfig.FlushSizeKB)
	}

	return settings, diags
}

func intPtrFromInt64(v types.Int64) *int {
	ptr := v.ValueInt64Pointer()
	if ptr == nil {
		return nil
	}
	return lib.ToPtr(int(*ptr))
}

func (p Pipeline) ToAPIPayload(ctx context.Context) (openapi.PayloadsPipelinePayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	tables := map[string]Table{}
	diags.Append(p.Tables.ElementsAs(ctx, &tables, false)...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	var apiTables []openapi.PayloadsTablePayload
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPIPayload(ctx)
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

	advancedSettings, advDiags := p.toAdvancedSettings(ctx)
	diags.Append(advDiags...)
	if diags.HasError() {
		return openapi.PayloadsPipelinePayload{}, diags
	}

	destCfg := p.DestinationConfig.ToAPIModel()
	return openapi.PayloadsPipelinePayload{
		Name:                     lib.ToPtr(p.Name.ValueString()),
		SourceReaderUUID:         sourceReaderUUID,
		Tables:                   &apiTables,
		DestinationUUID:          destinationUUID,
		SpecificDestCfg:          &destCfg,
		SnowflakeEcoScheduleUUID: snowflakeEcoScheduleUUID,
		EncryptionKeyUUID:        encryptionKeyUUID,
		DataPlaneName:            lib.ToPtr(p.DataPlaneName.ValueString()),
		AdvancedSettings:         advancedSettings,
	}, diags
}

func (p Pipeline) ToAPITablesForValidation(ctx context.Context) ([]openapi.PayloadsTable, diag.Diagnostics) {
	tables := map[string]Table{}
	diags := p.Tables.ElementsAs(ctx, &tables, false)
	if diags.HasError() {
		return nil, diags
	}

	var apiTables []openapi.PayloadsTable
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPITable(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiTables = append(apiTables, apiTable)
	}

	return apiTables, diags
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
	advSettings := apiModel.AdvancedSettings

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
	var appendOnly types.Bool
	var stagingSchema types.String
	var forceUTCTimezone types.Bool
	var writeRawBinaryValues types.Bool

	autoReplicateNewTables := types.BoolValue(false)
	staticColumns, staticColumnsDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: StaticColumnAttrTypes}, []StaticColumn{})
	diags.Append(staticColumnsDiags...)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	if advSettings.DropDeletedColumns != nil {
		dropDeletedColumns = types.BoolValue(*advSettings.DropDeletedColumns)
	}
	if advSettings.EnableSoftDelete != nil {
		softDeleteRows = types.BoolValue(*advSettings.EnableSoftDelete)
	}
	if advSettings.IncludeArtieUpdatedAtColumn != nil {
		includeArtieUpdatedAtColumn = types.BoolValue(*advSettings.IncludeArtieUpdatedAtColumn)
	}
	if advSettings.IncludeDatabaseUpdatedAtColumn != nil {
		includeDatabaseUpdatedAtColumn = types.BoolValue(*advSettings.IncludeDatabaseUpdatedAtColumn)
	}
	if advSettings.IncludeArtieOperationColumn != nil {
		includeArtieOperationColumn = types.BoolValue(*advSettings.IncludeArtieOperationColumn)
	}
	if advSettings.IncludeFullSourceTableNameColumn != nil {
		includeFullSourceTableNameColumn = types.BoolValue(*advSettings.IncludeFullSourceTableNameColumn)
	}
	if advSettings.IncludeFullSourceTableNameColumnAsPrimaryKey != nil {
		includeFullSourceTableNameColumnAsPrimaryKey = types.BoolValue(*advSettings.IncludeFullSourceTableNameColumnAsPrimaryKey)
	}
	if advSettings.DefaultSourceSchema != nil {
		defaultSourceSchema = types.StringValue(*advSettings.DefaultSourceSchema)
	}
	if advSettings.SplitEventsByType != nil {
		splitEventsByType = types.BoolValue(*advSettings.SplitEventsByType)
	}
	if advSettings.IncludeSourceMetadataColumn != nil {
		includeSourceMetadataColumn = types.BoolValue(*advSettings.IncludeSourceMetadataColumn)
	}
	if advSettings.AutoReplicateNewTables != nil {
		autoReplicateNewTables = types.BoolValue(*advSettings.AutoReplicateNewTables)
	}
	if advSettings.AppendOnly != nil {
		appendOnly = types.BoolValue(*advSettings.AppendOnly)
	}
	if advSettings.StagingSchema != nil {
		stagingSchema = types.StringValue(*advSettings.StagingSchema)
	}
	if advSettings.ForceUTCTimezone != nil {
		forceUTCTimezone = types.BoolValue(*advSettings.ForceUTCTimezone)
	}
	if advSettings.WriteRawBinaryValues != nil {
		writeRawBinaryValues = types.BoolValue(*advSettings.WriteRawBinaryValues)
	}

	flushConfigMap := map[string]attr.Value{}
	if advSettings.FlushIntervalSeconds != nil {
		flushConfigMap["flush_interval_seconds"] = types.Int64Value(int64(*advSettings.FlushIntervalSeconds))
	}
	if advSettings.BufferRows != nil {
		flushConfigMap["buffer_rows"] = types.Int64Value(int64(*advSettings.BufferRows))
	}
	if advSettings.FlushSizeKb != nil {
		flushConfigMap["flush_size_kb"] = types.Int64Value(int64(*advSettings.FlushSizeKb))
	}
	if len(flushConfigMap) > 0 {
		var flushConfigDiags diag.Diagnostics
		flushConfig, flushConfigDiags = types.ObjectValue(flushAttrTypes, flushConfigMap)
		diags.Append(flushConfigDiags...)
		if diags.HasError() {
			return Pipeline{}, diags
		}
	}

	var staticColumnsDiags2 diag.Diagnostics
	staticColumns, staticColumnsDiags2 = staticColumnsFromAPI(ctx, advSettings.StaticColumns)
	diags.Append(staticColumnsDiags2...)
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
		DataPlaneName:            types.StringValue(apiModel.DataPlaneName),

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
		AutoReplicateNewTables:                       autoReplicateNewTables,
		AppendOnly:                                   appendOnly,
		StaticColumns:                                staticColumns,
		StagingSchema:                                stagingSchema,
		ForceUTCTimezone:                             forceUTCTimezone,
		WriteRawBinaryValues:                         writeRawBinaryValues,
	}, diags
}
