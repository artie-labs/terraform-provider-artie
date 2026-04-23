package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type ColumnHashingSalt struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Salt        types.String `tfsdk:"salt"`
}

func (c ColumnHashingSalt) ToAPIBaseModel() artieclient.BaseColumnHashingSalt {
	return artieclient.BaseColumnHashingSalt{
		Name:        c.Name.ValueString(),
		Description: c.Description.ValueString(),
		Salt:        c.Salt.ValueString(),
	}
}

func ColumnHashingSaltFromAPIModel(apiModel artieclient.ColumnHashingSalt) ColumnHashingSalt {
	return ColumnHashingSalt{
		UUID:        types.StringValue(apiModel.UUID.String()),
		Name:        types.StringValue(apiModel.Name),
		Description: types.StringValue(apiModel.Description),
		Salt:        types.StringValue(apiModel.Salt),
	}
}
