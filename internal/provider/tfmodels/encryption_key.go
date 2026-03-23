package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type EncryptionKey struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	KMSKeyUUID  types.String `tfsdk:"kms_key_uuid"`
	Type        types.String `tfsdk:"type"`
	Key         types.String `tfsdk:"key"`
}

func (e EncryptionKey) ToAPIBaseModel() artieclient.BaseEncryptionKey {
	result := artieclient.BaseEncryptionKey{
		Name:        e.Name.ValueString(),
		Description: e.Description.ValueString(),
	}

	if kmsUUID, diags := parseOptionalUUID(e.KMSKeyUUID); !diags.HasError() && kmsUUID != nil {
		result.KMSKeyUUID = kmsUUID
	}

	return result
}

func EncryptionKeyFromAPIModel(apiModel artieclient.EncryptionKey) EncryptionKey {
	return EncryptionKey{
		UUID:        types.StringValue(apiModel.UUID.String()),
		Name:        types.StringValue(apiModel.Name),
		Description: types.StringValue(apiModel.Description),
		KMSKeyUUID:  optionalUUIDToStringValue(apiModel.KMSKeyUUID),
		Type:        types.StringValue(apiModel.Type),
		Key:         types.StringValue(apiModel.Key),
	}
}
