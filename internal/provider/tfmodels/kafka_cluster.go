package tfmodels

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type KafkaCluster struct {
	UUID                   types.String `tfsdk:"uuid"`
	Name                   types.String `tfsdk:"name"`
	Region                 types.String `tfsdk:"region"`
	InstanceType           types.String `tfsdk:"instance_type"`
	Brokers                types.Int64  `tfsdk:"brokers"`
	StorageGiB             types.Int64  `tfsdk:"storage_gib"`
	AllowedCIDRs           types.List   `tfsdk:"allowed_cidrs"`
	TieredStorage          types.Object `tfsdk:"tiered_storage"`
	StorageAutoscale       types.Object `tfsdk:"storage_autoscale"`
	CurrentStorageGiB      types.Int64  `tfsdk:"current_storage_gib"`
	Cloud                  types.String `tfsdk:"cloud"`
	Status                 types.String `tfsdk:"status"`
	ExternalBootstrapURL   types.String `tfsdk:"external_bootstrap_url"`
	KafkaVersion           types.String `tfsdk:"kafka_version"`
	TieredStorageAvailable types.Bool   `tfsdk:"tiered_storage_available"`
}

type KafkaClusterTieredStorage struct {
	Enabled             types.Bool  `tfsdk:"enabled"`
	LocalRetentionMS    types.Int64 `tfsdk:"local_retention_ms"`
	LocalRetentionBytes types.Int64 `tfsdk:"local_retention_bytes"`
}

var KafkaClusterTieredStorageAttrTypes = map[string]attr.Type{
	"enabled":               types.BoolType,
	"local_retention_ms":    types.Int64Type,
	"local_retention_bytes": types.Int64Type,
}

// DefaultKafkaLocalRetentionMS is the days on disk (3 days) Artie gives a new cluster.
const DefaultKafkaLocalRetentionMS = 259200000

// tieredStorage reads the tiered_storage object, reading a null or unknown one as on with the API's defaults.
func (k KafkaCluster) tieredStorage(ctx context.Context) (KafkaClusterTieredStorage, diag.Diagnostics) {
	tieredStorage, diags := parseOptionalObject[KafkaClusterTieredStorage](ctx, &k.TieredStorage)
	if tieredStorage == nil {
		return KafkaClusterTieredStorage{Enabled: types.BoolValue(true), LocalRetentionMS: types.Int64Unknown(), LocalRetentionBytes: types.Int64Null()}, diags
	}
	return *tieredStorage, diags
}

func tieredStorageFromAPIModel(apiModel openapi.PayloadsKafkaCluster) types.Object {
	return types.ObjectValueMust(KafkaClusterTieredStorageAttrTypes, map[string]attr.Value{
		"enabled":               types.BoolValue(apiModel.TierAllTopics),
		"local_retention_ms":    intPointerToInt64Value(apiModel.LocalRetentionMS),
		"local_retention_bytes": intPointerToInt64Value(apiModel.LocalRetentionBytes),
	})
}

type KafkaClusterStorageAutoscale struct {
	Enabled       types.Bool  `tfsdk:"enabled"`
	BumpAtPercent types.Int64 `tfsdk:"bump_at_percent"`
	TargetPercent types.Int64 `tfsdk:"target_percent"`
	MaxGiB        types.Int64 `tfsdk:"max_gib"`
}

var KafkaClusterStorageAutoscaleAttrTypes = map[string]attr.Type{
	"enabled":         types.BoolType,
	"bump_at_percent": types.Int64Type,
	"target_percent":  types.Int64Type,
	"max_gib":         types.Int64Type,
}

func (s KafkaClusterStorageAutoscale) toAPIModel() openapi.PayloadsKafkaClusterStorageAutoscale {
	return openapi.PayloadsKafkaClusterStorageAutoscale{
		Enabled:       s.Enabled.ValueBool(),
		BumpAtPercent: int(s.BumpAtPercent.ValueInt64()),
		TargetPercent: int(s.TargetPercent.ValueInt64()),
		MaxGiB:        int64ToIntPointer(s.MaxGiB),
	}
}

func (k KafkaCluster) storageAutoscaleToAPI(ctx context.Context) (*openapi.PayloadsKafkaClusterStorageAutoscale, diag.Diagnostics) {
	autoscale, diags := parseOptionalObject[KafkaClusterStorageAutoscale](ctx, &k.StorageAutoscale)
	if autoscale == nil || diags.HasError() {
		return nil, diags
	}
	return lib.ToPtr(autoscale.toAPIModel()), diags
}

