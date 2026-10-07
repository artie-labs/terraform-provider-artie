package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type ColumnHashingSalt struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Salt        types.String `tfsdk:"salt"`
}

func (c ColumnHashingSalt) ToAPICreateRequest() openapi.RouterCreateColumnHashingSaltRequest {
	return openapi.RouterCreateColumnHashingSaltRequest{
		Name:        c.Name.ValueString(),
		Description: nonEmptyStringPointer(c.Description),
		Salt:        nonEmptyStringPointer(c.Salt),
	}
}

func (c ColumnHashingSalt) ToAPIUpdateRequest() openapi.RouterUpdateColumnHashingSaltRequest {
	return openapi.RouterUpdateColumnHashingSaltRequest{
		Name:        c.Name.ValueString(),
		Description: nonEmptyStringPointer(c.Description),
	}
}

// ColumnHashingSaltFromAPIModel builds the Terraform model; salt is passed separately because the update response omits it.
func ColumnHashingSaltFromAPIModel(apiModel openapi.PayloadsColumnHashingSalt, salt string) ColumnHashingSalt {
	return ColumnHashingSalt{
		UUID:        types.StringValue(apiModel.Uuid.String()),
		Name:        types.StringValue(apiModel.Name),
		Description: types.StringValue(lib.RemovePtr(apiModel.Description)),
		Salt:        types.StringValue(salt),
	}
}

func ColumnHashingSaltFromAPIDetail(detail openapi.PayloadsColumnHashingSaltDetail) ColumnHashingSalt {
	return ColumnHashingSaltFromAPIModel(openapi.PayloadsColumnHashingSalt{
		Uuid:        detail.Uuid,
		Name:        detail.Name,
		Description: detail.Description,
	}, detail.Salt)
}
