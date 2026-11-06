package provider

import (
	"context"
	"errors"
	"fmt"
	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/openapi"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SnowflakeEcoScheduleResource{}
var _ resource.ResourceWithConfigure = &SnowflakeEcoScheduleResource{}
var _ resource.ResourceWithImportState = &SnowflakeEcoScheduleResource{}

func NewSnowflakeEcoScheduleResource() resource.Resource {
	return &SnowflakeEcoScheduleResource{}
}

type SnowflakeEcoScheduleResource struct {
	client *openapi.ClientWithResponses
}

func (r *SnowflakeEcoScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_eco_schedule"
}

func (r *SnowflakeEcoScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Snowflake Eco Mode schedule resource. A schedule sets a pipeline's peak hours; outside them (including all of any day not in `peak_days`), pipelines attached with `snowflake_eco_schedule_uuid` on `artie_pipeline` use the off-peak flush rules to save Snowflake credits.",
		Attributes: map[string]schema.Attribute{
			"uuid":     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":     schema.StringAttribute{Required: true, MarkdownDescription: "A human-readable label for this schedule."},
			"timezone": schema.StringAttribute{Required: true, MarkdownDescription: "The IANA time zone of the peak hours, like `America/New_York`."},
			"peak_days": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The days that have peak hours: `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday` or `Saturday`. Any other day is off-peak all day.",
			},
			"peak_start_time":                 schema.StringAttribute{Required: true, MarkdownDescription: "When peak hours start on each of `peak_days`, as 24-hour `HH:MM`, like `09:00`."},
			"peak_end_time":                   schema.StringAttribute{Required: true, MarkdownDescription: "When peak hours end on each of `peak_days`, as 24-hour `HH:MM`, like `18:00`. Must be later than `peak_start_time`; peak hours can't cross midnight."},
			"off_peak_flush_interval_seconds": schema.Int64Attribute{Required: true, MarkdownDescription: "How often, in seconds, pipelines flush outside peak hours (at least 10)."},
			"off_peak_buffer_rows":            schema.Int64Attribute{Required: true, MarkdownDescription: "How many rows pipelines buffer before flushing outside peak hours."},
			"off_peak_flush_size_kb":          schema.Int64Attribute{Required: true, MarkdownDescription: "How much data, in KB, pipelines buffer before flushing outside peak hours."},
		},
	}
}

func (r *SnowflakeEcoScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client, err := providerData.NewOpenAPIClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *SnowflakeEcoScheduleResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.SnowflakeEcoSchedule
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *SnowflakeEcoScheduleResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.SnowflakeEcoSchedule, bool) {
	var planData tfmodels.SnowflakeEcoSchedule
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *SnowflakeEcoScheduleResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, schedule openapi.PayloadsSnowflakeEcoSchedule) {
	// Translate API response type into Terraform model and save it into state
	diagnostics.Append(state.Set(ctx, tfmodels.SnowflakeEcoScheduleFromAPIModel(schedule))...)
}

func (r *SnowflakeEcoScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	schedule, err := artieclient.JSON200(r.client.SnowflakeEcoScheduleCreateWithResponse(ctx, planData.ToAPICreateRequest()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Snowflake Eco Schedule", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, *schedule)
}

func (r *SnowflakeEcoScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	scheduleUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	schedule, err := artieclient.JSON200(r.client.SnowflakeEcoScheduleDetailWithResponse(ctx, scheduleUUID))
	if errors.Is(err, artieclient.ErrNotFound) {
		resp.State.RemoveResource(ctx) // deleted outside Terraform, so the next apply recreates it
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Snowflake Eco Schedule", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, *schedule)
}

func (r *SnowflakeEcoScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	schedule, err := artieclient.JSON200(r.client.SnowflakeEcoScheduleUpdateWithResponse(ctx, planData.UUID.ValueString(), planData.ToAPIUpdateRequest()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Snowflake Eco Schedule", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, *schedule)
}

func (r *SnowflakeEcoScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	scheduleUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := artieclient.CheckResponse(r.client.SnowflakeEcoScheduleDeleteWithResponse(ctx, scheduleUUID)); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Snowflake Eco Schedule", err.Error())
	}
}

func (r *SnowflakeEcoScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
