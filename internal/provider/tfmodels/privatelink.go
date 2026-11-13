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
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	// PrivateLinkServiceAlias is the Azure Private Link target; VpcServiceName and AzIDs are the AWS one.
	PrivateLinkServiceAlias types.String `tfsdk:"private_link_service_alias"`
}

func (p PrivateLink) ToAPICreateRequest(ctx context.Context) (openapi.RouterPrivateLinkConnectionCreateRequest, diag.Diagnostics) {
	azIDs, diags := parseOptionalList[string](ctx, p.AzIDs)
	return openapi.RouterPrivateLinkConnectionCreateRequest{
		Name:                    stringPointer(p.Name),
		VpcServiceName:          p.VpcServiceName.ValueStringPointer(),
		AvailabilityZoneIds:     azIDs,
		PrivateLinkServiceAlias: nonEmptyStringPointer(p.PrivateLinkServiceAlias),
		DataPlaneName:           nonEmptyStringPointer(p.DataPlaneName),
	}, diags
}

func (p PrivateLink) ToAPIUpdateRequest(ctx context.Context) (openapi.RouterPrivateLinkConnectionUpdateRequest, diag.Diagnostics) {
	azIDs, diags := parseOptionalList[string](ctx, p.AzIDs)
	return openapi.RouterPrivateLinkConnectionUpdateRequest{
		Name:                    p.Name.ValueString(),
		VpcServiceName:          p.VpcServiceName.ValueStringPointer(),
		AvailabilityZoneIds:     azIDs,
		PrivateLinkServiceAlias: nonEmptyStringPointer(p.PrivateLinkServiceAlias),
		DataPlaneName:           nonEmptyStringPointer(p.DataPlaneName),
	}, diags
}

func PrivateLinkFromAPIModel(ctx context.Context, apiModel openapi.PayloadsPrivateLinkConnection) (PrivateLink, diag.Diagnostics) {
	// An Azure connection has no AWS target, and an AWS one no Azure target; read those back as null to match an unset
	// config. A legacy AWS connection can have no AZ IDs, so it keeps reading back as [].
	azIDs := types.ListNull(types.StringType)
	var diags diag.Diagnostics
	if apiModel.Provider != openapi.EnumsCloudProviderAzure {
		azIDs, diags = types.ListValueFrom(ctx, types.StringType, apiModel.AvailabilityZoneIds)
		if diags.HasError() {
			return PrivateLink{}, diags
		}
	}

	return PrivateLink{
		UUID:                    types.StringValue(apiModel.Uuid.String()),
		VpcServiceName:          nonEmptyStringValue(&apiModel.VpcServiceName),
		Region:                  types.StringValue(apiModel.Region),
		VpcEndpointID:           types.StringValue(lib.RemovePtr(apiModel.VpcEndpointId)),
		Name:                    types.StringValue(apiModel.Name),
		AzIDs:                   azIDs,
		Status:                  types.StringValue(apiModel.Status),
		DnsEntry:                types.StringValue(apiModel.DnsEntry),
		DataPlaneName:           types.StringValue(lib.RemovePtr(apiModel.DataPlaneName)),
		CloudProvider:           types.StringValue(string(apiModel.Provider)),
		PrivateLinkServiceAlias: nonEmptyStringValue(&apiModel.PrivateLinkServiceAlias),
	}, diags
}
