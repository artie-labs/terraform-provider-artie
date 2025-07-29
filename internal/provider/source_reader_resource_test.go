package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/provider/tfmodels"
)

func TestSourceReaderResource_ValidateConfig(t *testing.T) {
	connectorUUID := uuid.New().String()
	tableType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                        types.StringType,
			"schema":                      types.StringType,
			"is_partitioned":              types.BoolType,
			"columns_to_exclude":          types.ListType{ElemType: types.StringType},
			"columns_to_include":          types.ListType{ElemType: types.StringType},
			"child_partition_schema_name": types.StringType,
		},
	}

	{
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(false),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID:     types.StringValue(connectorUUID),
			IsShared:          types.BoolValue(false),
			BackfillBatchSize: types.Int64Value(60000),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "The maximum allowed value for `backfill_batch_size` is 50,000.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(true),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "You must specify a `tables` block if `is_shared` is set to true.")
	}
	{
		// no validation error if tables is unknown
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(true),
			Tables:        types.MapUnknown(tableType),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(false),
			Tables: types.MapValueMust(
				tableType,
				map[string]attr.Value{
					"test_table": types.ObjectValueMust(
						tableType.AttrTypes,
						map[string]attr.Value{
							"name":                        types.StringValue("test_table"),
							"schema":                      types.StringValue(""),
							"is_partitioned":              types.BoolValue(false),
							"columns_to_exclude":          types.ListValueMust(types.StringType, []attr.Value{}),
							"columns_to_include":          types.ListValueMust(types.StringType, []attr.Value{}),
							"child_partition_schema_name": types.StringValue(""),
						},
					),
				},
			),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "You should not specify a `tables` block if `is_shared` is set to false.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(true),
			Tables: types.MapValueMust(
				tableType,
				map[string]attr.Value{
					"wrong_key": types.ObjectValueMust(
						tableType.AttrTypes,
						map[string]attr.Value{
							"name":                        types.StringValue("test_table"),
							"schema":                      types.StringValue("public"),
							"is_partitioned":              types.BoolValue(false),
							"columns_to_exclude":          types.ListValueMust(types.StringType, []attr.Value{}),
							"columns_to_include":          types.ListValueMust(types.StringType, []attr.Value{}),
							"child_partition_schema_name": types.StringValue(""),
						},
					),
				},
			),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "Table key \"wrong_key\" should be \"public.test_table\" instead.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID: types.StringValue(connectorUUID),
			IsShared:      types.BoolValue(true),
			Tables: types.MapValueMust(
				tableType,
				map[string]attr.Value{
					"public.test_table": types.ObjectValueMust(
						tableType.AttrTypes,
						map[string]attr.Value{
							"name":                        types.StringValue("test_table"),
							"schema":                      types.StringValue("public"),
							"is_partitioned":              types.BoolValue(false),
							"columns_to_exclude":          types.ListValueMust(types.StringType, []attr.Value{types.StringValue("col1")}),
							"columns_to_include":          types.ListValueMust(types.StringType, []attr.Value{types.StringValue("col2")}),
							"child_partition_schema_name": types.StringValue(""),
						},
					),
				},
			),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "You can only use one of `columns_to_include` and `columns_to_exclude` within a source reader.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "You must specify `databases_to_unify` if `enable_unify_across_databases` is set to true.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db2")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "`databases_to_unify` should include the database you specified for `database_name`.")
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
}
