package provider

import (
	"context"
	"fmt"
	"slices"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SourceReaderResource{}
var _ resource.ResourceWithConfigure = &SourceReaderResource{}
var _ resource.ResourceWithImportState = &SourceReaderResource{}

func NewSourceReaderResource() resource.Resource {
	return &SourceReaderResource{}
}

type SourceReaderResource struct {
	client artieclient.Client
}

func (r *SourceReaderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_reader"
}

func (r *SourceReaderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Source Reader resource. This represents a process that reads data from a source connector and inserts it info Kafka. A Source Reader can be used by multiple Pipelines, e.g. to read from a single PostgreSQL replication slot and copy the data to multiple destinations.",
		Attributes: map[string]schema.Attribute{
			"uuid":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"connector_uuid": schema.StringAttribute{Required: true, MarkdownDescription: "The source connector that we should read data from."},
			"name":           schema.StringAttribute{Optional: true, MarkdownDescription: "An optional human-readable label for this source reader."},
			"data_plane_name": schema.StringAttribute{
				MarkdownDescription: "The name of the data plane to deploy this source reader in. If this is not set, we will use the default data plane for your account. To see the full list of supported data planes on your account, click on 'New pipeline' in our UI.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_shared":                          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, this source reader can be used by multiple pipelines."},
			"database_name":                      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The name of the database we should read data from in the source connector. This should be specified if the source connector's type is DocumentDB, MongoDB, MySQL, MS SQL, Oracle (this maps to the service name), or PostgreSQL."},
			"oracle_container_name":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The name of the container (pluggable database) if the source type is Oracle and you are using a container database."},
			"backfill_batch_size":                schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}, MarkdownDescription: "The number of rows to read from the source database in each batch while backfilling. Maximum allowed value is 50,000. Default is 5,000."},
			"enable_heartbeats":                  schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If the source database is a very low-traffic PostgreSQL database (e.g., a dev database) and is running on Amazon RDS, we recommend setting this to true to prevent WAL growth issues. This is only applicable if the source type is PostgreSQL."},
			"one_topic_per_schema":               schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will write all incoming CDC events into a single Kafka topic per schema. This is currently only supported if your source is Oracle and your account has this feature enabled."},
			"postgres_publication_name_override": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set, this will override the name of the PostgreSQL publication. Otherwise, we will use our default value, `dbz_publication`. This is only applicable if the source type is PostgreSQL."},
			"postgres_publication_mode":          schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This should be set to `filtered` if the PostgreSQL publication in the source database is not set to include `ALL TABLES`. If that's the case, you will need to explicitly add tables to the publication. Otherwise, this should be set to `\"\"`."},
			"postgres_replication_slot_override": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set, this will override the name of the PostgreSQL replication slot. Otherwise, we will use our default value, `artie`. This is only applicable if the source type is PostgreSQL."},
			"publish_via_partition_root":         schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, changes to partitioned tables will be published using the root partitioned table's identity rather than the actual partition that was changed (The API defaults this to true). This is only applicable if the source type is PostgreSQL."},
			"partition_suffix_regex_pattern":     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If this source reader is reading any partitioned tables, this regex pattern should describe the expected suffix of each partition's name so that we can consume data from all partitions. If not set, this defaults to `_((default)|([0-9]{4})_(0[1-9]|1[012]))$` - meaning that for a table called `my_table` that's partitioned by month, we will detect partitions such as `my_table_default`, `my_table_2025_01`, `my_table_2025_02`, etc."},
			"enable_unify_across_schemas":        schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, you can specify tables that should be generalized to all schemas, meaning we will sync all tables with the same name into the same destination table. This is useful if you have multiple identical schemas and want to fan-in the data. This is only applicable if the source type is PostgreSQL."},
			"unify_across_schemas_regex":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If unify across schemas is enabled, this is an additional regex pattern that you can use to filter which schemas should be unified. This is only applicable if the source type is PostgreSQL."},
			"mssql_replication_method":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If unset, we will use the default replication method (Capture Instances). If set to `fn_dblog`, we will stream data from transaction logs via SQL access. This is only applicable if the source type is Microsoft SQL Server."},
			"enable_unify_across_databases":      schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, you can specify multiple databases within your Microsoft SQL Server that we should sync data from, and we will unify tables with the same name and schema into a single destination table. This is useful if you have multiple identical databases and want to fan-in the data. This is only applicable if the source type is Microsoft SQL Server and `mssql_replication_method` is set to `fn_dblog` or `change_tracking`."},
			"databases_to_unify":                 schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If `enable_unify_across_databases` is set to true, this should be a list of databases within your Microsoft SQL Server that we should sync data from. All tables that you opt into being unified should exist in each of these databases. This is only applicable if the source type is Microsoft SQL Server."},
			"disable_auto_fetch_tables":          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will not automatically fetch tables from the source database on the UI. This is useful if you have a large number of tables and you want to manually specify the schema before we fetch all the tables."},
			"tables": schema.MapNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A map of tables from the source database that you want this source reader to include CDC events for. This should be specified if (and only if) the source reader has `is_shared` set to true, and it must include all tables that are specified in the `tables` attribute of any pipeline that uses this source reader. The key for each table should be formatted as `schema_name.table_name` if your source database uses schemas, otherwise just `table_name`.",
				NestedObject: schema.NestedAttributeObject{
					// All non-required table attributes must use UseNonNullStateForUnknown() to prevent errors when adding a new table (see https://github.com/hashicorp/terraform-plugin-framework/issues/1197)
					Attributes: map[string]schema.Attribute{
						"name":                        schema.StringAttribute{Required: true, MarkdownDescription: "The name of the table in the source database."},
						"schema":                      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()}, MarkdownDescription: "The name of the schema the table belongs to in the source database. This must be specified if your source database uses schemas (such as PostgreSQL), e.g. `public`."},
						"is_partitioned":              schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseNonNullStateForUnknown()}, MarkdownDescription: "If the source table is partitioned, set this to true and we will ingest data from all of its partitions. You may also need to customize `partition_suffix_regex_pattern` on the source reader."},
						"columns_to_exclude":          schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "An optional list of columns to exclude from CDC events. This cannot be used if `columns_to_include` is also specified."},
						"columns_to_include":          schema.ListAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "An optional list of columns to include in CDC events. If not provided, all columns will be included. This cannot be used if `columns_to_exclude` is also specified."},
						"child_partition_schema_name": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseNonNullStateForUnknown()}, DeprecationMessage: "This field is deprecated and no longer used. It will be removed in a future version.", MarkdownDescription: "If the source table is partitioned and its child partitions are in a different schema, this should specify the name of that schema."},
						"unify_across_schemas":        schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseNonNullStateForUnknown()}, MarkdownDescription: "This should be set to true for any tables that you intend to unify across schemas in any pipeline that uses this source reader. This is only applicable if the source reader has `enable_unify_across_schemas` set to true."},
						"unify_across_databases":      schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseNonNullStateForUnknown()}, MarkdownDescription: "This should be set to true for any tables that you intend to unify across databases in any pipeline that uses this source reader. This is only applicable if the source reader has `enable_unify_across_databases` set to true and `databases_to_unify` filled."},
					},
				},
			},
		},
	}
}

