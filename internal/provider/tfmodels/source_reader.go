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

func (s SourceReaderTable) ToAPIModel(ctx context.Context) (artieclient.SourceReaderTable, diag.Diagnostics) {
	colsToExclude, diags := parseList[string](ctx, s.ColumnsToExclude)
	colsToInclude, includeDiags := parseList[string](ctx, s.ColumnsToInclude)
	diags.Append(includeDiags...)

	return artieclient.SourceReaderTable{
		Name:                     s.Name.ValueString(),
		Schema:                   s.Schema.ValueString(),
		IsPartitioned:            s.IsPartitioned.ValueBool(),
		ColumnsToExclude:         colsToExclude,
		ColumnsToInclude:         colsToInclude,
		ChildPartitionSchemaName: s.ChildPartitionSchemaName.ValueString(),
		UnifyAcrossSchemas:       s.UnifyAcrossSchemas.ValueBool(),
		UnifyAcrossDatabases:     s.UnifyAcrossDatabases.ValueBool(),
	}, diags
}

func SourceReaderTablesFromAPIModel(ctx context.Context, apiTablesMap map[string]artieclient.SourceReaderTable) (types.Map, diag.Diagnostics) {
	tables := map[string]SourceReaderTable{}
	var diags diag.Diagnostics
	for key, apiTable := range apiTablesMap {
		colsToExclude, excludeDiags := types.ListValueFrom(ctx, types.StringType, apiTable.ColumnsToExclude)
		diags.Append(excludeDiags...)

		colsToInclude, includeDiags := types.ListValueFrom(ctx, types.StringType, apiTable.ColumnsToInclude)
		diags.Append(includeDiags...)
		tables[key] = SourceReaderTable{
			Name:                     types.StringValue(apiTable.Name),
			Schema:                   types.StringValue(apiTable.Schema),
			IsPartitioned:            types.BoolValue(apiTable.IsPartitioned),
			ColumnsToExclude:         colsToExclude,
			ColumnsToInclude:         colsToInclude,
			ChildPartitionSchemaName: types.StringValue(apiTable.ChildPartitionSchemaName),
			UnifyAcrossSchemas:       types.BoolValue(apiTable.UnifyAcrossSchemas),
			UnifyAcrossDatabases:     types.BoolValue(apiTable.UnifyAcrossDatabases),
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
	PartitionRegexPattern           types.String `tfsdk:"partition_suffix_regex_pattern"`
	EnableUnifyAcrossSchemas        types.Bool   `tfsdk:"enable_unify_across_schemas"`
	EnableUnifyAcrossDatabases      types.Bool   `tfsdk:"enable_unify_across_databases"`
	DatabasesToUnify                types.List   `tfsdk:"databases_to_unify"`
	Tables                          types.Map    `tfsdk:"tables"`
}

func (s SourceReader) ToAPIBaseModel(ctx context.Context) (artieclient.BaseSourceReader, diag.Diagnostics) {
	connectorUUID, diags := parseUUID(s.ConnectorUUID)
	if diags.HasError() {
		return artieclient.BaseSourceReader{}, diags
	}

	apiTablesMap := map[string]artieclient.SourceReaderTable{}
	if !s.Tables.IsNull() && !s.Tables.IsUnknown() {
		tablesMap := map[string]SourceReaderTable{}
		diags.Append(s.Tables.ElementsAs(ctx, &tablesMap, false)...)
		if diags.HasError() {
			return artieclient.BaseSourceReader{}, diags
		}

		for key, table := range tablesMap {
			apiTable, tableDiags := table.ToAPIModel(ctx)
			diags.Append(tableDiags...)
			apiTablesMap[key] = apiTable
		}
	}

	settings := artieclient.SourceReaderSettings{
		BackfillBatchSize:               s.BackfillBatchSize.ValueInt64(),
		EnableHeartbeats:                s.EnableHeartbeats.ValueBool(),
		OneTopicPerSchema:               s.OneTopicPerSchema.ValueBool(),
		PostgresPublicationNameOverride: s.PostgresPublicationNameOverride.ValueString(),
		PostgresPublicationMode:         s.PostgresPublicationMode.ValueString(),
		PostgresReplicationSlotOverride: s.PostgresReplicationSlotOverride.ValueString(),
		EnableUnifyAcrossSchemas:        s.EnableUnifyAcrossSchemas.ValueBool(),
		EnableUnifyAcrossDatabases:      s.EnableUnifyAcrossDatabases.ValueBool(),
	}

	if !s.PartitionRegexPattern.IsNull() && !s.PartitionRegexPattern.IsUnknown() {
		settings.PartitionRegex = &artieclient.PartitionRegex{
			Pattern: s.PartitionRegexPattern.ValueString(),
		}
	}

	if !s.DatabasesToUnify.IsNull() && !s.DatabasesToUnify.IsUnknown() {
		databasesToUnify, diags := parseList[string](ctx, s.DatabasesToUnify)
		diags.Append(diags...)
		settings.DatabasesToUnify = databasesToUnify
	}

	return artieclient.BaseSourceReader{
		Name:          s.Name.ValueString(),
		DataPlaneName: s.DataPlaneName.ValueString(),
		ConnectorUUID: connectorUUID,
		DatabaseName:  s.DatabaseName.ValueString(),
		ContainerName: s.OracleContainerName.ValueString(),
		IsShared:      s.IsShared.ValueBool(),
		Settings:      settings,
		Tables:        apiTablesMap,
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
	tablesMap, diags := SourceReaderTablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	databasesToUnify, diags := types.ListValueFrom(ctx, types.StringType, apiModel.Settings.DatabasesToUnify)
	if diags.HasError() {
		return SourceReader{}, diags
	}

	sourceReader := SourceReader{
		UUID:                            types.StringValue(apiModel.UUID.String()),
		Name:                            types.StringValue(apiModel.Name),
		DataPlaneName:                   types.StringValue(apiModel.DataPlaneName),
		ConnectorUUID:                   types.StringValue(apiModel.ConnectorUUID.String()),
		IsShared:                        types.BoolValue(apiModel.IsShared),
		DatabaseName:                    types.StringValue(apiModel.DatabaseName),
		OracleContainerName:             types.StringValue(apiModel.ContainerName),
		BackfillBatchSize:               types.Int64Value(apiModel.Settings.BackfillBatchSize),
		EnableHeartbeats:                types.BoolValue(apiModel.Settings.EnableHeartbeats),
		OneTopicPerSchema:               types.BoolValue(apiModel.Settings.OneTopicPerSchema),
		PostgresPublicationNameOverride: types.StringValue(apiModel.Settings.PostgresPublicationNameOverride),
		PostgresPublicationMode:         types.StringValue(apiModel.Settings.PostgresPublicationMode),
		PostgresReplicationSlotOverride: types.StringValue(apiModel.Settings.PostgresReplicationSlotOverride),
		EnableUnifyAcrossSchemas:        types.BoolValue(apiModel.Settings.EnableUnifyAcrossSchemas),
		EnableUnifyAcrossDatabases:      types.BoolValue(apiModel.Settings.EnableUnifyAcrossDatabases),
		DatabasesToUnify:                databasesToUnify,
		Tables:                          tablesMap,
	}

	if apiModel.Settings.PartitionRegex != nil {
		sourceReader.PartitionRegexPattern = types.StringValue(apiModel.Settings.PartitionRegex.Pattern)
	}

	return sourceReader, diags
}
