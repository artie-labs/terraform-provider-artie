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

func (s SSHTunnelResourceModel) ToBaseAPIModel() artieclient.BaseSSHTunnel {
	return artieclient.BaseSSHTunnel{
		Name:      s.Name.ValueString(),
		Host:      s.Host.ValueString(),
		Port:      s.Port.ValueInt32(),
		Username:  s.Username.ValueString(),
		PublicKey: s.PublicKey.ValueString(),
	}
}

func (s SSHTunnelResourceModel) ToAPIModel() artieclient.SSHTunnel {
	return artieclient.SSHTunnel{
		UUID:          parseUUID(s.UUID),
		BaseSSHTunnel: s.ToBaseAPIModel(),
	}
}
