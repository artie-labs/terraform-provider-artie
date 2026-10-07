package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/openapi"
)

type SSHTunnel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	Username  types.String `tfsdk:"username"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (s SSHTunnel) ToAPICreateRequest() openapi.RouterSSHTunnelCreateRequest {
	return openapi.RouterSSHTunnelCreateRequest{
		Name:     s.Name.ValueStringPointer(),
		Host:     s.Host.ValueString(),
		Port:     int(s.Port.ValueInt32()),
		Username: s.Username.ValueString(),
	}
}

func (s SSHTunnel) ToAPIUpdateRequest() openapi.RouterSSHTunnelUpdateRequest {
	return openapi.RouterSSHTunnelUpdateRequest{
		Name:     s.Name.ValueString(),
		Host:     s.Host.ValueString(),
		Port:     int(s.Port.ValueInt32()),
		Username: s.Username.ValueString(),
	}
}

func SSHTunnelFromAPIModel(apiModel openapi.PayloadsSSHTunnel) SSHTunnel {
	return SSHTunnel{
		UUID:      types.StringValue(apiModel.Uuid.String()),
		Name:      types.StringValue(apiModel.Name),
		Host:      types.StringValue(apiModel.Host),
		Port:      types.Int32Value(int32(apiModel.Port)),
		Username:  types.StringValue(apiModel.Username),
		PublicKey: types.StringValue(apiModel.PublicKey),
	}
}
