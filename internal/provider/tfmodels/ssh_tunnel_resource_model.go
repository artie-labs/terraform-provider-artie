package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type SSHTunnel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	Username  types.String `tfsdk:"username"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (s *SSHTunnel) UpdateFromAPIModel(apiModel artieclient.SSHTunnel) {
	s.UUID = types.StringValue(apiModel.UUID.String())
	s.Name = types.StringValue(apiModel.Name)
	s.Host = types.StringValue(apiModel.Host)
	s.Port = types.Int32Value(apiModel.Port)
	s.Username = types.StringValue(apiModel.Username)
	s.PublicKey = types.StringValue(apiModel.PublicKey)
}

func (s SSHTunnel) ToAPIBaseModel() artieclient.BaseSSHTunnel {
	return artieclient.BaseSSHTunnel{
		Name:      s.Name.ValueString(),
		Host:      s.Host.ValueString(),
		Port:      s.Port.ValueInt32(),
		Username:  s.Username.ValueString(),
		PublicKey: s.PublicKey.ValueString(),
	}
}

func (s SSHTunnel) ToAPIModel() artieclient.SSHTunnel {
	return artieclient.SSHTunnel{
		UUID:          parseUUID(s.UUID),
		BaseSSHTunnel: s.ToAPIBaseModel(),
	}
}
