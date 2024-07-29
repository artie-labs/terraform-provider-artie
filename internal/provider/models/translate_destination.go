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
	resourceModel.LastUpdatedAt = types.StringValue(apiModel.LastUpdatedAt)
	resourceModel.SSHTunnelUUID = types.StringValue(apiModel.SSHTunnelUUID)

	resourceModel.Config = &DestinationSharedConfigModel{
		Host:                types.StringValue(apiModel.Config.Host),
		Port:                types.Int64Value(apiModel.Config.Port),
		Endpoint:            types.StringValue(apiModel.Config.Endpoint),
		Username:            types.StringValue(apiModel.Config.Username),
		GCPProjectID:        types.StringValue(apiModel.Config.GCPProjectID),
		GCPLocation:         types.StringValue(apiModel.Config.GCPLocation),
		AWSAccessKeyID:      types.StringValue(apiModel.Config.AWSAccessKeyID),
		AWSRegion:           types.StringValue(apiModel.Config.AWSRegion),
		SnowflakeAccountURL: types.StringValue(apiModel.Config.SnowflakeAccountURL),
		SnowflakeVirtualDWH: types.StringValue(apiModel.Config.SnowflakeVirtualDWH),
		// TODO sensitive fields
		// Password:            types.StringValue(apiModel.Config.Password),
		// GCPCredentialsData:  types.StringValue(apiModel.Config.GCPCredentialsData),
		// AWSSecretAccessKey:  types.StringValue(apiModel.Config.AWSSecretAccessKey),
		// SnowflakePrivateKey: types.StringValue(apiModel.Config.SnowflakePrivateKey),
	}
}

func DestinationResourceToAPIModel(resourceModel DestinationResourceModel) DestinationAPIModel {
	sshTunnelUUID := resourceModel.SSHTunnelUUID.ValueString()
	if sshTunnelUUID == "" {
		sshTunnelUUID = uuid.Nil.String()
	}
	return DestinationAPIModel{
		UUID:          resourceModel.UUID.ValueString(),
		CompanyUUID:   resourceModel.CompanyUUID.ValueString(),
		Name:          resourceModel.Name.ValueString(),
		Label:         resourceModel.Label.ValueString(),
		LastUpdatedAt: resourceModel.LastUpdatedAt.ValueString(),
		SSHTunnelUUID: sshTunnelUUID,
		Config: DestinationSharedConfigAPIModel{
			Host:                resourceModel.Config.Host.ValueString(),
			Port:                resourceModel.Config.Port.ValueInt64(),
			Endpoint:            resourceModel.Config.Endpoint.ValueString(),
			Username:            resourceModel.Config.Username.ValueString(),
			GCPProjectID:        resourceModel.Config.GCPProjectID.ValueString(),
			GCPLocation:         resourceModel.Config.GCPLocation.ValueString(),
			AWSAccessKeyID:      resourceModel.Config.AWSAccessKeyID.ValueString(),
			AWSRegion:           resourceModel.Config.AWSRegion.ValueString(),
			SnowflakeAccountURL: resourceModel.Config.SnowflakeAccountURL.ValueString(),
			SnowflakeVirtualDWH: resourceModel.Config.SnowflakeVirtualDWH.ValueString(),
			// TODO sensitive fields
			// Password:            resourceModel.Config.Password.ValueString(),
			// GCPCredentialsData:  resourceModel.Config.GCPCredentialsData.ValueString(),
			// AWSSecretAccessKey:  resourceModel.Config.AWSSecretAccessKey.ValueString(),
			// SnowflakePrivateKey: resourceModel.Config.SnowflakePrivateKey.ValueString(),
		},
	}
}
