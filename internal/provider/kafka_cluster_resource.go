package provider

import (
	"context"
	"fmt"
	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/openapi"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaClusterResource{}
var _ resource.ResourceWithConfigure = &KafkaClusterResource{}
var _ resource.ResourceWithImportState = &KafkaClusterResource{}
var _ resource.ResourceWithModifyPlan = &KafkaClusterResource{}

var kafkaInstanceTypes = []string{
	string(openapi.PayloadsKafkaInstanceTypeM8gXlarge),
	string(openapi.PayloadsKafkaInstanceTypeM8g2xlarge),
	string(openapi.PayloadsKafkaInstanceTypeM8g4xlarge),
	string(openapi.PayloadsKafkaInstanceTypeM8iXlarge),
	string(openapi.PayloadsKafkaInstanceTypeM8i2xlarge),
	string(openapi.PayloadsKafkaInstanceTypeM8i4xlarge),
	string(openapi.PayloadsKafkaInstanceTypeR8gLarge),
	string(openapi.PayloadsKafkaInstanceTypeR8gXlarge),
	string(openapi.PayloadsKafkaInstanceTypeR8g2xlarge),
	string(openapi.PayloadsKafkaInstanceTypeR8iLarge),
	string(openapi.PayloadsKafkaInstanceTypeR8iXlarge),
	string(openapi.PayloadsKafkaInstanceTypeR8i2xlarge),
}

func NewKafkaClusterResource() resource.Resource {
	return &KafkaClusterResource{}
}

type KafkaClusterResource struct {
	client *openapi.ClientWithResponses
}

func (r *KafkaClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_cluster"
}

func (r *KafkaClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Managed Kafka cluster resource. This represents a Kafka cluster that Artie runs for your company. Creating one starts a deploy that finishes in the background; `status` is `deploying` until the cluster is `active`. Destroying one deletes the cluster, and its data can be restored from the Artie UI until it is deleted for good.",
		Attributes: map[string]schema.Attribute{
			"uuid":           schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The unique identifier for this Kafka cluster."},
			"name":           schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(3, 100)}, MarkdownDescription: "The cluster's name, as shown in the Artie UI. 3 to 100 characters. Renaming doesn't change the bootstrap address."},
			"region":         schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "The AWS region to run the cluster in (e.g. `us-east-1`). Changing this forces a new cluster."},
			"instance_type":  schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf(kafkaInstanceTypes...)}, MarkdownDescription: "The EC2 instance type for every broker. Changing it restarts the brokers one at a time onto the new machines."},
			"brokers":        schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(3)}, MarkdownDescription: "The number of brokers, at least 3. Brokers can only be added. Existing partitions stay on the brokers they're on; new topics use every broker."},
			"storage_gib":    schema.Int64Attribute{Required: true, MarkdownDescription: "Disk per broker in GiB: 20, 50 or 80, or any size from 100 to 1024 (1 TiB). Storage can only grow, to 50 GiB or more, at most 3 times in 24 hours."},
			"allowed_cidrs":  schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true, Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})), MarkdownDescription: "Client IPv4 ranges allowed to reach the cluster. Empty (the default) allows any IP; clients still need the password."},
			"tiered_storage": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether Artie turns tiered storage on for every topic. Turning it off pauses every topic tiered through Artie. Defaults to `true`."},
			"local_retention_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Validators:          []validator.Int64{int64validator.Between(86400000, 1209600000)},
				MarkdownDescription: "How long tiered topics keep data on broker disks, in ms, from 86400000 (one day) to 1209600000 (14 days). If not set, Artie uses 3 days.",
			},
			"local_retention_bytes": schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1073741824)}, MarkdownDescription: "The most broker disk each partition of a tiered topic keeps, in bytes, at least 1073741824 (1 GiB). If not set, there is no cap."},
			"storage_autoscale": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "How Artie grows the brokers' disks before they fill. If not set, Artie turns it on, growing at 65% full to 55% full, with no limit.",
				Attributes: map[string]schema.Attribute{
					"enabled":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether Artie grows the brokers' disks before they fill. Defaults to `true`."},
					"bump_at_percent": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(65), Validators: []validator.Int64{int64validator.Between(1, 99)}, MarkdownDescription: "How full, in %, a broker's disk gets before every broker's disk grows. At most 99. Defaults to 65."},
					"target_percent":  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(55), Validators: []validator.Int64{int64validator.Between(1, 98)}, MarkdownDescription: "How full, in %, the fullest broker's disk is after a grow. Must be less than `bump_at_percent`. Defaults to 55."},
					"max_gib":         schema.Int64Attribute{Optional: true, MarkdownDescription: "The most disk per broker, in GiB, a grow goes to. If not set, there is no limit."},
				},
			},
			"cloud":                    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The cloud the cluster runs in."},
			"status":                   schema.StringAttribute{Computed: true, MarkdownDescription: "The cluster's status: `deploying`, `active`, `failed`, `draft` or `deleted`."},
			"external_bootstrap_url":   schema.StringAttribute{Computed: true, MarkdownDescription: "The bootstrap address Kafka clients connect to."},
			"kafka_version":            schema.StringAttribute{Computed: true, MarkdownDescription: "The Kafka version the brokers run."},
			"tiered_storage_available": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether tiered storage is available on this cluster. Artie turns it on once the brokers support it."},
		},
	}
}

