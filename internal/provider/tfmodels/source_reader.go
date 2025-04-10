package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type SourceReaderTable struct {
	Name                     types.String `tfsdk:"name"`
	Schema                   types.String `tfsdk:"schema"`
	ColumnsToExclude         types.List   `tfsdk:"columns_to_exclude"`
	ColumnsToInclude         types.List   `tfsdk:"columns_to_include"`
	ChildPartitionSchemaName types.String `tfsdk:"child_partition_schema_name"`
}

var SourceReaderTableAttrTypes = map[string]attr.Type{
	"name":                        types.StringType,
	"schema":                      types.StringType,
	"columns_to_exclude":          types.ListType{ElemType: types.StringType},
	"columns_to_include":          types.ListType{ElemType: types.StringType},
	"child_partition_schema_name": types.StringType,
}

func (s SourceReaderTable) ToAPIModel(ctx context.Context) (artieclient.SourceReaderTable, diag.Diagnostics) {
	colsToExclude, diags := parseList[string](ctx, s.ColumnsToExclude)
	colsToInclude, includeDiags := parseList[string](ctx, s.ColumnsToInclude)
	diags.Append(includeDiags...)

	return artieclient.SourceReaderTable{
		Name:                     s.Name.ValueString(),
		Schema:                   s.Schema.ValueString(),
		ColumnsToExclude:         colsToExclude,
		ColumnsToInclude:         colsToInclude,
		ChildPartitionSchemaName: s.ChildPartitionSchemaName.ValueString(),
	}, diags
}

func SourceReaderTablesFromAPIModel(ctx context.Context, apiTablesMap map[string]artieclient.SourceReaderTable) (map[string]SourceReaderTable, diag.Diagnostics) {
	tables := map[string]SourceReaderTable{}
	var diags diag.Diagnostics
	for key, apiTable := range apiTablesMap {
		colsToExclude, excludeDiags := optionalStringListToStringValue(ctx, &apiTable.ColumnsToExclude)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := optionalStringListToStringValue(ctx, &apiTable.ColumnsToInclude)
		diags.Append(includeDiags...)
		tables[key] = SourceReaderTable{
			Name:                     types.StringValue(apiTable.Name),
			Schema:                   types.StringValue(apiTable.Schema),
			ColumnsToExclude:         colsToExclude,
			ColumnsToInclude:         colsToInclude,
			ChildPartitionSchemaName: types.StringValue(apiTable.ChildPartitionSchemaName),
		}
	}

	return tables, diags
}

type SourceReader struct {
	UUID                            types.String `tfsdk:"uuid"`
	Name                            types.String `tfsdk:"name"`
	DataPlaneName                   types.String `tfsdk:"data_plane_name"`
	ConnectorUUID                   types.String `tfsdk:"connector_uuid"`
	IsShared                        types.Bool   `tfsdk:"is_shared"`
	DatabaseName                    types.String `tfsdk:"database_name"`
	OracleContainerName             types.String `tfsdk:"oracle_container_name"`
	OneTopicPerSchema               types.Bool   `tfsdk:"one_topic_per_schema"`
	PostgresPublicationNameOverride types.String `tfsdk:"postgres_publication_name_override"`
	PostgresReplicationSlotOverride types.String `tfsdk:"postgres_replication_slot_override"`
	Tables                          types.Map    `tfsdk:"tables"`
}

func (s SourceReader) ToAPIBaseModel(ctx context.Context) (artieclient.BaseSourceReader, diag.Diagnostics) {
	connectorUUID, diags := parseUUID(s.ConnectorUUID)
	if diags.HasError() {
		return artieclient.BaseSourceReader{}, diags
	}

	tablesMap := map[string]SourceReaderTable{}
	tablesDiags := s.Tables.ElementsAs(ctx, &tablesMap, false)
	if tablesDiags.HasError() {
		return artieclient.BaseSourceReader{}, tablesDiags
	}

	apiTablesMap := map[string]artieclient.SourceReaderTable{}
	for key, table := range tablesMap {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		apiTablesMap[key] = apiTable
	}

	return artieclient.BaseSourceReader{
		Name:          s.Name.ValueString(),
		DataPlaneName: s.DataPlaneName.ValueString(),
		ConnectorUUID: connectorUUID,
		DatabaseName:  s.DatabaseName.ValueString(),
		ContainerName: s.OracleContainerName.ValueString(),
		IsShared:      s.IsShared.ValueBool(),
		Settings: artieclient.SourceReaderSettings{
			OneTopicPerSchema:               s.OneTopicPerSchema.ValueBool(),
			PostgresPublicationNameOverride: s.PostgresPublicationNameOverride.ValueString(),
			PostgresReplicationSlotOverride: s.PostgresReplicationSlotOverride.ValueString(),
		},
		Tables: apiTablesMap,
	}, diags
}

func (s SourceReader) ToAPIModel(ctx context.Context) (artieclient.SourceReader, diag.Diagnostics) {
	uuid, diags := parseUUID(s.UUID)
	if diags.HasError() {
		return artieclient.SourceReader{}, diags
	}

	baseSourceReader, diags := s.ToAPIBaseModel(ctx)
	if diags.HasError() {
		return artieclient.SourceReader{}, diags
	}

	return artieclient.SourceReader{
		UUID:             uuid,
		BaseSourceReader: baseSourceReader,
	}, diags
}

func SourceReaderFromAPIModel(ctx context.Context, apiModel artieclient.SourceReader) (SourceReader, diag.Diagnostics) {
	tables, diags := SourceReaderTablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: SourceReaderTableAttrTypes}, tables)
	diags.Append(mapDiags...)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	return SourceReader{
		UUID:                            types.StringValue(apiModel.UUID.String()),
		Name:                            types.StringValue(apiModel.Name),
		DataPlaneName:                   types.StringValue(apiModel.DataPlaneName),
		ConnectorUUID:                   types.StringValue(apiModel.ConnectorUUID.String()),
		IsShared:                        types.BoolValue(apiModel.IsShared),
		DatabaseName:                    types.StringValue(apiModel.DatabaseName),
		OracleContainerName:             types.StringValue(apiModel.ContainerName),
		OneTopicPerSchema:               types.BoolValue(apiModel.Settings.OneTopicPerSchema),
		PostgresPublicationNameOverride: types.StringValue(apiModel.Settings.PostgresPublicationNameOverride),
		PostgresReplicationSlotOverride: types.StringValue(apiModel.Settings.PostgresReplicationSlotOverride),
		Tables:                          tablesMap,
	}, diags
}
