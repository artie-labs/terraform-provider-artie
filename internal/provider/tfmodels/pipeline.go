package tfmodels

import (
	"context"

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
	FlushConfig                    types.Object `tfsdk:"flush_config"`
	DropDeletedColumns             types.Bool   `tfsdk:"drop_deleted_columns"`
	SoftDeleteRows                 types.Bool   `tfsdk:"soft_delete_rows"`
	IncludeArtieUpdatedAtColumn    types.Bool   `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool   `tfsdk:"include_database_updated_at_column"`
}

func (p Pipeline) ToAPIBaseModel(ctx context.Context) (artieclient.BaseDeployment, diag.Diagnostics) {
	tables := map[string]Table{}
	diags := p.Tables.ElementsAs(ctx, &tables, false)
	apiTables := []artieclient.Table{}
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return artieclient.BaseDeployment{}, diags
		}
		apiTables = append(apiTables, apiTable)
	}

	sourceReaderUUID, sourceReaderDiags := parseOptionalUUID(p.SourceReaderUUID)
	diags.Append(sourceReaderDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	destinationUUID, destDiags := parseOptionalUUID(p.DestinationUUID)
	diags.Append(destDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	snowflakeEcoScheduleUUID, snowflakeDiags := parseOptionalUUID(p.SnowflakeEcoScheduleUUID)
	diags.Append(snowflakeDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	flushConfig, flushConfigDiags := buildFlushConfig(ctx, p.FlushConfig)
	diags.Append(flushConfigDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	return artieclient.BaseDeployment{
		Name:                     p.Name.ValueString(),
		SourceReaderUUID:         sourceReaderUUID,
		Tables:                   apiTables,
		DestinationUUID:          destinationUUID,
		DestinationConfig:        p.DestinationConfig.ToAPIModel(),
		SnowflakeEcoScheduleUUID: snowflakeEcoScheduleUUID,
		DataPlaneName:            p.DataPlaneName.ValueString(),
		// Advanced settings:
		FlushConfig:                    flushConfig.ToAPIModel(),
		DropDeletedColumns:             p.DropDeletedColumns.ValueBoolPointer(),
		EnableSoftDelete:               p.SoftDeleteRows.ValueBoolPointer(),
		IncludeArtieUpdatedAtColumn:    p.IncludeArtieUpdatedAtColumn.ValueBoolPointer(),
		IncludeDatabaseUpdatedAtColumn: p.IncludeDatabaseUpdatedAtColumn.ValueBoolPointer(),
	}, diags
}

func (p Pipeline) ToAPIModel(ctx context.Context) (artieclient.Deployment, diag.Diagnostics) {
	apiBaseModel, diags := p.ToAPIBaseModel(ctx)
	if diags.HasError() {
		return artieclient.Deployment{}, diags
	}

	uuid, uuidDiags := parseUUID(p.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Deployment{}, diags
	}

	return artieclient.Deployment{
		UUID:           uuid,
		Status:         p.Status.ValueString(),
		BaseDeployment: apiBaseModel,
	}, diags
}

func PipelineFromAPIModel(ctx context.Context, apiModel artieclient.Deployment) (Pipeline, diag.Diagnostics) {
	tables, diags := TablesFromAPIModel(ctx, apiModel.Source.Tables)
	if diags.HasError() {
		return Pipeline{}, diags
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TableAttrTypes}, tables)
	diags.Append(mapDiags...)

	destinationConfig := DeploymentDestinationConfigFromAPIModel(apiModel.DestinationConfig)

	var flushConfig types.Object
	if apiModel.FlushConfig != nil {
		flushConfigObj, flushConfigDiags := DeploymentFlushConfigFromAPIModel(ctx, *apiModel.FlushConfig)
		diags.Append(flushConfigDiags...)
		if diags.HasError() {
			return Pipeline{}, diags
		}
		flushConfig = flushConfigObj
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
		FlushConfig:                    flushConfig,
		DropDeletedColumns:             types.BoolPointerValue(apiModel.DropDeletedColumns),
		SoftDeleteRows:                 types.BoolPointerValue(apiModel.EnableSoftDelete),
		IncludeArtieUpdatedAtColumn:    types.BoolPointerValue(apiModel.IncludeArtieUpdatedAtColumn),
		IncludeDatabaseUpdatedAtColumn: types.BoolPointerValue(apiModel.IncludeDatabaseUpdatedAtColumn),
	}, diags
}
