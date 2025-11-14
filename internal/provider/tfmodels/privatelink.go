package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type PrivateLink struct {
	UUID           types.String `tfsdk:"uuid"`
	VpcServiceName types.String `tfsdk:"vpc_service_name"`
	Region         types.String `tfsdk:"region"`
	VpcEndpointID  types.String `tfsdk:"vpc_endpoint_id"`
	Name           types.String `tfsdk:"name"`
	Status         types.String `tfsdk:"status"`
	DnsEntry       types.String `tfsdk:"dns_entry"`
	DataPlaneName  types.String `tfsdk:"data_plane_name"`
}

func (p PrivateLink) ToAPIBaseModel() artieclient.BasePrivateLinkConnection {
	return artieclient.BasePrivateLinkConnection{
		VpcServiceName: p.VpcServiceName.ValueString(),
		Region:         p.Region.ValueString(),
		VpcEndpointID:  p.VpcEndpointID.ValueString(),
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
		VpcServiceName: types.StringValue(apiModel.VpcServiceName),
		Region:         types.StringValue(apiModel.Region),
		VpcEndpointID:  types.StringValue(apiModel.VpcEndpointID),
		Name:           types.StringValue(apiModel.Name),
		Status:         types.StringValue(apiModel.Status),
		DnsEntry:       types.StringValue(apiModel.DnsEntry),
		DataPlaneName:  types.StringValue(apiModel.DataPlaneName),
	}
}
