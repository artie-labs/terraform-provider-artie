package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Pipeline struct {
	UUID                     types.String                 `tfsdk:"uuid"`
	Name                     types.String                 `tfsdk:"name"`
	Status                   types.String                 `tfsdk:"status"`
	SourceReaderUUID         types.String                 `tfsdk:"source_reader_uuid"`
	DestinationUUID          types.String                 `tfsdk:"destination_connector_uuid"`
	DestinationConfig        *DeploymentDestinationConfig `tfsdk:"destination_config"`
	SnowflakeEcoScheduleUUID types.String                 `tfsdk:"snowflake_eco_schedule_uuid"`
	DataPlaneName            types.String                 `tfsdk:"data_plane_name"`
	Tables                   types.Map                    `tfsdk:"tables"`

	// Advanced settings
	FlushConfig                    types.Object `tfsdk:"flush_rules"`
	DropDeletedColumns             types.Bool   `tfsdk:"drop_deleted_columns"`
	SoftDeleteRows                 types.Bool   `tfsdk:"soft_delete_rows"`
	IncludeArtieUpdatedAtColumn    types.Bool   `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool   `tfsdk:"include_database_updated_at_column"`
}

func (p Pipeline) ToAPIBaseModel(ctx context.Context) (artieclient.BasePipeline, diag.Diagnostics) {
	tables := map[string]Table{}
	diags := p.Tables.ElementsAs(ctx, &tables, false)
	if diags.HasError() {
		return artieclient.BasePipeline{}, diags
	}

	apiTables := []artieclient.Table{}
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return artieclient.BasePipeline{}, diags
		}
		apiTables = append(apiTables, apiTable)
	}

	sourceReaderUUID, sourceReaderDiags := parseOptionalUUID(p.SourceReaderUUID)
	diags.Append(sourceReaderDiags...)
	if diags.HasError() {
		return artieclient.BasePipeline{}, diags
	}

	destinationUUID, destDiags := parseOptionalUUID(p.DestinationUUID)
	diags.Append(destDiags...)
	if diags.HasError() {
		return artieclient.BasePipeline{}, diags
	}

	snowflakeEcoScheduleUUID, snowflakeDiags := parseOptionalUUID(p.SnowflakeEcoScheduleUUID)
	diags.Append(snowflakeDiags...)
	if diags.HasError() {
		return artieclient.BasePipeline{}, diags
	}

	flushConfig, flushConfigDiags := buildFlushConfig(ctx, p.FlushConfig)
	diags.Append(flushConfigDiags...)
	if diags.HasError() {
		return artieclient.BasePipeline{}, diags
	}

	advancedSettings := artieclient.AdvancedSettings{
		DropDeletedColumns:             p.DropDeletedColumns.ValueBoolPointer(),
		EnableSoftDelete:               p.SoftDeleteRows.ValueBoolPointer(),
		IncludeArtieUpdatedAtColumn:    p.IncludeArtieUpdatedAtColumn.ValueBoolPointer(),
		IncludeDatabaseUpdatedAtColumn: p.IncludeDatabaseUpdatedAtColumn.ValueBoolPointer(),
	}
	if flushConfig != nil {
		advancedSettings.FlushIntervalSeconds = flushConfig.FlushIntervalSeconds.ValueInt64Pointer()
		advancedSettings.BufferRows = flushConfig.BufferRows.ValueInt64Pointer()
		advancedSettings.FlushSizeKB = flushConfig.FlushSizeKB.ValueInt64Pointer()
	}

	return artieclient.BasePipeline{
		Name:                     p.Name.ValueString(),
		SourceReaderUUID:         sourceReaderUUID,
		Tables:                   apiTables,
		DestinationUUID:          destinationUUID,
		DestinationConfig:        p.DestinationConfig.ToAPIModel(),
		SnowflakeEcoScheduleUUID: snowflakeEcoScheduleUUID,
		DataPlaneName:            p.DataPlaneName.ValueString(),
		AdvancedSettings:         &advancedSettings,
	}, diags
}

func (p Pipeline) ToAPIModel(ctx context.Context) (artieclient.Pipeline, diag.Diagnostics) {
	apiBaseModel, diags := p.ToAPIBaseModel(ctx)
	if diags.HasError() {
		return artieclient.Pipeline{}, diags
	}

	uuid, uuidDiags := parseUUID(p.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Pipeline{}, diags
	}

	return artieclient.Pipeline{
		UUID:         uuid,
		Status:       p.Status.ValueString(),
		BasePipeline: apiBaseModel,
	}, diags
}

func PipelineFromAPIModel(ctx context.Context, apiModel artieclient.Pipeline) (Pipeline, diag.Diagnostics) {
	tables, diags := TablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TableAttrTypes}, tables)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	destinationConfig := DeploymentDestinationConfigFromAPIModel(apiModel.DestinationConfig)

	var flushConfig types.Object
	var dropDeletedColumns types.Bool
	var softDeleteRows types.Bool
	var includeArtieUpdatedAtColumn types.Bool
	var includeDatabaseUpdatedAtColumn types.Bool
	if apiModel.AdvancedSettings != nil {
		if apiModel.AdvancedSettings.DropDeletedColumns != nil {
			dropDeletedColumns = types.BoolValue(*apiModel.AdvancedSettings.DropDeletedColumns)
		}
		if apiModel.AdvancedSettings.EnableSoftDelete != nil {
			softDeleteRows = types.BoolValue(*apiModel.AdvancedSettings.EnableSoftDelete)
		}
		if apiModel.AdvancedSettings.IncludeArtieUpdatedAtColumn != nil {
			includeArtieUpdatedAtColumn = types.BoolValue(*apiModel.AdvancedSettings.IncludeArtieUpdatedAtColumn)
		}
		if apiModel.AdvancedSettings.IncludeDatabaseUpdatedAtColumn != nil {
			includeDatabaseUpdatedAtColumn = types.BoolValue(*apiModel.AdvancedSettings.IncludeDatabaseUpdatedAtColumn)
		}

		flushConfigMap := map[string]attr.Value{}
		if apiModel.AdvancedSettings.FlushIntervalSeconds != nil {
			flushConfigMap["flush_interval_seconds"] = types.Int64Value(*apiModel.AdvancedSettings.FlushIntervalSeconds)
		}
		if apiModel.AdvancedSettings.BufferRows != nil {
			flushConfigMap["buffer_rows"] = types.Int64Value(*apiModel.AdvancedSettings.BufferRows)
		}
		if apiModel.AdvancedSettings.FlushSizeKB != nil {
			flushConfigMap["flush_size_kb"] = types.Int64Value(*apiModel.AdvancedSettings.FlushSizeKB)
		}
		if len(flushConfigMap) > 0 {
			var flushConfigDiags diag.Diagnostics
			flushConfig, flushConfigDiags = types.ObjectValue(flushAttrTypes, flushConfigMap)
			diags.Append(flushConfigDiags...)
			if diags.HasError() {
				return Pipeline{}, diags
			}
		}
	}

	return Pipeline{
		UUID:                     types.StringValue(apiModel.UUID.String()),
		Name:                     types.StringValue(apiModel.Name),
		Status:                   types.StringValue(apiModel.Status),
		Tables:                   tablesMap,
		SourceReaderUUID:         optionalUUIDToStringValue(apiModel.SourceReaderUUID),
		DestinationUUID:          optionalUUIDToStringValue(apiModel.DestinationUUID),
		DestinationConfig:        &destinationConfig,
		SnowflakeEcoScheduleUUID: optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID),
		DataPlaneName:            types.StringValue(apiModel.DataPlaneName),

		// Advanced settings:
		DropDeletedColumns:             dropDeletedColumns,
		SoftDeleteRows:                 softDeleteRows,
		IncludeArtieUpdatedAtColumn:    includeArtieUpdatedAtColumn,
		IncludeDatabaseUpdatedAtColumn: includeDatabaseUpdatedAtColumn,
		FlushConfig:                    flushConfig,
	}, diags
}
