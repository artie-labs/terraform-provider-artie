package models

import (
	"strings"
	"terraform-provider-artie/internal/artieclient"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (resourceModel *DestinationResourceModel) FillFromAPIModel(apiModel artieclient.Destination) {
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

func (d DestinationResourceModel) ToBaseAPIModel() artieclient.BaseDestination {
	var sharedConfig artieclient.DestinationSharedConfig
	switch strings.ToLower(d.Type.ValueString()) {
	case string(Snowflake):
		sharedConfig = artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: d.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: d.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: d.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            d.SnowflakeConfig.Username.ValueString(),
			Password:            d.SnowflakeConfig.Password.ValueString(),
		}
	case string(BigQuery):
		sharedConfig = artieclient.DestinationSharedConfig{
			GCPProjectID:       d.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        d.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: d.BigQueryConfig.CredentialsData.ValueString(),
		}
	case string(Redshift):
		sharedConfig = artieclient.DestinationSharedConfig{
			Endpoint: d.RedshiftConfig.Endpoint.ValueString(),
			Username: d.RedshiftConfig.Username.ValueString(),
			Password: d.RedshiftConfig.Password.ValueString(),
		}
	default:
		sharedConfig = artieclient.DestinationSharedConfig{}
	}

	return artieclient.BaseDestination{
		Type:          d.Type.ValueString(),
		Label:         d.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: ParseOptionalUUID(d.SSHTunnelUUID),
	}
}

func (d DestinationResourceModel) ToAPIModel() artieclient.Destination {
	return artieclient.Destination{
		UUID:            parseUUID(d.UUID),
		BaseDestination: d.ToBaseAPIModel(),
	}
}
