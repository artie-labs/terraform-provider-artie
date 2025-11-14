package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type PrivateLink struct {
	UUID          types.String `tfsdk:"uuid"`
	Name          types.String `tfsdk:"name"`
	AWSAccountID  types.String `tfsdk:"aws_account_id"`
	AWSRegion     types.String `tfsdk:"aws_region"`
	VpcEndpointID types.String `tfsdk:"vpc_endpoint_id"`
	Status        types.String `tfsdk:"status"`
	ServiceName   types.String `tfsdk:"service_name"`
}

func (p PrivateLink) ToAPIBaseModel() artieclient.BasePrivateLinkConnection {
	return artieclient.BasePrivateLinkConnection{
		Name:          p.Name.ValueString(),
		AWSAccountID:  p.AWSAccountID.ValueString(),
		AWSRegion:     p.AWSRegion.ValueString(),
		VpcEndpointID: p.VpcEndpointID.ValueString(),
	}
}

func (p PrivateLink) ToAPIModel() (artieclient.PrivateLinkConnection, diag.Diagnostics) {
	uuid, diags := parseUUID(p.UUID)
	if diags.HasError() {
		return artieclient.PrivateLinkConnection{}, diags
	}

	return artieclient.PrivateLinkConnection{
		UUID:                      uuid,
		BasePrivateLinkConnection: p.ToAPIBaseModel(),
	}, diags
}

func PrivateLinkFromAPIModel(apiModel artieclient.PrivateLinkConnection) PrivateLink {
	return PrivateLink{
		UUID:          types.StringValue(apiModel.UUID.String()),
		Name:          types.StringValue(apiModel.Name),
		AWSAccountID:  types.StringValue(apiModel.AWSAccountID),
		AWSRegion:     types.StringValue(apiModel.AWSRegion),
		VpcEndpointID: types.StringValue(apiModel.VpcEndpointID),
		Status:        types.StringValue(apiModel.Status),
		ServiceName:   types.StringValue(apiModel.ServiceName),
	}
}
