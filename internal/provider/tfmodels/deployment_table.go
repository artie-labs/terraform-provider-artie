package tfmodels

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/artieclient"
)

type MergePredicate struct {
	PartitionField types.String `tfsdk:"partition_field"`
}

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`

	// Advanced table settings
	Alias           types.String `tfsdk:"alias"`
	ExcludeColumns  types.List   `tfsdk:"columns_to_exclude"`
	ColumnsToHash   types.List   `tfsdk:"columns_to_hash"`
	SkipDeletes     types.Bool   `tfsdk:"skip_deletes"`
	MergePredicates types.List   `tfsdk:"merge_predicates"`
}

func (t Table) ToAPIModel(ctx context.Context) (artieclient.Table, diag.Diagnostics) {
	tableUUID := uuid.Nil
	var diags diag.Diagnostics
	if t.UUID.ValueString() != "" {
		tableUUID, diags = parseUUID(t.UUID)
	}

	colsToExclude, excludeDiags := parseOptionalStringList(ctx, t.ExcludeColumns)
	diags.Append(excludeDiags...)

	colsToHash, hashDiags := parseOptionalStringList(ctx, t.ColumnsToHash)
	diags.Append(hashDiags...)

	mergePredicates, mergePredDiags := parseOptionalObjectList[MergePredicate](ctx, t.MergePredicates)
	diags.Append(mergePredDiags...)
	var clientMergePreds *[]artieclient.MergePredicate
	if mergePredicates != nil && len(*mergePredicates) > 0 {
		clientMPs := []artieclient.MergePredicate{}
		for _, mp := range *mergePredicates {
			clientMPs = append(clientMPs, artieclient.MergePredicate{PartitionField: mp.PartitionField.ValueString()})
		}
		clientMergePreds = &clientMPs
	}

	if diags.HasError() {
		return artieclient.Table{}, diags
	}

	return artieclient.Table{
		UUID:                 tableUUID,
		Name:                 t.Name.ValueString(),
		Schema:               t.Schema.ValueString(),
		EnableHistoryMode:    t.EnableHistoryMode.ValueBool(),
		IndividualDeployment: t.IndividualDeployment.ValueBool(),
		IsPartitioned:        t.IsPartitioned.ValueBool(),
		Alias:                t.Alias.ValueStringPointer(),
		ExcludeColumns:       colsToExclude,
		ColumnsToHash:        colsToHash,
		SkipDeletes:          t.SkipDeletes.ValueBoolPointer(),
		MergePredicates:      clientMergePreds,
	}, diags
}

func TablesFromAPIModel(ctx context.Context, apiModelTables []artieclient.Table) (map[string]Table, diag.Diagnostics) {
	tables := map[string]Table{}
	var diags diag.Diagnostics
	for _, apiTable := range apiModelTables {
		tableKey := apiTable.Name
		if apiTable.Schema != "" {
			tableKey = fmt.Sprintf("%s.%s", apiTable.Schema, apiTable.Name)
		}

		colsToExclude, excludeDiags := optionalStringListToStringValue(ctx, apiTable.ExcludeColumns)
		diags.Append(excludeDiags...)

		colsToHash, hashDiags := optionalStringListToStringValue(ctx, apiTable.ColumnsToHash)
		diags.Append(hashDiags...)

		var mergePredicates types.List
		attrTypes := map[string]attr.Type{"partition_field": types.StringType}
		if apiTable.MergePredicates != nil {
			preds := []attr.Value{}
			for _, mp := range *apiTable.MergePredicates {
				pred, predDiags := types.ObjectValueFrom(ctx, attrTypes, MergePredicate{PartitionField: types.StringValue(mp.PartitionField)})
				diags.Append(predDiags...)
				preds = append(preds, pred)
			}
			var mergePredDiags diag.Diagnostics
			mergePredicates, mergePredDiags = types.ListValue(basetypes.ObjectType{AttrTypes: attrTypes}, preds)
			diags.Append(mergePredDiags...)
		}

		tables[tableKey] = Table{
			UUID:                 types.StringValue(apiTable.UUID.String()),
			Name:                 types.StringValue(apiTable.Name),
			Schema:               types.StringValue(apiTable.Schema),
			EnableHistoryMode:    types.BoolValue(apiTable.EnableHistoryMode),
			IndividualDeployment: types.BoolValue(apiTable.IndividualDeployment),
			IsPartitioned:        types.BoolValue(apiTable.IsPartitioned),
			Alias:                types.StringPointerValue(apiTable.Alias),
			ExcludeColumns:       colsToExclude,
			ColumnsToHash:        colsToHash,
			SkipDeletes:          types.BoolPointerValue(apiTable.SkipDeletes),
			MergePredicates:      mergePredicates,
		}
	}

	if diags.HasError() {
		return map[string]Table{}, diags
	}

	return tables, diags
}
