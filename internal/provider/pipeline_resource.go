package provider

import (
	"context"
	"fmt"

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PipelineResource{}
var _ resource.ResourceWithConfigure = &PipelineResource{}
var _ resource.ResourceWithImportState = &PipelineResource{}

func NewPipelineResource() resource.Resource {
	return &PipelineResource{}
}

type PipelineResource struct {
	client artieclient.Client
}

func (r *PipelineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *PipelineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Pipeline resource. This represents a pipeline that syncs data from a single source (e.g., Postgres) to a single destination (e.g., Snowflake).",
		Attributes: map[string]schema.Attribute{
			"uuid":                        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":                        schema.StringAttribute{Required: true, MarkdownDescription: "The human-readable name of the pipeline. This is used only as a label and can contain any characters."},
			"status":                      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"source_reader_uuid":          schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This must point to an `artie_source_reader` resource."},
			"destination_connector_uuid":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "This must point to an `artie_connector` resource that represents the destination database."},
			"snowflake_eco_schedule_uuid": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tables": schema.MapNestedAttribute{
				Required:            true,
				MarkdownDescription: "A map of tables from the source database that you want to replicate to the destination. The key for each table should be formatted as `schema_name.table_name` if your source database uses schemas, otherwise just `table_name`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
						"name":                  schema.StringAttribute{Required: true, MarkdownDescription: "The name of the table in the source database."},
						"schema":                schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The name of the schema the table belongs to in the source database. This must be specified if your source database uses schemas (such as PostgreSQL), e.g. `public`."},
						"enable_history_mode":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, we will create an additional table in the destination (suffixed with `__history`) to store all changes to the source table over time."},
						"individual_deployment": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, we will spin up a separate Artie Transfer deployment to handle this table. This should only be used if this table has extremely high throughput (over 1M+ per hour) and has much higher throughput than other tables."},
						"is_partitioned":        schema.BoolAttribute{Computed: true, Default: booldefault.StaticBool(false), PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
						"alias":                 schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional alias for the table. If set, this will be the name of the destination table."},
						"columns_to_exclude":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional list of columns to exclude from syncing to the destination."},
						"columns_to_include":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional list of columns to include in replication. If not provided, all columns will be replicated. A pipeline can only have one of `columns_to_include` or `columns_to_exclude` set in any of its tables."},
						"columns_to_hash":       schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}, MarkdownDescription: "An optional list of columns to hash in the destination. Values for these columns will be obscured with a one-way hash."},
						"skip_deletes":          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, we will skip delete events for this table and only process insert and update events."},
						"merge_predicates": schema.ListNestedAttribute{
							Optional:            true,
							Computed:            true,
							PlanModifiers:       []planmodifier.List{listplanmodifier.UseStateForUnknown()},
							MarkdownDescription: "Optional: if the destination table is partitioned, specify the column(s) it's partitioned by. This will help with merge performance and currently only applies to Snowflake and BigQuery. For BigQuery, only one column can be specified and it must be a time column partitioned by day.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"partition_field": schema.StringAttribute{Required: true, MarkdownDescription: "The name of the column the destination table is partitioned by."},
								},
							}},
					},
				},
			},
			"flush_rules": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "This contains rules for how often Artie should flush data to the destination. If not specified, Artie will provide default values. A flush will happen when any of the rules are met (e.g. 30 seconds since the last flush OR 150k rows OR 50MB of data).",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"flush_interval_seconds": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The flush interval in seconds for how often Artie should flush data to the destination.",
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"buffer_rows": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The number of rows to buffer before flushing to the destination.",
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"flush_size_kb": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The size in kb of data to buffer before flushing to the destination.",
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
				},
			},
			"destination_config": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "This contains configuration that pertains to the destination database but is specific to this pipeline. The basic connection settings for the destination, which can be shared by multiple pipelines, are stored in the corresponding `artie_connector` resource.",
				Attributes: map[string]schema.Attribute{
					"database": schema.StringAttribute{
						MarkdownDescription: "The name of the database that data should be synced to in the destination. This should be filled if the destination is MS SQL or Snowflake, unless `use_same_schema_as_source` is set to true.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"schema": schema.StringAttribute{
						MarkdownDescription: "The name of the schema that data should be synced to in the destination. This should be filled if the destination is MS SQL, Redshift, or Snowflake (unless `use_same_schema_as_source` is set to true).",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"dataset": schema.StringAttribute{
						MarkdownDescription: "The name of the dataset that data should be synced to in the destination. This should be filled if the destination is BigQuery.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"use_same_schema_as_source": schema.BoolAttribute{
						MarkdownDescription: "If set to true, each table from the source database will be synced to a schema with the same name as its source schema. This can only be used if both the source and destination support multiple schemas (e.g. PostgreSQL, Redshift, Snowflake, etc).",
						Optional:            true,
						Computed:            true, Default: booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"schema_name_prefix": schema.StringAttribute{
						MarkdownDescription: "If `use_same_schema_as_source` is enabled, this prefix will be added to each schema name in the destination. This is useful if you want to namespace all of this pipeline's schemas in the destination.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"bucket": schema.StringAttribute{
						MarkdownDescription: "The name of the S3 bucket that data should be synced to. This should be filled if the destination is S3.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"folder": schema.StringAttribute{
						MarkdownDescription: "If provided, all files will be stored under this folder inside the S3 bucket. This is optional and only applies if the destination is S3.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString(""),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
			"data_plane_name": schema.StringAttribute{
				MarkdownDescription: "The name of the data plane to use for this pipeline. If this is not set, we will use the default data plane for your account. To see the full list of supported data planes on your account, click on 'New deployment' in our UI.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"drop_deleted_columns":               schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, when a column is dropped from the source it will also be dropped in the destination."},
			"soft_delete_rows":                   schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, a new boolean column called __artie_delete will be added to your destination to indicate if the row has been deleted."},
			"include_artie_updated_at_column":    schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will add a new column to your dataset called __artie_updated_at."},
			"include_database_updated_at_column": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will add a new column to your dataset called __artie_db_updated_at."},
		},
	}
}

func (r *PipelineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PipelineResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.Pipeline
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *PipelineResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.Pipeline, bool) {
	var planData tfmodels.Pipeline
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *PipelineResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiModel artieclient.Pipeline) {
	// Translate API response type into Terraform model and save it into state
	pipeline, diags := tfmodels.PipelineFromAPIModel(ctx, apiModel)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, pipeline)...)
}

