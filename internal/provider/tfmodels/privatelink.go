package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type PrivateLink struct {
	UUID           types.String `tfsdk:"uuid"`
	Name           types.String `tfsdk:"name"`
	AWSAccountID   types.String `tfsdk:"aws_account_id"`
	Region         types.String `tfsdk:"region"`
	VpcEndpointID  types.String `tfsdk:"vpc_endpoint_id"`
	VpcServiceName types.String `tfsdk:"vpc_service_name"`
	Status         types.String `tfsdk:"status"`
	DnsEntry       types.String `tfsdk:"dns_entry"`
}

func (p PrivateLink) ToAPIBaseModel() artieclient.BasePrivateLinkConnection {
	return artieclient.BasePrivateLinkConnection{
		Name:           p.Name.ValueString(),
		AWSAccountID:   p.AWSAccountID.ValueString(),
		Region:         p.Region.ValueString(),
		VpcEndpointID:  p.VpcEndpointID.ValueString(),
		VpcServiceName: p.VpcServiceName.ValueString(),
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
		UUID:           types.StringValue(apiModel.UUID.String()),
		Name:           types.StringValue(apiModel.Name),
		AWSAccountID:   types.StringValue(apiModel.AWSAccountID),
		Region:         types.StringValue(apiModel.Region),
		VpcEndpointID:  types.StringValue(apiModel.VpcEndpointID),
		VpcServiceName: types.StringValue(apiModel.VpcServiceName),
		Status:         types.StringValue(apiModel.Status),
		DnsEntry:       types.StringValue(apiModel.DnsEntry),
	}
}
