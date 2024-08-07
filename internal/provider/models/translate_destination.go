package models

import (
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DestinationAPIToResourceModel(apiModel artieclient.Destination, resourceModel *DestinationResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID)
	resourceModel.Type = types.StringValue(apiModel.Type)
	resourceModel.Label = types.StringValue(apiModel.Label)

	sshTunnelUUID := ""
	if apiModel.SSHTunnelUUID != nil {
		sshTunnelUUID = *apiModel.SSHTunnelUUID
	}
	resourceModel.SSHTunnelUUID = types.StringValue(sshTunnelUUID)

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

func DestinationResourceToAPIModel(resourceModel DestinationResourceModel) artieclient.Destination {
	apiModel := artieclient.Destination{
		UUID:  resourceModel.UUID.ValueString(),
		Type:  resourceModel.Type.ValueString(),
		Label: resourceModel.Label.ValueString(),
	}

	sshTunnelUUID := resourceModel.SSHTunnelUUID.ValueString()
	if sshTunnelUUID != "" {
		apiModel.SSHTunnelUUID = &sshTunnelUUID
	}

	switch resourceModel.Type.ValueString() {
	case string(Snowflake):
		apiModel.Config = artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: resourceModel.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: resourceModel.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: resourceModel.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            resourceModel.SnowflakeConfig.Username.ValueString(),
			Password:            resourceModel.SnowflakeConfig.Password.ValueString(),
		}
	case string(BigQuery):
		apiModel.Config = artieclient.DestinationSharedConfig{
			GCPProjectID:       resourceModel.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        resourceModel.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: resourceModel.BigQueryConfig.CredentialsData.ValueString(),
		}
	case string(Redshift):
		apiModel.Config = artieclient.DestinationSharedConfig{
			Endpoint: resourceModel.RedshiftConfig.Endpoint.ValueString(),
			Host:     resourceModel.RedshiftConfig.Host.ValueString(),
			Port:     resourceModel.RedshiftConfig.Port.ValueInt32(),
			Username: resourceModel.RedshiftConfig.Username.ValueString(),
			Password: resourceModel.RedshiftConfig.Password.ValueString(),
		}
	}

	return apiModel
}