func (k KafkaCluster) ToAPICreateRequest(ctx context.Context) (openapi.RouterCreateKafkaClusterRequest, diag.Diagnostics) {
	allowedCIDRs, diags := parseList[string](ctx, k.AllowedCIDRs)
	if allowedCIDRs == nil {
		allowedCIDRs = []string{}
	}
	tieredStorage, tieredDiags := k.tieredStorage(ctx)
	diags.Append(tieredDiags...)
	storageAutoscale, autoscaleDiags := k.storageAutoscaleToAPI(ctx)
	diags.Append(autoscaleDiags...)

	return openapi.RouterCreateKafkaClusterRequest{
		Name:                k.Name.ValueString(),
		Region:              k.Region.ValueString(),
		InstanceType:        openapi.PayloadsKafkaInstanceType(k.InstanceType.ValueString()),
		Brokers:             int(k.Brokers.ValueInt64()),
		StorageGiB:          int(k.StorageGiB.ValueInt64()),
		AllowedCIDRs:        allowedCIDRs,
		TieredStorage:       tieredStorage.Enabled.ValueBool(),
		LocalRetentionMS:    int64ToIntPointer(tieredStorage.LocalRetentionMS),
		LocalRetentionBytes: int64ToIntPointer(tieredStorage.LocalRetentionBytes),
		StorageAutoscale:    storageAutoscale,
	}, diags
}

// ToAPIUpdateRequest only sends the settings that differ from state: saving a setting again retries its last change,
// and each change the API applies can restart the brokers.
func (k KafkaCluster) ToAPIUpdateRequest(ctx context.Context, state KafkaCluster) (openapi.RouterUpdateKafkaClusterRequest, diag.Diagnostics) {
	var req openapi.RouterUpdateKafkaClusterRequest
	var diags diag.Diagnostics

	if !k.Name.Equal(state.Name) {
		req.Name = k.Name.ValueStringPointer()
	}
	if !k.InstanceType.Equal(state.InstanceType) {
		req.InstanceType = lib.ToPtr(openapi.PayloadsKafkaInstanceType(k.InstanceType.ValueString()))
	}
	if !k.Brokers.Equal(state.Brokers) {
		req.Brokers = int64ToIntPointer(k.Brokers)
	}
	if !k.StorageGiB.Equal(state.StorageGiB) {
		req.StorageGiB = int64ToIntPointer(k.StorageGiB)
	}
	if !k.AllowedCIDRs.Equal(state.AllowedCIDRs) {
		allowedCIDRs, listDiags := parseList[string](ctx, k.AllowedCIDRs)
		diags.Append(listDiags...)
		if allowedCIDRs == nil {
			allowedCIDRs = []string{}
		}
		req.AllowedCIDRs = &allowedCIDRs
	}
	if IsKnown(k.TieredStorage) {
		tieredStorage, tieredDiags := k.tieredStorageToAPIUpdate(ctx, state)
		diags.Append(tieredDiags...)
		req.TieredStorage = tieredStorage
	}
	if IsKnown(k.StorageAutoscale) && !k.StorageAutoscale.Equal(state.StorageAutoscale) {
		storageAutoscale, autoscaleDiags := k.storageAutoscaleToAPI(ctx)
		diags.Append(autoscaleDiags...)
		req.StorageAutoscale = storageAutoscale
	}

	return req, diags
}