func (r *KafkaClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan rejects removing brokers or shrinking storage at plan time, since the API can only grow either.
func (r *KafkaClusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return // creating or destroying
	}
	for _, attr := range []string{"brokers", "storage_gib"} {
		var planned, prior types.Int64
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(attr), &planned)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(attr), &prior)...)
		if shrinks(prior, planned) {
			resp.Diagnostics.AddAttributeError(path.Root(attr), "Kafka cluster can't shrink",
				fmt.Sprintf("`%s` can only grow (from %d to %d was planned).", attr, prior.ValueInt64(), planned.ValueInt64()))
		}
	}
}

func shrinks(prior, planned types.Int64) bool {
	return tfmodels.IsKnown(prior) && tfmodels.IsKnown(planned) && planned.ValueInt64() < prior.ValueInt64()
}

func (r *KafkaClusterResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.KafkaCluster
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *KafkaClusterResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.KafkaCluster, bool) {
	var planData tfmodels.KafkaCluster
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *KafkaClusterResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, cluster openapi.PayloadsKafkaCluster) {
	tfModel, diags := tfmodels.KafkaClusterFromAPIModel(ctx, cluster)
	diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diagnostics.Append(state.Set(ctx, tfModel)...)
}

// readIntoState fetches the cluster with Get a Kafka cluster, the only endpoint that returns its managed configuration.
func (r *KafkaClusterResource) readIntoState(ctx context.Context, uuid string, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	cluster, err := artieclient.JSON200(r.client.KafkaClusterDetailWithResponse(ctx, uuid))
	if err != nil {
		diagnostics.AddError("Unable to read Kafka cluster", err.Error())
		return
	}

	if cluster.Status == openapi.EnumsKafkaStatusDeleted {
		state.RemoveResource(ctx)
		return
	}

	r.SetStateData(ctx, state, diagnostics, *cluster)
}

func (r *KafkaClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	body, diags := planData.ToAPICreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	cluster, err := artieclient.JSON200(r.client.KafkaClusterCreateWithResponse(ctx, body))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Kafka cluster", err.Error())
		return
	}

	r.readIntoState(ctx, cluster.Uuid.String(), &resp.State, &resp.Diagnostics)
}

func (r *KafkaClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	uuid, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	r.readIntoState(ctx, uuid, &resp.State, &resp.Diagnostics)
}

func (r *KafkaClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	var stateData tfmodels.KafkaCluster
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := planData.ToAPIUpdateRequest(ctx, stateData)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	uuid := stateData.UUID.ValueString()
	if body != (openapi.RouterUpdateKafkaClusterRequest{}) {
		if _, err := artieclient.JSON200(r.client.KafkaClusterUpdateWithResponse(ctx, uuid, body)); err != nil {
			resp.Diagnostics.AddError("Unable to update Kafka cluster", err.Error())
			return
		}
	}

	r.readIntoState(ctx, uuid, &resp.State, &resp.Diagnostics)
}

func (r *KafkaClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	uuid, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := artieclient.CheckResponse(r.client.KafkaClusterDeleteWithResponse(ctx, uuid)); err != nil {
		resp.Diagnostics.AddError("Unable to delete Kafka cluster", err.Error())
		return
	}
}

func (r *KafkaClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
