package tfmodels

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
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

func (p PrivateLink) ToAPICreateRequest(ctx context.Context) (openapi.RouterPrivateLinkConnectionCreateRequest, diag.Diagnostics) {
	azIDs, diags := parseList[string](ctx, p.AzIDs)
	return openapi.RouterPrivateLinkConnectionCreateRequest{
		Name:                p.Name.ValueStringPointer(),
		VpcServiceName:      p.VpcServiceName.ValueString(),
		AvailabilityZoneIds: azIDs,
		DataPlaneName:       nonEmptyStringPointer(p.DataPlaneName),
	}, diags
}

func (p PrivateLink) ToAPIUpdateRequest(ctx context.Context) (openapi.RouterPrivateLinkConnectionUpdateRequest, diag.Diagnostics) {
	azIDs, diags := parseList[string](ctx, p.AzIDs)
	return openapi.RouterPrivateLinkConnectionUpdateRequest{
		Name:                p.Name.ValueString(),
		VpcServiceName:      p.VpcServiceName.ValueString(),
		AvailabilityZoneIds: azIDs,
		DataPlaneName:       nonEmptyStringPointer(p.DataPlaneName),
	}, diags
}

func PrivateLinkFromAPIModel(ctx context.Context, apiModel openapi.PayloadsPrivateLinkConnection) (PrivateLink, diag.Diagnostics) {
	azIDs, diags := types.ListValueFrom(ctx, types.StringType, apiModel.AvailabilityZoneIds)
	if diags.HasError() {
		return PrivateLink{}, diags
	}

	return PrivateLink{
		UUID:           types.StringValue(apiModel.Uuid.String()),
		VpcServiceName: types.StringValue(apiModel.VpcServiceName),
		Region:         types.StringValue(apiModel.Region),
		VpcEndpointID:  types.StringValue(lib.RemovePtr(apiModel.VpcEndpointId)),
		Name:           types.StringValue(apiModel.Name),
		AzIDs:          azIDs,
		Status:         types.StringValue(apiModel.Status),
		DnsEntry:       types.StringValue(apiModel.DnsEntry),
		DataPlaneName:  types.StringValue(lib.RemovePtr(apiModel.DataPlaneName)),
	}, diags
}