// tieredStorageToAPIUpdate returns the tiered storage change, or nil if there is none. An unknown days on disk (left out
// of the config, with none in state) isn't a change.
func (k KafkaCluster) tieredStorageToAPIUpdate(ctx context.Context, state KafkaCluster) (*openapi.RouterUpdateKafkaClusterTieredStorageRequest, diag.Diagnostics) {
	planned, diags := k.tieredStorage(ctx)
	prior, priorDiags := state.tieredStorage(ctx)
	diags.Append(priorDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if planned.Enabled.Equal(prior.Enabled) && planned.LocalRetentionBytes.Equal(prior.LocalRetentionBytes) &&
		(!IsKnown(planned.LocalRetentionMS) || planned.LocalRetentionMS.Equal(prior.LocalRetentionMS)) {
		return nil, diags
	}

	// The API replaces all three together, so a days on disk left out of the config keeps the one in state.
	localRetentionMS := planned.LocalRetentionMS
	if !IsKnown(localRetentionMS) {
		localRetentionMS = prior.LocalRetentionMS
	}
	if !IsKnown(localRetentionMS) {
		// Clusters made before days on disk existed don't return one, and the API needs one with every change.
		if planned.Enabled.ValueBool() {
			diags.AddAttributeError(path.Root("tiered_storage").AtName("local_retention_ms"), "Missing local_retention_ms",
				"Changing tiered storage on this cluster needs `tiered_storage.local_retention_ms` set in the config, since Artie didn't return its current value.")
			return nil, diags
		}
		// Turning tiered storage off, the days on disk can't be set and don't apply, so send Artie's default.
		localRetentionMS = types.Int64Value(DefaultKafkaLocalRetentionMS)
	}
	return &openapi.RouterUpdateKafkaClusterTieredStorageRequest{
		TierAllTopics:       planned.Enabled.ValueBool(),
		LocalRetentionMS:    localRetentionMS.ValueInt64(),
		LocalRetentionBytes: int64ToIntPointer(planned.LocalRetentionBytes),
	}, diags
}

// WithCreatedCluster is the state for a cluster that was created but couldn't be read back: the plan, with what the
// create response returns filled in and every other computed attribute null until the next refresh.
func (k KafkaCluster) WithCreatedCluster(apiModel openapi.PayloadsKafkaCluster) KafkaCluster {
	k.UUID = types.StringValue(apiModel.Uuid.String())
	k.Cloud = types.StringValue(apiModel.Cloud)
	k.Status = types.StringValue(string(apiModel.Status))
	k.ExternalBootstrapURL = types.StringValue(apiModel.ExternalBootstrapURL)
	k.TieredStorageAvailable = types.BoolValue(apiModel.TieredStorageAvailable)
	k.CurrentStorageGiB = types.Int64Null()
	k.KafkaVersion = types.StringNull()
	if IsKnown(k.TieredStorage) {
		tieredStorage := maps.Clone(k.TieredStorage.Attributes())
		if !IsKnown(tieredStorage["local_retention_ms"]) {
			tieredStorage["local_retention_ms"] = intPointerToInt64Value(apiModel.LocalRetentionMS)
		}
		k.TieredStorage = types.ObjectValueMust(KafkaClusterTieredStorageAttrTypes, tieredStorage)
	} else {
		k.TieredStorage = tieredStorageFromAPIModel(apiModel)
	}
	if !IsKnown(k.StorageAutoscale) {
		k.StorageAutoscale = types.ObjectNull(KafkaClusterStorageAutoscaleAttrTypes)
	}
	return k
}

// WithRequestedStorage keeps the storage_gib that was asked for while the brokers have at least that much, so a disk
// that storage autoscaling grew doesn't show up as a diff (and a plan to shrink it). current_storage_gib has the size
// the brokers have.
func (k KafkaCluster) WithRequestedStorage(requested types.Int64) KafkaCluster {
	if IsKnown(requested) && IsKnown(k.StorageGiB) && requested.ValueInt64() < k.StorageGiB.ValueInt64() {
		k.StorageGiB = requested
	}
	return k
}

func KafkaClusterFromAPIModel(ctx context.Context, apiModel openapi.PayloadsKafkaCluster) (KafkaCluster, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !apiModel.IsManaged || apiModel.Managed == nil {
		diags.AddError("Unsupported Kafka cluster", "Only Managed Kafka clusters that Artie runs can be managed with the artie_kafka_cluster resource.")
		return KafkaCluster{}, diags
	}
	managed := *apiModel.Managed

	allowedCIDRs, listDiags := optionalStringListToListValue(ctx, &managed.AllowedCIDRs)
	diags.Append(listDiags...)

	storageAutoscale := types.ObjectNull(KafkaClusterStorageAutoscaleAttrTypes)
	if apiModel.StorageAutoscale != nil {
		autoscale := *apiModel.StorageAutoscale
		var objDiags diag.Diagnostics
		storageAutoscale, objDiags = types.ObjectValueFrom(ctx, KafkaClusterStorageAutoscaleAttrTypes, KafkaClusterStorageAutoscale{
			Enabled:       types.BoolValue(autoscale.Enabled),
			BumpAtPercent: types.Int64Value(int64(autoscale.BumpAtPercent)),
			TargetPercent: types.Int64Value(int64(autoscale.TargetPercent)),
			MaxGiB:        intPointerToInt64Value(autoscale.MaxGiB),
		})
		diags.Append(objDiags...)
	}

	return KafkaCluster{
		UUID:                   types.StringValue(apiModel.Uuid.String()),
		Name:                   types.StringValue(apiModel.Name),
		Region:                 types.StringValue(apiModel.Region),
		InstanceType:           types.StringValue(managed.InstanceType),
		Brokers:                types.Int64Value(int64(managed.Brokers)),
		StorageGiB:             types.Int64Value(int64(managed.StorageGiB)),
		AllowedCIDRs:           allowedCIDRs,
		TieredStorage:          tieredStorageFromAPIModel(apiModel),
		StorageAutoscale:       storageAutoscale,
		CurrentStorageGiB:      types.Int64Value(int64(managed.StorageGiB)),
		Cloud:                  types.StringValue(apiModel.Cloud),
		Status:                 types.StringValue(string(apiModel.Status)),
		ExternalBootstrapURL:   types.StringValue(apiModel.ExternalBootstrapURL),
		KafkaVersion:           types.StringValue(managed.KafkaVersion),
		TieredStorageAvailable: types.BoolValue(apiModel.TieredStorageAvailable),
	}, diags
}

// intPointerToInt64Value is the inverse of int64ToIntPointer: nil reads back as null.
func intPointerToInt64Value(value *int) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*value))
}
