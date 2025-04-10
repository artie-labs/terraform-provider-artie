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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
				MarkdownDescription: "The name of the data plane to deploy this source reader in. If this is not set, we will use the default data plane for your account. To see the full list of supported data planes on your account, click on 'New deployment' in our UI.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_shared":                          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, this source reader can be used by multiple pipelines."},
			"database_name":                      schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The name of the database we should read data from in the source connector. This should be specified if the source connector's type is DocumentDB, MongoDB, MySQL, MS SQL, Oracle (this maps to the service name), or PostgreSQL."},
			"oracle_container_name":              schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The name of the container (pluggable database) if the source type is Oracle and you are using a container database."},
			"one_topic_per_schema":               schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set to true, Artie will write all incoming CDC events into a single Kafka topic per schema. This is currently only supported if your source is Oracle and your account has this feature enabled."},
			"postgres_publication_name_override": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set, this will override the name of the PostgreSQL publication. Otherwise, we will use our default value, `dbz_publication`. This is only applicable if the source type is PostgreSQL."},
			"postgres_replication_slot_override": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "If set, this will override the name of the PostgreSQL replication slot. Otherwise, we will use our default value, `artie`. This is only applicable if the source type is PostgreSQL."},
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
	diagnostics.Append(state.Set(ctx, tfmodels.SourceReaderFromAPIModel(apiSourceReader))...)
}

func (r *SourceReaderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	baseSourceReader, diags := planData.ToAPIBaseModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	sourceReader, err := r.client.SourceReaders().Create(ctx, baseSourceReader)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, sourceReader)
}

func (r *SourceReaderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	sourceReaderUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	sourceReader, err := r.client.SourceReaders().Get(ctx, sourceReaderUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, sourceReader)
}

func (r *SourceReaderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	apiModel, diags := planData.ToAPIModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	updatedSourceReader, err := r.client.SourceReaders().Update(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Source Reader", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, updatedSourceReader)
}

func (r *SourceReaderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	sourceReaderUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.SourceReaders().Delete(ctx, sourceReaderUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete Source Reader", err.Error())
		return
	}
}

func (r *SourceReaderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
