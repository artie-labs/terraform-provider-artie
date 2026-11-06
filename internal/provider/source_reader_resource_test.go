package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
			"unify_across_schemas":        types.BoolType,
			"unify_across_databases":      types.BoolType,
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
							"unify_across_schemas":        types.BoolValue(false),
							"unify_across_databases":      types.BoolValue(false),
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
							"unify_across_schemas":        types.BoolValue(false),
							"unify_across_databases":      types.BoolValue(false),
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
							"unify_across_schemas":        types.BoolValue(false),
							"unify_across_databases":      types.BoolValue(false),
						},
					),
				},
			),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
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
		assert.Contains(t, diags.Errors()[0].Detail(), "`enable_unify_across_databases` is only applicable if `mssql_replication_method` is set to `change_tracking` or `trn_log_backup`.")
	}
	{
		// fn_dblog was removed, so it can't unify
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			MSSQLReplicationMethod:     types.StringValue("fn_dblog"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "is set to `change_tracking` or `trn_log_backup`.")
	}
	{
		// Transaction log backups can unify
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			MSSQLReplicationMethod:     types.StringValue("trn_log_backup"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
	{
		// A method that isn't known yet isn't rejected
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			MSSQLReplicationMethod:     types.StringUnknown(),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
	{
		config := tfmodels.SourceReader{
			ConnectorUUID:              types.StringValue(connectorUUID),
			DatabaseName:               types.StringValue("test_db"),
			MSSQLReplicationMethod:     types.StringValue("change_tracking"),
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
			MSSQLReplicationMethod:     types.StringValue("change_tracking"),
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
			MSSQLReplicationMethod:     types.StringValue("change_tracking"),
			EnableUnifyAcrossDatabases: types.BoolValue(true),
			DatabasesToUnify:           types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test_db")}),
		}

		diags := validateSourceReaderConfig(t.Context(), config)
		assert.False(t, diags.HasError())
	}
}

func TestSourceReaderResource_ValidateArchiveConnector(t *testing.T) {
	config := func(method types.String) tfmodels.SourceReader {
		return tfmodels.SourceReader{
			ConnectorUUID:          types.StringValue(uuid.NewString()),
			MSSQLReplicationMethod: method,
			ArchiveConnectorUUID:   types.StringValue(uuid.NewString()),
		}
	}

	// Transaction log backups read from the archive connector, and a method that isn't known yet is checked later
	assert.False(t, validateSourceReaderConfig(t.Context(), config(types.StringValue("trn_log_backup"))).HasError())
	assert.False(t, validateSourceReaderConfig(t.Context(), config(types.StringUnknown())).HasError())

	for _, method := range []types.String{types.StringNull(), types.StringValue("change_tracking")} {
		diags := validateSourceReaderConfig(t.Context(), config(method))
		assert.True(t, diags.HasError())
		assert.Contains(t, diags.Errors()[0].Detail(), "`archive_connector_uuid` is only applicable if `mssql_replication_method` is set to `trn_log_backup`.")
	}

	// An explicit "" is rejected; an unknown value (a share created in the same apply) is checked later
	empty := config(types.StringValue("trn_log_backup"))
	empty.ArchiveConnectorUUID = types.StringValue("")
	diags := validateSourceReaderConfig(t.Context(), empty)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "`archive_connector_uuid` can't be empty.")
	empty.ArchiveConnectorUUID = types.StringUnknown()
	assert.False(t, validateSourceReaderConfig(t.Context(), empty).HasError())
}

func TestArchiveConnectorPlanModifier(t *testing.T) {
	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	(&SourceReaderResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	plan := func(config, state, configMethod, stateMethod types.String) types.String {
		newObj := func() tftypes.Value { return tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil) }
		// tfsdk.Config has no setter, so build its value as a State.
		cs := tfsdk.State{Schema: schemaResp.Schema, Raw: newObj()}
		assert.False(t, cs.SetAttribute(ctx, path.Root("mssql_replication_method"), configMethod).HasError())
		c := tfsdk.Config{Schema: schemaResp.Schema, Raw: cs.Raw}
		s := tfsdk.State{Schema: schemaResp.Schema, Raw: newObj()}
		assert.False(t, s.SetAttribute(ctx, path.Root("mssql_replication_method"), stateMethod).HasError())
		// The method's own plan value is left unknown, as it can be before UseStateForUnknown runs.
		p := tfsdk.Plan{Schema: schemaResp.Schema, Raw: newObj()}
		assert.False(t, p.SetAttribute(ctx, path.Root("mssql_replication_method"), types.StringUnknown()).HasError())
		planValue := config
		if config.IsNull() {
			// Unset optional+computed attributes start out unknown on update.
			planValue = types.StringUnknown()
		}
		req := planmodifier.StringRequest{ConfigValue: config, StateValue: state, PlanValue: planValue, Config: c, State: s, Plan: p}
		resp := &planmodifier.StringResponse{PlanValue: planValue}
		archiveConnectorPlanModifier{}.PlanModifyString(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError())
		return resp.PlanValue
	}
	archive := types.StringValue(uuid.NewString())
	trn := types.StringValue("trn_log_backup")

	// Left out of the config, a TRN source reader keeps its share, whether the method is set or left to state
	assert.Equal(t, archive, plan(types.StringNull(), archive, trn, trn))
	assert.Equal(t, archive, plan(types.StringNull(), archive, types.StringNull(), trn))
	// Moving off TRN unlinks it, as the API does
	assert.Equal(t, types.StringValue(""), plan(types.StringNull(), archive, types.StringValue("change_tracking"), trn))
	// A non-TRN reader with no method in config or state keeps "" instead of showing a diff
	assert.Equal(t, types.StringValue(""), plan(types.StringNull(), types.StringValue(""), types.StringNull(), types.StringNull()))
	// Not known until the method is, or until the source reader is created
	assert.True(t, plan(types.StringNull(), archive, types.StringUnknown(), trn).IsUnknown())
	assert.True(t, plan(types.StringNull(), types.StringNull(), trn, types.StringNull()).IsUnknown())
	// Set in the config, including to a share created in the same apply
	assert.Equal(t, archive, plan(archive, types.StringValue(""), trn, trn))
	assert.True(t, plan(types.StringUnknown(), archive, trn, trn).IsUnknown())
}
