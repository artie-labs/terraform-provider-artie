package provider

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaClusterResource{}
var _ resource.ResourceWithConfigure = &KafkaClusterResource{}
var _ resource.ResourceWithImportState = &KafkaClusterResource{}
var _ resource.ResourceWithModifyPlan = &KafkaClusterResource{}
var _ resource.ResourceWithValidateConfig = &KafkaClusterResource{}

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

const (
	defaultKafkaBumpAtPercent = 65
	defaultKafkaTargetPercent = 55
)

var defaultKafkaStorageAutoscale = types.ObjectValueMust(tfmodels.KafkaClusterStorageAutoscaleAttrTypes, map[string]attr.Value{
	"enabled":         types.BoolValue(true),
	"bump_at_percent": types.Int64Value(defaultKafkaBumpAtPercent),
	"target_percent":  types.Int64Value(defaultKafkaTargetPercent),
	"max_gib":         types.Int64Null(),
})

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
			"uuid":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The unique identifier for this Kafka cluster."},
			"name":          schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.LengthBetween(3, 100)}, MarkdownDescription: "The cluster's name, as shown in the Artie UI. 3 to 100 characters. Renaming doesn't change the bootstrap address."},
			"region":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "The AWS region to run the cluster in (e.g. `us-east-1`). Changing this forces a new cluster."},
			"instance_type": schema.StringAttribute{Required: true, Validators: []validator.String{stringvalidator.OneOf(kafkaInstanceTypes...)}, MarkdownDescription: "The EC2 instance type for every broker. Changing it restarts the brokers one at a time onto the new machines."},
			"brokers":       schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(3)}, MarkdownDescription: "The number of brokers, at least 3. Brokers can only be added. Existing partitions stay on the brokers they're on; new topics use every broker."},
			"storage_gib":   schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.Any(int64validator.OneOf(20, 50, 80), int64validator.Between(100, 1024))}, MarkdownDescription: "Disk per broker in GiB: 20, 50 or 80, or any size from 100 to 1024 (1 TiB). Storage can only grow, to 50 GiB or more, at most 3 times in 24 hours. Storage autoscaling can grow the disks past this size without a diff; `current_storage_gib` has the size they have."},
			"allowed_cidrs": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true, Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})), MarkdownDescription: "Client IPv4 ranges allowed to reach the cluster. Empty (the default) allows any IP; clients still need the password."},
			"tiered_storage": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Object{kafkaTieredStoragePlanModifier{}},
				MarkdownDescription: "How the cluster's topics tier data to remote storage. If not set, Artie turns it on for every topic, with no disk cap per partition.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether Artie turns tiered storage on for every topic. Turning it off pauses every topic tiered through Artie. Defaults to `true`."},
					"local_retention_ms": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						Validators:          []validator.Int64{int64validator.Between(86400000, 2592000000)},
						MarkdownDescription: "How long tiered topics keep data on broker disks, in ms, from 86400000 (one day) to 2592000000 (30 days). Can only be set while `enabled` is `true`. If not set, Artie uses 3 days.",
					},
					"local_retention_bytes": schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1073741824)}, MarkdownDescription: "The most broker disk each partition of a tiered topic keeps, in bytes, at least 1073741824 (1 GiB). Can only be set while `enabled` is `true`. If not set, there is no cap."},
				},
			},
			"storage_autoscale": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				// A default rather than UseStateForUnknown, so removing the block resets the cluster to these settings.
				Default:             objectdefault.StaticValue(defaultKafkaStorageAutoscale),
				MarkdownDescription: "How Artie grows the brokers' disks before they fill. If not set, Artie turns it on, growing at 65% full to 55% full, with no limit.",
				Attributes: map[string]schema.Attribute{
					"enabled":         schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether Artie grows the brokers' disks before they fill. Defaults to `true`."},
					"bump_at_percent": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(defaultKafkaBumpAtPercent), Validators: []validator.Int64{int64validator.Between(1, 99)}, MarkdownDescription: "How full, in %, a broker's disk gets before every broker's disk grows. At most 99. Defaults to 65."},
					"target_percent":  schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(defaultKafkaTargetPercent), Validators: []validator.Int64{int64validator.Between(1, 98)}, MarkdownDescription: "How full, in %, the fullest broker's disk is after a grow. Must be less than `bump_at_percent`. Defaults to 55."},
					"max_gib":         schema.Int64Attribute{Optional: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "The most disk per broker, in GiB, a grow goes to. If not set, there is no limit."},
				},
			},
			"cloud":                    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The cloud the cluster runs in."},
			"status":                   schema.StringAttribute{Computed: true, MarkdownDescription: "The cluster's status: `deploying`, `active`, `failed` or `draft`. A deleted cluster is removed from state."},
			"external_bootstrap_url":   schema.StringAttribute{Computed: true, MarkdownDescription: "The bootstrap address Kafka clients connect to."},
			"current_storage_gib":      schema.Int64Attribute{Computed: true, MarkdownDescription: "The disk per broker the brokers have, in GiB. Can be more than `storage_gib` once storage autoscaling grows the disks."},
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