func (r *SourceReaderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *SourceReaderResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.SourceReader
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *SourceReaderResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.SourceReader, bool) {
	var planData tfmodels.SourceReader
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *SourceReaderResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiSourceReader artieclient.SourceReader) {
	// Translate API response type into Terraform model and save it into state
	sourceReader, diags := tfmodels.SourceReaderFromAPIModel(ctx, apiSourceReader)
	diagnostics.Append(diags...)
	diagnostics.Append(state.Set(ctx, sourceReader)...)
}

func validateSourceReaderConfig(ctx context.Context, configData tfmodels.SourceReader) diag.Diagnostics {
	var diags diag.Diagnostics

	if configData.BackfillBatchSize.ValueInt64() > 50000 {
		diags.AddError("Invalid backfill batch size", "The maximum allowed value for `backfill_batch_size` is 50,000.")
	}

	if configData.IsShared.ValueBool() {
		if configData.Tables.IsNull() {
			diags.AddError("Invalid table configuration", "You must specify a `tables` block if `is_shared` is set to true.")
		} else if !configData.Tables.IsUnknown() && len(configData.Tables.Elements()) == 0 {
			diags.AddError("Invalid table configuration", "You must specify at least one table in the `tables` block if `is_shared` is set to true.")
		}
	} else {
		if !configData.Tables.IsNull() && !configData.Tables.IsUnknown() {
			diags.AddError("Invalid table configuration", "You should not specify a `tables` block if `is_shared` is set to false.")
		}
	}

	if !configData.Tables.IsNull() && !configData.Tables.IsUnknown() {
		tables := map[string]tfmodels.SourceReaderTable{}
		diags.Append(configData.Tables.ElementsAs(ctx, &tables, false)...)
		var usesIncludeColumns bool
		var usesExcludeColumns bool
		for tableKey, table := range tables {
			expectedKey := table.Name.ValueString()
			if table.Schema.ValueString() != "" {
				expectedKey = fmt.Sprintf("%s.%s", table.Schema.ValueString(), table.Name.ValueString())
			}
			if tableKey != expectedKey {
				diags.AddError("Table key mismatch", fmt.Sprintf("Table key %q should be %q instead.", tableKey, expectedKey))
			}
			if !table.ColumnsToInclude.IsNull() && len(table.ColumnsToInclude.Elements()) > 0 {
				usesIncludeColumns = true
			}
			if !table.ColumnsToExclude.IsNull() && len(table.ColumnsToExclude.Elements()) > 0 {
				usesExcludeColumns = true
			}
			if usesIncludeColumns && usesExcludeColumns {
				diags.AddError("Invalid table configuration", "You can only use one of `columns_to_include` and `columns_to_exclude` within a source reader.")
			}
		}
	}

	if configData.EnableUnifyAcrossDatabases.ValueBool() {
		if !slices.Contains([]string{"fn_dblog", "change_tracking"}, configData.MSSQLReplicationMethod.ValueString()) {
			diags.AddError("Invalid configuration", "`enable_unify_across_databases` is only applicable if `mssql_replication_method` is set to `fn_dblog` or `change_tracking`.")
		} else if configData.DatabasesToUnify.IsNull() || len(configData.DatabasesToUnify.Elements()) == 0 {
			diags.AddError("Invalid configuration", "You must specify `databases_to_unify` if `enable_unify_across_databases` is set to true.")
		} else if !configData.DatabaseName.IsUnknown() {
			databasesToUnify := []string{}
			diags.Append(configData.DatabasesToUnify.ElementsAs(ctx, &databasesToUnify, false)...)
			if !slices.Contains(databasesToUnify, configData.DatabaseName.ValueString()) {
				diags.AddError("Invalid configuration", "`databases_to_unify` should include the database you specified for `database_name`.")
			}
		}
	}

	return diags
}

