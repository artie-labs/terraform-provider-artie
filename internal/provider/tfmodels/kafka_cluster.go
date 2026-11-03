package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	TieredStorage          types.Bool   `tfsdk:"tiered_storage"`
	LocalRetentionMS       types.Int64  `tfsdk:"local_retention_ms"`
	LocalRetentionBytes    types.Int64  `tfsdk:"local_retention_bytes"`
	StorageAutoscale       types.Object `tfsdk:"storage_autoscale"`
	Cloud                  types.String `tfsdk:"cloud"`
	Status                 types.String `tfsdk:"status"`
	ExternalBootstrapURL   types.String `tfsdk:"external_bootstrap_url"`
	KafkaVersion           types.String `tfsdk:"kafka_version"`
	TieredStorageAvailable types.Bool   `tfsdk:"tiered_storage_available"`
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
	storageAutoscale, autoscaleDiags := k.storageAutoscaleToAPI(ctx)
	diags.Append(autoscaleDiags...)

	return openapi.RouterCreateKafkaClusterRequest{
		Name:                k.Name.ValueString(),
		Region:              k.Region.ValueString(),
		InstanceType:        openapi.PayloadsKafkaInstanceType(k.InstanceType.ValueString()),
		Brokers:             int(k.Brokers.ValueInt64()),
		StorageGiB:          int(k.StorageGiB.ValueInt64()),
		AllowedCIDRs:        allowedCIDRs,
		TieredStorage:       k.TieredStorage.ValueBool(),
		LocalRetentionMS:    int64ToIntPointer(k.LocalRetentionMS),
		LocalRetentionBytes: int64ToIntPointer(k.LocalRetentionBytes),
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
	if !k.TieredStorage.Equal(state.TieredStorage) || !k.LocalRetentionMS.Equal(state.LocalRetentionMS) || !k.LocalRetentionBytes.Equal(state.LocalRetentionBytes) {
		// The API replaces all three together, so a days-on-disk left out of the config keeps the one in state.
		localRetentionMS := k.LocalRetentionMS
		if !IsKnown(localRetentionMS) {
			localRetentionMS = state.LocalRetentionMS
		}
		req.TieredStorage = &openapi.RouterUpdateKafkaClusterTieredStorageRequest{
			TierAllTopics:       k.TieredStorage.ValueBool(),
			LocalRetentionMS:    localRetentionMS.ValueInt64(),
			LocalRetentionBytes: int64ToIntPointer(k.LocalRetentionBytes),
		}
	}
	if IsKnown(k.StorageAutoscale) && !k.StorageAutoscale.Equal(state.StorageAutoscale) {
		storageAutoscale, autoscaleDiags := k.storageAutoscaleToAPI(ctx)
		diags.Append(autoscaleDiags...)
		req.StorageAutoscale = storageAutoscale
	}

	return req, diags
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
		TieredStorage:          types.BoolValue(apiModel.TierAllTopics),
		LocalRetentionMS:       intPointerToInt64Value(apiModel.LocalRetentionMS),
		LocalRetentionBytes:    intPointerToInt64Value(apiModel.LocalRetentionBytes),
		StorageAutoscale:       storageAutoscale,
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
