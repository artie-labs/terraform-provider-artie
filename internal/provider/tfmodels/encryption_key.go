package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func (e EncryptionKey) ToAPIBaseModel() (artieclient.BaseEncryptionKey, diag.Diagnostics) {
	kmsKeyUUID, diags := parseOptionalUUID(e.KMSKeyUUID)
	if diags.HasError() {
		return artieclient.BaseEncryptionKey{}, diags
	}

	return artieclient.BaseEncryptionKey{
		Name:        e.Name.ValueString(),
		Description: e.Description.ValueString(),
		KMSKeyUUID:  kmsKeyUUID,
	}, nil
}

func EncryptionKeyFromAPIModel(apiModel artieclient.EncryptionKey) EncryptionKey {
	kmsKeyUUID := types.StringNull()
	if apiModel.KMSKeyUUID != nil {
		kmsKeyUUID = types.StringValue(apiModel.KMSKeyUUID.String())
	}

	return EncryptionKey{
		UUID:        types.StringValue(apiModel.UUID.String()),
		Name:        types.StringValue(apiModel.Name),
		Description: types.StringValue(apiModel.Description),
		KMSKeyUUID:  kmsKeyUUID,
		Type:        types.StringValue(apiModel.Type),
		Key:         types.StringValue(apiModel.Key),
	}
}
