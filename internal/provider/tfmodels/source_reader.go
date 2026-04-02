package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type SourceReaderTable struct {
	Name                     types.String `tfsdk:"name"`
	Schema                   types.String `tfsdk:"schema"`
	IsPartitioned            types.Bool   `tfsdk:"is_partitioned"`
	ColumnsToExclude         types.List   `tfsdk:"columns_to_exclude"`
	ColumnsToInclude         types.List   `tfsdk:"columns_to_include"`
	ChildPartitionSchemaName types.String `tfsdk:"child_partition_schema_name"`
	UnifyAcrossSchemas       types.Bool   `tfsdk:"unify_across_schemas"`
	UnifyAcrossDatabases     types.Bool   `tfsdk:"unify_across_databases"`
}

var SourceReaderTableAttrTypes = map[string]attr.Type{
	"name":                        types.StringType,
	"schema":                      types.StringType,
	"is_partitioned":              types.BoolType,
	"columns_to_exclude":          types.ListType{ElemType: types.StringType},
	"columns_to_include":          types.ListType{ElemType: types.StringType},
	"child_partition_schema_name": types.StringType,
	"unify_across_schemas":        types.BoolType,
	"unify_across_databases":      types.BoolType,
}

func (s SourceReaderTable) ToAPIModel(ctx context.Context) (openapi.PayloadsSourceReaderTable, diag.Diagnostics) {
	colsToExclude, diags := parseOptionalList[string](ctx, s.ColumnsToExclude)
	colsToInclude, includeDiags := parseOptionalList[string](ctx, s.ColumnsToInclude)
	diags.Append(includeDiags...)

	return openapi.PayloadsSourceReaderTable{
		Name:                 s.Name.ValueStringPointer(),
		Schema:               s.Schema.ValueStringPointer(),
		IsPartitioned:        s.IsPartitioned.ValueBoolPointer(),
		ExcludeColumns:       colsToExclude,
		IncludeColumns:       colsToInclude,
		UnifyAcrossSchemas:   s.UnifyAcrossSchemas.ValueBoolPointer(),
		UnifyAcrossDatabases: s.UnifyAcrossDatabases.ValueBoolPointer(),
	}, diags
}

func SourceReaderTablesFromAPIModel(ctx context.Context, apiTablesConfig *openapi.PayloadsSourceReaderTablesConfig) (types.Map, diag.Diagnostics) {
	tables := map[string]SourceReaderTable{}
	var diags diag.Diagnostics

	if apiTablesConfig != nil {
		for key, apiTable := range *apiTablesConfig {
			colsToExclude, excludeDiags := optionalStringListToListValue(ctx, apiTable.ExcludeColumns)
			diags.Append(excludeDiags...)

			colsToInclude, includeDiags := optionalStringListToListValue(ctx, apiTable.IncludeColumns)
			diags.Append(includeDiags...)

			tables[key] = SourceReaderTable{
				Name:                     types.StringValue(lib.RemovePtr(apiTable.Name)),
				Schema:                   types.StringValue(lib.RemovePtr(apiTable.Schema)),
				IsPartitioned:            types.BoolValue(lib.RemovePtr(apiTable.IsPartitioned)),
				ColumnsToExclude:         colsToExclude,
				ColumnsToInclude:         colsToInclude,
				ChildPartitionSchemaName: types.StringValue(""),
				UnifyAcrossSchemas:       types.BoolValue(lib.RemovePtr(apiTable.UnifyAcrossSchemas)),
				UnifyAcrossDatabases:     types.BoolValue(lib.RemovePtr(apiTable.UnifyAcrossDatabases)),
			}
		}
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: SourceReaderTableAttrTypes}, tables)
	diags.Append(mapDiags...)

	return tablesMap, diags
}

