package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/lib"
	"terraform-provider-artie/internal/openapi"
)

type EncryptionKey struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	KMSKeyUUID  types.String `tfsdk:"kms_key_uuid"`
	Type        types.String `tfsdk:"type"`
	Key         types.String `tfsdk:"key"`
}

func (e EncryptionKey) ToAPICreateRequest() (openapi.RouterCreateEncryptionKeyRequest, diag.Diagnostics) {
	kmsKeyUUID, diags := parseOptionalUUID(e.KMSKeyUUID)
	if diags.HasError() {
		return openapi.RouterCreateEncryptionKeyRequest{}, diags
	}

	return openapi.RouterCreateEncryptionKeyRequest{
		Name:        e.Name.ValueString(),
		Description: nonEmptyStringPointer(e.Description),
		KmsKeyUUID:  kmsKeyUUID,
	}, nil
}

func (e EncryptionKey) ToAPIUpdateRequest() openapi.RouterUpdateEncryptionKeyRequest {
	return openapi.RouterUpdateEncryptionKeyRequest{
		Name:        e.Name.ValueString(),
		Description: nonEmptyStringPointer(e.Description),
	}
}

// EncryptionKeyFromAPIModel builds the Terraform model; key is passed separately because the update response omits it.
func EncryptionKeyFromAPIModel(apiModel openapi.PayloadsEncryptionKey, key string) EncryptionKey {
	kmsKeyUUID := types.StringNull()
	if apiModel.KmsKeyUUID != nil {
		kmsKeyUUID = types.StringValue(apiModel.KmsKeyUUID.String())
	}

	return EncryptionKey{
		UUID:        types.StringValue(apiModel.Uuid.String()),
		Name:        types.StringValue(apiModel.Name),
		Description: types.StringValue(lib.RemovePtr(apiModel.Description)),
		KMSKeyUUID:  kmsKeyUUID,
		Type:        types.StringValue(apiModel.Type),
		Key:         types.StringValue(key),
	}
}

func EncryptionKeyFromAPIDetail(detail openapi.PayloadsEncryptionKeyDetail) EncryptionKey {
	return EncryptionKeyFromAPIModel(openapi.PayloadsEncryptionKey{
		Uuid:        detail.Uuid,
		Name:        detail.Name,
		Description: detail.Description,
		KmsKeyUUID:  detail.KmsKeyUUID,
		Type:        detail.Type,
	}, detail.Key)
}
