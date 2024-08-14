package models

import (
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DestinationAPIToResourceModel(apiModel artieclient.Destination, resourceModel *DestinationResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID.String())
	resourceModel.Type = types.StringValue(apiModel.Type)
	resourceModel.Label = types.StringValue(apiModel.Label)
	resourceModel.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)

	switch resourceModel.Type.ValueString() {
	case string(Snowflake):
		resourceModel.SnowflakeConfig = &SnowflakeSharedConfigModel{
			AccountURL: types.StringValue(apiModel.Config.SnowflakeAccountURL),
			VirtualDWH: types.StringValue(apiModel.Config.SnowflakeVirtualDWH),
			PrivateKey: types.StringValue(apiModel.Config.SnowflakePrivateKey),
			Username:   types.StringValue(apiModel.Config.Username),
			Password:   types.StringValue(apiModel.Config.Password),
		}
	case string(BigQuery):
		resourceModel.BigQueryConfig = &BigQuerySharedConfigModel{
			ProjectID:       types.StringValue(apiModel.Config.GCPProjectID),
			Location:        types.StringValue(apiModel.Config.GCPLocation),
			CredentialsData: types.StringValue(apiModel.Config.GCPCredentialsData),
		}
	case string(Redshift):
		resourceModel.RedshiftConfig = &RedshiftSharedConfigModel{
			Endpoint: types.StringValue(apiModel.Config.Endpoint),
			Host:     types.StringValue(apiModel.Config.Host),
			Port:     types.Int32Value(apiModel.Config.Port),
			Username: types.StringValue(apiModel.Config.Username),
			Password: types.StringValue(apiModel.Config.Password),
		}
	}
}

func DestinationResourceToAPISharedConfigModel(resourceModel DestinationResourceModel) artieclient.DestinationSharedConfig {
	switch resourceModel.Type.ValueString() {
	case string(Snowflake):
		return artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: resourceModel.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: resourceModel.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: resourceModel.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            resourceModel.SnowflakeConfig.Username.ValueString(),
			Password:            resourceModel.SnowflakeConfig.Password.ValueString(),
		}
	case string(BigQuery):
		return artieclient.DestinationSharedConfig{
			GCPProjectID:       resourceModel.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        resourceModel.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: resourceModel.BigQueryConfig.CredentialsData.ValueString(),
		}
	case string(Redshift):
		return artieclient.DestinationSharedConfig{
			Endpoint: resourceModel.RedshiftConfig.Endpoint.ValueString(),
			Host:     resourceModel.RedshiftConfig.Host.ValueString(),
			Port:     resourceModel.RedshiftConfig.Port.ValueInt32(),
			Username: resourceModel.RedshiftConfig.Username.ValueString(),
			Password: resourceModel.RedshiftConfig.Password.ValueString(),
		}
	default:
		return artieclient.DestinationSharedConfig{}
	}
}

func DestinationResourceToAPIModel(resourceModel DestinationResourceModel) artieclient.Destination {
	return artieclient.Destination{
		UUID:          parseUUID(resourceModel.UUID),
		Type:          resourceModel.Type.ValueString(),
		Label:         resourceModel.Label.ValueString(),
		Config:        DestinationResourceToAPISharedConfigModel(resourceModel),
		SSHTunnelUUID: ParseOptionalUUID(resourceModel.SSHTunnelUUID),
	}
}
