package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

func SSHTunnelAPIToResourceModel(apiModel artieclient.SSHTunnel, resourceModel *SSHTunnelResourceModel) {
	resourceModel.UUID = types.StringValue(apiModel.UUID.String())
	resourceModel.Name = types.StringValue(apiModel.Name)
	resourceModel.Host = types.StringValue(apiModel.Host)
	resourceModel.Port = types.Int32Value(apiModel.Port)
	resourceModel.Username = types.StringValue(apiModel.Username)
	resourceModel.PublicKey = types.StringValue(apiModel.PublicKey)
}

func SSHTunnelResourceToAPIModel(resourceModel SSHTunnelResourceModel) artieclient.SSHTunnel {
	return artieclient.SSHTunnel{
		UUID:      parseUUID(resourceModel.UUID),
		Name:      resourceModel.Name.ValueString(),
		Host:      resourceModel.Host.ValueString(),
		Port:      resourceModel.Port.ValueInt32(),
		Username:  resourceModel.Username.ValueString(),
		PublicKey: resourceModel.PublicKey.ValueString(),
	}
}