// ValidateConfig checks that the days on disk and disk cap are only set with tiered storage on, and that a grow ends
// below the fullness that starts it, filling in the defaults for unset values.
func (r *KafkaClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var tieredStorage tfmodels.KafkaClusterTieredStorage
	if getKnownObject(ctx, req.Config, path.Root("tiered_storage"), &tieredStorage, &resp.Diagnostics) {
		resp.Diagnostics.Append(validateKafkaTieredStorage(tieredStorage)...)
	}

	var autoscale tfmodels.KafkaClusterStorageAutoscale
	if getKnownObject(ctx, req.Config, path.Root("storage_autoscale"), &autoscale, &resp.Diagnostics) {
		resp.Diagnostics.Append(validateKafkaStorageAutoscale(autoscale)...)
	}
}

// getKnownObject reads the object at path into target, returning false if it's null or unknown or can't be read.
func getKnownObject(ctx context.Context, config tfsdk.Config, attrPath path.Path, target any, diagnostics *diag.Diagnostics) bool {
	var object types.Object
	diags := config.GetAttribute(ctx, attrPath, &object)
	if !diags.HasError() && tfmodels.IsKnown(object) {
		diags.Append(object.As(ctx, target, basetypes.ObjectAsOptions{})...)
	}
	diagnostics.Append(diags...)
	return !diags.HasError() && tfmodels.IsKnown(object)
}

// validateKafkaTieredStorage rejects a days on disk or disk cap with tiered storage off, since they only apply to
// tiered topics.
func validateKafkaTieredStorage(settings tfmodels.KafkaClusterTieredStorage) diag.Diagnostics {
	var diags diag.Diagnostics
	if !tfmodels.IsExplicitlyFalse(settings.Enabled) {
		return diags
	}
	for name, value := range map[string]types.Int64{"local_retention_ms": settings.LocalRetentionMS, "local_retention_bytes": settings.LocalRetentionBytes} {
		if !value.IsNull() {
			diags.AddAttributeError(path.Root("tiered_storage").AtName(name), "Invalid tiered storage settings",
				fmt.Sprintf("`%s` only applies to tiered topics, so it can only be set while `enabled` is `true`.", name))
		}
	}
	return diags
}

// kafkaTieredStoragePlanModifier plans a tiered_storage left out of the config the same as `tiered_storage = {}`: on,
// with no disk cap, keeping the days on disk from state. So removing the block turns tiered storage back on.
type kafkaTieredStoragePlanModifier struct{}

func (kafkaTieredStoragePlanModifier) Description(context.Context) string {
	return "Plans an unset tiered_storage as on, with no disk cap, keeping the prior days on disk."
}