type SourceReader struct {
	UUID                            types.String `tfsdk:"uuid"`
	Name                            types.String `tfsdk:"name"`
	DataPlaneName                   types.String `tfsdk:"data_plane_name"`
	ConnectorUUID                   types.String `tfsdk:"connector_uuid"`
	IsShared                        types.Bool   `tfsdk:"is_shared"`
	DatabaseName                    types.String `tfsdk:"database_name"`
	BackfillBatchSize               types.Int64  `tfsdk:"backfill_batch_size"`
	OracleContainerName             types.String `tfsdk:"oracle_container_name"`
	EnableHeartbeats                types.Bool   `tfsdk:"enable_heartbeats"`
	OneTopicPerSchema               types.Bool   `tfsdk:"one_topic_per_schema"`
	PostgresPublicationNameOverride types.String `tfsdk:"postgres_publication_name_override"`
	PostgresPublicationMode         types.String `tfsdk:"postgres_publication_mode"`
	PostgresReplicationSlotOverride types.String `tfsdk:"postgres_replication_slot_override"`
	PublishViaPartitionRoot         types.Bool   `tfsdk:"publish_via_partition_root"`
	EnableUnifyAcrossSchemas        types.Bool   `tfsdk:"enable_unify_across_schemas"`
	UnifyAcrossSchemasRegex         types.String `tfsdk:"unify_across_schemas_regex"`
	MSSQLReplicationMethod          types.String `tfsdk:"mssql_replication_method"`
	EnableUnifyAcrossDatabases      types.Bool   `tfsdk:"enable_unify_across_databases"`
	DatabasesToUnify                types.List   `tfsdk:"databases_to_unify"`
	DisableAutoFetchTables          types.Bool   `tfsdk:"disable_auto_fetch_tables"`
	Tables                          types.Map    `tfsdk:"tables"`
}

func (s SourceReader) toAPISettings(ctx context.Context) (openapi.PayloadsSourceReaderSettingsPayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := openapi.PayloadsSourceReaderSettingsPayload{
		BackfillBatchSize:         lib.ToPtr(int(s.BackfillBatchSize.ValueInt64())),
		EnableHeartbeats:          s.EnableHeartbeats.ValueBoolPointer(),
		OneTopicPerSchema:         s.OneTopicPerSchema.ValueBoolPointer(),
		PublicationNameOverride:   s.PostgresPublicationNameOverride.ValueStringPointer(),
		PublicationAutoCreateMode: s.PostgresPublicationMode.ValueStringPointer(),
		ReplicationSlotOverride:   s.PostgresReplicationSlotOverride.ValueStringPointer(),
		PublishViaPartitionRoot:   s.PublishViaPartitionRoot.ValueBoolPointer(),
		UnifyAcrossSchemas:        s.EnableUnifyAcrossSchemas.ValueBoolPointer(),
		UnifyAcrossSchemasRegex:   s.UnifyAcrossSchemasRegex.ValueStringPointer(),
		MssqlReplicationMethod:    s.MSSQLReplicationMethod.ValueStringPointer(),
		UnifyAcrossDatabases:      s.EnableUnifyAcrossDatabases.ValueBoolPointer(),
		DisableAutoFetchTables:    s.DisableAutoFetchTables.ValueBoolPointer(),
	}

	if !s.DatabasesToUnify.IsNull() && !s.DatabasesToUnify.IsUnknown() {
		databasesToUnify, listDiags := parseOptionalList[string](ctx, s.DatabasesToUnify)
		diags.Append(listDiags...)
		if diags.HasError() {
			return settings, diags
		}
		settings.DatabasesToSync = databasesToUnify
	}

	return settings, diags
}

func (s SourceReader) toAPITablesConfig(ctx context.Context) (*openapi.PayloadsSourceReaderTablesConfig, diag.Diagnostics) {
	if s.Tables.IsNull() || s.Tables.IsUnknown() {
		return nil, nil
	}

	tablesMap := map[string]SourceReaderTable{}
	diags := s.Tables.ElementsAs(ctx, &tablesMap, false)
	if diags.HasError() {
		return nil, diags
	}

	apiTables := openapi.PayloadsSourceReaderTablesConfig{}
	for key, table := range tablesMap {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		apiTables[key] = apiTable
	}

	return &apiTables, diags
}

func SourceReaderCreateRequestFromAPIModel(model openapi.PayloadsSourceReader) openapi.RouterSourceReaderCreateRequest {
	return openapi.RouterSourceReaderCreateRequest{
		ConnectorUUID: model.ConnectorUUID,
		Name:          &model.Name,
		DataPlaneName: &model.DataPlaneName,
		Database:      &model.Database,
		ContainerName: &model.ContainerName,
		IsShared:      model.IsShared,
		Settings:      &model.Settings,
		TablesConfig:  model.TablesConfig,
	}
}

