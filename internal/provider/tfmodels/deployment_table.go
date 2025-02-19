package tfmodels

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias          types.String `tfsdk:"alias"`
	ExcludeColumns types.String `tfsdk:"exclude_columns"`
	ColumnsToHash  types.String `tfsdk:"columns_to_hash"`
}

func (t Table) ToAPIModel() artieclient.Table {
	tableUUID := uuid.Nil
	if t.UUID.ValueString() != "" {
		tableUUID = uuid.MustParse(t.UUID.ValueString())
	}

	return artieclient.Table{
		UUID:                 tableUUID,
		Name:                 t.Name.ValueString(),
		Schema:               t.Schema.ValueString(),
		EnableHistoryMode:    t.EnableHistoryMode.ValueBool(),
		IndividualDeployment: t.IndividualDeployment.ValueBool(),
		IsPartitioned:        t.IsPartitioned.ValueBool(),
		Alias:                parseOptionalString(t.Alias),
		ExcludeColumns:       parseOptionalStringList(t.ExcludeColumns),
		ColumnsToHash:        parseOptionalStringList(t.ColumnsToHash),
	}
}

func TablesFromAPIModel(apiModelTables []artieclient.Table) map[string]Table {
	tables := map[string]Table{}
	for _, apiTable := range apiModelTables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}
		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                optionalStringToStringValue(apiTable.Alias),
			ExcludeColumns:       optionalStringListToStringValue(apiTable.ExcludeColumns),
			ColumnsToHash:        optionalStringListToStringValue(apiTable.ColumnsToHash),
		}
	}

	return tables
}