func (r *PipelineResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var configData tfmodels.Pipeline
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if configData.Tables.IsNull() && !configData.Tables.IsUnknown() {
		tables := map[string]tfmodels.Table{}
		resp.Diagnostics.Append(configData.Tables.ElementsAs(ctx, &tables, false)...)
		for tableKey, table := range tables {
			if !table.UUID.IsNull() {
				resp.Diagnostics.AddError("Table.uuid is Read-Only", fmt.Sprintf("%q table should not have `uuid` specified. Please remove this attribute from your config.", tableKey))
			}
			if !table.IsPartitioned.IsNull() {
				resp.Diagnostics.AddError("Table.is_partitioned is Read-Only", fmt.Sprintf("%q table should not have `is_partitioned` specified. Please remove this attribute from your config.", tableKey))
			}
		}
	}
}

func (r *PipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	pipeline, diags := planData.ToAPIBaseModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate config before creating the pipeline
	if err := r.client.Pipelines().ValidateSource(ctx, pipeline); err != nil {
		resp.Diagnostics.AddError("Unable to Create Pipeline", err.Error())
		return
	}

	if err := r.client.Pipelines().ValidateDestination(ctx, pipeline); err != nil {
		resp.Diagnostics.AddError("Unable to Create Pipeline", err.Error())
		return
	}

	createdPipeline, err := r.client.Pipelines().Create(ctx, pipeline)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Pipeline", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, createdPipeline)
}

func (r *PipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	pipelineUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	pipeline, err := r.client.Pipelines().Get(ctx, pipelineUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Pipeline", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, pipeline)
}

func (r *PipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	basePipeline, diags := planData.ToAPIBaseModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate source & destination config before updating the pipeline
	if err := r.client.Pipelines().ValidateSource(ctx, basePipeline); err != nil {
		resp.Diagnostics.AddError("Unable to Update Pipeline", err.Error())
		return
	}
	if err := r.client.Pipelines().ValidateDestination(ctx, basePipeline); err != nil {
		resp.Diagnostics.AddError("Unable to Update Pipeline", err.Error())
		return
	}

	apiModel, diags := planData.ToAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedPipeline, err := r.client.Pipelines().Update(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Pipeline", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedPipeline)
}

func (r *PipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	pipelineUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.Pipelines().Delete(ctx, pipelineUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Pipeline", err.Error())
	}
}

func (r *PipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
