package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/openapi"
)

const (
	KafkaUserAccessProfileProducerConsumer = "producer_consumer"
	KafkaUserAccessProfileAdmin            = "admin"
)

type KafkaUser struct {
	UUID                types.String `tfsdk:"uuid"`
	ClusterUUID         types.String `tfsdk:"cluster_uuid"`
	Username            types.String `tfsdk:"username"`
	AccessProfile       types.String `tfsdk:"access_profile"`
	ConsumerGroupPrefix types.String `tfsdk:"consumer_group_prefix"`
	Status              types.String `tfsdk:"status"`
	Password            types.String `tfsdk:"password"`
}

func (k KafkaUser) IsAdmin() bool {
	return k.AccessProfile.ValueString() == KafkaUserAccessProfileAdmin
}

func (k KafkaUser) ToAPICreateRequest() openapi.RouterCreateKafkaUserRequest {
	return openapi.RouterCreateKafkaUserRequest{
		AccessProfile: KafkaUserAccessProfileProducerConsumer,
		Username:      k.Username.ValueString(),
	}
}

func (k KafkaUser) ToAPICreateAdminRequest() openapi.RouterCreateKafkaAdminUserRequest {
	return openapi.RouterCreateKafkaAdminUserRequest{Username: k.Username.ValueString()}
}

// KafkaUserFromAPIModel builds the state for a Kafka user. The API never lists passwords, so the caller passes the
// one it already has (from state, the create response or the credentials endpoint).
func KafkaUserFromAPIModel(clusterUUID string, apiModel openapi.PayloadsKafkaUser, password types.String) KafkaUser {
	return KafkaUser{
		UUID:                types.StringValue(apiModel.Uuid.String()),
		ClusterUUID:         types.StringValue(clusterUUID),
		Username:            types.StringValue(apiModel.Username),
		AccessProfile:       types.StringValue(apiModel.AccessProfile),
		ConsumerGroupPrefix: types.StringValue(apiModel.ConsumerGroupPrefix),
		Status:              types.StringValue(apiModel.Status),
		Password:            password,
	}
}
