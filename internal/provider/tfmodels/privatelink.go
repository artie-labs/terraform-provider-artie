package tfmodels

import (
	"context"

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
	AzIDs          types.List   `tfsdk:"az_ids"`
	Status         types.String `tfsdk:"status"`
	DnsEntry       types.String `tfsdk:"dns_entry"`
	DataPlaneName  types.String `tfsdk:"data_plane_name"`
}

func (p PrivateLink) ToAPIBaseModel(ctx context.Context) (artieclient.BasePrivateLinkConnection, diag.Diagnostics) {
	azIDs, diags := parseList[string](ctx, p.AzIDs)
	if diags.HasError() {
		return artieclient.BasePrivateLinkConnection{}, diags
	}

	return artieclient.BasePrivateLinkConnection{
		Name:           p.Name.ValueString(),
		VpcServiceName: p.VpcServiceName.ValueString(),
		Region:         p.Region.ValueString(),
		VpcEndpointID:  p.VpcEndpointID.ValueString(),
		AzIDs:          azIDs,
		DataPlaneName:  p.DataPlaneName.ValueString(),
	}, diags
}

func (p PrivateLink) ToAPIModel(ctx context.Context) (artieclient.PrivateLinkConnection, diag.Diagnostics) {
	uuid, diags := parseUUID(p.UUID)
	if diags.HasError() {
		return artieclient.PrivateLinkConnection{}, diags
	}

	baseModel, baseDiags := p.ToAPIBaseModel(ctx)
	diags.Append(baseDiags...)
	if diags.HasError() {
		return artieclient.PrivateLinkConnection{}, diags
	}

	return artieclient.PrivateLinkConnection{
		UUID:                      uuid,
		BasePrivateLinkConnection: baseModel,
	}, diags
}

func PrivateLinkFromAPIModel(ctx context.Context, apiModel artieclient.PrivateLinkConnection) (PrivateLink, diag.Diagnostics) {
	azIDs, diags := types.ListValueFrom(ctx, types.StringType, apiModel.AzIDs)
	if diags.HasError() {
		return PrivateLink{}, diags
	}

	return PrivateLink{
		UUID:           types.StringValue(apiModel.UUID.String()),
		VpcServiceName: types.StringValue(apiModel.VpcServiceName),
		Region:         types.StringValue(apiModel.Region),
		VpcEndpointID:  types.StringValue(apiModel.VpcEndpointID),
		Name:           types.StringValue(apiModel.Name),
		AzIDs:          azIDs,
		Status:         types.StringValue(apiModel.Status),
		DnsEntry:       types.StringValue(apiModel.DnsEntry),
		DataPlaneName:  types.StringValue(apiModel.DataPlaneName),
	}, diags
}
