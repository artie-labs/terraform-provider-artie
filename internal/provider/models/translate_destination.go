package models

import (
	"strings"
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DestinationAPIToResourceModel(apiModel artieclient.Destination, resourceModel *DestinationResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID.String())
	resourceModel.Type = types.StringValue(apiModel.Type)
	resourceModel.Label = types.StringValue(apiModel.Label)
	resourceModel.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)

	switch strings.ToLower(resourceModel.Type.ValueString()) {
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
			Username: types.StringValue(apiModel.Config.Username),
			Password: types.StringValue(apiModel.Config.Password),
		}
	}
}

func (rm DestinationResourceModel) ToBaseAPIModel() artieclient.BaseDestination {
	var sharedConfig artieclient.DestinationSharedConfig
	switch strings.ToLower(rm.Type.ValueString()) {
	case string(Snowflake):
		sharedConfig = artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: rm.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: rm.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: rm.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            rm.SnowflakeConfig.Username.ValueString(),
			Password:            rm.SnowflakeConfig.Password.ValueString(),
		}
	case string(BigQuery):
		sharedConfig = artieclient.DestinationSharedConfig{
			GCPProjectID:       rm.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        rm.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: rm.BigQueryConfig.CredentialsData.ValueString(),
		}
	case string(Redshift):
		sharedConfig = artieclient.DestinationSharedConfig{
			Endpoint: rm.RedshiftConfig.Endpoint.ValueString(),
			Username: rm.RedshiftConfig.Username.ValueString(),
			Password: rm.RedshiftConfig.Password.ValueString(),
		}
	default:
		sharedConfig = artieclient.DestinationSharedConfig{}
	}

	return artieclient.BaseDestination{
		Type:          rm.Type.ValueString(),
		Label:         rm.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: ParseOptionalUUID(rm.SSHTunnelUUID),
	}
}

func (rm DestinationResourceModel) ToAPIModel() artieclient.Destination {
	return artieclient.Destination{
		UUID:            parseUUID(rm.UUID),
		BaseDestination: rm.ToBaseAPIModel(),
	}
}