func (m kafkaTieredStoragePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (kafkaTieredStoragePlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	var localRetentionMS attr.Value = types.Int64Unknown()
	if tfmodels.IsKnown(req.StateValue) {
		if prior := req.StateValue.Attributes()["local_retention_ms"]; tfmodels.IsKnown(prior) {
			localRetentionMS = prior
		}
	}
	resp.PlanValue = types.ObjectValueMust(tfmodels.KafkaClusterTieredStorageAttrTypes, map[string]attr.Value{
		"enabled":               types.BoolValue(true),
		"local_retention_ms":    localRetentionMS,
		"local_retention_bytes": types.Int64Null(),
	})
}

func validateKafkaStorageAutoscale(settings tfmodels.KafkaClusterStorageAutoscale) diag.Diagnostics {
	var diags diag.Diagnostics
	bumpAt, target := settings.BumpAtPercent, settings.TargetPercent
	if bumpAt.IsNull() {
		bumpAt = types.Int64Value(defaultKafkaBumpAtPercent)
	}
	if target.IsNull() {
		target = types.Int64Value(defaultKafkaTargetPercent)
	}
	if bumpAt.IsUnknown() || target.IsUnknown() {
		return diags
	}
	if target.ValueInt64() >= bumpAt.ValueInt64() {
		diags.AddAttributeError(path.Root("storage_autoscale").AtName("target_percent"), "Invalid storage autoscale settings",
			fmt.Sprintf("`target_percent` (%d) must be less than `bump_at_percent` (%d).", target.ValueInt64(), bumpAt.ValueInt64()))
	}
	return diags
}

// ModifyPlan rejects removing brokers or shrinking storage at plan time, since the API can only grow either.
func (r *KafkaClusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return // creating or destroying
	}
	var plannedRegion, priorRegion types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("region"), &plannedRegion)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("region"), &priorRegion)...)
	if !plannedRegion.Equal(priorRegion) {
		return // a new region replaces the cluster, and a new cluster can start at any size
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

func (r *KafkaClusterResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, cluster openapi.PayloadsKafkaCluster, requestedStorageGiB types.Int64) {
	tfModel, diags := tfmodels.KafkaClusterFromAPIModel(ctx, cluster)
	diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diagnostics.Append(state.Set(ctx, tfModel.WithRequestedStorage(requestedStorageGiB))...)
}

// readIntoState fetches the cluster with Get a Kafka cluster, the only endpoint that returns its managed configuration.
// requestedStorageGiB is the storage_gib from the config or prior state, null on import.
func (r *KafkaClusterResource) readIntoState(ctx context.Context, uuid string, requestedStorageGiB types.Int64, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	cluster, err := artieclient.JSON200(r.client.KafkaClusterDetailWithResponse(ctx, uuid))
	if errors.Is(err, artieclient.ErrNotFound) {
		state.RemoveResource(ctx) // deleted for good
		return
	} else if err != nil {
		diagnostics.AddError("Unable to read Kafka cluster", err.Error())
		return
	}

	if cluster.Status == openapi.EnumsKafkaStatusDeleted {
		state.RemoveResource(ctx)
		return
	}

	r.SetStateData(ctx, state, diagnostics, *cluster, requestedStorageGiB)
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

	// If the follow-up read fails or doesn't find the cluster yet, still record the cluster so a later apply doesn't
	// create a second one.
	var readDiags diag.Diagnostics
	r.readIntoState(ctx, cluster.Uuid.String(), planData.StorageGiB, &resp.State, &readDiags)
	if !readDiags.HasError() && !resp.State.Raw.IsNull() {
		resp.Diagnostics.Append(readDiags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, planData.WithCreatedCluster(*cluster))...)
	if !readDiags.HasError() {
		resp.Diagnostics.AddWarning("Kafka cluster created, but reading it back failed",
			"Artie didn't return the new cluster yet. The next refresh fills in the rest of its attributes.")
	}
	for _, d := range readDiags.Errors() {
		resp.Diagnostics.AddWarning("Kafka cluster created, but reading it back failed",
			fmt.Sprintf("%s: %s. The next refresh fills in the rest of its attributes.", d.Summary(), d.Detail()))
	}
}

func (r *KafkaClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfmodels.KafkaCluster
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIntoState(ctx, stateData.UUID.ValueString(), stateData.StorageGiB, &resp.State, &resp.Diagnostics)
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

	r.readIntoState(ctx, uuid, planData.StorageGiB, &resp.State, &resp.Diagnostics)
}

func (r *KafkaClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	uuid, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := artieclient.CheckResponse(r.client.KafkaClusterDeleteWithResponse(ctx, uuid)); err != nil && !errors.Is(err, artieclient.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete Kafka cluster", err.Error())
		return
	}
}

func (r *KafkaClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
