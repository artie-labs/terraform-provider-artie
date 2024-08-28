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

func SSHTunnelFromAPIModel(apiModel artieclient.SSHTunnel) SSHTunnel {
	return SSHTunnel{
		UUID:      types.StringValue(apiModel.UUID.String()),
		Name:      types.StringValue(apiModel.Name),
		Host:      types.StringValue(apiModel.Host),
		Port:      types.Int32Value(apiModel.Port),
		Username:  types.StringValue(apiModel.Username),
		PublicKey: types.StringValue(apiModel.PublicKey),
	}
}
