package tfmodels

import (
	"testing"

	"github.com/artie-labs/base/lib/ptr"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

func TestTablesFromAPIModel_NilBoolSettingsReadBackAsFalse(t *testing.T) {
	// When the API returns nil for "absent means off" toggles (which it does when they are
	// false), they must read back as an explicit `false`, not null. Otherwise an explicit
	// `false` in the Terraform config triggers a post-apply consistency error (false -> null).
	apiTables := []openapi.PayloadsTable{
		{
			Uuid:   ptr.ToPtr(uuid.New()),
			Name:   ptr.ToPtr("billing_counter_event"),
			Schema: ptr.ToPtr("task_mgmt"),
			AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
				EncryptJSONBColumns:        nil,
				SkipDelete:                 nil,
				UnifyAcrossSchemas:         nil,
				UnifyAcrossDatabases:       nil,
				ShouldBackfillHistoryTable: nil,
				SkipBackfill:               nil,
				SkipNoOpUpdates:            nil,
			},
		},
	}

	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	table, ok := tables["task_mgmt.billing_counter_event"]
	assert.True(t, ok, "expected table keyed by schema.name")

	for name, value := range map[string]bool{
		"encrypt_jsonb_columns":  table.EncryptJSONBColumns.IsNull(),
		"skip_deletes":           table.SkipDeletes.IsNull(),
		"unify_across_schemas":   table.UnifyAcrossSchemas.IsNull(),
		"unify_across_databases": table.UnifyAcrossDatabases.IsNull(),
		"backfill_history_table": table.BackfillHistoryTable.IsNull(),
		"skip_backfill":          table.SkipBackfill.IsNull(),
		"skip_no_op_updates":     table.SkipNoOpUpdates.IsNull(),
	} {
		assert.False(t, value, "%s should not read back as null", name)
	}

	assert.False(t, table.EncryptJSONBColumns.ValueBool())
	assert.False(t, table.SkipDeletes.ValueBool())
	assert.False(t, table.UnifyAcrossSchemas.ValueBool())
	assert.False(t, table.UnifyAcrossDatabases.ValueBool())
	assert.False(t, table.BackfillHistoryTable.ValueBool())
	assert.False(t, table.SkipBackfill.ValueBool())
	assert.False(t, table.SkipNoOpUpdates.ValueBool())
}

func TestTablesFromAPIModel_BoolSettingsRoundTripExplicitValues(t *testing.T) {
	trueVal := true
	falseVal := false
	apiTables := []openapi.PayloadsTable{
		{
			Uuid: ptr.ToPtr(uuid.New()),
			Name: ptr.ToPtr("orders"),
			AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
				EncryptJSONBColumns: &trueVal,
				SkipDelete:          &falseVal,
			},
		},
	}

	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	table := tables["orders"]
	assert.True(t, table.EncryptJSONBColumns.ValueBool(), "explicit true should round-trip as true")
	assert.False(t, table.SkipDeletes.IsNull(), "explicit false should not read back as null")
	assert.False(t, table.SkipDeletes.ValueBool(), "explicit false should round-trip as false")
}

func TestTableToAPIModel_RangeSettings(t *testing.T) {
	table := Table{
		Name:                types.StringValue("offers"),
		Schema:              types.StringValue("public"),
		RangeBackfill:       types.BoolValue(true),
		RangeChunkSize:      types.Int64Value(5000000),
		RangeMaxParallelism: types.Int64Value(5),
		RangeBatchSize:      types.Int64Value(0),
	}

	apiTable, diags := table.ToAPIModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotNil(t, apiTable.AdvancedSettings.RangeSettings)
	assert.True(t, *apiTable.AdvancedSettings.RangeSettings.Enabled)
	assert.Equal(t, 5000000, *apiTable.AdvancedSettings.RangeSettings.ChunksSize)
	assert.Equal(t, 5, *apiTable.AdvancedSettings.RangeSettings.MaxParallelism)
	assert.Equal(t, 0, *apiTable.AdvancedSettings.RangeSettings.BatchSize)
}

func TestTableToAPIModel_NullRangeBackfillOmitsRangeSettings(t *testing.T) {
	table := Table{
		Name:   types.StringValue("offers"),
		Schema: types.StringValue("public"),
	}

	apiTable, diags := table.ToAPIModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Nil(t, apiTable.AdvancedSettings.RangeSettings)
}

func TestTablesFromAPIModel_NilRangeSettingsReadBackAsDisabled(t *testing.T) {
	apiTables := []openapi.PayloadsTable{
		{
			Uuid:   ptr.ToPtr(uuid.New()),
			Name:   ptr.ToPtr("offers"),
			Schema: ptr.ToPtr("public"),
			AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
				RangeSettings: nil,
			},
		},
	}

	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	table := tables["public.offers"]
	assert.False(t, table.RangeBackfill.IsNull(), "range_backfill should not read back as null")
	assert.False(t, table.RangeBackfill.ValueBool())
	assert.Equal(t, int64(0), table.RangeChunkSize.ValueInt64())
	assert.Equal(t, int64(0), table.RangeMaxParallelism.ValueInt64())
	assert.Equal(t, int64(0), table.RangeBatchSize.ValueInt64())
}

