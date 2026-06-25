package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/artieclient"
)

func TestTablesFromAPIModel_NilBoolSettingsReadBackAsFalse(t *testing.T) {
	// When the API returns nil for "absent means off" toggles (which it does when they are
	// false), they must read back as an explicit `false`, not null. Otherwise an explicit
	// `false` in the Terraform config triggers a post-apply consistency error (false -> null).
	apiTables := []artieclient.Table{
		{
			UUID:   uuid.New(),
			Name:   "billing_counter_event",
			Schema: "task_mgmt",
			AdvancedSettings: artieclient.AdvancedTableSettings{
				EncryptJSONBColumns:        nil,
				SkipDeletes:                nil,
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
	apiTables := []artieclient.Table{
		{
			UUID: uuid.New(),
			Name: "orders",
			AdvancedSettings: artieclient.AdvancedTableSettings{
				EncryptJSONBColumns: &trueVal,
				SkipDeletes:         &falseVal,
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
		RangeEnabled:        types.BoolValue(true),
		RangeChunkSize:      types.Int64Value(5000000),
		RangeMaxParallelism: types.Int64Value(5),
		RangeBatchSize:      types.Int64Value(0),
	}

	apiTable, diags := table.ToAPIModel(t.Context())
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.NotNil(t, apiTable.AdvancedSettings.RangeSettings)
	assert.True(t, apiTable.AdvancedSettings.RangeSettings.Enabled)
	assert.Equal(t, 5000000, apiTable.AdvancedSettings.RangeSettings.ChunkSize)
	assert.Equal(t, 5, apiTable.AdvancedSettings.RangeSettings.MaxParallelism)
	assert.Equal(t, 0, apiTable.AdvancedSettings.RangeSettings.BatchSize)
}

func TestTablesFromAPIModel_RangeSettings(t *testing.T) {
	apiTables := []artieclient.Table{
		{
			UUID:   uuid.New(),
			Name:   "offers",
			Schema: "public",
			AdvancedSettings: artieclient.AdvancedTableSettings{
				RangeSettings: &artieclient.RangeSettings{
					Enabled:        true,
					ChunkSize:      5000000,
					MaxParallelism: 5,
					BatchSize:      0,
				},
			},
		},
	}

	tables, diags := TablesFromAPIModel(t.Context(), apiTables)
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	table := tables["public.offers"]
	assert.True(t, table.RangeEnabled.ValueBool(), "range_enabled should round-trip from API")
	assert.Equal(t, int64(5000000), table.RangeChunkSize.ValueInt64())
	assert.Equal(t, int64(5), table.RangeMaxParallelism.ValueInt64())
	assert.Equal(t, int64(0), table.RangeBatchSize.ValueInt64())
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
