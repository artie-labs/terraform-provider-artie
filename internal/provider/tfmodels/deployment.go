package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Deployment struct {
	UUID                     types.String                `tfsdk:"uuid"`
	Name                     types.String                `tfsdk:"name"`
	Status                   types.String                `tfsdk:"status"`
	Source                   Source                      `tfsdk:"source"`
	DestinationUUID          types.String                `tfsdk:"destination_uuid"`
	DestinationConfig        DeploymentDestinationConfig `tfsdk:"destination_config"`
	SSHTunnelUUID            types.String                `tfsdk:"ssh_tunnel_uuid"`
	SnowflakeEcoScheduleUUID types.String                `tfsdk:"snowflake_eco_schedule_uuid"`
}

func (d Deployment) ToAPIBaseModel() artieclient.BaseDeployment {
	return artieclient.BaseDeployment{
		Name:                     d.Name.ValueString(),
		Source:                   d.Source.ToAPIModel(),
		DestinationUUID:          ParseOptionalUUID(d.DestinationUUID),
		DestinationConfig:        d.DestinationConfig.ToAPIModel(),
		SSHTunnelUUID:            ParseOptionalUUID(d.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: ParseOptionalUUID(d.SnowflakeEcoScheduleUUID),
	}
}

func (d Deployment) ToAPIModel() artieclient.Deployment {
	return artieclient.Deployment{
		UUID:           parseUUID(d.UUID),
		Status:         d.Status.ValueString(),
		BaseDeployment: d.ToAPIBaseModel(),
	}
}

func DeploymentFromAPIModel(apiModel artieclient.Deployment) Deployment {
	return Deployment{
		UUID:                     types.StringValue(apiModel.UUID.String()),
		Name:                     types.StringValue(apiModel.Name),
		Status:                   types.StringValue(apiModel.Status),
		Source:                   SourceFromAPIModel(apiModel.Source),
		DestinationUUID:          optionalUUIDToStringValue(apiModel.DestinationUUID),
		DestinationConfig:        DeploymentDestinationConfigFromAPIModel(apiModel.DestinationConfig),
		SSHTunnelUUID:            optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
		SnowflakeEcoScheduleUUID: optionalUUIDToStringValue(apiModel.SnowflakeEcoScheduleUUID),
	}
}

type DeploymentDestinationConfig struct {
	Dataset               types.String `tfsdk:"dataset"`
	Database              types.String `tfsdk:"database"`
	Schema                types.String `tfsdk:"schema"`
	UseSameSchemaAsSource types.Bool   `tfsdk:"use_same_schema_as_source"`
	SchemaNamePrefix      types.String `tfsdk:"schema_name_prefix"`
}

func (d DeploymentDestinationConfig) ToAPIModel() artieclient.DestinationConfig {
	return artieclient.DestinationConfig{
		Dataset:               d.Dataset.ValueString(),
		Database:              d.Database.ValueString(),
		Schema:                d.Schema.ValueString(),
		UseSameSchemaAsSource: d.UseSameSchemaAsSource.ValueBool(),
		SchemaNamePrefix:      d.SchemaNamePrefix.ValueString(),
	}
}

func DeploymentDestinationConfigFromAPIModel(apiModel artieclient.DestinationConfig) DeploymentDestinationConfig {
	return DeploymentDestinationConfig{
		Dataset:               types.StringValue(apiModel.Dataset),
		Database:              types.StringValue(apiModel.Database),
		Schema:                types.StringValue(apiModel.Schema),
		UseSameSchemaAsSource: types.BoolValue(apiModel.UseSameSchemaAsSource),
		SchemaNamePrefix:      types.StringValue(apiModel.SchemaNamePrefix),
	}
}