func TestTablesFromAPIModel_RangeSettings(t *testing.T) {
	apiTables := []openapi.PayloadsTable{
		{
			Uuid:   ptr.ToPtr(uuid.New()),
			Name:   ptr.ToPtr("offers"),
			Schema: ptr.ToPtr("public"),
			AdvancedSettings: &openapi.PayloadsTableAdvancedSettings{
				RangeSettings: &openapi.PayloadsRangeSettings{
					Enabled:        ptr.ToPtr(true),
					ChunksSize:     ptr.ToPtr(5000000),
					MaxParallelism: ptr.ToPtr(5),
					BatchSize:      ptr.ToPtr(0),
				},
			},
		},
	}

	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	table := tables["public.offers"]
	assert.True(t, table.RangeBackfill.ValueBool())
	assert.Equal(t, int64(5000000), table.RangeChunkSize.ValueInt64())
	assert.Equal(t, int64(5), table.RangeMaxParallelism.ValueInt64())
	assert.Equal(t, int64(0), table.RangeBatchSize.ValueInt64())
}

func TestTablePrimaryKeysOverrideRoundTrip(t *testing.T) {
	primaryKeysOverride, primaryKeysOverrideDiags := types.ListValueFrom(t.Context(), types.StringType, []string{"account_id", "region"})
	assert.False(t, primaryKeysOverrideDiags.HasError(), "unexpected diagnostics: %v", primaryKeysOverrideDiags)

	table := Table{
		Name:                types.StringValue("accounts"),
		Schema:              types.StringValue("public"),
		PrimaryKeysOverride: primaryKeysOverride,
	}

	apiTable, diags := table.ToAPIModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, []string{"account_id", "region"}, *apiTable.AdvancedSettings.PrimaryKeysOverride)

	// Read back through ValidationTables, which reshapes the request payload into the response table type.
	apiTables, err := ValidationTables(&[]openapi.PayloadsTablePayload{apiTable})
	assert.NoError(t, err)
	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, primaryKeysOverride, tables["public.accounts"].PrimaryKeysOverride)
}

func TestBoolPointerValueOrFalse(t *testing.T) {
	trueVal := true
	falseVal := false

	assert.True(t, boolPointerValueOrFalse(&trueVal).IsNull() == false)
	assert.True(t, boolPointerValueOrFalse(&trueVal).ValueBool())

	assert.False(t, boolPointerValueOrFalse(&falseVal).IsNull())
	assert.False(t, boolPointerValueOrFalse(&falseVal).ValueBool())

	assert.False(t, boolPointerValueOrFalse(nil).IsNull(), "nil should coalesce to a known false, not null")
	assert.False(t, boolPointerValueOrFalse(nil).ValueBool())
}

func TestMergePredicatesPartitionByRoundTrip(t *testing.T) {
	{
		// partitionBy set in the dashboard survives a read and is sent back on apply
		apiPreds := []openapi.PayloadsMergePredicates{{PartitionField: ptr.ToPtr("created_at"), PartitionType: ptr.ToPtr("time"), PartitionBy: ptr.ToPtr("monthly")}}
		list, diags := MergePredicatesFromAPIModel(t.Context(), &apiPreds)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

		apiTable, diags := Table{Name: types.StringValue("events"), MergePredicates: list}.ToAPIModel(t.Context())
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Equal(t, &apiPreds, apiTable.AdvancedSettings.MergePredicates)
	}
	{
		// No partitionBy reads back as null and is omitted from the request
		apiPreds := []openapi.PayloadsMergePredicates{{PartitionField: ptr.ToPtr("id"), PartitionType: ptr.ToPtr("integer")}}
		list, diags := MergePredicatesFromAPIModel(t.Context(), &apiPreds)
		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

		var preds []MergePredicate
		assert.False(t, list.ElementsAs(t.Context(), &preds, false).HasError())
		assert.True(t, preds[0].PartitionBy.IsNull())
		assert.Nil(t, preds[0].ToAPIModel().PartitionBy)
	}
}

func TestMergePredicateValidate(t *testing.T) {
	{
		// partition_by on a time predicate
		assert.NoError(t, MergePredicate{PartitionField: types.StringValue("created_at"), PartitionType: types.StringValue("time"), PartitionBy: types.StringValue("hourly")}.Validate())
	}
	{
		// Integer predicate without partition_by
		assert.NoError(t, MergePredicate{PartitionField: types.StringValue("id"), PartitionType: types.StringValue("integer")}.Validate())
	}
	{
		// partition_by on an integer predicate
		err := MergePredicate{PartitionField: types.StringValue("id"), PartitionType: types.StringValue("integer"), PartitionBy: types.StringValue("daily")}.Validate()
		assert.ErrorContains(t, err, `merge predicate "id" is an integer column, so it cannot have a `+"`partition_by`")
	}
}
