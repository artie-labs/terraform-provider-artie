package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/openapi"
)

type KafkaAdminUser struct {
	UUID                types.String `tfsdk:"uuid"`
	ClusterUUID         types.String `tfsdk:"cluster_uuid"`
	Username            types.String `tfsdk:"username"`
	AccessProfile       types.String `tfsdk:"access_profile"`
	ConsumerGroupPrefix types.String `tfsdk:"consumer_group_prefix"`
	Status              types.String `tfsdk:"status"`
	Password            types.String `tfsdk:"password"`
}

func (k KafkaAdminUser) ToAPICreateRequest() openapi.RouterCreateKafkaAdminUserRequest {
	return openapi.RouterCreateKafkaAdminUserRequest{Username: k.Username.ValueString()}
}

// KafkaAdminUserFromAPIModel builds the Terraform model; the cluster UUID and password are passed separately because the user payload omits them.
func KafkaAdminUserFromAPIModel(apiModel openapi.PayloadsKafkaUser, clusterUUID string, password types.String) KafkaAdminUser {
	return KafkaAdminUser{
		UUID:                types.StringValue(apiModel.Uuid.String()),
		ClusterUUID:         types.StringValue(clusterUUID),
		Username:            types.StringValue(apiModel.Username),
		AccessProfile:       types.StringValue(apiModel.AccessProfile),
		ConsumerGroupPrefix: types.StringValue(apiModel.ConsumerGroupPrefix),
		Status:              types.StringValue(apiModel.Status),
		Password:            password,
	}
}
