package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

	// Advanced settings
	DropDeletedColumns             types.Bool `tfsdk:"drop_deleted_columns"`
	SoftDeleteRows                 types.Bool `tfsdk:"soft_delete_rows"`
	IncludeArtieUpdatedAtColumn    types.Bool `tfsdk:"include_artie_updated_at_column"`
	IncludeDatabaseUpdatedAtColumn types.Bool `tfsdk:"include_database_updated_at_column"`
	OneTopicPerSchema              types.Bool `tfsdk:"one_topic_per_schema"`
}

func (d Deployment) ToAPIBaseModel(ctx context.Context) (artieclient.BaseDeployment, diag.Diagnostics) {
	apiSource, diags := d.Source.ToAPIModel(ctx)
	if diags.HasError() {
		return artieclient.BaseDeployment{}, diags
	}

	return artieclient.BaseDeployment{
		Name:                           d.Name.ValueString(),
		Source:                         apiSource,
		DestinationUUID:                ParseOptionalUUID(d.DestinationUUID),
		DestinationConfig:              d.DestinationConfig.ToAPIModel(),
		SSHTunnelUUID:                  ParseOptionalUUID(d.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID:       ParseOptionalUUID(d.SnowflakeEcoScheduleUUID),
		DropDeletedColumns:             parseOptionalBool(d.DropDeletedColumns),
		EnableSoftDelete:               parseOptionalBool(d.SoftDeleteRows),
		IncludeArtieUpdatedAtColumn:    parseOptionalBool(d.IncludeArtieUpdatedAtColumn),
		IncludeDatabaseUpdatedAtColumn: parseOptionalBool(d.IncludeDatabaseUpdatedAtColumn),
		OneTopicPerSchema:              parseOptionalBool(d.OneTopicPerSchema),
	}, diags
}

func (d Deployment) ToAPIModel(ctx context.Context) (artieclient.Deployment, diag.Diagnostics) {
	apiBaseModel, diags := d.ToAPIBaseModel(ctx)
	if diags.HasError() {
		return artieclient.Deployment{}, diags
	}

	return artieclient.Deployment{
		UUID:           parseUUID(d.UUID),
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
	return Deployment{
		UUID:                           types.StringValue(apiModel.UUID.String()),
		Name:                           types.StringValue(apiModel.Name),
		Status:                         types.StringValue(apiModel.Status),
		Source:                         &source,
		DestinationUUID:                optionalUUIDToStringValue(apiModel.DestinationUUID),
		DestinationConfig:              &destinationConfig,
		SSHTunnelUUID:                  optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID:       optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID),
		DropDeletedColumns:             optionalBoolToBoolValue(apiModel.DropDeletedColumns),
		SoftDeleteRows:                 optionalBoolToBoolValue(apiModel.EnableSoftDelete),
		IncludeArtieUpdatedAtColumn:    optionalBoolToBoolValue(apiModel.IncludeArtieUpdatedAtColumn),
		IncludeDatabaseUpdatedAtColumn: optionalBoolToBoolValue(apiModel.IncludeDatabaseUpdatedAtColumn),
		OneTopicPerSchema:              optionalBoolToBoolValue(apiModel.OneTopicPerSchema),
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
