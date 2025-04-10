package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-artie/internal/artieclient"
)

type Deployment struct {
	UUID                     types.String                 `tfsdk:"uuid"`
	Name                     types.String                 `tfsdk:"name"`
	Status                   types.String                 `tfsdk:"status"`
	Source                   *Source                      `tfsdk:"source"`
	DestinationUUID          types.String                 `tfsdk:"destination_uuid"`
	DestinationConfig        *DeploymentDestinationConfig `tfsdk:"destination_config"`
	SSHTunnelUUID            types.String                 `tfsdk:"ssh_tunnel_uuid"`
	SnowflakeEcoScheduleUUID types.String                 `tfsdk:"snowflake_eco_schedule_uuid"`
	DataPlaneName            types.String                 `tfsdk:"data_plane_name"`

	// Advanced settings
	FlushConfig                    types.Object `tfsdk:"flush_config"`
	DropDeletedColumns             types.Bool   `tfsdk:"drop_deleted_columns"`
	SoftDeleteRows                 types.Bool   `tfsdk:"soft_delete_rows"`
	IncludeArtieUpdatedAtColumn    types.Bool   `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool   `tfsdk:"include_database_updated_at_column"`
	OneTopicPerSchema              types.Bool   `tfsdk:"one_topic_per_schema"`
	PublicationNameOverride        types.String `tfsdk:"postgres_publication_name_override"`
	ReplicationSlotOverride        types.String `tfsdk:"postgres_replication_slot_override"`
}

type DeploymentFlushConfig struct {
	FlushIntervalSeconds types.Int64 `tfsdk:"flush_interval_seconds"`
	BufferRows           types.Int64 `tfsdk:"buffer_rows"`
	FlushSizeKB          types.Int64 `tfsdk:"flush_size_kb"`
}

var flushAttrTypes = map[string]attr.Type{
	"flush_interval_seconds": types.Int64Type,
	"buffer_rows":            types.Int64Type,
	"flush_size_kb":          types.Int64Type,
}

func buildFlushConfig(ctx context.Context, d types.Object) (*DeploymentFlushConfig, diag.Diagnostics) {
	var flushConfig *DeploymentFlushConfig
	flushConfigDiags := d.As(ctx, &flushConfig, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})

	if flushConfigDiags.HasError() {
		return nil, flushConfigDiags
	}

	return flushConfig, nil
}

func (d Deployment) ToAPIBaseModel(ctx context.Context) (artieclient.BaseDeployment, diag.Diagnostics) {
	apiSource, diags := d.Source.ToAPIModel(ctx)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	destinationUUID, destDiags := parseOptionalUUID(d.DestinationUUID)
	diags.Append(destDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	sshTunnelUUID, sshDiags := parseOptionalUUID(d.SSHTunnelUUID)
	diags.Append(sshDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	snowflakeEcoScheduleUUID, snowflakeDiags := parseOptionalUUID(d.SnowflakeEcoScheduleUUID)
	diags.Append(snowflakeDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	flushConfig, flushConfigDiags := buildFlushConfig(ctx, d.FlushConfig)
	diags.Append(flushConfigDiags...)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	return artieclient.BaseDeployment{
		Name:                     d.Name.ValueString(),
		Source:                   apiSource,
		DestinationUUID:          destinationUUID,
		DestinationConfig:        d.DestinationConfig.ToAPIModel(),
		SSHTunnelUUID:            sshTunnelUUID,
		SnowflakeEcoScheduleUUID: snowflakeEcoScheduleUUID,
		DataPlaneName:            d.DataPlaneName.ValueString(),
		// Advanced settings:
		FlushConfig:                    flushConfig.ToAPIModel(),
		DropDeletedColumns:             d.DropDeletedColumns.ValueBoolPointer(),
		EnableSoftDelete:               d.SoftDeleteRows.ValueBoolPointer(),
		IncludeArtieUpdatedAtColumn:    d.IncludeArtieUpdatedAtColumn.ValueBoolPointer(),
		IncludeDatabaseUpdatedAtColumn: d.IncludeDatabaseUpdatedAtColumn.ValueBoolPointer(),
		OneTopicPerSchema:              d.OneTopicPerSchema.ValueBoolPointer(),
		PublicationNameOverride:        d.PublicationNameOverride.ValueStringPointer(),
		ReplicationSlotOverride:        d.ReplicationSlotOverride.ValueStringPointer(),
	}, diags
}

func (d Deployment) ToAPIModel(ctx context.Context) (artieclient.Deployment, diag.Diagnostics) {
	apiBaseModel, diags := d.ToAPIBaseModel(ctx)
	if diags.HasError() {
		return artieclient.Deployment{}, diags
	}

	uuid, uuidDiags := parseUUID(d.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Deployment{}, diags
	}

	return artieclient.Deployment{
		UUID:           uuid,
		Status:         d.Status.ValueString(),
		BaseDeployment: apiBaseModel,
	}, diags
}

func DeploymentFromAPIModel(ctx context.Context, apiModel artieclient.Deployment) (Deployment, diag.Diagnostics) {
	source, diags := SourceFromAPIModel(ctx, apiModel.Source)
	if diags.HasError() {
		return Deployment{}, diags
	}

	destinationConfig := DeploymentDestinationConfigFromAPIModel(apiModel.DestinationConfig)

	var flushConfig types.Object
	if apiModel.FlushConfig != nil {
		flushConfigObj, flushConfigDiags := DeploymentFlushConfigFromAPIModel(ctx, *apiModel.FlushConfig)
		diags.Append(flushConfigDiags...)
		if diags.HasError() {
			return Deployment{}, diags
		}
		flushConfig = flushConfigObj
	}

	return Deployment{
		UUID:                     types.StringValue(apiModel.UUID.String()),
		Name:                     types.StringValue(apiModel.Name),
		Status:                   types.StringValue(apiModel.Status),
		Source:                   &source,
		DestinationUUID:          optionalUUIDToStringValue(apiModel.DestinationUUID),
		DestinationConfig:        &destinationConfig,
		SSHTunnelUUID:            optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID),
		DataPlaneName:            types.StringValue(apiModel.DataPlaneName),

		// Advanced settings:
		FlushConfig:                    flushConfig,
		DropDeletedColumns:             types.BoolPointerValue(apiModel.DropDeletedColumns),
		SoftDeleteRows:                 types.BoolPointerValue(apiModel.EnableSoftDelete),
		IncludeArtieUpdatedAtColumn:    types.BoolPointerValue(apiModel.IncludeArtieUpdatedAtColumn),
		IncludeDatabaseUpdatedAtColumn: types.BoolPointerValue(apiModel.IncludeDatabaseUpdatedAtColumn),
		OneTopicPerSchema:              types.BoolPointerValue(apiModel.OneTopicPerSchema),
		PublicationNameOverride:        types.StringPointerValue(apiModel.PublicationNameOverride),
		ReplicationSlotOverride:        types.StringPointerValue(apiModel.ReplicationSlotOverride),
	}, diags
}

type DeploymentDestinationConfig struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
	Bucket                types.String `tfsdk:"bucket"`
	Folder                types.String `tfsdk:"folder"`
}

func (d *DeploymentFlushConfig) ToAPIModel() *artieclient.FlushConfig {
	if d == nil {
		// Support unknown.
		return nil
	}

	return &artieclient.FlushConfig{
		FlushIntervalSeconds: d.FlushIntervalSeconds.ValueInt64(),
		BufferRows:           d.BufferRows.ValueInt64(),
		FlushSizeKB:          d.FlushSizeKB.ValueInt64(),
	}
}

func (d DeploymentDestinationConfig) ToAPIModel() artieclient.DestinationConfig {
	return artieclient.DestinationConfig{
		Dataset:               d.Dataset.ValueString(),
		Database:              d.Database.ValueString(),
		Schema:                d.Schema.ValueString(),
		UseSameSchemaAsSource: d.UseSameSchemaAsSource.ValueBool(),
		SchemaNamePrefix:      d.SchemaNamePrefix.ValueString(),
		Bucket:                d.Bucket.ValueString(),
		Folder:                d.Folder.ValueString(),
	}
}

func DeploymentDestinationConfigFromAPIModel(apiModel artieclient.DestinationConfig) DeploymentDestinationConfig {
	return DeploymentDestinationConfig{
		Dataset:               types.StringValue(apiModel.Dataset),
		Database:              types.StringValue(apiModel.Database),
		Schema:                types.StringValue(apiModel.Schema),
		UseSameSchemaAsSource: types.BoolValue(apiModel.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.SchemaNamePrefix),
		Bucket:                types.StringValue(apiModel.Bucket),
		Folder:                types.StringValue(apiModel.Folder),
	}
}

func DeploymentFlushConfigFromAPIModel(ctx context.Context, apiModel artieclient.FlushConfig) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, flushAttrTypes, DeploymentFlushConfig{
		FlushIntervalSeconds: types.Int64Value(apiModel.FlushIntervalSeconds),
		BufferRows:           types.Int64Value(apiModel.BufferRows),
		FlushSizeKB:          types.Int64Value(apiModel.FlushSizeKB),
	})
}
