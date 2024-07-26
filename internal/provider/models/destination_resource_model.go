package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type DestinationResourceModel struct {
	UUID        types.String `tfsdk:"uuid"`
	CompanyUUID types.String `tfsdk:"company_uuid"`
	Name        types.String `tfsdk:"name"`
	Label       types.String `tfsdk:"label"`
}
