package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
	"terraform-provider-artie/internal/provider/tfmodels"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaTopicResource{}
var _ resource.ResourceWithConfigure = &KafkaTopicResource{}
var _ resource.ResourceWithImportState = &KafkaTopicResource{}
var _ resource.ResourceWithModifyPlan = &KafkaTopicResource{}

const (
	kafkaTopicMinRetentionMS     = 86400000   // one day
	kafkaTopicMaxTieredRetention = 1209600000 // 14 days
	kafkaTopicRetentionForever   = -1         // keep messages forever

	kafkaClusterDeployTimeout = time.Hour
)

// kafkaClusterDeployPollInterval is how often Create retries while the topic's cluster is deploying (a var for tests).
var kafkaClusterDeployPollInterval = 30 * time.Second

var kafkaTopicTieredStorageStates = []string{
	string(openapi.EnumsKafkaTopicTieredStorageStateOff),
	string(openapi.EnumsKafkaTopicTieredStorageStateOn),
	string(openapi.EnumsKafkaTopicTieredStorageStatePaused),
}

// kafkaTopicReservedConfigKeys are the topic configs this resource manages through its own attributes.
var kafkaTopicReservedConfigKeys = []string{"retention.ms", "local.retention.ms", "remote.storage.enable", "remote.log.copy.disable"}

func NewKafkaTopicResource() resource.Resource {
	return &KafkaTopicResource{}
}

type KafkaTopicResource struct {
	client *openapi.ClientWithResponses
}

func (r *KafkaTopicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_topic"
}

