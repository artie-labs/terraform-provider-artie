package models

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DestinationAPIToResourceModel(apiModel DestinationAPIModel, resourceModel *DestinationResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID)
	resourceModel.CompanyUUID = types.StringValue(apiModel.CompanyUUID)
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Label = types.StringValue(apiModel.Label)
	resourceModel.SSHTunnelUUID = types.StringValue(apiModel.SSHTunnelUUID)

	switch resourceModel.Name.ValueString() {
	case "Snowflake":
		resourceModel.SnowflakeConfig = &SnowflakeSharedConfigModel{
			AccountURL: types.StringValue(apiModel.Config.SnowflakeAccountURL),
			VirtualDWH: types.StringValue(apiModel.Config.SnowflakeVirtualDWH),
			PrivateKey: types.StringValue(apiModel.Config.SnowflakePrivateKey),
			Username:   types.StringValue(apiModel.Config.Username),
			Password:   types.StringValue(apiModel.Config.Password),
		}
	case "BigQuery":
		resourceModel.BigQueryConfig = &BigQuerySharedConfigModel{
			ProjectID:       types.StringValue(apiModel.Config.GCPProjectID),
			Location:        types.StringValue(apiModel.Config.GCPLocation),
			CredentialsData: types.StringValue(apiModel.Config.GCPCredentialsData),
		}
	case "Redshift":
		resourceModel.RedshiftConfig = &RedshiftSharedConfigModel{
			Endpoint: types.StringValue(apiModel.Config.Endpoint),
			Host:     types.StringValue(apiModel.Config.Host),
			Port:     types.Int64Value(apiModel.Config.Port),
			Username: types.StringValue(apiModel.Config.Username),
			Password: types.StringValue(apiModel.Config.Password),
		}
	}
}

func DestinationResourceToAPIModel(resourceModel DestinationResourceModel) DestinationAPIModel {
	sshTunnelUUID := resourceModel.SSHTunnelUUID.ValueString()
	if sshTunnelUUID == "" {
		sshTunnelUUID = uuid.Nil.String()
	}
	apiModel := DestinationAPIModel{
		UUID:          resourceModel.UUID.ValueString(),
		CompanyUUID:   resourceModel.CompanyUUID.ValueString(),
		Name:          resourceModel.Name.ValueString(),
		Label:         resourceModel.Label.ValueString(),
		SSHTunnelUUID: sshTunnelUUID,
	}

	switch resourceModel.Name.ValueString() {
	case "Snowflake":
		apiModel.Config = DestinationSharedConfigAPIModel{
			SnowflakeAccountURL: resourceModel.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: resourceModel.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: resourceModel.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            resourceModel.SnowflakeConfig.Username.ValueString(),
			Password:            resourceModel.SnowflakeConfig.Password.ValueString(),
		}
	case "BigQuery":
		apiModel.Config = DestinationSharedConfigAPIModel{
			GCPProjectID:       resourceModel.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        resourceModel.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: resourceModel.BigQueryConfig.CredentialsData.ValueString(),
		}
	case "Redshift":
		apiModel.Config = DestinationSharedConfigAPIModel{
			Endpoint: resourceModel.RedshiftConfig.Endpoint.ValueString(),
			Host:     resourceModel.RedshiftConfig.Host.ValueString(),
			Port:     resourceModel.RedshiftConfig.Port.ValueInt64(),
			Username: resourceModel.RedshiftConfig.Username.ValueString(),
			Password: resourceModel.RedshiftConfig.Password.ValueString(),
		}
	}

	return apiModel
}
