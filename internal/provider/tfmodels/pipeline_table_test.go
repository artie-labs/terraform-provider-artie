package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	rangeSettings, rangeDiags := types.ObjectValueFrom(t.Context(), RangeSettingsAttrTypes, RangeSettings{
		Enabled:        types.BoolValue(true),
		ChunkSize:      types.Int64Value(5000000),
		MaxParallelism: types.Int64Value(5),
		BatchSize:      types.Int64Value(0),
	})
	assert.False(t, rangeDiags.HasError(), "unexpected diagnostics: %v", rangeDiags)

	table := Table{
		Name:          types.StringValue("offers"),
		Schema:        types.StringValue("public"),
		RangeSettings: rangeSettings,
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
	assert.False(t, table.RangeSettings.IsNull(), "range_settings should round-trip from API")

	var rangeSettings RangeSettings
	diags = table.RangeSettings.As(t.Context(), &rangeSettings, basetypes.ObjectAsOptions{})
	assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.True(t, rangeSettings.Enabled.ValueBool())
	assert.Equal(t, int64(5000000), rangeSettings.ChunkSize.ValueInt64())
	assert.Equal(t, int64(5), rangeSettings.MaxParallelism.ValueInt64())
	assert.Equal(t, int64(0), rangeSettings.BatchSize.ValueInt64())
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
