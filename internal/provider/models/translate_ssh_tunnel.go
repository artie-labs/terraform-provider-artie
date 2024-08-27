package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

func (s *SSHTunnelResourceModel) UpdateFromAPIModel(apiModel artieclient.SSHTunnel) {
	s.UUID = types.StringValue(apiModel.UUID.String())
	s.Name = types.StringValue(apiModel.Name)
	s.Host = types.StringValue(apiModel.Host)
	s.Port = types.Int32Value(apiModel.Port)
	s.Username = types.StringValue(apiModel.Username)
	s.PublicKey = types.StringValue(apiModel.PublicKey)
}

func (s SSHTunnelResourceModel) ToAPIBaseModel() artieclient.BaseSSHTunnel {
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
		BaseSSHTunnel: s.ToAPIBaseModel(),
	}
}