func (r *SourceReaderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var configData tfmodels.SourceReader
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateSourceReaderConfig(ctx, configData)...)
}

func (r *SourceReaderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	baseSourceReader, diags := planData.ToAPIBaseModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if err := r.client.SourceReaders().Validate(ctx, baseSourceReader); err != nil {
		resp.Diagnostics.AddError("Unable to create Source Reader", err.Error())
		return
	}

	sourceReader, err := r.client.SourceReaders().Create(ctx, baseSourceReader)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, sourceReader)

	if sourceReader.IsShared {
		if err := r.client.SourceReaders().Deploy(ctx, sourceReader.UUID.String()); err != nil {
			resp.Diagnostics.AddError("Unable to deploy Source Reader", err.Error())
		}
	}
}

func (r *SourceReaderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	sourceReaderUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	sourceReader, err := r.client.SourceReaders().Get(ctx, sourceReaderUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, sourceReader)
}

func (r *SourceReaderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	apiBaseModel, diags := planData.ToAPIBaseModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if err := r.client.SourceReaders().Validate(ctx, apiBaseModel); err != nil {
		resp.Diagnostics.AddError("Unable to update Source Reader", err.Error())
		return
	}

	apiModel, diags := planData.ToAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	updatedSourceReader, err := r.client.SourceReaders().Update(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedSourceReader)

	if updatedSourceReader.IsShared {
		if err := r.client.SourceReaders().Deploy(ctx, updatedSourceReader.UUID.String()); err != nil {
			resp.Diagnostics.AddError("Unable to deploy Source Reader", err.Error())
		}
	}
}

func (r *SourceReaderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	sourceReaderUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.SourceReaders().Delete(ctx, sourceReaderUUID); err != nil {
		resp.Diagnostics.AddError("Unable to delete Source Reader", err.Error())
		return
	}
}

func (r *SourceReaderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