func (s SourceReader) ToAPIPayload(ctx context.Context) (openapi.PayloadsSourceReader, diag.Diagnostics) {
	var diags diag.Diagnostics

	connectorUUID, connDiags := parseUUID(s.ConnectorUUID)
	diags.Append(connDiags...)
	if diags.HasError() {
		return openapi.PayloadsSourceReader{}, diags
	}

	settings, settingsDiags := s.toAPISettings(ctx)
	diags.Append(settingsDiags...)
	if diags.HasError() {
		return openapi.PayloadsSourceReader{}, diags
	}

	tablesConfig, tablesDiags := s.toAPITablesConfig(ctx)
	diags.Append(tablesDiags...)

	return openapi.PayloadsSourceReader{
		ConnectorUUID: connectorUUID,
		Name:          s.Name.ValueString(),
		DataPlaneName: s.DataPlaneName.ValueString(),
		Database:      s.DatabaseName.ValueString(),
		ContainerName: s.OracleContainerName.ValueString(),
		IsShared:      s.IsShared.ValueBoolPointer(),
		Settings:      settings,
		TablesConfig:  tablesConfig,
	}, diags
}

func (s SourceReader) ToAPIModel(ctx context.Context) (openapi.PayloadsSourceReader, diag.Diagnostics) {
	uuid, diags := parseUUID(s.UUID)
	if diags.HasError() {
		return openapi.PayloadsSourceReader{}, diags
	}

	model, baseDiags := s.ToAPIPayload(ctx)
	diags.Append(baseDiags...)
	if diags.HasError() {
		return openapi.PayloadsSourceReader{}, diags
	}

	model.Uuid = uuid
	return model, diags
}

func SourceReaderFromAPIModel(ctx context.Context, apiModel openapi.PayloadsSourceReader) (SourceReader, diag.Diagnostics) {
	tablesMap, diags := SourceReaderTablesFromAPIModel(ctx, apiModel.TablesConfig)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	databasesToUnify, listDiags := optionalStringListToListValue(ctx, apiModel.Settings.DatabasesToSync)
	diags.Append(listDiags...)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	sourceReader := SourceReader{
		UUID:                            types.StringValue(apiModel.Uuid.String()),
		Name:                            types.StringValue(apiModel.Name),
		DataPlaneName:                   types.StringValue(apiModel.DataPlaneName),
		ConnectorUUID:                   types.StringValue(apiModel.ConnectorUUID.String()),
		IsShared:                        types.BoolValue(lib.RemovePtr(apiModel.IsShared)),
		DatabaseName:                    types.StringValue(apiModel.Database),
		OracleContainerName:             types.StringValue(apiModel.ContainerName),
		BackfillBatchSize:               types.Int64Value(int64(lib.RemovePtr(apiModel.Settings.BackfillBatchSize))),
		EnableHeartbeats:                types.BoolValue(lib.RemovePtr(apiModel.Settings.EnableHeartbeats)),
		OneTopicPerSchema:               types.BoolValue(lib.RemovePtr(apiModel.Settings.OneTopicPerSchema)),
		PostgresPublicationNameOverride: types.StringValue(lib.RemovePtr(apiModel.Settings.PublicationNameOverride)),
		PostgresPublicationMode:         types.StringValue(lib.RemovePtr(apiModel.Settings.PublicationAutoCreateMode)),
		PostgresReplicationSlotOverride: types.StringValue(lib.RemovePtr(apiModel.Settings.ReplicationSlotOverride)),
		PublishViaPartitionRoot:         types.BoolPointerValue(apiModel.Settings.PublishViaPartitionRoot),
		EnableUnifyAcrossSchemas:        types.BoolValue(lib.RemovePtr(apiModel.Settings.UnifyAcrossSchemas)),
		UnifyAcrossSchemasRegex:         types.StringPointerValue(apiModel.Settings.UnifyAcrossSchemasRegex),
		MSSQLReplicationMethod:          types.StringValue(lib.RemovePtr(apiModel.Settings.MssqlReplicationMethod)),
		EnableUnifyAcrossDatabases:      types.BoolValue(lib.RemovePtr(apiModel.Settings.UnifyAcrossDatabases)),
		DatabasesToUnify:                databasesToUnify,
		DisableAutoFetchTables:          types.BoolValue(lib.RemovePtr(apiModel.Settings.DisableAutoFetchTables)),
		Tables:                          tablesMap,
	}

	return sourceReader, diags
}