func (r *KafkaTopicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Kafka topic resource. This represents a topic on a Kafka cluster that Artie runs, such as an `artie_kafka_cluster`. Topics that Artie manages itself (CDC topics and Kafka-internal topics) can't be managed with this resource.",
		Attributes: map[string]schema.Attribute{
			"cluster_uuid": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "The UUID of the Kafka cluster the topic is on. Changing this forces a new topic."},
			"topic": schema.StringAttribute{
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 249), stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`), "must contain only letters, digits, `.`, `_` and `-`"), stringvalidator.NoneOf(".", "..")},
				MarkdownDescription: "The topic's name: letters, digits, `.`, `_` and `-`, at most 249 characters, and not `.` or `..`. Changing this forces a new topic.",
			},
			"partition_count": schema.Int64Attribute{Required: true, Validators: []validator.Int64{int64validator.AtLeast(1)}, MarkdownDescription: "The number of partitions. Partitions can only be added. Adding them changes which partition each key goes to."},
			"replication_factor": schema.Int64Attribute{
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplaceIf(priorIsSetInt64, "Changing the replication factor forces a new topic.", "Changing the replication factor forces a new topic.")},
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
				MarkdownDescription: "How many copies of each partition the brokers keep. At most the cluster's number of brokers. Changing this forces a new topic. Only Managed Kafka clusters return it, so a topic that's imported from another cluster keeps the replication factor it has, and the first value in the config is taken as is.",
			},
			"config_entries": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers:       []planmodifier.Map{mapplanmodifier.RequiresReplaceIf(configEntriesRequireReplace, "Changing the config entries forces a new topic.", "Changing the config entries forces a new topic.")},
				Validators:          []validator.Map{mapvalidator.KeysAre(stringvalidator.NoneOf(kafkaTopicReservedConfigKeys...))},
				MarkdownDescription: "Kafka topic configs to create the topic with (e.g. `cleanup.policy`). The API can't change them afterwards, so changing this forces a new topic. Use `retention_ms`, `local_retention_ms` and `tiered_storage` instead of `retention.ms`, `local.retention.ms`, `remote.storage.enable` and `remote.log.copy.disable`. Not read back from the API, so a topic that's imported keeps the config entries it has.",
			},
			"retention_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Validators:          []validator.Int64{int64validator.Any(int64validator.OneOf(kafkaTopicRetentionForever), int64validator.AtLeast(kafkaTopicMinRetentionMS))},
				MarkdownDescription: "How long the topic keeps messages (`retention.ms`), at least 86400000 (one day), or `-1` to keep messages forever. At most 1209600000 (14 days) while tiered storage is `on` or `paused`, so `-1` needs tiered storage `off`. If not set when the topic is created, the topic keeps the cluster's default. Removing it from the config later keeps the topic's current retention.",
			},
			"local_retention_ms": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				Validators:          []validator.Int64{int64validator.AtLeast(kafkaTopicMinRetentionMS)},
				MarkdownDescription: "How long a tiered topic keeps data on broker disks (`local.retention.ms`), from 86400000 (one day) up to `retention_ms`. If not set when the topic is created, the topic uses the cluster's days on disk. Removing it from the config later keeps the topic's current value. `-2`, which can only be read back, means the same as `retention_ms`.",
			},
			"tiered_storage": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:          []validator.String{stringvalidator.OneOf(kafkaTopicTieredStorageStates...)},
				MarkdownDescription: "Whether the topic keeps older data in remote storage: `off`, `on` or `paused`. A topic that's `on` can only be paused, and only a `paused` topic can be turned `off`, which deletes its data in remote storage. Pausing keeps that data readable until retention expires it. If not set when the topic is created, the topic follows the cluster's `tiered_storage`. Removing it from the config later keeps the topic's current state.",
			},
		},
	}
}

// kafkaTopicImportedKey marks a topic in private state as imported, so its null config entries are the ones the API
// doesn't return, not ones it was created without.
const kafkaTopicImportedKey = "imported"

// priorIsSetInt64 doesn't replace a topic whose prior replication factor is null, which is how an imported topic
// starts out outside Managed Kafka clusters.
func priorIsSetInt64(ctx context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
	resp.RequiresReplace = !req.StateValue.IsNull()
}

func configEntriesRequireReplace(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
	imported, diags := isImportedTopic(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	resp.RequiresReplace = configEntriesChangeReplaces(req.StateValue, req.PlanValue, imported)
}

// configEntriesChangeReplaces reports whether changing the config entries replaces the topic. An imported topic's
// first config entries are adopted in place, since the API doesn't return the ones it has.
func configEntriesChangeReplaces(prior, planned types.Map, imported bool) bool {
	return !prior.Equal(planned) && (!prior.IsNull() || !imported)
}

func isImportedTopic(ctx context.Context, private interface {
	GetKey(context.Context, string) ([]byte, diag.Diagnostics)
}) (bool, diag.Diagnostics) {
	value, diags := private.GetKey(ctx, kafkaTopicImportedKey)
	return value != nil, diags
}

func (r *KafkaTopicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan rejects changes the API refuses at plan time: removing partitions, turning tiered storage off without
// pausing it first, and keeping more than 14 days while tiered storage is on or paused.
func (r *KafkaTopicResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // destroying
	}

	var plan tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if msg := tieredRetentionError(plan); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("retention_ms"), "Kafka topic retention too long for tiered storage", msg)
	}
	if msg := localRetentionError(plan); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("local_retention_ms"), "Kafka topic local retention longer than retention", msg)
	}

	if req.State.Raw.IsNull() {
		return // creating
	}
	var state tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	imported, diags := isImportedTopic(ctx, req.Private)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if replacesTopic(state, plan, imported) {
		return // a new topic can start with any partitions and tiered storage
	}
	if shrinks(state.PartitionCount, plan.PartitionCount) {
		resp.Diagnostics.AddAttributeError(path.Root("partition_count"), "Kafka topic partitions can't be removed",
			fmt.Sprintf("`partition_count` can only grow (from %d to %d was planned). To use fewer partitions, create a new topic.", state.PartitionCount.ValueInt64(), plan.PartitionCount.ValueInt64()))
	}
	if msg := tieredStorageTransitionError(state.TieredStorage, plan.TieredStorage); msg != "" {
		resp.Diagnostics.AddAttributeError(path.Root("tiered_storage"), "Invalid Kafka topic tiered storage change", msg)
	}
}

// replacesTopic mirrors the attributes' RequiresReplace plan modifiers, which ModifyPlan can't see.
func replacesTopic(state, plan tfmodels.KafkaTopic, imported bool) bool {
	return !plan.ClusterUUID.Equal(state.ClusterUUID) || !plan.Topic.Equal(state.Topic) ||
		(!state.ReplicationFactor.IsNull() && !plan.ReplicationFactor.Equal(state.ReplicationFactor)) ||
		configEntriesChangeReplaces(state.ConfigEntries, plan.ConfigEntries, imported)
}

func tieredRetentionError(plan tfmodels.KafkaTopic) string {
	if !tfmodels.IsKnown(plan.RetentionMS) || !tfmodels.IsKnown(plan.TieredStorage) {
		return ""
	}
	retentionMS := plan.RetentionMS.ValueInt64()
	// -1 keeps messages forever, which is longer than the tiered cap.
	if plan.TieredStorage.ValueString() == string(openapi.EnumsKafkaTopicTieredStorageStateOff) || (retentionMS != kafkaTopicRetentionForever && retentionMS <= kafkaTopicMaxTieredRetention) {
		return ""
	}
	return fmt.Sprintf("`retention_ms` can be at most %d (14 days) while `tiered_storage` is `%s` (%d was planned).", kafkaTopicMaxTieredRetention, plan.TieredStorage.ValueString(), plan.RetentionMS.ValueInt64())
}

// localRetentionError rejects keeping data on broker disks longer than the topic keeps it at all. A negative value is
// one read back from the API (`-2` for local retention), which isn't checked.
func localRetentionError(plan tfmodels.KafkaTopic) string {
	if !tfmodels.IsKnown(plan.RetentionMS) || !tfmodels.IsKnown(plan.LocalRetentionMS) || plan.TieredStorage.ValueString() == string(openapi.EnumsKafkaTopicTieredStorageStateOff) {
		return ""
	}
	retentionMS, localRetentionMS := plan.RetentionMS.ValueInt64(), plan.LocalRetentionMS.ValueInt64()
	if retentionMS < 0 || localRetentionMS < 0 || localRetentionMS <= retentionMS {
		return ""
	}
	return fmt.Sprintf("`local_retention_ms` can be at most `retention_ms` (%d was planned, with a retention of %d).", localRetentionMS, retentionMS)
}

func tieredStorageTransitionError(prior, planned types.String) string {
	if !tfmodels.IsKnown(prior) || !tfmodels.IsKnown(planned) {
		return ""
	}
	if prior.ValueString() == string(openapi.EnumsKafkaTopicTieredStorageStateOn) && planned.ValueString() == string(openapi.EnumsKafkaTopicTieredStorageStateOff) {
		return "Tiered storage that's `on` can't be turned `off` directly. Set `tiered_storage` to `paused` and apply, then set it to `off`. Turning it off deletes the topic's data in remote storage."
	}
	return ""
}

// fetchTopic returns nil without an error when the topic doesn't exist.
func (r *KafkaTopicResource) fetchTopic(ctx context.Context, clusterUUID, topic string) (*openapi.PayloadsKafkaTopicDetail, error) {
	resp, err := r.client.KafkaTopicDetailWithResponse(ctx, clusterUUID, topic, &openapi.KafkaTopicDetailParams{ForceRefresh: lib.ToPtr(true)})
	if err == nil && resp.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	return artieclient.JSON200(resp, err)
}

// readIntoState fetches the topic and stores it, keeping the settings the API doesn't return from prior.
func (r *KafkaTopicResource) readIntoState(ctx context.Context, prior tfmodels.KafkaTopic, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	clusterUUID := prior.ClusterUUID.ValueString()
	detail, err := r.fetchTopic(ctx, clusterUUID, prior.Topic.ValueString())
	if err != nil {
		diagnostics.AddError("Unable to read Kafka topic", err.Error())
		return
	}
	if detail == nil {
		state.RemoveResource(ctx)
		return
	}

	tfModel, diags := tfmodels.KafkaTopicFromAPIModel(clusterUUID, *detail, prior)
	diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diagnostics.Append(state.Set(ctx, tfModel)...)
}

func (r *KafkaTopicResource) updateSettings(ctx context.Context, plan tfmodels.KafkaTopic, body openapi.RouterUpdateKafkaTopicSettingsRequest, diagnostics *diag.Diagnostics) {
	if body == (openapi.RouterUpdateKafkaTopicSettingsRequest{}) {
		return
	}
	if _, err := artieclient.JSON200(r.client.KafkaTopicSettingsUpdateWithResponse(ctx, plan.ClusterUUID.ValueString(), plan.Topic.ValueString(), body)); err != nil {
		diagnostics.AddError("Unable to update Kafka topic settings", err.Error())
	}
}

func (r *KafkaTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := planData.ToAPICreateRequest(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	clusterUUID := planData.ClusterUUID.ValueString()
	if err := r.createTopic(ctx, clusterUUID, body); err != nil {
		resp.Diagnostics.AddError("Unable to create Kafka topic", err.Error())
		return
	}

	// The topic exists from here on, so it's kept in state even if a later step fails: Terraform then marks it tainted
	// and replaces it, instead of trying to create it again.
	r.setCreatedTopicSettings(ctx, planData, &resp.Diagnostics)
	r.readIntoState(ctx, planData, &resp.State, &resp.Diagnostics)
	if resp.State.Raw.IsNull() {
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.AddError("Unable to read Kafka topic", fmt.Sprintf("topic %q not found after creating it", planData.Topic.ValueString()))
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), planData.ClusterUUID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("topic"), planData.Topic)...)
	}
}

// setCreatedTopicSettings sets retention and tiered storage on a new topic. The create request only takes config
// entries, so they're set afterwards, through the same endpoint (and validation) as updates.
func (r *KafkaTopicResource) setCreatedTopicSettings(ctx context.Context, planData tfmodels.KafkaTopic, diagnostics *diag.Diagnostics) {
	if !tfmodels.IsKnown(planData.RetentionMS) && !tfmodels.IsKnown(planData.LocalRetentionMS) && !tfmodels.IsKnown(planData.TieredStorage) {
		return
	}

	clusterUUID := planData.ClusterUUID.ValueString()
	detail, err := r.fetchTopic(ctx, clusterUUID, planData.Topic.ValueString())
	if err == nil && detail == nil {
		err = fmt.Errorf("topic %q not found after creating it", planData.Topic.ValueString())
	}
	if err != nil {
		diagnostics.AddError("Unable to read Kafka topic", err.Error())
		return
	}
	created, diags := tfmodels.KafkaTopicFromAPIModel(clusterUUID, *detail, planData)
	diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	// A topic that's `on` (from the cluster's default) has to be paused before it can be turned off. Paused, the 14-day
	// retention cap still applies, so the retentions are sent with `off`.
	paused := string(openapi.EnumsKafkaTopicTieredStorageStatePaused)
	if tieredStorageTransitionError(created.TieredStorage, planData.TieredStorage) != "" {
		r.updateSettings(ctx, planData, openapi.RouterUpdateKafkaTopicSettingsRequest{
			TieredStorage: lib.ToPtr(openapi.RouterUpdateKafkaTopicSettingsRequestTieredStorage(paused)),
		}, diagnostics)
		if diagnostics.HasError() {
			return
		}
		created.TieredStorage = types.StringValue(paused)
	}
	r.updateSettings(ctx, planData, planData.ToAPICreateSettingsRequest(created), diagnostics)
}

// createTopic creates the topic, first waiting for a new Managed Kafka cluster to finish deploying, since the API refuses
// (409) to create topics on a cluster that isn't active yet.
func (r *KafkaTopicResource) createTopic(ctx context.Context, clusterUUID string, body openapi.RouterCreateKafkaTopicRequest) error {
	deadline := time.Now().Add(kafkaClusterDeployTimeout)
	retriedActive := false
	for {
		resp, err := r.client.KafkaTopicCreateWithResponse(ctx, clusterUUID, body)
		if err != nil || resp.StatusCode() != http.StatusConflict || time.Now().After(deadline) {
			_, err = artieclient.JSON200(resp, err)
			return err
		}
		cluster, clusterErr := artieclient.JSON200(r.client.KafkaClusterDetailWithResponse(ctx, clusterUUID))
		if clusterErr == nil && cluster.Status == openapi.EnumsKafkaStatusActive && !retriedActive {
			// The deploy may have finished between the create and this read, so retry once before calling it a real conflict.
			retriedActive = true
			continue
		}
		if clusterErr != nil || cluster.Status != openapi.EnumsKafkaStatusDeploying {
			_, err = artieclient.JSON200(resp, err) // the conflict isn't the cluster deploying
			return err
		}

		tflog.Info(ctx, fmt.Sprintf("Kafka cluster %s is still deploying, retrying the topic create in %s", clusterUUID, kafkaClusterDeployPollInterval))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kafkaClusterDeployPollInterval):
		}
	}
}

func (r *KafkaTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readIntoState(ctx, stateData, &resp.State, &resp.Diagnostics)
}

func (r *KafkaTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData, stateData tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateSettings(ctx, planData, planData.ToAPISettingsRequest(stateData), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if planData.PartitionCount.ValueInt64() > stateData.PartitionCount.ValueInt64() {
		body := openapi.RouterIncreaseKafkaTopicPartitionsRequest{
			CurrentPartitionCount: int(stateData.PartitionCount.ValueInt64()),
			NewPartitionCount:     int(planData.PartitionCount.ValueInt64()),
		}
		if err := artieclient.CheckResponse(r.client.KafkaTopicPartitionsIncreaseWithResponse(ctx, planData.ClusterUUID.ValueString(), planData.Topic.ValueString(), body)); err != nil {
			resp.Diagnostics.AddError("Unable to add Kafka topic partitions", err.Error())
			return
		}
	}

	r.readIntoState(ctx, planData, &resp.State, &resp.Diagnostics)
}

func (r *KafkaTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateData tfmodels.KafkaTopic
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteResp, err := r.client.KafkaTopicDeleteWithResponse(ctx, stateData.ClusterUUID.ValueString(), stateData.Topic.ValueString())
	if err == nil && deleteResp.StatusCode() == http.StatusNotFound {
		return // already gone
	}
	if err := artieclient.CheckResponse(deleteResp, err); err != nil {
		resp.Diagnostics.AddError("Unable to delete Kafka topic", err.Error())
		return
	}
}

func (r *KafkaTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterUUID, topic, err := parseKafkaTopicImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), clusterUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("topic"), topic)...)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, kafkaTopicImportedKey, []byte("true"))...)
}

func parseKafkaTopicImportID(id string) (string, string, error) {
	clusterUUID, topic, ok := strings.Cut(id, "/")
	if !ok || clusterUUID == "" || topic == "" || strings.Contains(topic, "/") {
		return "", "", fmt.Errorf("expected an import ID of the form `<cluster_uuid>/<topic>`, got: %q", id)
	}
	return clusterUUID, topic, nil
}
